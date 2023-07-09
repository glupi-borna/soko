package ui

import (
	"github.com/veandco/go-sdl2/sdl"
)

type StyleVariant[K any] struct {
	Normal  K
	Active K
	Hot K
}

func ColHex(col uint32) sdl.Color {
	r := uint8(col >> 24 & 0xff)
	g := uint8(col >> 16 & 0xff)
	b := uint8(col >> 8 & 0xff)
	a := uint8(col & 0xff)
	return sdl.Color{r,g,b,a}
}

func Col(col uint8) sdl.Color {
	return sdl.Color{col, col, col, 255}
}

func StyleVar[K any](val K) StyleVariant[K] {
	return StyleVariant[K]{ Normal: val, Active: val, Hot: val }
}

func StyleVar2[K any](normal K, active_hot K) StyleVariant[K] {
	return StyleVariant[K]{ Normal: normal, Active: active_hot, Hot: active_hot }
}

func StyleVars[K any](norm, hov K) StyleVariant[K] {
	return StyleVariant[K]{ Normal: norm, Active: hov }
}

type PaddingType struct {
	Left, Right, Top, Bottom float32
}

func Padding1(pad float32) PaddingType {
	return PaddingType{pad, pad, pad, pad}
}

func Padding2(x float32, y float32) PaddingType {
	return PaddingType{x, x, y, y}
}

func Padding(args ...float32) PaddingType {
	if len(args) == 1 { return Padding1(args[0]) }
	if len(args) == 2 { return Padding2(args[0], args[1]) }
	if len(args) == 4 {
		return PaddingType{
			Left: args[0],
			Top: args[1],
			Right: args[2],
			Bottom: args[3],
		}
	}
	println("Padding(): unsupported number of arguments:", len(args))
	return PaddingType{}
}

type Style struct {
	Foreground   StyleVariant[sdl.Color]
	Background   StyleVariant[sdl.Color]
	Border   	 StyleVariant[sdl.Color]
	CornerRadius StyleVariant[float32]
	Padding      PaddingType
	Font         string
	FontSize     int
}

func (s *Style) Copy() *Style {
	new := *s
	return &new
}

func (s *Style) SetBorder(border StyleVariant[sdl.Color]) *Style {
	s.Border = border
	return s
}

func (s *Style) Invert() *Style {
	bg := s.Background
	s.Background = s.Foreground
	s.Foreground = bg
	return s
}

var DefaultStyle = Style{
	Foreground: StyleVar(Col(255)),
	Background: StyleVar(ColHex(0x0)),
	Border: StyleVar(ColHex(0x0)),
	Font: "Sans",
	FontSize: 16,
}

var ButtonStyle = Style{
	Foreground: StyleVar(Col(0)),
	Background: StyleVar2(Col(255), Col(200)),
	CornerRadius: StyleVar[float32](5),
}

var SliderStyle = ButtonStyle.Copy().
	Invert().
	SetBorder(
		StyleVar(Col(255)))
