package hub

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/RamizMohammad/dockup/internal/runner"
	"github.com/RamizMohammad/dockup/internal/store"
)

type agentConn struct {
	serverID string
	ch       chan []byte
	quit     chan struct{}
}

func (a *agentConn) send(event string, v any) bool {
	b, _ := json.Marshal(v)
	msg := []byte("event: " + event + "\ndata: " + string(b) + "\n\n")
	select {
	case a.ch <- msg:
		return true
	case <-time.After(3 * time.Second):
		return false
	}
}

func (h *Hub) agentAuth(r *http.Request) *store.Server {
	tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tok == "" {
		return nil
	}
	hash := store.Hash(tok)
	var out *store.Server
	h.db.View(func(s *store.State) {
		for _, sv := range s.Servers {
			if !sv.Local && sv.TokenHash != "" && subtle.ConstantTimeCompare([]byte(sv.TokenHash), []byte(hash)) == 1 {
				c := *sv
				out = &c
			}
		}
	})
	return out
}

// GET /api/agent/stream : server-sent events from hub to agent (jobs, cancels).
func (h *Hub) agentStream(w http.ResponseWriter, r *http.Request) {
	sv := h.agentAuth(r)
	if sv == nil {
		fail(w, 401, "bad agent token")
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
	ac := &agentConn{serverID: sv.ID, ch: make(chan []byte, 32), quit: make(chan struct{})}
	h.mu.Lock()
	if old := h.agents[sv.ID]; old != nil {
		close(old.quit)
	}
	h.agents[sv.ID] = ac
	h.mu.Unlock()
	h.touch(sv.ID, nil)
	fmt.Fprint(w, ": connected\n\n")
	fl.Flush()
	h.pump()

	tick := time.NewTicker(15 * time.Second)
	defer tick.Stop()
	defer func() {
		h.mu.Lock()
		if h.agents[sv.ID] == ac {
			delete(h.agents, sv.ID)
		}
		h.mu.Unlock()
		// builds keep running on the agent while it reconnects; give it time
		time.AfterFunc(agentGrace, func() { h.reapLost(sv.ID) })
	}()
	for {
		select {
		case m := <-ac.ch:
			if _, err := w.Write(m); err != nil {
				return
			}
			fl.Flush()
		case <-tick.C:
			fmt.Fprint(w, ": ka\n\n")
			fl.Flush()
			h.touch(sv.ID, nil)
		case <-ac.quit:
			return
		case <-r.Context().Done():
			return
		}
	}
}

// agentGrace is how long a disconnected agent's running builds are kept alive.
var agentGrace = 3 * time.Minute

type agentInfo struct {
	Hostname string    `json:"hostname"`
	OS       string    `json:"os"`
	Docker   string    `json:"docker"`
	Version  string    `json:"version"`
	Running  *[]string `json:"running"` // build IDs executing on the agent (nil = old agent)
}

// touch refreshes LastSeen in memory; the state file is only rewritten when
// the agent's reported host details actually change.
func (h *Hub) touch(id string, in *agentInfo) {
	h.db.Mutate(func(s *store.State) bool {
		sv := h.serverByID(s, id)
		if sv == nil {
			return false
		}
		sv.LastSeen = time.Now()
		if in == nil {
			return false
		}
		changed := sv.Hostname != in.Hostname || sv.OS != in.OS || sv.Docker != in.Docker || sv.Version != in.Version
		sv.Hostname, sv.OS, sv.Docker, sv.Version = in.Hostname, in.OS, in.Docker, in.Version
		return changed
	})
	if in != nil && in.Running != nil {
		h.reconcile(id, *in.Running)
	}
}

// reconcile fails builds the hub thinks are running on an agent which the
// agent no longer knows about (e.g. the agent restarted mid-build).
func (h *Hub) reconcile(serverID string, running []string) {
	set := map[string]bool{}
	for _, id := range running {
		set[id] = true
	}
	h.mu.Lock()
	h.agentRun[serverID] = set
	h.mu.Unlock()
	var lost []string
	h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			if b.ServerID == serverID && b.Status == "running" && !set[b.ID] && time.Since(b.Started) > time.Minute {
				lost = append(lost, b.ID)
			}
		}
	})
	for _, id := range lost {
		h.logs.Append(id, "==> the agent is no longer running this build (agent restarted?)")
		h.finish(id, "", "agent lost the build")
	}
}

func (h *Hub) agentPing(w http.ResponseWriter, r *http.Request) {
	sv := h.agentAuth(r)
	if sv == nil {
		fail(w, 401, "bad agent token")
		return
	}
	var in agentInfo
	readJSON(r, &in)
	h.touch(sv.ID, &in)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) ownsBuild(sv *store.Server, id string) bool {
	ok := false
	h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			if b.ID == id && b.ServerID == sv.ID && b.Status == "running" {
				ok = true
			}
		}
	})
	return ok
}

func (h *Hub) agentLog(w http.ResponseWriter, r *http.Request) {
	sv := h.agentAuth(r)
	if sv == nil {
		fail(w, 401, "bad agent token")
		return
	}
	id := r.URL.Query().Get("build")
	if !h.ownsBuild(sv, id) {
		fail(w, 404, "no such running build")
		return
	}
	b, _ := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	txt := strings.TrimRight(string(b), "\n")
	if txt != "" {
		h.logs.Append(id, strings.Split(txt, "\n")...)
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (h *Hub) agentDone(w http.ResponseWriter, r *http.Request) {
	sv := h.agentAuth(r)
	if sv == nil {
		fail(w, 401, "bad agent token")
		return
	}
	var res struct {
		BuildID string `json:"buildId"`
		runner.Result
	}
	if readJSON(r, &res) != nil || !h.ownsBuild(sv, res.BuildID) {
		fail(w, 404, "no such running build")
		return
	}
	h.finish(res.BuildID, res.Commit, res.Err)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// reapLost fails builds of an agent that never came back.
func (h *Hub) reapLost(serverID string) {
	h.mu.Lock()
	_, up := h.agents[serverID]
	h.mu.Unlock()
	if up {
		return
	}
	var ids []string
	h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			if b.ServerID == serverID && b.Status == "running" {
				ids = append(ids, b.ID)
			}
		}
	})
	for _, id := range ids {
		h.logs.Append(id, "==> agent disconnected and did not come back within "+agentGrace.String())
		h.finish(id, "", "agent disconnected")
	}
}

func (h *Hub) isOnline(sv *store.Server) bool {
	if sv.Local {
		return true
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.agents[sv.ID] != nil
}
