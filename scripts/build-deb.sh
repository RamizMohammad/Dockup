#!/usr/bin/env bash
# Build .deb packages.  usage: scripts/build-deb.sh [version] [arch ...]   (arch: amd64 arm64)
set -euo pipefail
cd "$(dirname "$0")/.."
VERSION="${1:-0.3.0}"; shift || true
ARCHES="${*:-amd64 arm64}"
OUT=dist; mkdir -p "$OUT"
for ARCH in $ARCHES; do
  case "$ARCH" in amd64|arm64) ;; *) echo "unsupported arch $ARCH" >&2; exit 1;; esac
  ROOT="$(mktemp -d)"; chmod 0755 "$ROOT"; trap 'rm -rf "$ROOT"' EXIT
  install -d "$ROOT/DEBIAN" "$ROOT/usr/bin" "$ROOT/lib/systemd/system" "$ROOT/etc/dockup"
  CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build -trimpath -ldflags "-s -w -X main.version=$VERSION" -o "$ROOT/usr/bin/dockup" ./cmd/dockup
  install -m 0644 packaging/debian/dockup.service packaging/debian/dockup-agent.service "$ROOT/lib/systemd/system/"
  install -m 0600 packaging/debian/dockup.env "$ROOT/etc/dockup/dockup.env"
  install -m 0600 packaging/debian/agent.env "$ROOT/etc/dockup/agent.env"
  install -m 0755 packaging/debian/postinst packaging/debian/prerm packaging/debian/postrm "$ROOT/DEBIAN/"
  printf '/etc/dockup/dockup.env\n/etc/dockup/agent.env\n' > "$ROOT/DEBIAN/conffiles"
  SIZE=$(du -sk "$ROOT/usr" | cut -f1)
  cat > "$ROOT/DEBIAN/control" <<CTL
Package: dockup
Version: $VERSION
Section: admin
Priority: optional
Architecture: $ARCH
Installed-Size: $SIZE
Maintainer: Dockup <noreply@example.com>
Conflicts: deployhub
Replaces: deployhub
Provides: deployhub
Recommends: docker.io | docker-ce, docker-compose-v2 | docker-compose-plugin, git, openssh-client
Description: Minimal CI/CD for Docker Compose with a hub and agents
 Dockup deploys Docker Compose stacks from a local path or a git
 repository, with live logs, webhooks, cron/SCM polling, OAuth login and
 remote agents. One static binary runs as the hub (web UI) or as an agent.
CTL
  dpkg-deb --root-owner-group -Zxz --build "$ROOT" "$OUT/dockup_${VERSION}_${ARCH}.deb" >/dev/null
  echo "built $OUT/dockup_${VERSION}_${ARCH}.deb"
  rm -rf "$ROOT"; trap - EXIT
done
