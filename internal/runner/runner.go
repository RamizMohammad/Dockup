// Package runner executes a deploy job on the machine it runs on. It is used
// by the hub for the "local" server and by the remote agent, so behaviour is
// identical everywhere.
package runner

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/RamizMohammad/dockup/internal/envx"
)

type Job struct {
	BuildID        string            `json:"buildId"`
	Action         string            `json:"action"` // deploy | rollback | restart | stop | down | pull
	ProjectID      string            `json:"projectId"`
	ComposeProject string            `json:"composeProject"`
	SourceKind     string            `json:"sourceKind"` // path | git
	Path           string            `json:"path"`
	Repo           string            `json:"repo"`
	Branch         string            `json:"branch"`
	Commit         string            `json:"commit"` // pin to commit (rollback)
	ComposeFile    string            `json:"composeFile"`
	GitToken       string            `json:"gitToken"`
	SSHKey         string            `json:"sshKey"`
	EnvFile        string            `json:"envFile"`
	EnvInline      string            `json:"envInline"`
	Params         map[string]string `json:"params"`
	SecretParams   []string          `json:"secretParams"` // names whose values are masked
	Profiles       []string          `json:"profiles"`
	PreCmd         string            `json:"preCmd"`
	PostCmd        string            `json:"postCmd"`
	HealthURL      string            `json:"healthUrl"`
	HealthSecs     int               `json:"healthSecs"`
	NoCache        bool              `json:"noCache"`
	Pull           bool              `json:"pull"`
	CheckOnly      bool              `json:"checkOnly"` // SCM poll: just report remote HEAD
}

type Result struct {
	Commit string `json:"commit"`
	Err    string `json:"err"`
}

type Runner struct {
	WorkDir string // where git checkouts and env files live
}

var safe = regexp.MustCompile(`[^a-z0-9_-]+`)

func ComposeName(n string) string {
	s := safe.ReplaceAllString(strings.ToLower(n), "-")
	s = strings.Trim(s, "-_")
	if s == "" {
		s = "project"
	}
	return s
}

type masker struct {
	secrets []string
	out     func(string)
	stage   string
}

func (m *masker) add(v string) {
	if len(v) >= 4 {
		m.secrets = append(m.secrets, v)
	}
}

func (m *masker) log(s string) {
	for _, x := range m.secrets {
		s = strings.ReplaceAll(s, x, "****")
	}
	m.out(s)
}

// StagePrefix marks pipeline stage transitions in the console stream. The hub
// turns these lines into structured stage records (and keeps them in the log
// so the console can fold output per stage).
const StagePrefix = "##[stage] "

// begin closes the running stage (as success) and opens a new one.
func (m *masker) begin(name string) {
	m.end("success")
	m.stage = name
	m.out(StagePrefix + name + " running")
}

// end closes the running stage with the given status.
func (m *masker) end(status string) {
	if m.stage != "" {
		m.out(StagePrefix + m.stage + " " + status)
		m.stage = ""
	}
}

var secretKey = regexp.MustCompile(`(?i)(secret|token|passw|pwd|key|auth|credential|private|cert|dsn|api)`)

// envSecrets returns the values of KEY=VAL lines that should be masked: any
// value whose key looks sensitive, plus any long value (12+ chars). Short
// non-sensitive values such as "production" or "8080" stay readable.
func envSecrets(env string) []string {
	var out []string
	for _, l := range strings.Split(env, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		i := strings.Index(l, "=")
		if i <= 0 {
			continue
		}
		k := strings.TrimPrefix(strings.TrimSpace(l[:i]), "export ")
		v := strings.Trim(strings.TrimSpace(l[i+1:]), `"'`)
		if secretKey.MatchString(k) || len(v) >= 12 {
			out = append(out, v)
		}
	}
	return out
}

// buildEnv is the environment for every command a build runs. Dockup's own
// configuration (secret key, OAuth secrets, agent token, admin password) is
// removed so compose files, hooks and git never see it.
func buildEnv(params map[string]string) []string {
	var env []string
	for _, kv := range os.Environ() {
		if envx.Internal(kv) {
			continue
		}
		env = append(env, kv)
	}
	for k, v := range params {
		env = append(env, k+"="+v)
	}
	return env
}

// Run executes the job, streaming console lines to logf.
func (r *Runner) Run(ctx context.Context, j Job, logf func(string)) (res Result) {
	m := &masker{out: logf}
	m.add(j.GitToken)
	for _, v := range envSecrets(j.EnvInline) {
		m.add(v)
	}
	for _, n := range j.SecretParams {
		m.add(j.Params[n])
	}
	defer func() {
		if p := recover(); p != nil {
			res.Err = fmt.Sprint("internal error: ", p)
		}
		switch {
		case res.Err != "" && ctx.Err() != nil:
			m.end("cancelled")
		case res.Err != "":
			m.end("failed")
		default:
			m.end("success")
		}
	}()
	commit, err := r.run(ctx, j, m)
	res.Commit = commit
	if err != nil {
		res.Err = err.Error()
	}
	return
}

func (r *Runner) run(ctx context.Context, j Job, m *masker) (string, error) {
	started := time.Now()
	dir, composeFile, commit := j.Path, j.ComposeFile, ""
	env := buildEnv(j.Params)

	// ---- git checkout ----
	if j.SourceKind == "git" {
		var cleanup func()
		var gitEnv []string
		gitEnv, cleanup = r.gitEnv(j, m)
		defer cleanup()
		dir = filepath.Join(r.WorkDir, "checkouts", ComposeName(j.ProjectID))
		if j.CheckOnly {
			m.begin("poll")
			out, err := r.capture(ctx, "", append(env, gitEnv...), "git", "ls-remote", j.Repo, "refs/heads/"+j.Branch)
			if err != nil {
				return "", err
			}
			f := strings.Fields(out)
			if len(f) > 0 {
				return f[0], nil
			}
			return "", errors.New("branch not found on remote")
		}
		m.begin("checkout")
		if err := os.MkdirAll(filepath.Dir(dir), 0o750); err != nil {
			return "", err
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
			m.log("==> cloning " + j.Repo + " (" + j.Branch + ")")
			if err := r.stream(ctx, "", append(env, gitEnv...), m, "git", "clone", "--branch", j.Branch, j.Repo, dir); err != nil {
				return "", fmt.Errorf("git clone: %w", err)
			}
		} else {
			m.log("==> updating checkout")
			// follow repo URL changes made in the project settings
			if cur, _ := r.capture(ctx, dir, env, "git", "remote", "get-url", "origin"); strings.TrimSpace(cur) != j.Repo {
				m.log("==> repository URL changed, re-pointing origin")
				if err := r.stream(ctx, dir, env, m, "git", "remote", "set-url", "origin", j.Repo); err != nil {
					return "", fmt.Errorf("git remote set-url: %w", err)
				}
			}
			if err := r.stream(ctx, dir, append(env, gitEnv...), m, "git", "fetch", "--prune", "origin"); err != nil {
				return "", fmt.Errorf("git fetch: %w", err)
			}
		}
		target := "origin/" + j.Branch
		if j.Commit != "" {
			target = j.Commit
		}
		if err := r.stream(ctx, dir, env, m, "git", "reset", "--hard", target); err != nil {
			return "", fmt.Errorf("git reset: %w", err)
		}
		if out, err := r.capture(ctx, dir, env, "git", "rev-parse", "HEAD"); err == nil {
			commit = strings.TrimSpace(out)
			subj, _ := r.capture(ctx, dir, env, "git", "log", "-1", "--pretty=%an: %s")
			m.log("==> at " + short(commit) + "  " + strings.TrimSpace(subj))
		}
	} else if j.CheckOnly {
		return "", nil
	}

	// ---- locate compose file ----
	m.begin("prepare")
	if j.SourceKind == "path" {
		if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
			composeFile, dir = filepath.Base(dir), filepath.Dir(dir)
		}
	}
	if composeFile == "" {
		for _, c := range []string{"compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"} {
			if _, err := os.Stat(filepath.Join(dir, c)); err == nil {
				composeFile = c
				break
			}
		}
	}
	if composeFile == "" {
		return commit, fmt.Errorf("no compose file found in %s", dir)
	}
	full := composeFile
	if !filepath.IsAbs(full) {
		full = filepath.Join(dir, composeFile)
	}
	if _, err := os.Stat(full); err != nil {
		return commit, fmt.Errorf("compose file not found: %s", full)
	}
	cdir := filepath.Dir(full)

	// ---- env files ----
	args := []string{"-p", j.ComposeProject, "-f", full}
	if j.EnvFile != "" {
		ef := j.EnvFile
		if !filepath.IsAbs(ef) {
			ef = filepath.Join(cdir, ef)
		}
		if _, err := os.Stat(ef); err != nil {
			return commit, fmt.Errorf("env file not found: %s", ef)
		}
		args = append(args, "--env-file", ef)
		m.log("==> env file: " + ef)
	}
	if strings.TrimSpace(j.EnvInline) != "" {
		ed := filepath.Join(r.WorkDir, "env")
		os.MkdirAll(ed, 0o700)
		ep := filepath.Join(ed, ComposeName(j.ProjectID)+".env")
		if err := os.WriteFile(ep, []byte(j.EnvInline+"\n"), 0o600); err != nil {
			return commit, err
		}
		// the decrypted copy only lives for the duration of the build
		defer os.Remove(ep)
		args = append(args, "--env-file", ep)
		m.log("==> inline env applied")
	}
	for _, p := range j.Profiles {
		if p = strings.TrimSpace(p); p != "" {
			args = append(args, "--profile", p)
		}
	}

	bin, pre := dockerCmd()
	run := func(extra ...string) error {
		a := append(append([]string{}, pre...), args...)
		a = append(a, extra...)
		m.log("$ " + filepath.Base(bin) + " " + strings.Join(a, " "))
		if err := r.stream(ctx, cdir, env, m, bin, a...); err != nil {
			if ctx.Err() != nil {
				return err
			}
			return fmt.Errorf("compose %s: %w", extra[0], err)
		}
		return nil
	}

	if j.PreCmd != "" && (j.Action == "deploy" || j.Action == "rollback") {
		m.begin("pre-deploy")
		m.log("==> pre-deploy command")
		if err := r.stream(ctx, cdir, env, m, "sh", "-c", j.PreCmd); err != nil {
			return commit, fmt.Errorf("pre command: %w", err)
		}
	}

	switch j.Action {
	case "deploy", "rollback":
		if j.Pull {
			m.begin("pull")
			if err := run("pull", "--ignore-buildable"); err != nil {
				m.log("(pull failed, continuing)")
			}
		}
		if j.NoCache {
			m.begin("build")
			if err := run("build", "--no-cache"); err != nil {
				return commit, err
			}
		}
		m.begin("up")
		if err := run("up", "-d", "--build", "--remove-orphans"); err != nil {
			return commit, err
		}
		_ = run("ps")
	case "restart":
		m.begin("restart")
		if err := run("restart"); err != nil {
			return commit, err
		}
	case "stop":
		m.begin("stop")
		if err := run("stop"); err != nil {
			return commit, err
		}
	case "down":
		m.begin("down")
		if err := run("down"); err != nil {
			return commit, err
		}
	case "pull":
		m.begin("pull")
		if err := run("pull"); err != nil {
			return commit, err
		}
	default:
		return commit, fmt.Errorf("unknown action %q", j.Action)
	}

	if j.HealthURL != "" && (j.Action == "deploy" || j.Action == "rollback" || j.Action == "restart") {
		m.begin("health")
		if err := health(ctx, j.HealthURL, j.HealthSecs, m); err != nil {
			return commit, err
		}
	}
	if j.PostCmd != "" && (j.Action == "deploy" || j.Action == "rollback") {
		m.begin("post-deploy")
		m.log("==> post-deploy command")
		if err := r.stream(ctx, cdir, env, m, "sh", "-c", j.PostCmd); err != nil {
			return commit, fmt.Errorf("post command: %w", err)
		}
	}
	m.end("success")
	m.log(fmt.Sprintf("==> finished in %s", time.Since(started).Round(time.Millisecond)))
	return commit, nil
}

func short(c string) string {
	if len(c) > 8 {
		return c[:8]
	}
	return c
}

func health(ctx context.Context, url string, secs int, m *masker) error {
	if secs <= 0 {
		secs = 60
	}
	m.log(fmt.Sprintf("==> health check %s (up to %ds)", url, secs))
	deadline := time.Now().Add(time.Duration(secs) * time.Second)
	cl := &http.Client{Timeout: 5 * time.Second}
	for {
		resp, err := cl.Get(url)
		if err == nil {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
			resp.Body.Close()
			if resp.StatusCode < 400 {
				m.log(fmt.Sprintf("health OK (%d)", resp.StatusCode))
				return nil
			}
		}
		if time.Now().After(deadline) {
			return errors.New("health check timed out")
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// dockerCmd returns the docker compose invocation (plugin preferred).
func dockerCmd() (string, []string) {
	if exec.Command("docker", "compose", "version").Run() == nil {
		return "docker", []string{"compose"}
	}
	if p, err := exec.LookPath("docker-compose"); err == nil {
		return p, nil
	}
	return "docker", []string{"compose"}
}

func (r *Runner) gitEnv(j Job, m *masker) ([]string, func()) {
	env := []string{"GIT_TERMINAL_PROMPT=0"}
	cleanup := func() {}
	if j.GitToken != "" {
		b := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + j.GitToken))
		// scope the header to the repository's origin so submodules or
		// redirects to other hosts never receive the token
		key := "http.extraHeader"
		if u, err := url.Parse(j.Repo); err == nil && (u.Scheme == "https" || u.Scheme == "http") && u.Host != "" {
			key = "http." + u.Scheme + "://" + u.Host + "/.extraHeader"
		}
		env = append(env, "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0="+key,
			"GIT_CONFIG_VALUE_0=Authorization: Basic "+b)
		m.add(b)
	}
	if j.SSHKey != "" {
		dir, err := os.MkdirTemp("", "dockup-ssh")
		if err == nil {
			kp := filepath.Join(dir, "id")
			key := j.SSHKey
			if !strings.HasSuffix(key, "\n") {
				key += "\n"
			}
			os.WriteFile(kp, []byte(key), 0o600)
			env = append(env, "GIT_SSH_COMMAND=ssh -i "+kp+" -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new")
			cleanup = func() { os.RemoveAll(dir) }
		}
	}
	return env, cleanup
}

func (r *Runner) capture(ctx context.Context, dir string, env []string, name string, args ...string) (string, error) {
	c := exec.CommandContext(ctx, name, args...)
	c.Dir, c.Env = dir, env
	out, err := c.Output()
	return string(out), err
}

// stream runs a command in its own process group, forwards combined output
// line by line, and kills the whole group when ctx is cancelled.
func (r *Runner) stream(ctx context.Context, dir string, env []string, m *masker, name string, args ...string) error {
	c := exec.CommandContext(ctx, name, args...)
	c.Dir, c.Env = dir, env
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	c.Cancel = func() error { return syscall.Kill(-c.Process.Pid, syscall.SIGKILL) }
	c.WaitDelay = 3 * time.Second
	pr, pw := io.Pipe()
	c.Stdout, c.Stderr = pw, pw
	if err := c.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	var last string
	go func() {
		sc := bufio.NewScanner(pr)
		sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for sc.Scan() {
			line := strings.TrimRight(sc.Text(), "\r")
			if strings.TrimSpace(line) != "" {
				last = line
			}
			m.log(line)
		}
		close(done)
	}()
	err := c.Wait()
	pw.Close()
	<-done
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil && last != "" {
		// the last output line usually says why (shown as the build's failure reason)
		for _, x := range m.secrets {
			last = strings.ReplaceAll(last, x, "****")
		}
		if len(last) > 300 {
			last = last[:300] + "…"
		}
		return fmt.Errorf("%s (%w)", strings.TrimSpace(last), err)
	}
	return err
}

// Info describes the host for the dashboard.
func Info() (host, osName, dockerVer string) {
	host, _ = os.Hostname()
	osName = runtimeOS()
	if out, err := exec.Command("docker", "version", "--format", "{{.Server.Version}}").Output(); err == nil {
		dockerVer = strings.TrimSpace(string(out))
	}
	return
}

// ComposeVersion returns the docker compose (plugin or standalone) version, or "".
func ComposeVersion() string {
	bin, pre := dockerCmd()
	out, err := exec.Command(bin, append(pre, "version", "--short")...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
}

func runtimeOS() string {
	if b, err := os.ReadFile("/etc/os-release"); err == nil {
		for _, l := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(l, "PRETTY_NAME=") {
				return strings.Trim(strings.TrimPrefix(l, "PRETTY_NAME="), `"`)
			}
		}
	}
	return "linux"
}
