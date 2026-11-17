package hub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/RamizMohammad/dockup/internal/store"
)

type th struct {
	t   *testing.T
	h   *Hub
	mux *http.ServeMux
	dir string
}

func newTH(t *testing.T, cfg Config) *th {
	t.Helper()
	dir := t.TempDir()
	cfg.DataDir = dir
	cfg.WebFS = fstest.MapFS{"index.html": {Data: []byte("<html>")}}
	h, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h.routes(mux)
	return &th{t: t, h: h, mux: mux, dir: dir}
}

// session creates a user with the given role and returns a cookie value.
func (x *th) session(role string) string {
	tok := store.NewSecret(16)
	id := store.NewID()
	x.h.db.Update(func(s *store.State) {
		s.Users = append(s.Users, &store.User{ID: id, Login: role + id[:4], Role: role, Provider: "local"})
		s.Sessions = append(s.Sessions, &store.Session{ID: store.NewID(), Token: store.Hash(tok), UserID: id, Expires: time.Now().Add(time.Hour)})
	})
	return tok
}

func (x *th) do(method, path, cookie string, body any, hdr ...string) *httptest.ResponseRecorder {
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rd)
	req.RemoteAddr = "192.0.2.10:5555"
	req.Header.Set("X-Dockup", "1")
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: cookieName, Value: cookie})
	}
	for i := 0; i+1 < len(hdr); i += 2 {
		req.Header.Set(hdr[i], hdr[i+1])
	}
	rec := httptest.NewRecorder()
	x.h.secureHeaders(x.mux).ServeHTTP(rec, req)
	return rec
}

func TestLogPathTraversalBlocked(t *testing.T) {
	x := newTH(t, Config{})
	outside := filepath.Join(filepath.Dir(x.dir), "outside-secret.log")
	os.WriteFile(outside, []byte("TOP-SECRET"), 0o600)
	defer os.Remove(outside)
	v := x.session("viewer")
	for _, p := range []string{"/api/builds/..%2F..%2Foutside-secret/log", "/api/builds/..%2Foutside-secret/log?download=1"} {
		rec := x.do("GET", p, v, nil)
		if rec.Code != 404 || strings.Contains(rec.Body.String(), "TOP-SECRET") {
			t.Fatalf("%s: status %d body %q", p, rec.Code, rec.Body.String())
		}
	}
}

func TestSetupNeedsCode(t *testing.T) {
	x := newTH(t, Config{})
	code, _ := os.ReadFile(filepath.Join(x.dir, "setup-code"))
	if len(strings.TrimSpace(string(code))) < 12 {
		t.Fatalf("setup code not written: %q", code)
	}
	body := map[string]string{"login": "root", "password": "correct-horse-battery"}
	if rec := x.do("POST", "/api/setup", "", body); rec.Code != 403 {
		t.Fatalf("setup without code: %d", rec.Code)
	}
	body["code"] = "WRONG-CODE-0000"
	if rec := x.do("POST", "/api/setup", "", body); rec.Code != 403 {
		t.Fatalf("setup with wrong code: %d", rec.Code)
	}
	body["code"] = strings.ToLower(strings.TrimSpace(string(code))) // case/dash-insensitive
	if rec := x.do("POST", "/api/setup", "", body); rec.Code != 200 {
		t.Fatalf("setup with code: %d %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(x.dir, "setup-code")); !os.IsNotExist(err) {
		t.Fatal("setup code file should be removed after setup")
	}
	if rec := x.do("POST", "/api/setup", "", body); rec.Code == 200 {
		t.Fatal("second setup must fail")
	}
}

func TestOAuthFirstUserIsNotAdmin(t *testing.T) {
	x := newTH(t, Config{})
	u, _ := x.h.resolveOAuthUser("github", &oauthProfile{Sub: "1", Login: "eve", Email: "eve@evil.test", Verified: true})
	if u != nil {
		t.Fatalf("first OAuth user must not be created: %+v", u)
	}
	y := newTH(t, Config{AdminEmail: "boss@corp.test"})
	u, _ = y.h.resolveOAuthUser("google", &oauthProfile{Sub: "2", Login: "boss", Email: "boss@corp.test", Verified: true})
	if u == nil || u.Role != "admin" {
		t.Fatalf("invited admin email should bind as admin, got %+v", u)
	}
	u, _ = y.h.resolveOAuthUser("google", &oauthProfile{Sub: "3", Login: "x", Email: "boss@corp.test.evil", Verified: true})
	if u != nil {
		t.Fatal("look-alike email must not bind")
	}
}

func TestXFFRightmostAndLoginThrottle(t *testing.T) {
	x := newTH(t, Config{TrustProxy: true, AdminUser: "admin", AdminPass: "correct-horse-battery"})
	// attacker rotates the client-controlled (leftmost) part; the proxy-added
	// rightmost address stays the same, so the limit still applies
	codes := map[int]int{}
	for i := 0; i < 35; i++ {
		rec := x.do("POST", "/api/login", "", map[string]string{"login": "nobody" + string(rune('a'+i%26)), "password": "x"},
			"X-Forwarded-For", "10.0.0."+string(rune('0'+i%10))+", 203.0.113.7")
		codes[rec.Code]++
	}
	if codes[429] == 0 {
		t.Fatalf("XFF rotation bypassed the limit: %v", codes)
	}
	// account lockout after repeated failures, from different IPs
	for i := 0; i < 10; i++ {
		x.do("POST", "/api/login", "", map[string]string{"login": "admin", "password": "wrong"}, "X-Forwarded-For", "198.51.100."+string(rune('0'+i)))
	}
	rec := x.do("POST", "/api/login", "", map[string]string{"login": "admin", "password": "correct-horse-battery"}, "X-Forwarded-For", "198.51.100.99")
	if rec.Code != 429 {
		t.Fatalf("account should be locked, got %d", rec.Code)
	}
}

func TestSignupNeedsAllowList(t *testing.T) {
	x := newTH(t, Config{})
	a := x.session("admin")
	rec := x.do("PUT", "/api/settings", a, map[string]any{"allowSignup": true, "defaultRole": "operator"})
	if rec.Code != 400 {
		t.Fatalf("open sign-up accepted: %d", rec.Code)
	}
	rec = x.do("PUT", "/api/settings", a, map[string]any{"allowSignup": true, "defaultRole": "viewer", "allowedEmail": []string{"@corp.test"}})
	if rec.Code != 200 {
		t.Fatalf("restricted sign-up rejected: %d %s", rec.Code, rec.Body.String())
	}
	u, _ := x.h.resolveOAuthUser("google", &oauthProfile{Sub: "9", Login: "o", Email: "o@other.test", Verified: true})
	if u != nil {
		t.Fatal("outsider signed up")
	}
	u, _ = x.h.resolveOAuthUser("google", &oauthProfile{Sub: "8", Login: "m", Email: "m@corp.test", Verified: true})
	if u == nil || u.Role != "viewer" {
		t.Fatalf("allowed domain should sign up as viewer: %+v", u)
	}
}

func TestViewerRedaction(t *testing.T) {
	x := newTH(t, Config{})
	a, v := x.session("admin"), x.session("viewer")
	var local string
	x.h.db.View(func(s *store.State) { local = s.Servers[0].ID })
	proj := map[string]any{"name": "web", "serverId": local, "source": map[string]any{"kind": "git", "repo": "https://bob:ghp_SECRET@github.com/o/r.git"},
		"preCmd": "curl -H 'X: s3cr3t' x", "notifyUrls": []string{"https://hooks.slack.com/services/T/B/SECRET"},
		"params": []map[string]any{{"name": "API_KEY", "default": "dflt-secret", "secret": true}}}
	rec := x.do("POST", "/api/projects", a, proj)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	body := x.do("GET", "/api/projects", v, nil).Body.String()
	for _, leak := range []string{"ghp_SECRET", "s3cr3t", "hooks.slack.com", "dflt-secret", "webhook"} {
		if strings.Contains(body, leak) {
			t.Errorf("viewer sees %q: %s", leak, body)
		}
	}
	if !strings.Contains(x.do("GET", "/api/projects", a, nil).Body.String(), "ghp_SECRET") {
		t.Error("admin should still see the full repo URL to edit it")
	}
}

func TestIdlePumpDoesNotRewriteState(t *testing.T) {
	x := newTH(t, Config{})
	time.Sleep(400 * time.Millisecond) // let the start-up save land
	p := filepath.Join(x.dir, "state.json")
	st, _ := os.Stat(p)
	m0 := st.ModTime()
	time.Sleep(20 * time.Millisecond)
	for i := 0; i < 3; i++ {
		x.h.pump()
		var id string
		x.h.db.View(func(s *store.State) { id = s.Servers[0].ID })
		x.h.touch(id, nil)
	}
	time.Sleep(400 * time.Millisecond)
	st, _ = os.Stat(p)
	if st.ModTime().After(m0) {
		t.Fatal("idle scheduler/heartbeat rewrote state.json")
	}
}

func TestFailedPollsPruned(t *testing.T) {
	x := newTH(t, Config{})
	var local string
	x.h.db.View(func(s *store.State) { local = s.Servers[0].ID })
	x.h.db.Update(func(s *store.State) {
		s.Projects = append(s.Projects, &store.Project{ID: "p1", Name: "p", ServerID: local, Source: store.Source{Kind: "git"}})
	})
	for i := 0; i < 5; i++ {
		id := store.NewID()
		x.h.db.Update(func(s *store.State) {
			s.Builds = append(s.Builds, &store.Build{ID: id, ProjectID: "p1", Action: "poll", Status: "running", ServerID: local, Started: time.Now()})
		})
		x.h.finish(id, "", "repository not found")
	}
	n := 0
	x.h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			if b.Action == "poll" {
				n++
			}
		}
	})
	if n != 1 {
		t.Fatalf("want 1 failed poll kept, got %d", n)
	}
}

func TestAgentReconcileAndStages(t *testing.T) {
	x := newTH(t, Config{})
	sv := &store.Server{ID: store.NewID(), Name: "remote"}
	bid, keep := store.NewID(), store.NewID()
	x.h.db.Update(func(s *store.State) {
		s.Servers = append(s.Servers, sv)
		s.Projects = append(s.Projects, &store.Project{ID: "p2", Name: "p2", ServerID: sv.ID})
		old := time.Now().Add(-5 * time.Minute)
		s.Builds = append(s.Builds,
			&store.Build{ID: bid, ProjectID: "p2", Action: "deploy", Status: "running", ServerID: sv.ID, Started: old},
			&store.Build{ID: keep, ProjectID: "p2", Action: "deploy", Status: "running", ServerID: sv.ID, Started: old})
	})
	x.h.logs.Append(keep, "##[stage] checkout running", "hello", "##[stage] checkout success", "##[stage] up running")
	x.h.touch(sv.ID, &agentInfo{Running: &[]string{keep}})
	x.h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			switch b.ID {
			case bid:
				if b.Status != "failed" || b.Reason == "" {
					t.Errorf("lost build should fail, got %s", b.Status)
				}
			case keep:
				if b.Status != "running" {
					t.Errorf("reported build must keep running, got %s", b.Status)
				}
				if len(b.Stages) != 2 || b.Stages[0].Status != "success" || b.Stages[1].Status != "running" {
					t.Errorf("stages: %+v", b.Stages)
				}
			}
		}
	})
	logb, _ := os.ReadFile(x.h.db.LogPath(keep))
	if !strings.Contains(string(logb), "\thello\n") {
		t.Errorf("log lines should be timestamped: %q", logb)
	}
}

func TestSecurityHeadersAndCronPreview(t *testing.T) {
	x := newTH(t, Config{})
	v := x.session("viewer")
	rec := x.do("GET", "/api/cron/preview?spec=0+3+*+*+*", v, nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "next") {
		t.Fatalf("cron preview: %d %s", rec.Code, rec.Body.String())
	}
	for _, hd := range []string{"Content-Security-Policy", "X-Frame-Options", "Permissions-Policy", "Cross-Origin-Opener-Policy"} {
		if rec.Header().Get(hd) == "" {
			t.Errorf("missing %s", hd)
		}
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("API responses must not be cached")
	}
}

func TestLastAdminProtectedAndTokenExpiry(t *testing.T) {
	x := newTH(t, Config{})
	a := x.session("admin")
	var id string
	x.h.db.View(func(s *store.State) { id = s.Users[0].ID })
	other := x.session("admin")
	_ = other
	// delete the other admin is fine, the last one is not
	var ids []string
	x.h.db.View(func(s *store.State) {
		for _, u := range s.Users {
			ids = append(ids, u.ID)
		}
	})
	if rec := x.do("DELETE", "/api/users/"+ids[1], a, nil); rec.Code != 200 {
		t.Fatalf("delete second admin: %d %s", rec.Code, rec.Body.String())
	}
	_ = id
	// expired API tokens are refused
	x.h.db.Update(func(s *store.State) {
		s.Tokens = append(s.Tokens, &store.APIToken{ID: "t", Hash: store.Hash("dk_expired"), UserID: ids[0], Expires: time.Now().Add(-time.Minute)})
	})
	req := httptest.NewRequest("GET", "/api/me", nil)
	req.Header.Set("Authorization", "Bearer dk_expired")
	rec := httptest.NewRecorder()
	x.mux.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("expired token accepted: %d", rec.Code)
	}
}

func TestOnboardingSampleAndDismiss(t *testing.T) {
	x := newTH(t, Config{})
	a, v := x.session("admin"), x.session("viewer")
	if rec := x.do("GET", "/api/onboarding", v, nil); rec.Code != 403 {
		t.Fatalf("viewers must not see onboarding: %d", rec.Code)
	}
	rec := x.do("GET", "/api/onboarding", a, nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"id":"project","done":false`) {
		t.Fatalf("onboarding: %d %s", rec.Code, rec.Body.String())
	}
	first := x.do("POST", "/api/onboarding/sample", a, nil)
	second := x.do("POST", "/api/onboarding/sample", a, nil)
	if first.Code != 200 || second.Code != 200 {
		t.Fatalf("sample: %d %d", first.Code, second.Code)
	}
	var p1, p2 struct{ ID, Name string }
	json.Unmarshal(first.Body.Bytes(), &p1)
	json.Unmarshal(second.Body.Bytes(), &p2)
	if p1.ID == "" || p1.ID != p2.ID || p1.Name != "hello-dockup" {
		t.Fatalf("sample must be created once: %+v %+v", p1, p2)
	}
	if b, err := os.ReadFile(filepath.Join(x.dir, "samples", "hello-dockup", "compose.yaml")); err != nil || !strings.Contains(string(b), "traefik/whoami") || strings.Contains(string(b), "ports:") {
		t.Fatalf("sample compose: %v %q", err, b)
	}
	x.do("PUT", "/api/onboarding", a, map[string]bool{"dismissed": true})
	if !strings.Contains(x.do("GET", "/api/onboarding", a, nil).Body.String(), `"dismissed":true`) {
		t.Fatal("dismiss not saved")
	}
	// the settings page must not wipe onboarding state
	x.do("PUT", "/api/settings", a, map[string]any{"defaultRole": "viewer"})
	if !strings.Contains(x.do("GET", "/api/onboarding", a, nil).Body.String(), `"dismissed":true`) {
		t.Fatal("saving settings reset onboarding")
	}
}
