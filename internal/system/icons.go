package system

import (
	"fmt"
	"os/exec"
	"path"
	"os"
	"strings"
)

func getIconFolder() string {
	return "/usr/share/icons"
}

// GetIconTheme does its best to find
// the current icon theme name.
func getIconThemeName() string {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "icon-theme")
	out, err := cmd.Output()
	if err != nil { return "" }
	return strings.Trim(string(out), "'\"\n")
}

func GetIconPath(icon string) string {
	if !icons_indexed {
		root_path := path.Join(getIconFolder(), getIconThemeName())
		indexIcons(root_path)
		icons_indexed = true
	}
	return icon_index[icon]
}

var icon_index = make(map[string]string)
var icons_indexed = false

func indexIcons(dir_path string) {
	entries, err := os.ReadDir(dir_path)
	if err != nil {
		fmt.Println("Failed to index", dir_path, ": ", err)
	}

	for _, e := range entries {
		name := e.Name()
		entry_fullpath := path.Join(dir_path, name)
		if e.IsDir() {
			indexIcons(entry_fullpath)
			continue
		}
		icon_index[name[:len(name) - len(path.Ext(name))]] = entry_fullpath
	}
}
