package system

import (
	"fmt"
	"os/exec"
	"path"
	"os"
	"strings"
	. "github.com/glupi-borna/soko/internal/debug"
	. "github.com/glupi-borna/soko/internal/utils"
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

var warnIconNotFound = Once1(func(icon string) bool {
	fmt.Println("Icon not found: ", icon)
	fmt.Println("Similar icon names:")

	for k := range icon_index {
		if strings.Contains(k, icon) {
			fmt.Println("\t", k)
		}
	}

	return true
})

func GetIconPath(icon string) string {
	if !icons_indexed {
		root_path := path.Join(getIconFolder(), getIconThemeName())
		indexIcons(root_path)
		icons_indexed = true
	}

	path, ok := icon_index[icon]
	if DEBUG && !ok { warnIconNotFound(icon) }

	return path
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

		if e.Name() == "index.theme" {
			b, err := os.ReadFile(entry_fullpath)
			if err != nil { continue }

			lines := strings.Split(string(b), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				before, after, found := strings.Cut(line, "=")
				if !found { continue }
				if before != "Inherits" { continue }
				parts := strings.Split(after, ",")
				for _, part := range parts {
					inherited_path := path.Join(getIconFolder(), strings.TrimSpace(part))
					indexIcons(inherited_path)
				}
				break
			}
			continue
		}

		icon_index[name[:len(name) - len(path.Ext(name))]] = entry_fullpath
	}
}
