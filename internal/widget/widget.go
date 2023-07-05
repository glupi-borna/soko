package widget

import (
	"io/ioutil"
	"path"
	"strings"
	"errors"
)

type Widget interface {
	Name()    string
	Path()    string
	Type()    string
	Init()    error
	Frame()   error
	Cleanup() error
}

func RootPath() string {
	return "/home/borna"
}

func ExtSupported(ext string) bool {
	switch ext {
	case ".lua": return true
	default: return false
	}
}

// Finds widget definition files.
// Widget definitions have filenames starting with 'soko_', and ending with one
// of the supported extensions (currently only lua).
func FindWidgets() ([]Widget, error) {
	out := []Widget{}

	items, err := ioutil.ReadDir(".")
	if err != nil { return nil, err }

	for _, item := range items {
		if item.IsDir() { continue }

		filename := item.Name()
		if !strings.HasPrefix(filename, "soko_") { continue }

		ext := path.Ext(filename)
		if !ExtSupported(ext) { continue }
		name := filename[5:len(filename)-len(ext)]

		switch ext {
		case ".lua":
			out = append(out, MakeLuaWidget(name, filename))
		}
	}

	return out, nil
}

func Load(name string) (Widget, error) {
	all_widgets, err := FindWidgets()
	if err != nil { return nil, err }

	for _, w := range all_widgets {
		w.Name()
		if w.Name() != name { continue }
		return w, nil
	}

	return nil, errors.New("Widget '" + name + "' not found!")
}
