package platform

import (
	"strconv"
	. "github.com/glupi-borna/wiggo/internal/utils"
	"github.com/glupi-borna/wiggo/internal/lru"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fontPath = "assets/test.ttf"
	fontSize = 16
)

var Platform platform

type Font struct {
	SDLFont *ttf.Font
	CacheName string
	Height   float32
}

type platform struct {
	Window *sdl.Window
	Renderer *sdl.Renderer
	Font Font
	Mouse map[uint8]BUTTON_STATE
	MousePos V2
	MouseDelta V2
	Keyboard map[uint32]BUTTON_STATE
	AnyKeyPressed bool
	Close func()
}

func (p *platform) Init(x, y, w, h int32) {
	var window_flags uint32 =
		sdl.WINDOW_SHOWN |
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
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1)

	window, err := sdl.CreateShapedWindow("", uint32(x), uint32(y), uint32(w), uint32(h), window_flags)
	Die(err)
	p.Window = window

	renderer, err := sdl.CreateRenderer(window, -1, renderer_flags)
	Die(err)
	p.Renderer = renderer

	p.Window.SetPosition(300, 300)

	font, err := ttf.OpenFont(fontPath, fontSize)
	Die(err)

	fh := font.Height()
	p.Font = Font{
		SDLFont: font,
		Height: float32(fh),
		CacheName: font.FaceFamilyName() + "|" + font.FaceStyleName() + strconv.Itoa(fh) + "|",
	}

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
func fontCacheKey(f Font, text string, c sdl.Color) string {
	return f.CacheName + ColStr(c) + "|" + text
}

func GetTextTexture(f Font, t string, c sdl.Color) *sdl.Texture {
	key := fontCacheKey(f, t, c)
	tex, ok := fontTextureCache.Get(key)
	if ok { return tex }
	surf, _ := f.SDLFont.RenderUTF8Blended(t, c)
	tex, _ = Platform.Renderer.CreateTextureFromSurface(surf)
	surf.Free()
	fontTextureCache.Set(key, tex)
	return tex
}

var textMetricsCache = make(map[string]V2, 100)
func textMetricsCacheKey(f Font, text string) string {
	return f.CacheName + "|" + text
}

func TextMetrics(text string) V2 {
	key := textMetricsCacheKey(Platform.Font, text)
	m, ok := textMetricsCache[key]
	if !ok {
		w, h, _ := Platform.Font.SDLFont.SizeUTF8(text)
		m = V2{float32(w), float32(h)}
		textMetricsCache[key] = m
	}
	return m
}

