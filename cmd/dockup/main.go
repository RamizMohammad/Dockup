package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/RamizMohammad/dockup/internal/agent"
	"github.com/RamizMohammad/dockup/internal/envx"
	"github.com/RamizMohammad/dockup/internal/hub"
)

var version = "dev"

//go:embed all:web
var webFS embed.FS

func usage() {
	fmt.Fprintf(os.Stderr, `dockup %s - minimal CI/CD for Docker Compose

usage:
  dockup serve     run the hub (web UI + API)
  dockup agent     run a remote agent that connects to a hub
  dockup version

hub environment:
  DOCKUP_ADDR=:8080           listen address
  DOCKUP_DATA=/var/lib/dockup
  DOCKUP_URL=https://deploy.example.com   public URL (OAuth callbacks)
  DOCKUP_TRUST_PROXY=1        trust X-Forwarded-* from a reverse proxy
  DOCKUP_ADMIN_USER / DOCKUP_ADMIN_PASSWORD   bootstrap first admin
  DOCKUP_ADMIN_EMAIL          invite this email as admin (binds on first OAuth sign-in)
  (otherwise the first admin is created in the browser with the one-time
   setup code printed to the log and written to <data>/setup-code)
  DOCKUP_SECRET_KEY           key for encrypting stored secrets
  DOCKUP_GITHUB_CLIENT_ID / _SECRET
  DOCKUP_GOOGLE_CLIENT_ID / _SECRET
  DOCKUP_OIDC_ISSUER / _CLIENT_ID / _CLIENT_SECRET / _NAME

(DEPLOYHUB_* names from before the rename are still accepted.)

agent environment:
  DOCKUP_HUB=https://deploy.example.com
  DOCKUP_TOKEN=dks_...        from Servers page
  DOCKUP_DATA=/var/lib/dockup-agent
`, version)
}

func main() {
	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch cmd {
	case "version", "--version", "-v":
		fmt.Println(version)
	case "serve", "hub":
		sub, err := fs.Sub(webFS, "web")
		if err != nil {
			log.Fatal(err)
		}
		h, err := hub.New(hub.Config{
			Addr:       envx.Or("ADDR", ":8080"),
			DataDir:    envx.Or("DATA", "./data"),
			PublicURL:  envx.Get("URL"),
			Version:    version,
			AdminUser:  envx.Get("ADMIN_USER"),
			AdminPass:  envx.Get("ADMIN_PASSWORD"),
			AdminEmail: envx.Get("ADMIN_EMAIL"),
			OAuth:      hub.OAuthFromEnv(),
			TrustProxy: envx.Get("TRUST_PROXY") == "1",
			WebFS:      sub,
		})
		if err != nil {
			log.Fatal(err)
		}
		if err := h.Run(ctx); err != nil {
			log.Fatal(err)
		}
	case "agent":
		hubURL, tok := envx.Get("HUB"), envx.Get("TOKEN")
		if hubURL == "" || tok == "" {
			log.Fatal("DOCKUP_HUB and DOCKUP_TOKEN are required")
		}
		if !strings.HasPrefix(hubURL, "http") {
			log.Fatal("DOCKUP_HUB must start with http:// or https://")
		}
		a := &agent.Agent{Hub: hubURL, Token: tok, DataDir: envx.Or("DATA", "./agent-data"), Version: version}
		a.Run(ctx)
	default:
		usage()
		os.Exit(2)
	}
}
