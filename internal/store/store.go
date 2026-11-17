// Package store is Dockup's persistence layer: one JSON state file (atomic
// writes), per-build log files, and AES-GCM sealing for secrets.
package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/RamizMohammad/dockup/internal/envx"
)

type User struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"` // local | github | google | oidc ("" = invited, not yet bound)
	ExtID     string    `json:"extId,omitempty"`
	Role      string    `json:"role"` // admin | operator | viewer
	PassHash  string    `json:"passHash,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	LastLogin time.Time `json:"lastLogin"`
}

type Session struct {
	ID      string    `json:"id"`
	Token   string    `json:"token"` // sha256 of the cookie value
	UserID  string    `json:"userId"`
	Created time.Time `json:"created"`
	Seen    time.Time `json:"seen"`
	IP      string    `json:"ip"`
	Agent   string    `json:"agent"`
	Expires time.Time `json:"expires"`
}

type APIToken struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Hash    string    `json:"hash"`
	UserID  string    `json:"userId"`
	Created time.Time `json:"created"`
	Expires time.Time `json:"expires"` // zero = never
	Used    time.Time `json:"used"`
}

type Server struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Local     bool      `json:"local"`
	TokenHash string    `json:"tokenHash,omitempty"`
	LastSeen  time.Time `json:"lastSeen"`
	Hostname  string    `json:"hostname,omitempty"`
	OS        string    `json:"os,omitempty"`
	Docker    string    `json:"docker,omitempty"`
	Version   string    `json:"version,omitempty"`
	Created   time.Time `json:"created"`
}

type Credential struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Kind    string    `json:"kind"` // git-token | ssh-key | secret-text
	Enc     string    `json:"enc"`
	Hint    string    `json:"hint"` // non-secret fingerprint shown in the UI
	Created time.Time `json:"created"`
	Used    time.Time `json:"used"`
	UsedBy  string    `json:"usedBy"` // "project #N"
}

type Source struct {
	Kind         string `json:"kind"` // path | git
	Path         string `json:"path"` // dir containing compose file (path mode) or checkout dir override (git mode)
	Repo         string `json:"repo"`
	Branch       string `json:"branch"`
	CredentialID string `json:"credentialId"`
	ComposeFile  string `json:"composeFile"` // relative to dir; default docker-compose.yml / compose.yaml auto-detected
}

type Param struct {
	Name    string `json:"name"`
	Default string `json:"default"`
	Choices string `json:"choices"` // comma separated, optional
	Secret  bool   `json:"secret"`  // masked in logs and the UI
	Help    string `json:"help"`
}

type Project struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Desc       string    `json:"desc"`
	Folder     string    `json:"folder"`
	ServerID   string    `json:"serverId"`
	Source     Source    `json:"source"`
	EnvFile    string    `json:"envFile"`    // path to an .env on the target server
	EnvEnc     string    `json:"envEnc"`     // sealed inline env (KEY=VAL lines)
	Params     []Param   `json:"params"`     // build parameters, exported as env at deploy
	Profiles   string    `json:"profiles"`   // compose --profile list, comma separated
	PreCmd     string    `json:"preCmd"`     // shell, runs before deploy
	PostCmd    string    `json:"postCmd"`    // shell, runs after successful deploy
	HealthURL  string    `json:"healthUrl"`  // polled after deploy
	HealthSecs int       `json:"healthSecs"` // timeout
	Cron       string    `json:"cron"`       // schedule (5 field), empty = none
	PollSCM    bool      `json:"pollScm"`    // with Cron: only build if git HEAD changed
	Webhook    string    `json:"webhook"`    // secret token for /hooks/<id>
	NotifyURLs []string  `json:"notifyUrls"` // slack/discord/generic webhooks
	NotifyOn   string    `json:"notifyOn"`   // always | failure | change
	Keep       int       `json:"keep"`       // builds to retain (0 = 50)
	NoCache    bool      `json:"noCache"`
	Pull       bool      `json:"pull"`
	Disabled   bool      `json:"disabled"`
	LastCommit string    `json:"lastCommit"`
	Created    time.Time `json:"created"`
}

type Build struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"projectId"`
	Number    int               `json:"number"`
	Action    string            `json:"action"` // deploy | restart | stop | down | pull
	Status    string            `json:"status"` // queued | running | success | failed | cancelled
	Trigger   string            `json:"trigger"`
	By        string            `json:"by"`
	Commit    string            `json:"commit"`
	Params    map[string]string `json:"params,omitempty"`
	ServerID  string            `json:"serverId"`
	Queued    time.Time         `json:"queued"`
	Started   time.Time         `json:"started"`
	Finished  time.Time         `json:"finished"`
	Reason    string            `json:"reason,omitempty"` // failure reason
	Stages    []Stage           `json:"stages,omitempty"`
}

// Stage is one pipeline step of a build (checkout, pull, build, up, health, post).
type Stage struct {
	Name   string    `json:"name"`
	Status string    `json:"status"` // running | success | failed | skipped
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

type Audit struct {
	Time   time.Time `json:"time"`
	User   string    `json:"user"`
	Action string    `json:"action"`
	Target string    `json:"target"`
	IP     string    `json:"ip,omitempty"`
}

type Settings struct {
	AllowSignup  bool     `json:"allowSignup"`  // auto-create viewer accounts for new OAuth users
	AllowedEmail []string `json:"allowedEmail"` // allowed emails/domains for OAuth ("@example.com")
	DefaultRole  string   `json:"defaultRole"`
}

// Onboarding tracks the first-run guide shown to admins.
type Onboarding struct {
	Dismissed bool      `json:"dismissed"`
	At        time.Time `json:"at"`
}

type State struct {
	Users       []*User        `json:"users"`
	Sessions    []*Session     `json:"sessions"`
	Tokens      []*APIToken    `json:"tokens"`
	Servers     []*Server      `json:"servers"`
	Credentials []*Credential  `json:"credentials"`
	Projects    []*Project     `json:"projects"`
	Builds      []*Build       `json:"builds"`
	Audit       []*Audit       `json:"audit"`
	Settings    Settings       `json:"settings"`
	Onboarding  Onboarding     `json:"onboarding"`
	NextNum     map[string]int `json:"nextNum"`
}

type Store struct {
	mu   sync.RWMutex
	wmu  sync.Mutex // serialises writes of state.json
	dir  string
	st   State
	key  []byte
	save chan struct{}
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(dir, "logs"), 0o750); err != nil {
		return nil, err
	}
	s := &Store{dir: dir, save: make(chan struct{}, 1)}
	s.st.NextNum = map[string]int{}
	if b, err := os.ReadFile(filepath.Join(dir, "state.json")); err == nil {
		if err := json.Unmarshal(b, &s.st); err != nil {
			return nil, err
		}
		if s.st.NextNum == nil {
			s.st.NextNum = map[string]int{}
		}
		for _, se := range s.st.Sessions { // sessions from 0.1.x had no ID
			if se.ID == "" {
				se.ID = NewID()
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if s.st.Settings.DefaultRole == "" {
		s.st.Settings.DefaultRole = "viewer"
	}
	k, err := loadKey(dir)
	if err != nil {
		return nil, err
	}
	s.key = k
	go s.flusher()
	return s, nil
}

func loadKey(dir string) ([]byte, error) {
	if v := envx.Get("SECRET_KEY"); v != "" {
		h := sha256.Sum256([]byte(v))
		return h[:], nil
	}
	p := filepath.Join(dir, "secret.key")
	if b, err := os.ReadFile(p); err == nil && len(b) == 32 {
		return b, nil
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, os.WriteFile(p, b, 0o600)
}

func (s *Store) flusher() {
	for range s.save {
		time.Sleep(150 * time.Millisecond)
		s.Flush()
	}
}

// View runs fn with a read lock. fn must not retain pointers after return.
func (s *Store) View(fn func(*State)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn(&s.st)
}

// Update runs fn with the write lock and schedules a flush.
func (s *Store) Update(fn func(*State)) {
	s.Mutate(func(st *State) bool { fn(st); return true })
}

// Mutate runs fn with the write lock; state is only persisted when fn reports a
// change. Volatile fields (heartbeats) can be updated without disk writes.
func (s *Store) Mutate(fn func(*State) bool) bool {
	s.mu.Lock()
	changed := fn(&s.st)
	s.mu.Unlock()
	if changed {
		select {
		case s.save <- struct{}{}:
		default:
		}
	}
	return changed
}

// Flush writes state synchronously (also used on shutdown).
func (s *Store) Flush() {
	s.wmu.Lock()
	defer s.wmu.Unlock()
	s.mu.RLock()
	b, err := json.MarshalIndent(&s.st, "", " ")
	s.mu.RUnlock()
	if err != nil {
		return
	}
	tmp := filepath.Join(s.dir, "state.json.tmp")
	if os.WriteFile(tmp, b, 0o600) == nil {
		os.Rename(tmp, filepath.Join(s.dir, "state.json"))
	}
}

// LogPath returns the console log of a build. IDs are always generated by
// NewID (hex); anything else maps to an impossible name so callers can never
// escape the logs directory.
func (s *Store) LogPath(buildID string) string {
	if !ValidID(buildID) {
		buildID = "invalid"
	}
	return filepath.Join(s.dir, "logs", buildID+".log")
}

// ValidID reports whether v looks like an ID produced by NewID.
func ValidID(v string) bool {
	if len(v) != 16 {
		return false
	}
	for _, c := range v {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
func (s *Store) Dir() string { return s.dir }

// Seal / Open protect credentials at rest.
func (s *Store) Seal(plain string) string {
	if plain == "" {
		return ""
	}
	blk, _ := aes.NewCipher(s.key)
	g, _ := cipher.NewGCM(blk)
	n := make([]byte, g.NonceSize())
	rand.Read(n)
	return base64.StdEncoding.EncodeToString(g.Seal(n, n, []byte(plain), nil))
}

func (s *Store) Open(enc string) (string, error) {
	if enc == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	blk, _ := aes.NewCipher(s.key)
	g, _ := cipher.NewGCM(blk)
	if len(raw) < g.NonceSize() {
		return "", errors.New("short ciphertext")
	}
	p, err := g.Open(nil, raw[:g.NonceSize()], raw[g.NonceSize():], nil)
	return string(p), err
}

func NewID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func NewSecret(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Hash(v string) string {
	h := sha256.Sum256([]byte(v))
	return hex.EncodeToString(h[:])
}
