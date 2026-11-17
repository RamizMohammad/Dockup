package hub

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/RamizMohammad/dockup/internal/cron"
	"github.com/RamizMohammad/dockup/internal/envx"
	"github.com/RamizMohammad/dockup/internal/store"
)

var (
	reRepo   = regexp.MustCompile(`^(https?://|ssh://|git://|git@)[^\s]+$`)
	reBranch = regexp.MustCompile(`^[A-Za-z0-9._/@+-]+$`)
	reCommit = regexp.MustCompile(`^[0-9a-fA-F]{4,64}$`)
	reParam  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

func (h *Hub) routes(m *http.ServeMux) {
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	m.HandleFunc("GET /api/boot", h.handleBoot)
	m.HandleFunc("POST /api/setup", h.handleSetup)
	m.HandleFunc("POST /api/login", h.handleLogin)
	m.HandleFunc("POST /api/logout", h.handleLogout)
	m.HandleFunc("GET /auth/{provider}/login", h.oauthLogin)
	m.HandleFunc("GET /auth/{provider}/callback", h.oauthCallback)

	m.HandleFunc("GET /api/me", h.auth("viewer", h.me))
	m.HandleFunc("PUT /api/me/password", h.auth("viewer", h.changePassword))
	m.HandleFunc("GET /api/me/sessions", h.auth("viewer", h.listSessions))
	m.HandleFunc("DELETE /api/me/sessions/{id}", h.auth("viewer", h.revokeSession))
	m.HandleFunc("GET /api/cron/preview", h.auth("viewer", h.cronPreview))
	m.HandleFunc("GET /api/onboarding", h.auth("admin", h.onboarding))
	m.HandleFunc("PUT /api/onboarding", h.auth("admin", h.setOnboarding))
	m.HandleFunc("POST /api/onboarding/sample", h.auth("admin", h.createSample))
	m.HandleFunc("GET /api/dashboard", h.auth("viewer", h.dashboard))
	m.HandleFunc("GET /api/servers", h.auth("viewer", h.listServers))
	m.HandleFunc("POST /api/servers", h.auth("admin", h.createServer))
	m.HandleFunc("POST /api/servers/{id}/rotate", h.auth("admin", h.rotateServer))
	m.HandleFunc("DELETE /api/servers/{id}", h.auth("admin", h.deleteServer))
	m.HandleFunc("GET /api/credentials", h.auth("admin", h.listCreds))
	m.HandleFunc("POST /api/credentials", h.auth("admin", h.createCred))
	m.HandleFunc("DELETE /api/credentials/{id}", h.auth("admin", h.deleteCred))

	m.HandleFunc("GET /api/projects", h.auth("viewer", h.listProjects))
	m.HandleFunc("POST /api/projects", h.auth("admin", h.saveProject))
	m.HandleFunc("GET /api/projects/{id}", h.auth("viewer", h.getProject))
	m.HandleFunc("PUT /api/projects/{id}", h.auth("admin", h.saveProject))
	m.HandleFunc("DELETE /api/projects/{id}", h.auth("admin", h.deleteProject))
	m.HandleFunc("GET /api/projects/{id}/env", h.auth("admin", h.getEnv))
	m.HandleFunc("POST /api/projects/{id}/builds", h.auth("operator", h.triggerBuild))
	m.HandleFunc("GET /api/builds", h.auth("viewer", h.listBuilds))
	m.HandleFunc("GET /api/builds/{id}", h.auth("viewer", h.getBuild))
	m.HandleFunc("GET /api/builds/{id}/log", h.auth("viewer", h.buildLog))
	m.HandleFunc("GET /api/builds/{id}/stream", h.auth("viewer", h.buildStream))
	m.HandleFunc("POST /api/builds/{id}/cancel", h.auth("operator", h.cancel))
	m.HandleFunc("POST /hooks/{id}", h.webhook)

	m.HandleFunc("GET /api/users", h.auth("admin", h.listUsers))
	m.HandleFunc("POST /api/users", h.auth("admin", h.createUser))
	m.HandleFunc("PUT /api/users/{id}", h.auth("admin", h.updateUser))
	m.HandleFunc("DELETE /api/users/{id}", h.auth("admin", h.deleteUser))
	m.HandleFunc("GET /api/tokens", h.auth("viewer", h.listTokens))
	m.HandleFunc("POST /api/tokens", h.auth("viewer", h.createToken))
	m.HandleFunc("DELETE /api/tokens/{id}", h.auth("viewer", h.deleteToken))
	m.HandleFunc("GET /api/settings", h.auth("admin", h.getSettings))
	m.HandleFunc("PUT /api/settings", h.auth("admin", h.putSettings))
	m.HandleFunc("GET /api/audit", h.auth("admin", h.listAudit))

	m.HandleFunc("GET /api/agent/stream", h.agentStream)
	m.HandleFunc("POST /api/agent/ping", h.agentPing)
	m.HandleFunc("POST /api/agent/log", h.agentLog)
	m.HandleFunc("POST /api/agent/done", h.agentDone)

	m.Handle("/", h.static())
}

func (h *Hub) static() http.Handler {
	fsrv := http.FileServerFS(h.cfg.WebFS)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			fail(w, 404, "not found")
			return
		}
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if _, err := fs.Stat(h.cfg.WebFS, p); err != nil {
			r.URL.Path = "/"
		}
		// fonts and brand images never change within a version; the app shell
		// is revalidated on every load so an upgrade shows up immediately
		if strings.HasPrefix(p, "fonts/") || strings.HasPrefix(p, "brand/") {
			w.Header().Set("Cache-Control", "public, max-age=604800")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fsrv.ServeHTTP(w, r)
	})
}

// ---------------- dashboard / servers ----------------

type serverView struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Local    bool      `json:"local"`
	Online   bool      `json:"online"`
	LastSeen time.Time `json:"lastSeen"`
	Hostname string    `json:"hostname"`
	OS       string    `json:"os"`
	Docker   string    `json:"docker"`
	Version  string    `json:"version"`
	Projects int       `json:"projects"`
	Running  []string  `json:"running"` // "project #N"
	Created  time.Time `json:"created"`
}

func (h *Hub) serverViews() []serverView {
	var out []serverView
	h.db.View(func(s *store.State) {
		for _, sv := range s.Servers {
			v := serverView{ID: sv.ID, Name: sv.Name, Local: sv.Local, LastSeen: sv.LastSeen, Hostname: sv.Hostname, OS: sv.OS,
				Docker: sv.Docker, Version: sv.Version, Created: sv.Created, Running: []string{}}
			for _, p := range s.Projects {
				if p.ServerID == sv.ID {
					v.Projects++
				}
			}
			for _, b := range s.Builds {
				if b.ServerID == sv.ID && b.Status == "running" && b.Action != "poll" {
					if p := h.projectByID(s, b.ProjectID); p != nil {
						v.Running = append(v.Running, fmt.Sprintf("%s #%d", p.Name, b.Number))
					}
				}
			}
			out = append(out, v)
		}
	})
	for i := range out {
		h.mu.Lock()
		out[i].Online = out[i].Local || h.agents[out[i].ID] != nil
		h.mu.Unlock()
	}
	return out
}

type buildView struct {
	store.Build
	Project string `json:"project"`
	Server  string `json:"server"`
}

func (h *Hub) buildViews(s *store.State, bs []*store.Build) []buildView {
	pn, sn := map[string]string{}, map[string]string{}
	secret := map[string]bool{} // projectID/param
	for _, p := range s.Projects {
		pn[p.ID] = p.Name
		for _, d := range p.Params {
			if d.Secret {
				secret[p.ID+"/"+d.Name] = true
			}
		}
	}
	for _, v := range s.Servers {
		sn[v.ID] = v.Name
	}
	out := []buildView{}
	for _, b := range bs {
		if b.Action == "poll" && b.Status != "failed" {
			continue
		}
		v := buildView{*b, pn[b.ProjectID], sn[b.ServerID]}
		v.Build.Params = maskParams(b.ProjectID, b.Params, secret)
		out = append(out, v)
	}
	return out
}

func maskParams(pid string, in map[string]string, secret map[string]bool) map[string]string {
	if len(in) == 0 {
		return in
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		if secret[pid+"/"+k] {
			v = "****"
		}
		out[k] = v
	}
	return out
}

// redactURL removes credentials embedded in a repository URL.
func redactURL(v string) string {
	u, err := url.Parse(v)
	if err != nil || u.User == nil {
		return v
	}
	u.User = nil
	return u.String()
}

func (h *Hub) dashboard(w http.ResponseWriter, r *http.Request, u *store.User) {
	servers := h.serverViews()
	type pv struct {
		ID, Name, Folder, ServerID, Server, Kind string
		Disabled                                 bool
		Last                                     *buildView
		Running                                  *buildView
		Cron                                     string
		Health                                   []string `json:"health"`
	}
	var projects []map[string]any
	var recent []buildView
	stats := map[string]int{"running": 0, "queued": 0, "today": 0, "projects": 0, "servers": 0, "online": 0}
	var week []map[string]int
	h.db.View(func(s *store.State) {
		for _, p := range s.Projects {
			var last, run *store.Build
			var hist []string
			for _, b := range s.Builds {
				if b.ProjectID != p.ID || b.Action == "poll" {
					continue
				}
				if b.Status == "running" || b.Status == "queued" {
					if run == nil || b.Number > run.Number {
						run = b
					}
				} else if last == nil || b.Number > last.Number {
					last = b
				}
			}
			var done []*store.Build
			for _, b := range s.Builds {
				if b.ProjectID == p.ID && b.Action != "poll" && (b.Status == "success" || b.Status == "failed") {
					done = append(done, b)
				}
			}
			sort.Slice(done, func(i, j int) bool { return done[i].Number > done[j].Number })
			for i := 0; i < len(done) && i < 12; i++ {
				hist = append([]string{done[i].Status}, hist...)
			}
			sv := h.serverByID(s, p.ServerID)
			src := p.Source.Path
			if p.Source.Kind == "git" {
				src = redactURL(p.Source.Repo) + " @ " + p.Source.Branch
			}
			trig := "manual"
			switch {
			case p.Cron != "" && p.PollSCM:
				trig = "scm"
			case p.Cron != "":
				trig = "cron"
			case p.Source.Kind == "git":
				trig = "hook"
			}
			m := map[string]any{"id": p.ID, "name": p.Name, "folder": p.Folder, "serverId": p.ServerID, "kind": p.Source.Kind,
				"disabled": p.Disabled, "cron": p.Cron, "history": hist, "source": src, "trigger": trig, "hasParams": len(p.Params) > 0}
			if sv != nil {
				m["server"] = sv.Name
			}
			if last != nil {
				m["last"] = last
			}
			if run != nil {
				m["running"] = run
			}
			projects = append(projects, m)
		}
		bs := append([]*store.Build{}, s.Builds...)
		sort.Slice(bs, func(i, j int) bool { return bs[i].Queued.After(bs[j].Queued) })
		if len(bs) > 40 {
			bs = bs[:40]
		}
		recent = h.buildViews(s, bs)
		if len(recent) > 12 {
			recent = recent[:12]
		}
		// headline stats: deployments today, 7-day pass rate per day, queue
		now := time.Now()
		day0 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		days := make([]map[string]int, 7)
		for i := range days {
			days[i] = map[string]int{"success": 0, "failed": 0}
		}
		for _, b := range s.Builds {
			if b.Action == "poll" {
				continue
			}
			switch b.Status {
			case "running":
				stats["running"]++
			case "queued":
				stats["queued"]++
			}
			if !b.Queued.Before(day0) {
				stats["today"]++
			}
			if b.Status == "success" || b.Status == "failed" {
				f := b.Finished.In(now.Location())
				d := int(day0.Sub(time.Date(f.Year(), f.Month(), f.Day(), 0, 0, 0, 0, now.Location())).Hours()+1) / 24
				if d >= 0 && d < 7 {
					days[6-d][b.Status]++
				}
			}
		}
		stats["projects"] = len(s.Projects)
		week = days
	})
	if projects == nil {
		projects = []map[string]any{}
	}
	online := 0
	for _, sv := range servers {
		if sv.Online {
			online++
		}
	}
	stats["servers"], stats["online"] = len(servers), online
	writeJSON(w, 200, map[string]any{"servers": servers, "projects": projects, "recent": recent, "stats": stats, "week": week})
}

func (h *Hub) listServers(w http.ResponseWriter, r *http.Request, u *store.User) {
	writeJSON(w, 200, h.serverViews())
}

func (h *Hub) createServer(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct{ Name string }
	if readJSON(r, &b) != nil || strings.TrimSpace(b.Name) == "" {
		fail(w, 400, "name required")
		return
	}
	tok := "dks_" + store.NewSecret(24)
	sv := &store.Server{ID: store.NewID(), Name: strings.TrimSpace(b.Name), TokenHash: store.Hash(tok), Created: time.Now()}
	h.db.Update(func(s *store.State) { s.Servers = append(s.Servers, sv) })
	h.audit(r, u.Login, "server.create", sv.Name)
	writeJSON(w, 200, map[string]any{"id": sv.ID, "name": sv.Name, "token": tok, "hub": h.baseURL(r)})
}

func (h *Hub) cronPreview(w http.ResponseWriter, r *http.Request, u *store.User) {
	sc, err := cron.Parse(r.URL.Query().Get("spec"))
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"next": sc.Next(time.Now(), 3), "zone": time.Now().Format("MST -07:00")})
}

func (h *Hub) rotateServer(w http.ResponseWriter, r *http.Request, u *store.User) {
	id, tok := r.PathValue("id"), "dks_"+store.NewSecret(24)
	ok := false
	h.db.Update(func(s *store.State) {
		if sv := h.serverByID(s, id); sv != nil && !sv.Local {
			sv.TokenHash, ok = store.Hash(tok), true
		}
	})
	if !ok {
		fail(w, 404, "server not found")
		return
	}
	h.mu.Lock()
	if ag := h.agents[id]; ag != nil {
		close(ag.quit)
		delete(h.agents, id)
	}
	h.mu.Unlock()
	h.audit(r, u.Login, "server.rotate", id)
	writeJSON(w, 200, map[string]any{"id": id, "token": tok, "hub": h.baseURL(r)})
}

func (h *Hub) deleteServer(w http.ResponseWriter, r *http.Request, u *store.User) {
	id := r.PathValue("id")
	var err error
	h.db.Update(func(s *store.State) {
		sv := h.serverByID(s, id)
		switch {
		case sv == nil:
			err = errors.New("not found")
		case sv.Local:
			err = errors.New("the local server cannot be removed")
		default:
			for _, p := range s.Projects {
				if p.ServerID == id {
					err = errors.New("server still has projects")
					return
				}
			}
			out := s.Servers[:0]
			for _, x := range s.Servers {
				if x.ID != id {
					out = append(out, x)
				}
			}
			s.Servers = out
		}
	})
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "server.delete", id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---------------- credentials ----------------

func (h *Hub) listCreds(w http.ResponseWriter, r *http.Request, u *store.User) {
	out := []map[string]any{}
	h.db.View(func(s *store.State) {
		for _, c := range s.Credentials {
			var used []string
			for _, p := range s.Projects {
				if p.Source.CredentialID == c.ID {
					used = append(used, p.Name)
				}
			}
			out = append(out, map[string]any{"id": c.ID, "name": c.Name, "kind": c.Kind, "created": c.Created, "hint": c.Hint,
				"used": c.Used, "usedBy": c.UsedBy, "projects": used})
		}
	})
	writeJSON(w, 200, out)
}

func (h *Hub) createCred(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct{ Name, Kind, Secret string }
	if readJSON(r, &b) != nil || b.Name == "" || b.Secret == "" || (b.Kind != "git-token" && b.Kind != "ssh-key") {
		fail(w, 400, "name, kind (git-token|ssh-key) and secret required")
		return
	}
	if b.Kind == "ssh-key" && !strings.Contains(b.Secret, "PRIVATE KEY") {
		fail(w, 400, "that does not look like a private key (expected -----BEGIN ... PRIVATE KEY-----)")
		return
	}
	c := &store.Credential{ID: store.NewID(), Name: strings.TrimSpace(b.Name), Kind: b.Kind, Enc: h.db.Seal(b.Secret),
		Hint: credHint(b.Kind, b.Secret), Created: time.Now()}
	h.db.Update(func(s *store.State) { s.Credentials = append(s.Credentials, c) })
	h.audit(r, u.Login, "credential.create", b.Name)
	writeJSON(w, 200, map[string]string{"id": c.ID})
}

// credHint is a non-secret identifier for a credential: a SHA-256 digest
// prefix of the secret, plus the token's well-known prefix (ghp_, glpat-).
func credHint(kind, secret string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(secret)))
	fp := "SHA256:" + hex.EncodeToString(sum[:])[:12]
	if kind == "git-token" {
		for _, p := range []string{"github_pat_", "ghp_", "gho_", "ghs_", "glpat-", "gldt-"} {
			if strings.HasPrefix(secret, p) {
				return p + "… " + fp
			}
		}
	}
	if kind == "ssh-key" {
		for _, t := range []string{"OPENSSH", "RSA", "EC", "DSA"} {
			if strings.Contains(secret, "BEGIN "+t+" PRIVATE KEY") {
				return strings.ToLower(t) + " " + fp
			}
		}
	}
	return fp
}

func (h *Hub) deleteCred(w http.ResponseWriter, r *http.Request, u *store.User) {
	id := r.PathValue("id")
	var err error
	h.db.Update(func(s *store.State) {
		for _, p := range s.Projects {
			if p.Source.CredentialID == id {
				err = errors.New("credential is used by project " + p.Name)
				return
			}
		}
		out := s.Credentials[:0]
		for _, c := range s.Credentials {
			if c.ID != id {
				out = append(out, c)
			}
		}
		s.Credentials = out
	})
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "credential.delete", id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---------------- projects ----------------

type projectView struct {
	store.Project
	EnvEnc     string `json:"-"`
	HasEnv     bool   `json:"hasEnv"`
	Webhook    string `json:"webhook,omitempty"`
	WebhookURL string `json:"webhookUrl,omitempty"`
}

// view is what a user may see of a project. Only admins (who can edit it)
// get the hook commands, env file paths, notification webhooks (Slack URLs
// are secrets), secret parameter defaults and repository credentials.
func (h *Hub) view(p *store.Project, u *store.User, r *http.Request) projectView {
	v := projectView{Project: *p, HasEnv: p.EnvEnc != ""}
	v.Project.EnvEnc = ""
	v.Project.Webhook = ""
	v.Project.Params = append([]store.Param{}, p.Params...)
	if u.Role == "admin" {
		v.Webhook = p.Webhook
		v.WebhookURL = h.baseURL(r) + "/hooks/" + p.ID
		return v
	}
	v.Project.Source.Repo = redactURL(p.Source.Repo)
	v.Project.PreCmd, v.Project.PostCmd, v.Project.EnvFile = redacted(p.PreCmd), redacted(p.PostCmd), redacted(p.EnvFile)
	v.Project.NotifyURLs = nil
	for i := range v.Project.Params {
		if v.Project.Params[i].Secret {
			v.Project.Params[i].Default = ""
		}
	}
	return v
}

func redacted(v string) string {
	if v == "" {
		return ""
	}
	return "(hidden)"
}

func (h *Hub) listProjects(w http.ResponseWriter, r *http.Request, u *store.User) {
	out := []projectView{}
	h.db.View(func(s *store.State) {
		for _, p := range s.Projects {
			out = append(out, h.view(p, u, r))
		}
	})
	writeJSON(w, 200, out)
}

func (h *Hub) getProject(w http.ResponseWriter, r *http.Request, u *store.User) {
	var v *projectView
	h.db.View(func(s *store.State) {
		if p := h.projectByID(s, r.PathValue("id")); p != nil {
			x := h.view(p, u, r)
			v = &x
		}
	})
	if v == nil {
		fail(w, 404, "not found")
		return
	}
	writeJSON(w, 200, v)
}

func (h *Hub) getEnv(w http.ResponseWriter, r *http.Request, u *store.User) {
	var enc string
	h.db.View(func(s *store.State) {
		if p := h.projectByID(s, r.PathValue("id")); p != nil {
			enc = p.EnvEnc
		}
	})
	v, err := h.db.Open(enc)
	if err != nil {
		fail(w, 500, "cannot decrypt")
		return
	}
	h.audit(r, u.Login, "project.env.read", r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"env": v})
}

type projectInput struct {
	Name         string        `json:"name"`
	Desc         string        `json:"desc"`
	Folder       string        `json:"folder"`
	ServerID     string        `json:"serverId"`
	Source       store.Source  `json:"source"`
	EnvFile      string        `json:"envFile"`
	Env          *string       `json:"env"`
	Params       []store.Param `json:"params"`
	Profiles     string        `json:"profiles"`
	PreCmd       string        `json:"preCmd"`
	PostCmd      string        `json:"postCmd"`
	HealthURL    string        `json:"healthUrl"`
	HealthSecs   int           `json:"healthSecs"`
	Cron         string        `json:"cron"`
	PollSCM      bool          `json:"pollScm"`
	NotifyURLs   []string      `json:"notifyUrls"`
	NotifyOn     string        `json:"notifyOn"`
	Keep         int           `json:"keep"`
	NoCache      bool          `json:"noCache"`
	Pull         bool          `json:"pull"`
	Disabled     bool          `json:"disabled"`
	RegenWebhook bool          `json:"regenWebhook"`
}

func validate(in *projectInput) error {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 80 {
		return errors.New("name required (max 80 chars)")
	}
	src := &in.Source
	switch src.Kind {
	case "path":
		if !strings.HasPrefix(src.Path, "/") {
			return errors.New("path must be an absolute path to the compose directory or file")
		}
	case "git":
		if !reRepo.MatchString(src.Repo) {
			return errors.New("repo must be an https://, ssh:// or git@ URL")
		}
		if src.Branch == "" {
			src.Branch = "main"
		}
		if !reBranch.MatchString(src.Branch) || strings.HasPrefix(src.Branch, "-") {
			return errors.New("invalid branch name")
		}
		if strings.Contains(src.ComposeFile, "..") || strings.HasPrefix(src.ComposeFile, "/") {
			return errors.New("compose file must be a relative path inside the repo")
		}
	default:
		return errors.New("source kind must be path or git")
	}
	if in.Cron != "" {
		if _, err := cron.Parse(in.Cron); err != nil {
			return fmt.Errorf("cron: %w", err)
		}
	}
	seen := map[string]bool{}
	for _, p := range in.Params {
		if !reParam.MatchString(p.Name) {
			return fmt.Errorf("invalid parameter name %q", p.Name)
		}
		if envx.Internal(p.Name) || seen[p.Name] {
			return fmt.Errorf("parameter %q is reserved or duplicated", p.Name)
		}
		seen[p.Name] = true
	}
	for _, n := range in.NotifyURLs {
		if !strings.HasPrefix(n, "https://") && !strings.HasPrefix(n, "http://") {
			return fmt.Errorf("notification URL must be http(s): %q", n)
		}
	}
	if in.HealthURL != "" && !strings.HasPrefix(in.HealthURL, "http://") && !strings.HasPrefix(in.HealthURL, "https://") {
		return errors.New("health check URL must start with http:// or https://")
	}
	if in.HealthSecs < 0 || in.HealthSecs > 3600 {
		return errors.New("health timeout must be between 0 and 3600 seconds")
	}
	if in.Keep < 0 || in.Keep > 1000 {
		return errors.New("retention must be between 0 and 1000 builds")
	}
	if len(in.Desc) > 500 {
		return errors.New("description is limited to 500 characters")
	}
	switch in.NotifyOn {
	case "", "always", "failure", "change":
	default:
		return errors.New("notify on must be always, failure or change")
	}
	if in.NotifyOn == "" {
		in.NotifyOn = "always"
	}
	return nil
}

func (h *Hub) saveProject(w http.ResponseWriter, r *http.Request, u *store.User) {
	var in projectInput
	if err := readJSON(r, &in); err != nil {
		fail(w, 400, "bad request")
		return
	}
	if err := validate(&in); err != nil {
		fail(w, 400, err.Error())
		return
	}
	id := r.PathValue("id")
	var err error
	var saved store.Project
	h.db.Update(func(s *store.State) {
		if h.serverByID(s, in.ServerID) == nil {
			err = errors.New("unknown server")
			return
		}
		var p *store.Project
		if id != "" {
			if p = h.projectByID(s, id); p == nil {
				err = errors.New("not found")
				return
			}
		} else {
			p = &store.Project{ID: store.NewID(), Created: time.Now(), Webhook: store.NewSecret(16)}
			s.Projects = append(s.Projects, p)
		}
		for _, o := range s.Projects {
			if o != p && strings.EqualFold(o.Name, in.Name) {
				err = errors.New("a project with that name exists")
				return
			}
		}
		p.Name, p.Desc, p.Folder, p.ServerID, p.Source, p.EnvFile = in.Name, in.Desc, strings.TrimSpace(in.Folder), in.ServerID, in.Source, in.EnvFile
		p.Params, p.Profiles, p.PreCmd, p.PostCmd = in.Params, in.Profiles, in.PreCmd, in.PostCmd
		p.HealthURL, p.HealthSecs, p.Cron, p.PollSCM = in.HealthURL, in.HealthSecs, in.Cron, in.PollSCM
		p.NotifyURLs, p.NotifyOn, p.Keep, p.NoCache, p.Pull, p.Disabled = in.NotifyURLs, in.NotifyOn, in.Keep, in.NoCache, in.Pull, in.Disabled
		if in.Env != nil {
			p.EnvEnc = h.db.Seal(strings.TrimSpace(*in.Env))
		}
		if in.RegenWebhook {
			p.Webhook = store.NewSecret(16)
		}
		saved = *p
	})
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "project.save", saved.Name)
	writeJSON(w, 200, h.view(&saved, u, r))
}

func (h *Hub) deleteProject(w http.ResponseWriter, r *http.Request, u *store.User) {
	id := r.PathValue("id")
	var name string
	var drop []string
	h.db.Update(func(s *store.State) {
		out := s.Projects[:0]
		for _, p := range s.Projects {
			if p.ID == id {
				name = p.Name
				continue
			}
			out = append(out, p)
		}
		s.Projects = out
		bo := s.Builds[:0]
		for _, b := range s.Builds {
			if b.ProjectID == id {
				drop = append(drop, b.ID)
				continue
			}
			bo = append(bo, b)
		}
		s.Builds = bo
	})
	if name == "" {
		fail(w, 404, "not found")
		return
	}
	for _, d := range drop {
		os.Remove(h.db.LogPath(d))
	}
	h.audit(r, u.Login, "project.delete", name)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---------------- builds ----------------

func (h *Hub) triggerBuild(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct {
		Action string            `json:"action"`
		Params map[string]string `json:"params"`
		Commit string            `json:"commit"`
	}
	if readJSON(r, &b) != nil {
		fail(w, 400, "bad request")
		return
	}
	if b.Action == "" {
		b.Action = "deploy"
	}
	if b.Action == "poll" {
		fail(w, 400, "unknown action")
		return
	}
	if b.Commit != "" && !reCommit.MatchString(b.Commit) {
		fail(w, 400, "invalid commit")
		return
	}
	if b.Action == "rollback" && b.Commit == "" {
		fail(w, 400, "rollback needs a commit")
		return
	}
	bd, err := h.enqueue(r.PathValue("id"), b.Action, "manual", u.Login, b.Params, b.Commit)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "build."+b.Action, r.PathValue("id"))
	writeJSON(w, 200, bd)
}

func (h *Hub) listBuilds(w http.ResponseWriter, r *http.Request, u *store.User) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	pid := r.URL.Query().Get("project")
	var out []buildView
	h.db.View(func(s *store.State) {
		var bs []*store.Build
		for _, b := range s.Builds {
			if pid == "" || b.ProjectID == pid {
				bs = append(bs, b)
			}
		}
		sort.Slice(bs, func(i, j int) bool { return bs[i].Queued.After(bs[j].Queued) })
		out = h.buildViews(s, bs)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	writeJSON(w, 200, out)
}

func (h *Hub) getBuild(w http.ResponseWriter, r *http.Request, u *store.User) {
	var out []buildView
	h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			if b.ID == r.PathValue("id") {
				pn, sn := "", ""
				if p := h.projectByID(s, b.ProjectID); p != nil {
					pn = p.Name
				}
				if sv := h.serverByID(s, b.ServerID); sv != nil {
					sn = sv.Name
				}
				out = []buildView{{*b, pn, sn}}
			}
		}
	})
	if len(out) == 0 {
		fail(w, 404, "not found")
		return
	}
	writeJSON(w, 200, out[0])
}

func (h *Hub) buildExists(id string) (store.Build, string, bool) {
	var b store.Build
	var pname string
	ok := false
	if !store.ValidID(id) {
		return b, "", false
	}
	h.db.View(func(s *store.State) {
		for _, x := range s.Builds {
			if x.ID == id {
				b, ok = *x, true
				if p := h.projectByID(s, x.ProjectID); p != nil {
					pname = p.Name
				}
			}
		}
	})
	return b, pname, ok
}

var reFileSafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func (h *Hub) buildLog(w http.ResponseWriter, r *http.Request, u *store.User) {
	id := r.PathValue("id")
	b, pname, ok := h.buildExists(id)
	if !ok {
		fail(w, 404, "not found")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.URL.Query().Get("download") != "" {
		name := reFileSafe.ReplaceAllString(fmt.Sprintf("%s-%d.log", pname, b.Number), "_")
		w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	}
	f, err := os.Open(h.db.LogPath(id))
	if err != nil {
		return
	}
	defer f.Close()
	io.Copy(w, f)
}

func (h *Hub) buildStream(w http.ResponseWriter, r *http.Request, u *store.User) {
	id := r.PathValue("id")
	var lastBuild, pendingBuild []byte
	status := func() string {
		st := ""
		var js []byte
		h.db.View(func(s *store.State) {
			for _, b := range s.Builds {
				if b.ID == id {
					st = b.Status
					if bv := h.buildViews(s, []*store.Build{b}); len(bv) > 0 {
						js, _ = json.Marshal(bv[0])
					}
				}
			}
		})
		if js != nil && string(js) != string(lastBuild) {
			lastBuild, pendingBuild = js, js
		}
		return st
	}
	if !store.ValidID(id) || status() == "" {
		fail(w, 404, "not found")
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		fail(w, 500, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	ch, unsub := h.logs.Sub(id)
	defer unsub()
	flushBuild := func() {
		if pendingBuild != nil {
			fmt.Fprintf(w, "event: build\ndata: %s\n\n", pendingBuild)
			pendingBuild = nil
			fl.Flush()
		}
	}
	flushBuild()
	var off int64
	buf := make([]byte, 64*1024)
	for {
		// read what is new
		for {
			f, err := os.Open(h.db.LogPath(id))
			if err != nil {
				break
			}
			n, _ := f.ReadAt(buf, off)
			f.Close()
			if n == 0 {
				break
			}
			// only ship complete lines (a write may be in progress)
			if k := strings.LastIndexByte(string(buf[:n]), '\n'); k >= 0 {
				n = k + 1
			} else if n < len(buf) {
				break
			}
			off += int64(n)
			j, _ := json.Marshal(string(buf[:n]))
			fmt.Fprintf(w, "event: log\ndata: %s\n\n", j)
			fl.Flush()
		}
		st := status()
		flushBuild()
		if st != "queued" && st != "running" {
			// one last drain happened above; confirm no more bytes
			if fi, err := os.Stat(h.db.LogPath(id)); err != nil || fi.Size() <= off {
				fmt.Fprintf(w, "event: end\ndata: %q\n\n", st)
				fl.Flush()
				return
			}
			continue
		}
		select {
		case <-ch:
		case <-time.After(1500 * time.Millisecond):
			fmt.Fprint(w, ": ka\n\n")
			fl.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (h *Hub) cancel(w http.ResponseWriter, r *http.Request, u *store.User) {
	if err := h.cancelBuild(r.PathValue("id")); err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "build.cancel", r.PathValue("id"))
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /hooks/{id}: GitHub/Gitea (HMAC), GitLab (X-Gitlab-Token) or ?token=
func (h *Hub) webhook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var secret, branch, kind string
	h.db.View(func(s *store.State) {
		if p := h.projectByID(s, id); p != nil {
			secret, branch, kind = p.Webhook, p.Source.Branch, p.Source.Kind
		}
	})
	body, _ := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	okAuth := false
	if secret != "" {
		eq := func(a, b string) bool { return hmac.Equal([]byte(a), []byte(b)) }
		if sig := r.Header.Get("X-Hub-Signature-256"); sig != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write(body)
			okAuth = eq(sig, "sha256="+hex.EncodeToString(mac.Sum(nil)))
		} else if t := firstNonEmpty(r.Header.Get("X-Gitlab-Token"), r.Header.Get("X-Webhook-Token"), r.URL.Query().Get("token")); t != "" {
			okAuth = eq(t, secret)
		}
	}
	if !okAuth {
		fail(w, 401, "invalid webhook secret")
		return
	}
	if r.Header.Get("X-GitHub-Event") == "ping" {
		writeJSON(w, 200, map[string]string{"ok": "pong"})
		return
	}
	var payload struct {
		Ref string `json:"ref"`
	}
	json.Unmarshal(body, &payload)
	if kind == "git" && payload.Ref != "" && payload.Ref != "refs/heads/"+branch {
		writeJSON(w, 200, map[string]string{"skipped": "other branch"})
		return
	}
	b, err := h.enqueue(id, "deploy", "webhook", "", nil, "")
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "build": b.Number})
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if s != "" {
			return s
		}
	}
	return ""
}

// ---------------- users / tokens / settings / audit ----------------

func (h *Hub) listUsers(w http.ResponseWriter, r *http.Request, u *store.User) {
	out := []map[string]any{}
	h.db.View(func(s *store.State) {
		for _, x := range s.Users {
			m := publicUser(x)
			m["created"] = x.CreatedAt
			m["lastLogin"] = x.LastLogin
			m["invited"] = x.Provider == ""
			out = append(out, m)
		}
	})
	writeJSON(w, 200, out)
}

func (h *Hub) createUser(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct{ Login, Name, Email, Role, Password string }
	if readJSON(r, &b) != nil || roleRank[b.Role] == 0 {
		fail(w, 400, "role must be admin, operator or viewer")
		return
	}
	b.Login, b.Email = strings.TrimSpace(b.Login), strings.TrimSpace(b.Email)
	if b.Email != "" {
		if a, err := mail.ParseAddress(b.Email); err != nil || a.Address != b.Email {
			fail(w, 400, "invalid email address")
			return
		}
	}
	nu := &store.User{ID: store.NewID(), Login: b.Login, Name: b.Name, Email: b.Email, Role: b.Role, CreatedAt: time.Now()}
	switch {
	case b.Password != "":
		if len(b.Password) < 10 || len(b.Login) < 2 {
			fail(w, 400, "login (2+) and password (10+) required")
			return
		}
		nu.Provider, nu.PassHash = "local", hashPassword(b.Password)
	case b.Email != "":
		if nu.Login == "" {
			nu.Login = strings.Split(b.Email, "@")[0]
		}
	default:
		fail(w, 400, "give a password (local user) or an email (OAuth invite)")
		return
	}
	var err error
	h.db.Update(func(s *store.State) {
		for _, x := range s.Users {
			if strings.EqualFold(x.Login, nu.Login) || (nu.Email != "" && strings.EqualFold(x.Email, nu.Email)) {
				err = errors.New("user already exists")
				return
			}
		}
		s.Users = append(s.Users, nu)
	})
	if err != nil {
		fail(w, 409, err.Error())
		return
	}
	h.audit(r, u.Login, "user.create", nu.Login)
	writeJSON(w, 200, map[string]string{"id": nu.ID})
}

func (h *Hub) updateUser(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct{ Name, Role, Password string }
	if readJSON(r, &b) != nil || (b.Role != "" && roleRank[b.Role] == 0) {
		fail(w, 400, "bad request")
		return
	}
	id := r.PathValue("id")
	var err error
	h.db.Update(func(s *store.State) {
		admins := 0
		for _, x := range s.Users {
			if x.Role == "admin" {
				admins++
			}
		}
		for _, x := range s.Users {
			if x.ID != id {
				continue
			}
			if b.Role != "" && x.Role == "admin" && b.Role != "admin" && admins == 1 {
				err = errors.New("cannot demote the last admin")
				return
			}
			if b.Name != "" {
				x.Name = b.Name
			}
			if b.Role != "" {
				x.Role = b.Role
			}
			if b.Password != "" {
				if len(b.Password) < 10 || x.Provider != "local" {
					err = errors.New("password (10+) can only be set on local users")
					return
				}
				x.PassHash = hashPassword(b.Password)
				so := s.Sessions[:0] // a reset password signs the user out everywhere
				for _, se := range s.Sessions {
					if se.UserID != x.ID {
						so = append(so, se)
					}
				}
				s.Sessions = so
			}
			return
		}
		err = errors.New("not found")
	})
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "user.update", id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) deleteUser(w http.ResponseWriter, r *http.Request, u *store.User) {
	id := r.PathValue("id")
	if id == u.ID {
		fail(w, 400, "you cannot delete yourself")
		return
	}
	var err error
	h.db.Update(func(s *store.State) {
		admins, target := 0, (*store.User)(nil)
		for _, x := range s.Users {
			if x.Role == "admin" {
				admins++
			}
			if x.ID == id {
				target = x
			}
		}
		if target == nil {
			err = errors.New("not found")
			return
		}
		if target.Role == "admin" && admins <= 1 {
			err = errors.New("cannot delete the last admin")
			return
		}
		out := s.Users[:0]
		for _, x := range s.Users {
			if x.ID != id {
				out = append(out, x)
			}
		}
		s.Users = out
		so := s.Sessions[:0]
		for _, se := range s.Sessions {
			if se.UserID != id {
				so = append(so, se)
			}
		}
		s.Sessions = so
		to := s.Tokens[:0]
		for _, t := range s.Tokens {
			if t.UserID != id {
				to = append(to, t)
			}
		}
		s.Tokens = to
	})
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	h.audit(r, u.Login, "user.delete", id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) listTokens(w http.ResponseWriter, r *http.Request, u *store.User) {
	out := []map[string]any{}
	h.db.View(func(s *store.State) {
		for _, t := range s.Tokens {
			if t.UserID == u.ID {
				out = append(out, map[string]any{"id": t.ID, "name": t.Name, "created": t.Created, "expires": t.Expires, "used": t.Used})
			}
		}
	})
	writeJSON(w, 200, out)
}

func (h *Hub) createToken(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct {
		Name string `json:"name"`
		Days int    `json:"days"` // 0 = never expires
	}
	if readJSON(r, &b) != nil || strings.TrimSpace(b.Name) == "" || b.Days < 0 || b.Days > 3650 {
		fail(w, 400, "name required; expiry 0-3650 days")
		return
	}
	tok := "dk_" + store.NewSecret(24)
	t := &store.APIToken{ID: store.NewID(), Name: strings.TrimSpace(b.Name), Hash: store.Hash(tok), UserID: u.ID, Created: time.Now()}
	if b.Days > 0 {
		t.Expires = t.Created.Add(time.Duration(b.Days) * 24 * time.Hour)
	}
	h.db.Update(func(s *store.State) { s.Tokens = append(s.Tokens, t) })
	h.audit(r, u.Login, "token.create", b.Name)
	writeJSON(w, 200, map[string]string{"token": tok})
}

func (h *Hub) deleteToken(w http.ResponseWriter, r *http.Request, u *store.User) {
	h.db.Update(func(s *store.State) {
		out := s.Tokens[:0]
		for _, t := range s.Tokens {
			if !(t.ID == r.PathValue("id") && t.UserID == u.ID) {
				out = append(out, t)
			}
		}
		s.Tokens = out
	})
	h.audit(r, u.Login, "token.delete", r.PathValue("id"))
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) getSettings(w http.ResponseWriter, r *http.Request, u *store.User) {
	var st store.Settings
	h.db.View(func(s *store.State) { st = s.Settings })
	provs := []string{}
	for k := range h.cfg.OAuth {
		provs = append(provs, k)
	}
	sort.Strings(provs)
	cb := map[string]string{}
	for _, p := range provs {
		cb[p] = h.baseURL(r) + "/auth/" + p + "/callback"
	}
	writeJSON(w, 200, map[string]any{"settings": st, "providers": provs, "publicUrl": h.cfg.PublicURL, "callbacks": cb,
		"trustProxy": h.cfg.TrustProxy, "version": h.cfg.Version})
}

func (h *Hub) putSettings(w http.ResponseWriter, r *http.Request, u *store.User) {
	var in store.Settings
	if readJSON(r, &in) != nil || (in.DefaultRole != "viewer" && in.DefaultRole != "operator") {
		fail(w, 400, "default role must be viewer or operator")
		return
	}
	clean := []string{}
	for _, a := range in.AllowedEmail {
		if a = strings.ToLower(strings.TrimSpace(a)); a != "" {
			if strings.Count(a, "@") != 1 {
				fail(w, 400, "allow-list entries must be an email or @domain: "+a)
				return
			}
			clean = append(clean, a)
		}
	}
	in.AllowedEmail = clean
	if in.AllowSignup && len(clean) == 0 {
		// otherwise any GitHub/Google account in the world could sign in
		fail(w, 400, "auto sign-up needs at least one allowed email or @domain")
		return
	}
	h.db.Update(func(s *store.State) { s.Settings = in })
	h.audit(r, u.Login, "settings.update", "")
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) listAudit(w http.ResponseWriter, r *http.Request, u *store.User) {
	csvOut := r.URL.Query().Get("format") == "csv"
	limit := 300
	if csvOut {
		limit = 1 << 30
	}
	q := strings.ToLower(r.URL.Query().Get("q"))
	var out []store.Audit
	h.db.View(func(s *store.State) {
		for i := len(s.Audit) - 1; i >= 0 && len(out) < limit; i-- {
			a := s.Audit[i]
			if q != "" && !strings.Contains(strings.ToLower(a.User+" "+a.Action+" "+a.Target+" "+a.IP), q) {
				continue
			}
			out = append(out, *a)
		}
	})
	if out == nil {
		out = []store.Audit{}
	}
	if csvOut {
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="dockup-audit-`+time.Now().Format("20060102")+`.csv"`)
		cw := csv.NewWriter(w)
		cw.Write([]string{"time", "user", "action", "target", "ip"})
		for _, a := range out {
			cw.Write([]string{a.Time.UTC().Format(time.RFC3339), csvSafe(a.User), csvSafe(a.Action), csvSafe(a.Target), csvSafe(a.IP)})
		}
		cw.Flush()
		return
	}
	writeJSON(w, 200, out)
}

// csvSafe neutralises spreadsheet formula injection in exported cells.
func csvSafe(v string) string {
	if v != "" && strings.ContainsRune("=+-@\t\r", rune(v[0])) {
		return "'" + v
	}
	return v
}
