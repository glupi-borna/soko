package ui

import (
	"github.com/veandco/go-sdl2/sdl"
)

type StyleVariant[K any] struct {
	Normal  K
	Hovered K
}

func StyleVar[K any](val K) StyleVariant[K] {
	return StyleVariant[K]{ Normal: val, Hovered: val }
}

func StyleVars[K any](norm, hov K) StyleVariant[K] {
	return StyleVariant[K]{ Normal: norm, Hovered: hov }
}

type Padding struct {
	Left, Right, Top, Bottom float32
}

func Padding1(pad float32) Padding {
	return Padding{pad, pad, pad, pad}
}

func Padding2(x float32, y float32) Padding {
	return Padding{x, x, y, y}
}

type Style struct {
	Foreground   StyleVariant[sdl.Color]
	Background   StyleVariant[sdl.Color]
	Border   	 StyleVariant[sdl.Color]
	CornerRadius StyleVariant[float32]
	Padding      Padding
}

func (s *Style) Copy() *Style {
	new := *s
	return &new
}

var DefaultStyle = Style{
	Foreground: StyleVariant[sdl.Color]{
		Normal: sdl.Color{255, 255, 255, 255},
		Hovered: sdl.Color{255, 255, 255, 255},
	},
	Background: StyleVariant[sdl.Color]{
		Normal: sdl.Color{0, 0, 0, 0},
		Hovered: sdl.Color{0, 0, 0, 0},
	},
	Border: StyleVariant[sdl.Color]{
		Normal: sdl.Color{0, 0, 0, 0},
		Hovered: sdl.Color{0, 0, 0, 0},
	},
	CornerRadius: StyleVariant[float32]{
		Normal: 3,
		Hovered: 3,
	},
}

var ButtonStyle = Style{
	Foreground: StyleVariant[sdl.Color]{
		Normal: sdl.Color{0, 0, 0, 255},
		Hovered: sdl.Color{0, 0, 0, 255},
	},
	Background: StyleVariant[sdl.Color]{
		Normal: sdl.Color{255, 255, 255, 255},
		Hovered: sdl.Color{50, 50, 50, 255},
	},
	CornerRadius: StyleVariant[float32]{
		Normal: 5,
		Hovered: 5,
	},
}
