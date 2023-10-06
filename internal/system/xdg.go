package system

import (
	"os"
)

type XDG struct {
	// User specific data files, fallback: $HOME/.local/share
	DataHome string
	// User specific config files, fallback: $HOME/.config
	ConfigHome string
	// User specific state data files, fallback: $HOME/.local/state
	StateHome string
	// User specific cache files, fallback: $HOME/.cache
	CacheHome string
	// User specific runtme files (sockets, named pipes), fallback: /run/user/$UID
	RuntimeDir string
	// Colon separated list of strings identify the current desktop environment
	CurrentDesktop string
}

func envFallback(varname string, fallback string) string {
	value := os.Getenv(varname)
	if value == "" { value = fallback }
	return value
}

var _xdg XDG
var _xdg_set = false

func GetXDG() XDG {
	if !_xdg_set {
		_xdg_set = true
		uid := os.Getenv("UID")
		home := os.Getenv("HOME")
		_xdg = XDG{
			DataHome: envFallback("XDG_DATA_HOME", home + "/.local/share"),
			ConfigHome: envFallback("XDG_CONFIG_HOME", home + "/.config"),
			StateHome: envFallback("XDG_STATE_HOME", home + "/.local/state"),
			CacheHome: envFallback("XDG_CACHE_HOME", home + "/.cache"),
			RuntimeDir: envFallback("XDG_CACHE_HOME", "/run/user/" + uid),
			CurrentDesktop: os.Getenv("XDG_CURRENT_DESKTOP"),
		}
	}
	return _xdg
}

