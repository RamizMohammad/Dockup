package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeDocker puts a stub `docker` first on PATH that records its arguments.
func fakeDocker(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	rec := filepath.Join(dir, "calls.log")
	script := "#!/bin/sh\n" +
		"if [ \"$1\" = compose ] && [ \"$2\" = version ]; then exit 0; fi\n" +
		"echo \"$@\" >> " + rec + "\n" +
		"echo \"out: $@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "docker"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return rec
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	c := exec.Command("git", args...)
	c.Dir = dir
	c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	out, err := c.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestGitDeployAndRollback(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	rec := fakeDocker(t)
	repo := t.TempDir()
	git(t, repo, "init", "-b", "main")
	os.WriteFile(filepath.Join(repo, "compose.yaml"), []byte("services: {}\n"), 0o644)
	os.WriteFile(filepath.Join(repo, "v.txt"), []byte("1"), 0o644)
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-m", "one")
	c1 := git(t, repo, "rev-parse", "HEAD")
	os.WriteFile(filepath.Join(repo, "v.txt"), []byte("2"), 0o644)
	git(t, repo, "commit", "-am", "two")
	c2 := git(t, repo, "rev-parse", "HEAD")

	r := &Runner{WorkDir: t.TempDir()}
	var lines []string
	logf := func(s string) { lines = append(lines, s) }
	job := Job{BuildID: "b1", Action: "deploy", ProjectID: "p1", ComposeProject: "demo", SourceKind: "git", Repo: repo, Branch: "main",
		EnvInline: "SECRET=hunter2hunter2", HealthSecs: 1}

	res := r.Run(context.Background(), job, logf)
	if res.Err != "" || res.Commit != c2 {
		t.Fatalf("deploy: err=%q commit=%q want %q\n%s", res.Err, res.Commit, c2, strings.Join(lines, "\n"))
	}
	// roll back to the first commit
	job.Commit = c1
	res = r.Run(context.Background(), job, logf)
	if res.Err != "" || res.Commit != c1 {
		t.Fatalf("rollback: err=%q commit=%q want %q", res.Err, res.Commit, c1)
	}
	calls, _ := os.ReadFile(rec)
	if !strings.Contains(string(calls), "up -d --build --remove-orphans") || strings.Contains(string(calls), "compose compose") {
		t.Fatalf("unexpected docker calls:\n%s", calls)
	}
	// CheckOnly reports the remote head without deploying
	job.CheckOnly, job.Commit = true, ""
	res = r.Run(context.Background(), job, logf)
	if res.Err != "" || res.Commit != c2 {
		t.Fatalf("poll: err=%q commit=%q want %q", res.Err, res.Commit, c2)
	}
}

func TestSecretsMaskedAndCancel(t *testing.T) {
	fakeDocker(t)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("services: {}\n"), 0o644)
	r := &Runner{WorkDir: t.TempDir()}
	var lines []string
	job := Job{BuildID: "b", Action: "deploy", ProjectID: "p", ComposeProject: "x", SourceKind: "path", Path: dir,
		GitToken: "tok_supersecret", PreCmd: "echo token=tok_supersecret"}
	res := r.Run(context.Background(), job, func(s string) { lines = append(lines, s) })
	if res.Err != "" {
		t.Fatal(res.Err)
	}
	if strings.Contains(strings.Join(lines, "\n"), "tok_supersecret") {
		t.Fatalf("secret leaked in log:\n%s", strings.Join(lines, "\n"))
	}

	// cancellation kills the running command group
	job.PreCmd, job.GitToken = "sleep 30", ""
	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(300 * time.Millisecond); cancel() }()
	start := time.Now()
	res = r.Run(ctx, job, func(string) {})
	if res.Err == "" || time.Since(start) > 5*time.Second {
		t.Fatalf("cancel did not stop the job quickly: err=%q after %s", res.Err, time.Since(start))
	}
}

func TestMissingComposeFile(t *testing.T) {
	fakeDocker(t)
	r := &Runner{WorkDir: t.TempDir()}
	res := r.Run(context.Background(), Job{Action: "deploy", ProjectID: "p", ComposeProject: "x", SourceKind: "path", Path: t.TempDir()}, func(string) {})
	if !strings.Contains(res.Err, "no compose file") {
		t.Fatalf("got %q", res.Err)
	}
}

func TestRepoURLChangeFollowed(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	fakeDocker(t)
	mk := func(content string) (string, string) {
		repo := t.TempDir()
		git(t, repo, "init", "-b", "main")
		os.WriteFile(filepath.Join(repo, "compose.yaml"), []byte("services: {}\n"), 0o644)
		os.WriteFile(filepath.Join(repo, "who.txt"), []byte(content), 0o644)
		git(t, repo, "add", ".")
		git(t, repo, "commit", "-m", content)
		return repo, git(t, repo, "rev-parse", "HEAD")
	}
	a, _ := mk("a")
	b, cb := mk("b")
	r := &Runner{WorkDir: t.TempDir()}
	job := Job{Action: "deploy", ProjectID: "p", ComposeProject: "x", SourceKind: "git", Repo: a, Branch: "main"}
	if res := r.Run(context.Background(), job, func(string) {}); res.Err != "" {
		t.Fatal(res.Err)
	}
	job.Repo = b
	res := r.Run(context.Background(), job, func(string) {})
	if res.Err != "" || res.Commit != cb {
		t.Fatalf("after repo change: err=%q commit=%q want %q", res.Err, res.Commit, cb)
	}
}

func TestEnvHygieneAndStages(t *testing.T) {
	fakeDocker(t)
	t.Setenv("DOCKUP_SECRET_KEY", "must-not-leak")
	t.Setenv("DEPLOYHUB_SECRET_KEY", "must-not-leak-either")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("services: {}\n"), 0o644)
	work := t.TempDir()
	r := &Runner{WorkDir: work}
	var lines []string
	job := Job{Action: "deploy", ProjectID: "p", ComposeProject: "x", SourceKind: "path", Path: dir,
		EnvInline: "NODE_ENV=production\nDB_PASSWORD=abcd1234\nLONG=0123456789abcdef",
		Params:    map[string]string{"TOKEN_P": "param-secret-value"}, SecretParams: []string{"TOKEN_P"},
		PreCmd:  "env | grep -c -e DOCKUP_ -e DEPLOYHUB_ || true; echo production abcd1234 0123456789abcdef $TOKEN_P",
		PostCmd: "true"}
	res := r.Run(context.Background(), job, func(s string) { lines = append(lines, s) })
	if res.Err != "" {
		t.Fatal(res.Err)
	}
	out := strings.Join(lines, "\n")
	for _, leak := range []string{"must-not-leak", "must-not-leak-either", "abcd1234", "0123456789abcdef", "param-secret-value"} {
		if strings.Contains(out, leak) {
			t.Errorf("%q leaked:\n%s", leak, out)
		}
	}
	if !strings.Contains(out, "\n0\n") {
		t.Errorf("DOCKUP_ vars reached the build:\n%s", out)
	}
	if !strings.Contains(out, "production") {
		t.Errorf("short non-secret value should stay readable:\n%s", out)
	}
	if ents, _ := os.ReadDir(filepath.Join(work, "env")); len(ents) != 0 {
		t.Errorf("decrypted env file left on disk: %v", ents)
	}
	var stages []string
	for _, l := range lines {
		if strings.HasPrefix(l, StagePrefix) {
			stages = append(stages, strings.TrimPrefix(l, StagePrefix))
		}
	}
	want := "prepare running|prepare success|pre-deploy running|pre-deploy success|up running|up success|post-deploy running|post-deploy success"
	if strings.Join(stages, "|") != want {
		t.Errorf("stages = %v", stages)
	}
}
