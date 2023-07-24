package platform

import (
	"errors"
	"os/exec"
	"strconv"
	. "github.com/glupi-borna/soko/internal/utils"
	"github.com/glupi-borna/soko/internal/lru"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type WindowAnchorFlag struct {
	V2
}

func (w *WindowAnchorFlag) Help() string {
	return `Available values:
	top-left       top[-center]        top-right
	[middle-]left middle[-center] [middle]-right
	bottom-left   bottom[-center]   bottom-right`
}

func (w *WindowAnchorFlag) String() string {
	return w.V2.String()
}

func (w *WindowAnchorFlag) Set(val string) error {
	switch val {
	case "top-left": w.X = 0 ; w.Y = 0

	case "top-center": fallthrough
	case "top": w.X = -0.5 ; w.Y = 0

	case "top-right": w.X = -1 ; w.Y = 0

	case "middle-left": fallthrough
	case "left": w.X = 0 ; w.Y = -0.5

	case "middle-center": fallthrough
	case "middle": w.X = -0.5 ; w.Y = -0.5

	case "middle-right": fallthrough
	case "right": w.X = -1 ; w.Y = -0.5

	case "bottom-left": w.X = 0 ; w.Y = -1

	case "bottom-center": fallthrough
	case "bottom": w.X = -0.5 ; w.Y = -1

	case "bottom-right": w.X = -1 ; w.Y = -1

	default:
		return errors.New("Aborting")
	}

	return nil
}

var Platform platform

type PlatformInitOptions struct {
	Display int
	X int32
	Y int32
	Anchor WindowAnchorFlag
}


type Font struct {
	SDLFont *ttf.Font
	CacheName string
	Height   float32
}

var loadedFontCache = lru.New(50, func (f *Font) { f.SDLFont.Close() })
func loadedFontCacheKey(font_name string, font_size int) string {
	return font_name + "|" + strconv.Itoa(font_size)
}

func GetFont(font_name string, font_size int) *Font {
	key := loadedFontCacheKey(font_name, font_size)
	font, ok := loadedFontCache.Get(key)
	if ok { return font }

	ttf_font, err := ttf.OpenFont(GetFontPath(font_name), font_size)
	Die(err)

	font = &Font{
		SDLFont: ttf_font,
		CacheName: key,
		Height: float32(font_size),
	}

	loadedFontCache.Set(key, font)
	return font
}

var fc_lookup = make(map[string]string)
func GetFontPath(font_name string) string {
	font_path, ok := fc_lookup[font_name]
	if ok { return font_path }

	cmd := exec.Command("fc-match", "--format=%{file}", font_name)
	stdout, err := cmd.Output()
	Die(err)

	font_path = string(stdout)
	fc_lookup[font_name] = font_path
	return font_path
}


type V2i struct { X, Y int32 }

type platform struct {
	Window *sdl.Window
	TargetDisplay int
	TargetPosition V2i
	AnchorOffset V2
	Renderer *sdl.Renderer
	Mouse map[uint8]BUTTON_STATE
	MousePos V2
	MouseDelta V2
	Keyboard map[uint32]BUTTON_STATE
	AnyKeyPressed bool

	Font *Font
	FontSize float64
	Color sdl.Color

	shapeSurf *sdl.Surface
	cornerRadius float32
}

func (p *platform) Init(opts PlatformInitOptions) {
	p.TargetDisplay = opts.Display
	p.TargetPosition = V2i{X: opts.X, Y: opts.Y}
	p.AnchorOffset = opts.Anchor.V2

	var window_flags uint32 =
		sdl.WINDOW_HIDDEN |
		sdl.WINDOW_BORDERLESS |
		sdl.WINDOW_UTILITY |
		// NOTE: if this flag is not provided, shaped
		// windows do not work - presumably because the
		// window gets recreated.
		sdl.WINDOW_OPENGL |
		sdl.WINDOW_ALWAYS_ON_TOP

	var renderer_flags uint32 =
		sdl.RENDERER_PRESENTVSYNC |
		sdl.RENDERER_ACCELERATED |
		sdl.RENDERER_TARGETTEXTURE

	sdl.SetHint("SDL_X11_FORCE_OVERRIDE_REDIRECT", "1")
	sdl.SetHint(sdl.HINT_FRAMEBUFFER_ACCELERATION, "0")
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "1")
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)

	window, err := sdl.CreateShapedWindow(
		"", 0, 0, 200, 200, window_flags)
	Die(err)
	p.Window = window

	renderer, err := sdl.CreateRenderer(window, -1, renderer_flags)
	Die(err)
	p.Renderer = renderer

	p.ResizeWindow(200, 200)
	p.ReshapeWindow(0, true)

	p.SetFont("Sans", 16)

	p.Mouse = make(map[uint8]BUTTON_STATE)
	p.Keyboard = make(map[uint32]BUTTON_STATE)
	p.MousePos.X = -1
	p.MousePos.Y = -1
}

func (p *platform) Cleanup() {
	p.Window.Destroy()
	p.Font.SDLFont.Close()
}

var fontTextureCache = lru.New(200, func (t *sdl.Texture) { t.Destroy() })
func fontCacheKey(f *Font, text string, c sdl.Color) string {
	return f.CacheName + ColStr(c) + "|" + text
}

func (p *platform) GetTextTexture(f *Font, t string, c sdl.Color) *sdl.Texture {
	key := fontCacheKey(f, t, c)
	tex, ok := fontTextureCache.Get(key)
	if ok { return tex }
	surf, _ := f.SDLFont.RenderUTF8Blended(t, c)
	tex, _ = p.Renderer.CreateTextureFromSurface(surf)
	surf.Free()
	fontTextureCache.Set(key, tex)
	return tex
}

var textMetricsCache = make(map[string]V2, 100)
func textMetricsCacheKey(f *Font, text string) string {
	return f.CacheName + "|" + text
}

func (p *platform) TextMetrics(text string) V2 {
	key := textMetricsCacheKey(p.Font, text)
	m, ok := textMetricsCache[key]
	if !ok {
		w, h, _ := p.Font.SDLFont.SizeUTF8(text)
		m = V2{float32(w), float32(h)}
		textMetricsCache[key] = m
	}
	return m
}

func (p *platform) TargetDisplayBounds() (sdl.Rect, error) {
	if p.TargetDisplay != -1 {
		return sdl.GetDisplayBounds(p.TargetDisplay)
	}

	switch p.TargetDisplay {
	case -1:
		x, y, _ := sdl.GetGlobalMouseState()
		displays, err := sdl.GetNumVideoDisplays()
		return sdl.Rect{}, err
		for i:=0 ; i<displays ; i++ {
			bounds, err := sdl.GetDisplayBounds(i)
			if err != nil { return sdl.Rect{}, err }
			if x < bounds.X { continue }
			if y < bounds.Y { continue }
			if x > bounds.X + bounds.W { continue }
			if y > bounds.Y + bounds.H { continue }
			return bounds, nil
		}
	default:
		return sdl.GetDisplayBounds(p.TargetDisplay)
	}

	return sdl.Rect{}, errors.New("Invalid -display constant: " + strconv.Itoa(p.TargetDisplay))
}

func (p *platform) ReshapeWindow(radius float32, force bool) {
	window_resized := false
	radius_changed := radius != p.cornerRadius
	p.cornerRadius = radius

	w, h := p.Window.GetSize()
	if p.shapeSurf != nil {
		window_resized = p.shapeSurf.W != w || p.shapeSurf.H != h
	}

	if !window_resized && !radius_changed && !force { return }

	if (window_resized && p.shapeSurf != nil) || p.shapeSurf == nil {
		var err error
		p.shapeSurf, err = sdl.CreateRGBSurfaceWithFormat(
			0, w, h,
			32, sdl.PIXELFORMAT_RGBA8888)
		Die(err)
	}

	shape := p.shapeSurf

	shape.FillRect(nil, 0)
	ir := int32(radius)

	if radius > 0 {
		pxls := shape.Pixels()
		rsq := ir*ir
		bpp := int32(shape.BytesPerPixel())

		for x := -ir; x < ir ; x++ {
			xsq := x*x
			for y := -ir; y < ir ; y++ {
				if xsq + y*y < rsq {
					i := shape.Pitch * (y+ir) + bpp * (x+ir)
					pxls[i] = 255
					i = shape.Pitch * (y+ir) + bpp * (x+w-ir-1)
					pxls[i] = 255
					i = shape.Pitch * (y+h-ir-1) + bpp * (x+ir)
					pxls[i] = 255
					i = shape.Pitch * (y+h-ir-1) + bpp * (x+w-ir-1)
					pxls[i] = 255
				}
			}
		}
	}

	r := sdl.Rect{ir, 0, w-ir*2, h}
	shape.FillRect(&r, 255)
	r = sdl.Rect{0, ir, w, h-ir*2}
	shape.FillRect(&r, 255)

	x, y := Platform.Window.GetPosition()
	Platform.Window.SetShape(shape, sdl.ShapeModeDefault{})
	Platform.Window.SetPosition(x, y)
}

func (p *platform) ResizeWindow(width int32, height int32) {
	ow, oh := p.Window.GetSize()
	ox, oy := p.Window.GetPosition()

	if ox != -1000 && oy != -1000 {
		if ow == width && oh == height { return }
	}

	bounds, err := p.TargetDisplayBounds()
	Die(err)
	dx := bounds.X
	dy := bounds.Y

	ax := p.AnchorOffset.X
	ay := p.AnchorOffset.Y

	if p.TargetPosition.X < 0 {
		dx += bounds.W + p.TargetPosition.X + 1
		ax *= -1
	} else {
		dx += p.TargetPosition.X
	}

	if p.TargetPosition.Y < 0 {
		dy += bounds.H + p.TargetPosition.Y + 1
		ay *= -1
	} else {
		dy += p.TargetPosition.Y
	}

	x := int32(dx) + int32(float32(width) * p.AnchorOffset.X)
	y := int32(dy) + int32(float32(height) * p.AnchorOffset.Y)
	p.Window.SetSize(width, height)
	p.Window.SetPosition(x, y)
}

func (p *platform) WindowWidth() float32 {
	w, _ := p.Window.GetSize()
	return float32(w)
}

func (p *platform) WindowHeight() float32 {
	_, h := p.Window.GetSize()
	return float32(h)
}

func (p *platform) TextWidth(text string) float32 {
	return p.TextMetrics(text).X
}

func (p *platform) TextHeight(text string) float32 {
	return p.TextMetrics(text).Y
}
