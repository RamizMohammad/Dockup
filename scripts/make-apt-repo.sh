#!/usr/bin/env bash
# Turn dist/*.deb into a static, flat apt repository you can host anywhere
# (nginx, S3, GitHub Pages...).   usage: scripts/make-apt-repo.sh [debs_dir] [out_dir]
# Optional signing:  GPG_KEY=<key id> scripts/make-apt-repo.sh
set -euo pipefail
cd "$(dirname "$0")/.."
SRC="${1:-dist}"; OUT="${2:-dist/apt}"
rm -rf "$OUT"; mkdir -p "$OUT"
cp "$SRC"/*.deb "$OUT"/
( cd "$OUT"
  dpkg-scanpackages --multiversion . /dev/null > Packages 2>/dev/null
  gzip -9 -k -f Packages
  ARCHS=$(grep '^Architecture:' Packages | awk '{print $2}' | sort -u | tr '\n' ' ')
  hashes() { for f in Packages Packages.gz; do printf ' %s %16d %s\n' "$($1 "$f" | cut -d' ' -f1)" "$(stat -c %s "$f")" "$f"; done; }
  {
    echo "Origin: Dockup"; echo "Label: Dockup"; echo "Suite: stable"; echo "Codename: stable"
    echo "Date: $(LC_ALL=C date -Ru)"; echo "Architectures: $ARCHS"
    echo "MD5Sum:";    hashes md5sum
    echo "SHA256:";    hashes sha256sum
  } > Release
  if [ -n "${GPG_KEY:-}" ]; then
    gpg --batch --yes --default-key "$GPG_KEY" -abs -o Release.gpg Release
    gpg --batch --yes --default-key "$GPG_KEY" --clearsign -o InRelease Release
    gpg --export "$GPG_KEY" | gpg --dearmor > dockup.gpg 2>/dev/null || gpg --export "$GPG_KEY" > dockup.gpg
    echo "signed with $GPG_KEY; publish dockup.gpg as the repo key"
  fi )
echo "apt repo ready in $OUT"
echo "client:  echo 'deb [trusted=yes] https://YOUR-HOST/apt ./' | sudo tee /etc/apt/sources.list.d/dockup.list   # or use signed-by=/usr/share/keyrings/dockup.gpg"
echo "then:    sudo apt update && sudo apt install dockup"
