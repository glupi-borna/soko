package widget

import (
	"io/ioutil"
	"path"
	"strings"
	"errors"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/glupi-borna/soko/internal/ui"
	"github.com/glupi-borna/soko/internal/globals"
	"github.com/glupi-borna/soko/internal/sound"
	"github.com/glupi-borna/soko/internal/player"
)

type Widget interface {
	// Returns the name of the widget
	Name()    string

	// Returns the path of the widget file
	Path()    string

	// Returns a string that describes the type of widget (e.g. "lua")
	Type()    string

	// Exposes a named value to the widget environment
	Expose(name string, val any)

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

func ExposeEnvironment(w Widget) {
	w.Expose("UI", func() *ui.UI_State { return ui.CurrentUI })
	w.Expose("TextButton", ui.TextButton)
	w.Expose("Text", ui.Text)
	w.Expose("Animate", ui.Animate)
	w.Expose("Column", ui.Column)
	w.Expose("Row", ui.Row)
	w.Expose("Button", ui.Button)
	w.Expose("Slider", ui.Slider)
	w.Expose("VSlider", ui.VSlider)
	w.Expose("Invisible", ui.Invisible)
	w.Expose("Col", ui.Col)
	w.Expose("RGBA", func (r, g, b, a uint8) sdl.Color { return sdl.Color{r, g, b, a} })
	w.Expose("ColHex", ui.ColHex)
	w.Expose("Fr", ui.Fr)
	w.Expose("Px", ui.Px)
	w.Expose("Em", ui.Em)
	w.Expose("Auto", ui.Auto)
	w.Expose("ChildrenSize", ui.ChildrenSize)
	w.Expose("FitText", ui.FitText)
	w.Expose("Close", globals.Close)
	w.Expose("Padding", ui.Padding)
	w.Expose("Padding1", ui.Padding1)
	w.Expose("Padding2", ui.Padding2)
	w.Expose("Tick", ui.Tick)
	w.Expose("Marquee", ui.Marquee)
	for key, val := range sound.WidgetFns { w.Expose(key, val) }
	for key, val := range player.WidgetFns { w.Expose(key, val) }
}
