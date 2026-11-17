package hub

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RamizMohammad/dockup/internal/store"
)

const cookieName = "dockup_session"

func hashPassword(pw string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	k, _ := pbkdf2.Key(sha256.New, pw, salt, 210000, 32)
	return "pbkdf2$" + hex.EncodeToString(salt) + "$" + hex.EncodeToString(k)
}

// dummyHash is checked against when a login name does not exist, so a failed
// login takes the same time whether or not the account exists.
var dummyHash = hashPassword("dockup-timing-equaliser")

func checkPassword(pw, enc string) bool {
	p := strings.Split(enc, "$")
	if len(p) != 3 || p[0] != "pbkdf2" {
		return false
	}
	salt, err := hex.DecodeString(p[1])
	if err != nil {
		return false
	}
	want, _ := hex.DecodeString(p[2])
	got, _ := pbkdf2.Key(sha256.New, pw, salt, 210000, 32)
	return subtle.ConstantTimeCompare(got, want) == 1
}

var roleRank = map[string]int{"viewer": 1, "operator": 2, "admin": 3}

type loginBucket struct {
	n     int
	reset time.Time
}

// clientIP returns the caller's address. Behind a trusted proxy the RIGHTMOST
// X-Forwarded-For entry is used: it is the one appended by our own proxy,
// while everything to its left is supplied by the client and can be forged.
func (h *Hub) clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if h.cfg.TrustProxy {
		if v := r.Header.Get("X-Forwarded-For"); v != "" {
			parts := strings.Split(v, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(ip) != nil {
				return ip
			}
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// hit records one attempt for key and reports whether the limit is exceeded.
func (h *Hub) hit(key string, limit int, window time.Duration) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	b := h.logins[key]
	if b == nil || time.Now().After(b.reset) {
		b = &loginBucket{reset: time.Now().Add(window)}
		h.logins[key] = b
	}
	b.n++
	return b.n > limit
}

func (h *Hub) limited(key string, limit int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	b := h.logins[key]
	return b != nil && time.Now().Before(b.reset) && b.n >= limit
}

func (h *Hub) clearHits(key string) {
	h.mu.Lock()
	delete(h.logins, key)
	h.mu.Unlock()
}

func (h *Hub) secureCookie(r *http.Request) bool {
	return strings.HasPrefix(h.baseURL(r), "https://")
}

func (h *Hub) newSession(w http.ResponseWriter, r *http.Request, userID string) {
	raw := make([]byte, 32)
	rand.Read(raw)
	tok := base64.RawURLEncoding.EncodeToString(raw)
	now := time.Now()
	exp := now.Add(time.Duration(h.cfg.SessionDays) * 24 * time.Hour)
	ua := r.UserAgent()
	if len(ua) > 200 {
		ua = ua[:200]
	}
	h.db.Update(func(s *store.State) {
		s.Sessions = append(s.Sessions, &store.Session{ID: store.NewID(), Token: store.Hash(tok), UserID: userID,
			Created: now, Seen: now, IP: h.clientIP(r), Agent: ua, Expires: exp})
		for _, u := range s.Users {
			if u.ID == userID {
				u.LastLogin = now
			}
		}
	})
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: tok, Path: "/", Expires: exp, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: h.secureCookie(r)})
}

// principal is the authenticated caller.
type principal struct {
	user      store.User
	sessionID string // empty for API tokens
	viaCookie bool
}

// authenticate resolves the caller from a bearer API token or session cookie.
func (h *Hub) authenticate(r *http.Request) *principal {
	now := time.Now()
	// API tokens: dk_… (dh_… from before the rename keeps working)
	if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer dk_") || strings.HasPrefix(a, "Bearer dh_") {
		hash := store.Hash(strings.TrimPrefix(a, "Bearer "))
		var out *principal
		h.db.Mutate(func(s *store.State) bool {
			for _, t := range s.Tokens {
				if subtle.ConstantTimeCompare([]byte(t.Hash), []byte(hash)) != 1 {
					continue
				}
				if !t.Expires.IsZero() && now.After(t.Expires) {
					return false
				}
				for _, u := range s.Users {
					if u.ID == t.UserID {
						out = &principal{user: *u}
						t.Used = now // volatile: persisted with the next real change
					}
				}
			}
			return false
		})
		return out
	}
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	hash := store.Hash(c.Value)
	var out *principal
	h.db.Mutate(func(s *store.State) bool {
		for _, se := range s.Sessions {
			if subtle.ConstantTimeCompare([]byte(se.Token), []byte(hash)) == 1 && se.Expires.After(now) {
				for _, u := range s.Users {
					if u.ID == se.UserID {
						out = &principal{user: *u, sessionID: se.ID, viaCookie: true}
						se.Seen = now
					}
				}
			}
		}
		return false
	})
	return out
}

type sessionKey struct{}

// sessionID returns the caller's session ID ("" for API tokens).
func sessionID(r *http.Request) string {
	v, _ := r.Context().Value(sessionKey{}).(string)
	return v
}

type authed func(w http.ResponseWriter, r *http.Request, u *store.User)

func (h *Hub) auth(min string, fn authed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := h.authenticate(r)
		if p == nil {
			fail(w, 401, "not signed in")
			return
		}
		if p.viaCookie && r.Method != "GET" && r.Method != "HEAD" && r.Header.Get("X-Dockup") == "" && r.Header.Get("X-DH") == "" {
			fail(w, 403, "missing X-Dockup header")
			return
		}
		if roleRank[p.user.Role] < roleRank[min] {
			fail(w, 403, "insufficient role")
			return
		}
		fn(w, r.WithContext(context.WithValue(r.Context(), sessionKey{}, p.sessionID)), &p.user)
	}
}

func publicUser(u *store.User) map[string]any {
	return map[string]any{"id": u.ID, "login": u.Login, "name": u.Name, "email": u.Email, "role": u.Role, "provider": u.Provider}
}

// ---------- first-run setup ----------

// initSetupCode creates a one-time code while no account exists. Creating the
// first admin requires it, so a freshly installed hub reachable from the
// network cannot be claimed by whoever finds it first.
func (h *Hub) initSetupCode() {
	var empty bool
	h.db.View(func(s *store.State) { empty = len(s.Users) == 0 })
	p := filepath.Join(h.cfg.DataDir, "setup-code")
	if !empty {
		os.Remove(p)
		return
	}
	code := ""
	if b, err := os.ReadFile(p); err == nil && len(strings.TrimSpace(string(b))) >= 12 {
		code = strings.TrimSpace(string(b))
	} else {
		raw := make([]byte, 9)
		rand.Read(raw)
		code = strings.ToUpper(hex.EncodeToString(raw)[:12])
		code = code[:4] + "-" + code[4:8] + "-" + code[8:]
		os.WriteFile(p, []byte(code+"\n"), 0o600)
	}
	h.mu.Lock()
	h.setupCode = code
	h.mu.Unlock()
	log.Printf("first-run setup code: %s   (also in %s)", code, p)
}

func (h *Hub) handleSetup(w http.ResponseWriter, r *http.Request) {
	if h.hit("setup|"+h.clientIP(r), 10, 10*time.Minute) {
		fail(w, 429, "too many attempts, try later")
		return
	}
	var b struct{ Login, Password, Name, Email, Code string }
	if readJSON(r, &b) != nil || len(b.Login) < 2 || len(b.Password) < 10 {
		fail(w, 400, "login (2+ chars) and password (10+ chars) required")
		return
	}
	h.mu.Lock()
	code := h.setupCode
	h.mu.Unlock()
	norm := func(v string) string { return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(v), "-", "")) }
	if code == "" || subtle.ConstantTimeCompare([]byte(norm(b.Code)), []byte(norm(code))) != 1 {
		fail(w, 403, "invalid setup code: see the hub log (journalctl -u dockup) or <data dir>/setup-code")
		return
	}
	created := false
	var id string
	h.db.Update(func(s *store.State) {
		if len(s.Users) > 0 {
			return
		}
		id = store.NewID()
		s.Users = append(s.Users, &store.User{ID: id, Login: b.Login, Name: b.Name, Email: b.Email, Provider: "local", Role: "admin",
			PassHash: hashPassword(b.Password), CreatedAt: time.Now()})
		created = true
	})
	if !created {
		fail(w, 409, "setup already completed")
		return
	}
	h.mu.Lock()
	h.setupCode = ""
	h.mu.Unlock()
	os.Remove(filepath.Join(h.cfg.DataDir, "setup-code"))
	h.audit(r, b.Login, "setup", "first admin created")
	h.newSession(w, r, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// ---------- login / logout ----------

func (h *Hub) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := h.clientIP(r)
	if h.hit("ip|"+ip, 30, 10*time.Minute) {
		fail(w, 429, "too many attempts, try later")
		return
	}
	var b struct{ Login, Password string }
	if readJSON(r, &b) != nil || b.Login == "" {
		fail(w, 400, "bad request")
		return
	}
	acct := "acct|" + strings.ToLower(b.Login)
	if h.limited(acct, 10) {
		fail(w, 429, "account temporarily locked after repeated failures, try again in 15 minutes")
		return
	}
	var uid, hash string
	h.db.View(func(s *store.State) {
		for _, u := range s.Users {
			if u.Provider == "local" && strings.EqualFold(u.Login, b.Login) {
				uid, hash = u.ID, u.PassHash
			}
		}
	})
	// hash outside the store lock; always hash, even for unknown users
	if hash == "" {
		hash = dummyHash
	}
	if !checkPassword(b.Password, hash) || uid == "" {
		h.hit(acct, 10, 15*time.Minute)
		h.audit(r, b.Login, "login.failed", "local")
		fail(w, 401, "invalid credentials")
		return
	}
	h.clearHits(acct)
	h.newSession(w, r, uid)
	h.audit(r, b.Login, "login", "local")
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		hash := store.Hash(c.Value)
		h.db.Update(func(s *store.State) {
			out := s.Sessions[:0]
			for _, se := range s.Sessions {
				if se.Token != hash {
					out = append(out, se)
				}
			}
			s.Sessions = out
		})
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: h.secureCookie(r)})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) handleBoot(w http.ResponseWriter, r *http.Request) {
	var needSetup bool
	h.db.View(func(s *store.State) { needSetup = len(s.Users) == 0 })
	provs := []map[string]string{}
	for _, k := range sortedKeys(h.cfg.OAuth) {
		provs = append(provs, map[string]string{"id": k, "name": h.cfg.OAuth[k].Label, "kind": h.cfg.OAuth[k].Kind})
	}
	out := map[string]any{"needSetup": needSetup, "providers": provs, "version": h.cfg.Version, "instance": h.instanceName(r),
		"tls": h.secureCookie(r)}
	if p := h.authenticate(r); p != nil {
		out["user"] = publicUser(&p.user)
	}
	writeJSON(w, 200, out)
}

func (h *Hub) instanceName(r *http.Request) string {
	u := h.baseURL(r)
	u = strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
	return u
}

// ---------- self service ----------

func (h *Hub) me(w http.ResponseWriter, r *http.Request, u *store.User) {
	m := publicUser(u)
	m["lastLogin"] = u.LastLogin
	writeJSON(w, 200, m)
}

func (h *Hub) changePassword(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct{ Current, New string }
	if readJSON(r, &b) != nil || len(b.New) < 10 {
		fail(w, 400, "new password must be at least 10 characters")
		return
	}
	if u.Provider != "local" {
		fail(w, 400, "password sign-in is not enabled for this account (it uses "+u.Provider+")")
		return
	}
	if h.hit("pw|"+u.ID, 10, 15*time.Minute) {
		fail(w, 429, "too many attempts, try later")
		return
	}
	if !checkPassword(b.Current, u.PassHash) {
		fail(w, 403, "current password is wrong")
		return
	}
	cur := sessionID(r)
	nh := hashPassword(b.New)
	h.db.Update(func(s *store.State) {
		for _, x := range s.Users {
			if x.ID == u.ID {
				x.PassHash = nh
			}
		}
		// sign out every other session of this user
		out := s.Sessions[:0]
		for _, se := range s.Sessions {
			if se.UserID != u.ID || se.ID == cur {
				out = append(out, se)
			}
		}
		s.Sessions = out
	})
	h.audit(r, u.Login, "user.password", u.Login)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) listSessions(w http.ResponseWriter, r *http.Request, u *store.User) {
	cur := sessionID(r)
	out := []map[string]any{}
	h.db.View(func(s *store.State) {
		for _, se := range s.Sessions {
			if se.UserID == u.ID && se.Expires.After(time.Now()) {
				out = append(out, map[string]any{"id": se.ID, "created": se.Created, "seen": se.Seen, "ip": se.IP,
					"agent": se.Agent, "expires": se.Expires, "current": se.ID == cur && cur != ""})
			}
		}
	})
	writeJSON(w, 200, out)
}

// DELETE /api/me/sessions/{id} ; id "others" signs out every other session.
func (h *Hub) revokeSession(w http.ResponseWriter, r *http.Request, u *store.User) {
	id, cur := r.PathValue("id"), sessionID(r)
	n := 0
	h.db.Update(func(s *store.State) {
		out := s.Sessions[:0]
		for _, se := range s.Sessions {
			kill := se.UserID == u.ID && ((id == "others" && se.ID != cur) || se.ID == id)
			if kill {
				n++
				continue
			}
			out = append(out, se)
		}
		s.Sessions = out
	})
	h.audit(r, u.Login, "session.revoke", id)
	writeJSON(w, 200, map[string]int{"revoked": n})
}

// janitor purges expired sessions and stale throttles.
func (h *Hub) janitor(ctx interface{ Done() <-chan struct{} }) {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			now := time.Now()
			h.db.Mutate(func(s *store.State) bool {
				out := s.Sessions[:0]
				for _, se := range s.Sessions {
					if se.Expires.After(now) {
						out = append(out, se)
					}
				}
				changed := len(out) != len(s.Sessions)
				s.Sessions = out
				return changed
			})
			h.mu.Lock()
			for k, b := range h.logins {
				if now.After(b.reset) {
					delete(h.logins, k)
				}
			}
			for k, st := range h.oauthSt {
				if now.After(st.exp) {
					delete(h.oauthSt, k)
				}
			}
			h.mu.Unlock()
		}
	}
}
