package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/RamizMohammad/dockup/internal/cron"
	"github.com/RamizMohammad/dockup/internal/runner"
	"github.com/RamizMohammad/dockup/internal/store"
)

// ---------------- log hub (file backed, live fan-out) ----------------

type logHub struct {
	db      *store.Store
	mu      sync.Mutex
	files   map[string]*os.File
	subs    map[string]map[chan struct{}]struct{}
	onStage func(buildID, name, status string, at time.Time)
}

// TimeFormat prefixes every console line on disk ("<ts>\t<text>").
const TimeFormat = "2006-01-02T15:04:05.000Z07:00"

func newLogHub(db *store.Store) *logHub {
	return &logHub{db: db, files: map[string]*os.File{}, subs: map[string]map[chan struct{}]struct{}{}}
}

func (l *logHub) Append(id string, lines ...string) {
	if len(lines) == 0 {
		return
	}
	l.mu.Lock()
	f := l.files[id]
	if f == nil {
		var err error
		f, err = os.OpenFile(l.db.LogPath(id), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			l.mu.Unlock()
			return
		}
		l.files[id] = f
	}
	var b bytes.Buffer
	now := time.Now()
	ts := now.UTC().Format(TimeFormat)
	var stages [][2]string
	for _, s := range lines {
		if strings.HasPrefix(s, runner.StagePrefix) {
			if f := strings.Fields(strings.TrimPrefix(s, runner.StagePrefix)); len(f) == 2 {
				stages = append(stages, [2]string{f[0], f[1]})
			}
		}
		b.WriteString(ts)
		b.WriteByte('\t')
		b.WriteString(strings.ReplaceAll(s, "\r", ""))
		b.WriteByte('\n')
	}
	f.Write(b.Bytes())
	for ch := range l.subs[id] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	l.mu.Unlock()
	if l.onStage != nil {
		for _, st := range stages {
			l.onStage(id, st[0], st[1], now)
		}
	}
}

func (l *logHub) Close(id string) {
	l.mu.Lock()
	if f := l.files[id]; f != nil {
		f.Close()
		delete(l.files, id)
	}
	for ch := range l.subs[id] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	l.mu.Unlock()
}

func (l *logHub) Sub(id string) (chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	l.mu.Lock()
	if l.subs[id] == nil {
		l.subs[id] = map[chan struct{}]struct{}{}
	}
	l.subs[id][ch] = struct{}{}
	l.mu.Unlock()
	return ch, func() {
		l.mu.Lock()
		delete(l.subs[id], ch)
		l.mu.Unlock()
	}
}

// stageEvent records a pipeline stage transition reported in the log stream.
func (h *Hub) stageEvent(id, name, status string, at time.Time) {
	h.db.Update(func(s *store.State) {
		for _, b := range s.Builds {
			if b.ID != id {
				continue
			}
			for i := range b.Stages {
				if b.Stages[i].Name == name && b.Stages[i].End.IsZero() {
					b.Stages[i].Status = status
					if status != "running" {
						b.Stages[i].End = at
					}
					return
				}
			}
			if status == "running" {
				b.Stages = append(b.Stages, store.Stage{Name: name, Status: status, Start: at})
			}
		}
	})
}

// ---------------- queue ----------------

func (h *Hub) serverByID(s *store.State, id string) *store.Server {
	for _, sv := range s.Servers {
		if sv.ID == id {
			return sv
		}
	}
	return nil
}

func (h *Hub) projectByID(s *store.State, id string) *store.Project {
	for _, p := range s.Projects {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (h *Hub) enqueue(pid, action, trigger, by string, params map[string]string, commit string) (*store.Build, error) {
	var out *store.Build
	var err error
	h.db.Update(func(s *store.State) {
		p := h.projectByID(s, pid)
		if p == nil {
			err = errors.New("project not found")
			return
		}
		if p.Disabled && trigger != "manual" {
			err = errors.New("project disabled")
			return
		}
		switch action {
		case "deploy", "restart", "stop", "down", "pull", "rollback", "poll":
		default:
			err = errors.New("unknown action")
			return
		}
		// validate params against declaration
		clean := map[string]string{}
		for _, d := range p.Params {
			v, ok := params[d.Name]
			if !ok {
				v = d.Default
			}
			if d.Choices != "" {
				okc := false
				for _, c := range strings.Split(d.Choices, ",") {
					if strings.TrimSpace(c) == v {
						okc = true
					}
				}
				if !okc {
					err = fmt.Errorf("parameter %s must be one of %s", d.Name, d.Choices)
					return
				}
			}
			clean[d.Name] = v
		}
		// automatic triggers coalesce: one pending deploy is enough
		if trigger != "manual" && action == "deploy" {
			for _, x := range s.Builds {
				if x.ProjectID == pid && x.Status == "queued" && x.Action == "deploy" && x.Trigger != "manual" && x.Commit == "" {
					c := *x
					out = &c
					return
				}
			}
		}
		s.NextNum[pid]++
		b := &store.Build{ID: store.NewID(), ProjectID: pid, Number: s.NextNum[pid], Action: action, Status: "queued",
			Trigger: trigger, By: by, Params: clean, ServerID: p.ServerID, Queued: time.Now(), Commit: commit}
		s.Builds = append(s.Builds, b)
		c := *b
		out = &c
	})
	if err == nil {
		h.pump()
	}
	return out, err
}

func (h *Hub) pumpLoop(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			h.pump()
		}
	}
}

type startItem struct {
	b  store.Build
	p  store.Project
	sv store.Server
}

func (h *Hub) pump() {
	h.mu.Lock()
	online := map[string]bool{}
	for id := range h.agents {
		online[id] = true
	}
	h.mu.Unlock()

	var start []startItem
	var timedOut []string
	h.db.Mutate(func(s *store.State) bool {
		busy := map[string]bool{}
		for _, b := range s.Builds {
			if b.Status == "running" {
				busy[b.ProjectID] = true
			}
		}
		for _, b := range s.Builds {
			if b.Status != "queued" || busy[b.ProjectID] {
				continue
			}
			p, sv := h.projectByID(s, b.ProjectID), h.serverByID(s, b.ServerID)
			if p == nil || sv == nil {
				b.Status, b.Finished, b.Reason = "failed", time.Now(), "project or server removed"
				timedOut = append(timedOut, b.ID)
				continue
			}
			if !sv.Local && !online[sv.ID] {
				if time.Since(b.Queued) > 5*time.Minute {
					b.Status, b.Finished, b.Reason = "failed", time.Now(), "server "+sv.Name+" offline for 5 minutes"
					timedOut = append(timedOut, b.ID)
				}
				continue
			}
			b.Status, b.Started = "running", time.Now()
			busy[b.ProjectID] = true
			start = append(start, startItem{*b, *p, *sv})
		}
		return len(start) > 0 || len(timedOut) > 0
	})
	for _, id := range timedOut {
		h.logs.Append(id, "==> failed: server offline, or project/server removed")
		h.logs.Close(id)
	}
	for _, it := range start {
		h.startBuild(it)
	}
}

func (h *Hub) jobFor(b store.Build, p store.Project) (runner.Job, error) {
	j := runner.Job{BuildID: b.ID, Action: b.Action, ProjectID: p.ID, ComposeProject: runner.ComposeName(p.Name),
		SourceKind: p.Source.Kind, Path: p.Source.Path, Repo: p.Source.Repo, Branch: p.Source.Branch, ComposeFile: p.Source.ComposeFile,
		EnvFile: p.EnvFile, Params: b.Params, Profiles: strings.Split(p.Profiles, ","), PreCmd: p.PreCmd, PostCmd: p.PostCmd,
		HealthURL: p.HealthURL, HealthSecs: p.HealthSecs, NoCache: p.NoCache, Pull: p.Pull}
	if j.Branch == "" {
		j.Branch = "main"
	}
	if b.Action == "poll" {
		j.CheckOnly = true
	}
	for _, d := range p.Params {
		if d.Secret {
			j.SecretParams = append(j.SecretParams, d.Name)
		}
	}
	if b.Action == "rollback" || (b.Action == "deploy" && b.Commit != "") {
		j.Commit = b.Commit
		j.Action = "deploy"
	}
	if p.EnvEnc != "" {
		v, err := h.db.Open(p.EnvEnc)
		if err != nil {
			return j, errors.New("cannot decrypt project env (secret key changed?)")
		}
		j.EnvInline = v
	}
	if p.Source.CredentialID != "" {
		var enc, kind string
		h.db.Update(func(s *store.State) {
			for _, c := range s.Credentials {
				if c.ID == p.Source.CredentialID {
					enc, kind = c.Enc, c.Kind
					c.Used, c.UsedBy = time.Now(), fmt.Sprintf("%s #%d", p.Name, b.Number)
				}
			}
		})
		if enc == "" {
			return j, errors.New("the project's git credential no longer exists")
		}
		v, err := h.db.Open(enc)
		if err != nil {
			return j, errors.New("cannot decrypt credential")
		}
		if kind == "ssh-key" {
			j.SSHKey = v
		} else {
			j.GitToken = v
		}
	}
	return j, nil
}

func (h *Hub) startBuild(it startItem) {
	h.logs.Append(it.b.ID, fmt.Sprintf("Started by %s (%s) on %s", orDash(it.b.By), it.b.Trigger, it.sv.Name))
	job, err := h.jobFor(it.b, it.p)
	if err != nil {
		h.logs.Append(it.b.ID, "==> "+err.Error())
		h.finish(it.b.ID, "", err.Error())
		return
	}
	if it.sv.Local {
		ctx, cancel := context.WithCancel(context.Background())
		h.mu.Lock()
		h.running[it.b.ID] = cancel
		h.mu.Unlock()
		go func() {
			res := h.local.Run(ctx, job, func(l string) { h.logs.Append(it.b.ID, l) })
			h.mu.Lock()
			delete(h.running, it.b.ID)
			h.mu.Unlock()
			cancel()
			h.finish(it.b.ID, res.Commit, res.Err)
		}()
		return
	}
	h.mu.Lock()
	ag := h.agents[it.sv.ID]
	h.mu.Unlock()
	if ag == nil || !ag.send("job", job) {
		h.logs.Append(it.b.ID, "==> agent not reachable")
		h.finish(it.b.ID, "", "agent not reachable")
	}
}

func orDash(s string) string {
	if s == "" {
		return "system"
	}
	return s
}

func (h *Hub) cancelBuild(id string) error {
	var status, server string
	h.db.View(func(s *store.State) {
		for _, b := range s.Builds {
			if b.ID == id {
				status, server = b.Status, b.ServerID
			}
		}
	})
	switch status {
	case "queued":
		h.db.Update(func(s *store.State) {
			for _, b := range s.Builds {
				if b.ID == id && b.Status == "queued" {
					b.Status, b.Finished = "cancelled", time.Now()
				}
			}
		})
		h.logs.Append(id, "==> cancelled while queued")
		h.logs.Close(id)
		return nil
	case "running":
		h.mu.Lock()
		h.cancelReq[id] = true
		c := h.running[id]
		ag := h.agents[server]
		h.mu.Unlock()
		if c != nil {
			c()
		} else if ag != nil {
			ag.send("cancel", id)
		} else {
			return errors.New("cannot reach the build runner")
		}
		return nil
	}
	return errors.New("build is not active")
}

// finish records the outcome, fires follow-ups and notifications.
func (h *Hub) finish(id, commit, errStr string) {
	h.mu.Lock()
	cancelled := h.cancelReq[id]
	delete(h.cancelReq, id)
	h.mu.Unlock()

	var b store.Build
	var p store.Project
	var prev string
	var changed bool
	h.db.Update(func(s *store.State) {
		var bp *store.Build
		for _, x := range s.Builds {
			if x.ID == id {
				bp = x
			}
		}
		if bp == nil || bp.Status == "success" || bp.Status == "failed" || bp.Status == "cancelled" {
			return
		}
		bp.Finished = time.Now()
		switch {
		case cancelled:
			bp.Status = "cancelled"
		case errStr != "":
			bp.Status, bp.Reason = "failed", errStr
		default:
			bp.Status = "success"
		}
		for i := range bp.Stages { // close stages a lost runner never reported
			if bp.Stages[i].End.IsZero() {
				bp.Stages[i].End = bp.Finished
				if bp.Status == "success" {
					bp.Stages[i].Status = "success"
				} else {
					bp.Stages[i].Status = bp.Status
				}
			}
		}
		if commit != "" {
			bp.Commit = commit
		}
		pp := h.projectByID(s, bp.ProjectID)
		if pp == nil {
			return
		}
		if bp.Status == "success" && commit != "" {
			if bp.Action == "poll" {
				changed = pp.LastCommit != commit
			} else {
				pp.LastCommit = commit
			}
		}
		for i := len(s.Builds) - 1; i >= 0; i-- {
			x := s.Builds[i]
			if x != bp && x.ProjectID == bp.ProjectID && x.Action != "poll" && (x.Status == "success" || x.Status == "failed") && x.Finished.Before(bp.Finished) {
				prev = x.Status
				break
			}
		}
		b, p = *bp, *pp
	})
	if b.ID == "" {
		return
	}
	switch {
	case cancelled:
		h.logs.Append(id, "==> cancelled")
	case errStr != "":
		h.logs.Append(id, "==> FAILED: "+errStr)
	case b.Action != "poll":
		h.logs.Append(id, "==> SUCCESS")
	}
	h.logs.Close(id)

	if b.Action == "poll" {
		if b.Status == "success" {
			h.dropBuild(id)
			if changed {
				h.enqueue(p.ID, "deploy", "scm", "", b.Params, "")
			}
		} else {
			h.pruneFailedPolls(p.ID, id)
		}
		h.pump()
		return
	}
	h.retain(p)
	h.notify(p, b, prev)
	h.pump()
}

func (h *Hub) dropBuild(id string) {
	h.db.Update(func(s *store.State) {
		out := s.Builds[:0]
		for _, x := range s.Builds {
			if x.ID != id {
				out = append(out, x)
			}
		}
		s.Builds = out
	})
	os.Remove(h.db.LogPath(id))
}

// pruneFailedPolls keeps only the newest failed SCM poll of a project, so a
// broken repository polled every minute cannot flood the state or the history.
func (h *Hub) pruneFailedPolls(pid, keep string) {
	var drop []string
	h.db.Update(func(s *store.State) {
		out := s.Builds[:0]
		for _, x := range s.Builds {
			if x.ProjectID == pid && x.Action == "poll" && x.ID != keep && (x.Status == "failed" || x.Status == "cancelled") {
				drop = append(drop, x.ID)
				continue
			}
			out = append(out, x)
		}
		s.Builds = out
	})
	for _, id := range drop {
		os.Remove(h.db.LogPath(id))
	}
}

func (h *Hub) retain(p store.Project) {
	keep := p.Keep
	if keep <= 0 {
		keep = 50
	}
	var drop []string
	h.db.Update(func(s *store.State) {
		var mine []*store.Build
		for _, b := range s.Builds {
			if b.ProjectID == p.ID && b.Action != "poll" && (b.Status == "success" || b.Status == "failed" || b.Status == "cancelled") {
				mine = append(mine, b)
			}
		}
		if len(mine) <= keep {
			return
		}
		sort.Slice(mine, func(i, j int) bool { return mine[i].Number > mine[j].Number })
		kill := map[string]bool{}
		for _, b := range mine[keep:] {
			kill[b.ID] = true
			drop = append(drop, b.ID)
		}
		out := s.Builds[:0]
		for _, b := range s.Builds {
			if !kill[b.ID] {
				out = append(out, b)
			}
		}
		s.Builds = out
	})
	for _, id := range drop {
		os.Remove(h.db.LogPath(id))
	}
}

func (h *Hub) notify(p store.Project, b store.Build, prev string) {
	if len(p.NotifyURLs) == 0 || b.Status == "cancelled" {
		return
	}
	switch p.NotifyOn {
	case "failure":
		if b.Status != "failed" {
			return
		}
	case "change":
		if prev == b.Status {
			return
		}
	}
	icon := map[string]string{"success": "✅", "failed": "❌"}[b.Status]
	link := h.cfg.PublicURL
	if link != "" {
		link = strings.TrimRight(link, "/") + "/#/builds/" + b.ID
	}
	msg := fmt.Sprintf("%s %s #%d %s (%s)", icon, p.Name, b.Number, b.Status, b.Action)
	if b.Reason != "" && b.Status == "failed" {
		msg += ": " + b.Reason
	}
	if link != "" {
		msg += " " + link
	}
	body, _ := json.Marshal(map[string]any{"text": msg, "content": msg, "project": p.Name, "build": b.Number,
		"status": b.Status, "action": b.Action, "commit": b.Commit, "url": link, "reason": b.Reason})
	for _, u := range p.NotifyURLs {
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			continue
		}
		go func(u string) {
			c := &http.Client{Timeout: 10 * time.Second}
			resp, err := c.Post(u, "application/json", bytes.NewReader(body))
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}(u)
	}
}

// ---------------- scheduler ----------------

func (h *Hub) cronLoop(ctx context.Context) {
	cache := map[string]*cron.Schedule{}
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			key := now.Format("200601021504")
			if key == h.cronSeen {
				continue
			}
			h.cronSeen = key
			var due []store.Project
			h.db.View(func(s *store.State) {
				for _, p := range s.Projects {
					if p.Cron == "" || p.Disabled {
						continue
					}
					sc := cache[p.Cron]
					if sc == nil {
						var err error
						if sc, err = cron.Parse(p.Cron); err != nil {
							continue
						}
						cache[p.Cron] = sc
					}
					if sc.Match(now) {
						due = append(due, *p)
					}
				}
			})
			for _, p := range due {
				if p.PollSCM && p.Source.Kind == "git" {
					h.enqueue(p.ID, "poll", "scm", "", nil, "")
				} else {
					h.enqueue(p.ID, "deploy", "schedule", "", nil, "")
				}
			}
		}
	}
}
