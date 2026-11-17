package hub

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/RamizMohammad/dockup/internal/envx"
	"github.com/RamizMohammad/dockup/internal/store"
)

// OAuthProvider is a standard authorization-code (+PKCE) provider.
type OAuthProvider struct {
	ID, Label              string
	ClientID, ClientSecret string
	AuthURL, TokenURL      string
	UserURL, EmailsURL     string
	Scopes                 string
	Kind                   string // github | google | oidc
}

type oauthState struct {
	provider, verifier string
	exp                time.Time
}

// OAuthFromEnv builds providers from DOCKUP_* environment variables.
func OAuthFromEnv() map[string]*OAuthProvider {
	out := map[string]*OAuthProvider{}
	if id, sec := envx.Get("GITHUB_CLIENT_ID"), envx.Get("GITHUB_CLIENT_SECRET"); id != "" && sec != "" {
		out["github"] = &OAuthProvider{ID: "github", Label: "GitHub", ClientID: id, ClientSecret: sec, Kind: "github",
			AuthURL: "https://github.com/login/oauth/authorize", TokenURL: "https://github.com/login/oauth/access_token",
			UserURL: "https://api.github.com/user", EmailsURL: "https://api.github.com/user/emails", Scopes: "read:user user:email"}
	}
	if id, sec := envx.Get("GOOGLE_CLIENT_ID"), envx.Get("GOOGLE_CLIENT_SECRET"); id != "" && sec != "" {
		out["google"] = &OAuthProvider{ID: "google", Label: "Google", ClientID: id, ClientSecret: sec, Kind: "google",
			AuthURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token",
			UserURL: "https://openidconnect.googleapis.com/v1/userinfo", Scopes: "openid email profile"}
	}
	if iss := strings.TrimRight(envx.Get("OIDC_ISSUER"), "/"); iss != "" {
		id, sec := envx.Get("OIDC_CLIENT_ID"), envx.Get("OIDC_CLIENT_SECRET")
		if id != "" && sec != "" {
			if p, err := discoverOIDC(iss); err == nil {
				p.ClientID, p.ClientSecret = id, sec
				if n := envx.Get("OIDC_NAME"); n != "" {
					p.Label = n
				}
				out["oidc"] = p
			} else {
				fmt.Fprintln(os.Stderr, "dockup: OIDC discovery failed:", err)
			}
		}
	}
	return out
}

func discoverOIDC(iss string) (*OAuthProvider, error) {
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(iss + "/.well-known/openid-configuration")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var d struct {
		Auth     string `json:"authorization_endpoint"`
		Token    string `json:"token_endpoint"`
		UserInfo string `json:"userinfo_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil || d.Auth == "" || d.Token == "" {
		return nil, errors.New("invalid discovery document")
	}
	return &OAuthProvider{ID: "oidc", Label: "SSO", Kind: "oidc", AuthURL: d.Auth, TokenURL: d.Token, UserURL: d.UserInfo,
		Scopes: "openid email profile"}, nil
}

func rnd(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (h *Hub) oauthLogin(w http.ResponseWriter, r *http.Request) {
	p := h.cfg.OAuth[r.PathValue("provider")]
	if p == nil {
		http.NotFound(w, r)
		return
	}
	state, verifier := rnd(24), rnd(48)
	h.mu.Lock()
	h.oauthSt[state] = oauthState{provider: p.ID, verifier: verifier, exp: time.Now().Add(10 * time.Minute)}
	h.mu.Unlock()
	chal := sha256.Sum256([]byte(verifier))
	q := url.Values{
		"client_id": {p.ClientID}, "response_type": {"code"}, "scope": {p.Scopes}, "state": {state},
		"redirect_uri":          {h.baseURL(r) + "/auth/" + p.ID + "/callback"},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(chal[:])},
		"code_challenge_method": {"S256"},
	}
	// bind the state to this browser
	http.SetCookie(w, &http.Cookie{Name: "dockup_oauth", Value: state, Path: "/", MaxAge: 600, HttpOnly: true,
		SameSite: http.SameSiteLaxMode, Secure: strings.HasPrefix(h.baseURL(r), "https://")})
	http.Redirect(w, r, p.AuthURL+"?"+q.Encode(), http.StatusFound)
}

func (h *Hub) oauthCallback(w http.ResponseWriter, r *http.Request) {
	p := h.cfg.OAuth[r.PathValue("provider")]
	if p == nil {
		http.NotFound(w, r)
		return
	}
	state := r.URL.Query().Get("state")
	ck, err := r.Cookie("dockup_oauth")
	h.mu.Lock()
	st, ok := h.oauthSt[state]
	delete(h.oauthSt, state)
	h.mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: "dockup_oauth", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: h.secureCookie(r)})
	if err != nil || ck.Value != state || !ok || st.provider != p.ID || time.Now().After(st.exp) {
		h.loginError(w, r, "login session expired, try again")
		return
	}
	if e := r.URL.Query().Get("error"); e != "" {
		h.loginError(w, r, e)
		return
	}
	prof, err := h.oauthExchange(p, r.URL.Query().Get("code"), st.verifier, h.baseURL(r)+"/auth/"+p.ID+"/callback")
	if err != nil {
		h.loginError(w, r, "sign-in failed: "+err.Error())
		return
	}
	u, why := h.resolveOAuthUser(p.ID, prof)
	if u == nil {
		h.audit(r, prof.Email, "login.denied", p.ID)
		h.loginError(w, r, why)
		return
	}
	h.newSession(w, r, u.ID)
	h.audit(r, u.Login, "login", p.ID)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *Hub) loginError(w http.ResponseWriter, r *http.Request, msg string) {
	http.Redirect(w, r, "/?error="+url.QueryEscape(msg), http.StatusFound)
}

type oauthProfile struct {
	Sub, Login, Name, Email string
	Verified                bool
}

func (h *Hub) oauthExchange(p *OAuthProvider, code, verifier, redirect string) (*oauthProfile, error) {
	if code == "" {
		return nil, errors.New("no code")
	}
	form := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "redirect_uri": {redirect},
		"client_id": {p.ClientID}, "client_secret": {p.ClientSecret}, "code_verifier": {verifier}}
	req, _ := http.NewRequest("POST", p.TokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	cl := &http.Client{Timeout: 15 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tok struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if json.Unmarshal(b, &tok) != nil || tok.AccessToken == "" {
		return nil, fmt.Errorf("token exchange failed %s", tok.Error)
	}
	get := func(u string, v any) error {
		rq, _ := http.NewRequest("GET", u, nil)
		rq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
		rq.Header.Set("Accept", "application/json")
		rs, err := cl.Do(rq)
		if err != nil {
			return err
		}
		defer rs.Body.Close()
		if rs.StatusCode != 200 {
			return fmt.Errorf("%s returned %d", u, rs.StatusCode)
		}
		return json.NewDecoder(io.LimitReader(rs.Body, 1<<20)).Decode(v)
	}
	prof := &oauthProfile{}
	switch p.Kind {
	case "github":
		var u struct {
			ID    int64  `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := get(p.UserURL, &u); err != nil {
			return nil, err
		}
		prof.Sub, prof.Login, prof.Name = fmt.Sprint(u.ID), u.Login, u.Name
		var emails []struct {
			Email             string `json:"email"`
			Primary, Verified bool
		}
		if get(p.EmailsURL, &emails) == nil {
			for _, e := range emails {
				if e.Primary && e.Verified {
					prof.Email, prof.Verified = e.Email, true
				}
			}
		}
	default:
		var u struct {
			Sub      string `json:"sub"`
			Name     string `json:"name"`
			Email    string `json:"email"`
			Verified bool   `json:"email_verified"`
			Username string `json:"preferred_username"`
		}
		if err := get(p.UserURL, &u); err != nil {
			return nil, err
		}
		prof.Sub, prof.Name, prof.Email, prof.Verified, prof.Login = u.Sub, u.Name, u.Email, u.Verified, u.Username
		if prof.Login == "" {
			prof.Login = strings.Split(u.Email, "@")[0]
		}
	}
	if prof.Sub == "" {
		return nil, errors.New("provider returned no user id")
	}
	return prof, nil
}

func emailAllowed(list []string, email string) bool {
	if len(list) == 0 || email == "" {
		return false
	}
	email = strings.ToLower(email)
	for _, a := range list {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == email || (strings.HasPrefix(a, "@") && strings.HasSuffix(email, a)) {
			return true
		}
	}
	return false
}

// resolveOAuthUser binds or creates the account according to policy:
//  1. existing identity -> sign in
//  2. admin pre-invited this verified email -> bind and sign in
//  3. signup enabled and email on the allow-list -> default role
//
// There is deliberately no "first OAuth user becomes admin" rule: the first
// admin comes from the setup code, DOCKUP_ADMIN_USER/PASSWORD, or an
// invite created by DOCKUP_ADMIN_EMAIL.
func (h *Hub) resolveOAuthUser(prov string, p *oauthProfile) (*store.User, string) {
	var out *store.User
	why := "this account is not allowed. Ask an admin to invite " + p.Email
	h.db.Update(func(s *store.State) {
		for _, u := range s.Users {
			if u.Provider == prov && u.ExtID == p.Sub {
				c := *u
				out = &c
				return
			}
		}
		if p.Email != "" && p.Verified {
			for _, u := range s.Users {
				if u.Provider == "" && strings.EqualFold(u.Email, p.Email) {
					u.Provider, u.ExtID = prov, p.Sub
					if u.Name == "" {
						u.Name = p.Name
					}
					c := *u
					out = &c
					return
				}
			}
		}
		role := ""
		switch {
		case s.Settings.AllowSignup && len(s.Settings.AllowedEmail) > 0 && p.Verified && emailAllowed(s.Settings.AllowedEmail, p.Email):
			role = s.Settings.DefaultRole
			if role == "" {
				role = "viewer"
			}
		}
		if role == "" {
			return
		}
		login := p.Login
		for _, u := range s.Users {
			if strings.EqualFold(u.Login, login) {
				login = prov + ":" + login
			}
		}
		nu := &store.User{ID: store.NewID(), Login: login, Name: p.Name, Email: p.Email, Provider: prov, ExtID: p.Sub, Role: role, CreatedAt: time.Now()}
		s.Users = append(s.Users, nu)
		c := *nu
		out = &c
	})
	if out == nil {
		return nil, why
	}
	return out, ""
}
