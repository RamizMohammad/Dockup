package hub

import (
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/RamizMohammad/dockup/internal/runner"
	"github.com/RamizMohammad/dockup/internal/store"
)

// sampleCompose is the stack behind "Try a sample stack" in onboarding. It
// publishes no ports (so it cannot clash with anything on the host) and uses a
// tiny, well-known image that answers HTTP with request details.
const sampleCompose = `# Created by Dockup's first-run guide. Safe to delete (use "Down" first).
services:
  hello:
    image: traefik/whoami:latest
    restart: unless-stopped
    labels:
      dev.dockup.sample: "true"
`

const sampleName = "hello-dockup"

type onboardStep struct {
	ID       string `json:"id"`
	Done     bool   `json:"done"`
	Optional bool   `json:"optional"`
}

// GET /api/onboarding : progress of the first-run guide and a host check.
func (h *Hub) onboarding(w http.ResponseWriter, r *http.Request, u *store.User) {
	var st store.Onboarding
	var projects, servers, creds, users int
	var deployed, webhook, remoteOnline bool
	var sampleID, firstProject string
	h.db.View(func(s *store.State) {
		st = s.Onboarding
		projects, creds, users = len(s.Projects), len(s.Credentials), len(s.Users)
		servers = len(s.Servers)
		for _, p := range s.Projects {
			if firstProject == "" {
				firstProject = p.ID
			}
			if p.Name == sampleName {
				sampleID = p.ID
			}
		}
		for _, b := range s.Builds {
			if b.Action != "poll" && b.Status == "success" {
				deployed = true
			}
			if b.Trigger == "webhook" {
				webhook = true
			}
		}
	})
	h.mu.Lock()
	remoteOnline = len(h.agents) > 0
	h.mu.Unlock()
	_, _, docker := runner.Info()
	compose := ""
	if docker != "" {
		compose = runner.ComposeVersion()
	}
	steps := []onboardStep{
		{ID: "admin", Done: true},
		{ID: "server", Done: docker != "" || remoteOnline},
		{ID: "project", Done: projects > 0},
		{ID: "deploy", Done: deployed},
		{ID: "webhook", Done: webhook, Optional: true},
		{ID: "team", Done: users > 1, Optional: true},
		{ID: "agent", Done: servers > 1, Optional: true},
		{ID: "credential", Done: creds > 0, Optional: true},
		{ID: "sso", Done: len(h.cfg.OAuth) > 0, Optional: true},
	}
	writeJSON(w, 200, map[string]any{
		"dismissed": st.Dismissed, "steps": steps, "sampleId": sampleID, "firstProject": firstProject,
		"host": map[string]any{"docker": docker, "compose": compose, "https": h.secureCookie(r), "publicUrl": h.cfg.PublicURL,
			"version": h.cfg.Version, "sso": len(h.cfg.OAuth)},
	})
}

// PUT /api/onboarding {"dismissed": bool}
func (h *Hub) setOnboarding(w http.ResponseWriter, r *http.Request, u *store.User) {
	var b struct {
		Dismissed bool `json:"dismissed"`
	}
	if readJSON(r, &b) != nil {
		fail(w, 400, "bad request")
		return
	}
	h.db.Update(func(s *store.State) { s.Onboarding = store.Onboarding{Dismissed: b.Dismissed, At: time.Now()} })
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /api/onboarding/sample : create (or reuse) the sample project on the hub.
func (h *Hub) createSample(w http.ResponseWriter, r *http.Request, u *store.User) {
	dir := filepath.Join(h.cfg.DataDir, "samples", sampleName)
	if !filepath.IsAbs(dir) {
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		fail(w, 500, "cannot create sample directory: "+err.Error())
		return
	}
	if err := os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte(sampleCompose), 0o640); err != nil {
		fail(w, 500, "cannot write sample compose file: "+err.Error())
		return
	}
	var out store.Project
	h.db.Update(func(s *store.State) {
		var local *store.Server
		for _, sv := range s.Servers {
			if sv.Local {
				local = sv
			}
		}
		for _, p := range s.Projects {
			if p.Name == sampleName {
				out = *p
				return
			}
		}
		p := &store.Project{ID: store.NewID(), Name: sampleName, Folder: "Samples", Created: time.Now(), Webhook: store.NewSecret(16),
			Desc:     "A tiny demo stack (traefik/whoami) created by the first-run guide. It publishes no ports. Safe to remove.",
			ServerID: local.ID, Source: store.Source{Kind: "path", Path: dir}, NotifyOn: "always", Pull: true}
		s.Projects = append(s.Projects, p)
		out = *p
	})
	h.audit(r, u.Login, "project.save", out.Name+" (sample)")
	writeJSON(w, 200, h.view(&out, u, r))
}
