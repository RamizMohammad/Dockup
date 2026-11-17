// Package hub is the Dockup control plane: HTTP API, auth, build queue,
// scheduler and the link to remote agents.
package hub

import (
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/RamizMohammad/dockup/internal/runner"
	"github.com/RamizMohammad/dockup/internal/store"
)

type Config struct {
	Addr        string
	DataDir     string
	PublicURL   string // e.g. https://deploy.example.com (used for OAuth callbacks & agent commands)
	Version     string
	AdminUser   string
	AdminPass   string
	AdminEmail  string // invite this email as admin (OAuth) on first start
	OAuth       map[string]*OAuthProvider
	TrustProxy  bool
	WebFS       fs.FS
	SessionDays int
}

type Hub struct {
	cfg   Config
	db    *store.Store
	local *runner.Runner

	mu        sync.Mutex
	running   map[string]context.CancelFunc // local builds
	cancelReq map[string]bool
	cronSeen  string
	agents    map[string]*agentConn
	logs      *logHub
	logins    map[string]*loginBucket
	oauthSt   map[string]oauthState
	setupCode string
	agentRun  map[string]map[string]bool // serverID -> build IDs the agent reports as running
	httpSrv   *http.Server
}

func New(cfg Config) (*Hub, error) {
	db, err := store.Open(cfg.DataDir)
	if err != nil {
		return nil, err
	}
	if cfg.SessionDays == 0 {
		cfg.SessionDays = 14
	}
	h := &Hub{
		cfg:       cfg,
		db:        db,
		local:     &runner.Runner{WorkDir: cfg.DataDir + "/work"},
		running:   map[string]context.CancelFunc{},
		cancelReq: map[string]bool{},
		agents:    map[string]*agentConn{},
		logs:      newLogHub(db),
		logins:    map[string]*loginBucket{},
		oauthSt:   map[string]oauthState{},
		agentRun:  map[string]map[string]bool{},
	}
	h.logs.onStage = h.stageEvent
	os.MkdirAll(cfg.DataDir+"/work", 0o750)
	h.bootstrap()
	h.initSetupCode()
	return h, nil
}

func (h *Hub) bootstrap() {
	host, osn, dv := runner.Info()
	h.db.Update(func(s *store.State) {
		var local *store.Server
		for _, sv := range s.Servers {
			if sv.Local {
				local = sv
			}
		}
		if local == nil {
			local = &store.Server{ID: store.NewID(), Name: "local", Local: true, Created: time.Now()}
			s.Servers = append(s.Servers, local)
		}
		local.Hostname, local.OS, local.Docker, local.Version = host, osn, dv, h.cfg.Version
		local.LastSeen = time.Now()
		if len(s.Users) == 0 && h.cfg.AdminUser != "" && h.cfg.AdminPass != "" {
			s.Users = append(s.Users, &store.User{ID: store.NewID(), Login: h.cfg.AdminUser, Name: h.cfg.AdminUser,
				Provider: "local", Role: "admin", PassHash: hashPassword(h.cfg.AdminPass), CreatedAt: time.Now()})
		}
		if len(s.Users) == 0 && h.cfg.AdminEmail != "" {
			// invited admin: binds on first OAuth sign-in with this verified email
			s.Users = append(s.Users, &store.User{ID: store.NewID(), Login: strings.Split(h.cfg.AdminEmail, "@")[0],
				Email: h.cfg.AdminEmail, Role: "admin", CreatedAt: time.Now()})
		}
		// builds left "running"/"queued" by a crash are failed on boot
		for _, b := range s.Builds {
			if b.Status == "running" {
				b.Status, b.Finished, b.Reason = "failed", time.Now(), "hub restarted while the build was running"
				for i := range b.Stages {
					if b.Stages[i].Status == "running" {
						b.Stages[i].Status, b.Stages[i].End = "failed", b.Finished
					}
				}
			}
		}
	})
}

func (h *Hub) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	h.routes(mux)
	h.httpSrv = &http.Server{Addr: h.cfg.Addr, Handler: h.secureHeaders(mux), ReadHeaderTimeout: 15 * time.Second}
	go h.pumpLoop(ctx)
	go h.cronLoop(ctx)
	go h.janitor(ctx)
	errc := make(chan error, 1)
	go func() { errc <- h.httpSrv.ListenAndServe() }()
	log.Printf("dockup %s listening on %s (data: %s)", h.cfg.Version, h.cfg.Addr, h.cfg.DataDir)
	select {
	case <-ctx.Done():
		sc, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		h.httpSrv.Shutdown(sc)
		h.mu.Lock()
		for _, c := range h.running {
			c()
		}
		h.mu.Unlock()
		time.Sleep(500 * time.Millisecond)
		h.db.Flush()
		return nil
	case err := <-errc:
		return err
	}
}

func (h *Hub) secureHeaders(n http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hd := w.Header()
		hd.Set("X-Content-Type-Options", "nosniff")
		hd.Set("X-Frame-Options", "DENY")
		hd.Set("Referrer-Policy", "same-origin")
		hd.Set("Cross-Origin-Opener-Policy", "same-origin")
		hd.Set("Cross-Origin-Resource-Policy", "same-origin")
		hd.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		hd.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; "+
			"font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		if h.secureCookie(r) {
			hd.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			hd.Set("Cache-Control", "no-store")
		}
		n.ServeHTTP(w, r)
	})
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *Hub) baseURL(r *http.Request) string {
	if h.cfg.PublicURL != "" {
		return strings.TrimRight(h.cfg.PublicURL, "/")
	}
	sch := "http"
	if r.TLS != nil || (h.cfg.TrustProxy && r.Header.Get("X-Forwarded-Proto") == "https") {
		sch = "https"
	}
	host := r.Host
	if h.cfg.TrustProxy && r.Header.Get("X-Forwarded-Host") != "" {
		host = r.Header.Get("X-Forwarded-Host")
	}
	return sch + "://" + host
}

// audit records who did what. r may be nil for system actions.
func (h *Hub) audit(r *http.Request, user, action, target string) {
	ip := h.clientIP(r)
	if user == "" {
		user = "system"
	}
	h.db.Update(func(s *store.State) {
		s.Audit = append(s.Audit, &store.Audit{Time: time.Now(), User: user, Action: action, Target: target, IP: ip})
		if len(s.Audit) > 5000 {
			s.Audit = s.Audit[len(s.Audit)-5000:]
		}
	})
}
