// Package envx reads Dockup configuration from the environment.
//
// Every setting is read as DOCKUP_<NAME>. The pre-rename DEPLOYHUB_<NAME>
// spelling is still accepted as a fallback so existing installs, compose files
// and systemd environment files keep working after the upgrade.
package envx

import (
	"os"
	"strings"
)

// Prefixes lists every prefix Dockup reads its own configuration from.
// Builds never inherit variables with these prefixes.
var Prefixes = []string{"DOCKUP_", "DEPLOYHUB_"}

// Get returns DOCKUP_<name>, falling back to DEPLOYHUB_<name>.
func Get(name string) string {
	for _, p := range Prefixes {
		if v := os.Getenv(p + name); v != "" {
			return v
		}
	}
	return ""
}

// Or returns Get(name), or def when unset.
func Or(name, def string) string {
	if v := Get(name); v != "" {
		return v
	}
	return def
}

// Internal reports whether an environment entry ("KEY=VALUE" or "KEY") is
// Dockup's own configuration.
func Internal(kv string) bool {
	k := strings.ToUpper(kv)
	for _, p := range Prefixes {
		if strings.HasPrefix(k, p) {
			return true
		}
	}
	return false
}
