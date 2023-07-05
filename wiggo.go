package main

import (
	"net/http"
	_ "net/http/pprof"
	"strings"
	"runtime"
	"flag"
	"os"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	. "github.com/glupi-borna/wiggo/internal/utils"
	. "github.com/glupi-borna/wiggo/internal/platform"
	. "github.com/glupi-borna/wiggo/internal/ui"
	// . "github.com/glupi-borna/wiggo/internal/debug"
	"github.com/glupi-borna/wiggo/internal/widget"
	"github.com/glupi-borna/wiggo/internal/globals"
)

var widget_name string

var profile = flag.Bool("profile", false, "run profiling webserver")
var timeout = flag.Uint64("timeout", 0, "stop running after this number of milliseconds (0 = no timeout)")

var display = flag.Int(
	"display", 0,
	"The number of the display that the widget should appear on.\n"+
	"-1 -> the display that currently contains the mouse cursor.")

var window_x = flag.Int(
	"x", 0,
	"The x-position of the widget.\n"+
	"Negative values are offset from the right side of the display.")

var window_y = flag.Int(
	"y", 0,
	"The y-position of the widget\n"+
	"Negative values are offset from the right side of the display.")

var window_anchor WindowAnchorFlag

func UsageHandler() {
	b := strings.Builder{}
	b.WriteString("Usage: wiggo [options] widget_name\n")
	b.WriteString("options:")

	flag.VisitAll(func (f *flag.Flag) {
		b.WriteString("\n-")
		b.WriteString(f.Name)
		t, usage := flag.UnquoteUsage(f)
		b.WriteString(" ")
		b.WriteString(t)
		usage_lines := strings.Split(usage, "\n")
		for _, line := range usage_lines {
			b.WriteString("\n\t")
			b.WriteString(line)
		}
	})

	println(b.String())
}

func main() {
	flag.Usage = UsageHandler

	flag.Var(window_anchor, "anchor",
	"Alignment of the widget against it's position\n" +
	"Supported values:\n" +
	"	top-left\n" +
	"	center")

	flag.Parse()
	if *profile {
		go func() {
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				println(err.Error())
			}
		}()
	}

	if flag.NArg() != 1 {
		println("Widget name not provided!")
		flag.Usage()
		os.Exit(1)
	}

	widget_name = flag.Arg(0)

	w, err := widget.Load(widget_name)
	Die(err)

	runtime.LockOSThread()

	err = sdl.Init(sdl.INIT_EVERYTHING)
	Die(err)
	defer sdl.Quit()

	err = ttf.Init()
	Die(err)
	defer ttf.Quit()

	running := true
	Platform.Init(PlatformInitOptions{
		X: int32(*window_x),
		Y: int32(*window_y),
		Anchor: window_anchor,
		Display: *display,
	})
	globals.Close = func () { running = false }

	UI := MakeUI()

	// LastFrameStart := uint64(0)
	// FrameTime := uint64(0)
	// count := 0
	// val := float32(5)

	err = w.Init()
	Die(err)
	defer func() {
		Die(w.Cleanup())
	}()

	last_err_text := ""

	for running {
		FrameStart := sdl.GetTicks64()
		if *timeout > 0 && FrameStart > *timeout { running = false }
		if Platform.KeyboardPressed(sdl.SCANCODE_Q) { running = false }

		ButtonMapUpdate(Platform.Keyboard)
		ButtonMapUpdate(Platform.Mouse)
		Platform.AnyKeyPressed = false
		Platform.MouseDelta.X = 0
		Platform.MouseDelta.Y = 0

		for event := sdl.PollEvent() ; event != nil ; event = sdl.PollEvent() {
			switch e := event.(type) {

			case *sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONDOWN {
					Platform.Mouse[e.Button] = BS_PRESSED
				} else {
					Platform.Mouse[e.Button] = BS_RELEASED
				}
				Platform.MousePos.X = float32(e.X)
				Platform.MousePos.Y = float32(e.Y)

			case *sdl.MouseMotionEvent:
				Platform.MousePos.X = float32(e.X)
				Platform.MousePos.Y = float32(e.Y)
				Platform.MouseDelta.X += float32(e.XRel)
				Platform.MouseDelta.Y += float32(e.YRel)

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					Platform.Keyboard[uint32(e.Keysym.Scancode)] = BS_PRESSED
					Platform.AnyKeyPressed = true
				} else {
					Platform.Keyboard[uint32(e.Keysym.Scancode)] = BS_RELEASED
				}

			case *sdl.QuitEvent:
				running = false
				break
			}
		}

		Platform.Renderer.SetDrawColor(0, 0, 0, 255)
		Platform.Renderer.Clear()

		Platform.Renderer.SetDrawColor(255, 0, 0, 255)

		UI.Begin(); {
			err := w.Frame()
			if err != nil {
				if err.Error() != last_err_text {
					last_err_text = err.Error()
					println(last_err_text)
				}
				UI.Root.Children = nil
				UI.Current = UI.Root
				UI.Root.Style = DefaultStyle.Copy()
				UI.Root.Style.Background = StyleVar(ColHex(0xff0000ff))
				WithNode(Column(), func(n *Node) {
					Text("Error in " + w.Name())
					Text("Check output for trace")
				})
			}
		} ; UI.End()

		UI.Render()
		Platform.Renderer.Present()
		// FrameTime = FrameStart - LastFrameStart
		// LastFrameStart = FrameStart
	}
}

func PrintTree(n *Node, indent string) {
	child_indent := indent + "  "
	println(indent + n.Type, n.RealSize.String())
	for _, child := range n.Children {
		PrintTree(child, child_indent)
	}
}
