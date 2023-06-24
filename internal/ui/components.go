package ui

import (
	"math"
	"golang.org/x/exp/constraints"
	"github.com/veandco/go-sdl2/sdl"
	. "github.com/glupi-borna/wiggo/internal/platform"
	. "github.com/glupi-borna/wiggo/internal/utils"
)

func drawNodeRectBg(pos V2, size V2, s *Style, hovered bool) {
	drawNodeRect(pos, size, s.CornerRadius, s.Background, s.Border, hovered)
}

func drawNodeRectFg(pos V2, size V2, s *Style, hovered bool) {
	drawNodeRect(pos, size, s.CornerRadius, s.Foreground, s.Border, hovered)
}

func drawNodeRect(
	pos V2, size V2,
	corner_style StyleVariant[float32],
	bg_color StyleVariant[sdl.Color],
	border_color StyleVariant[sdl.Color],
	hovered bool,
) {
	var corner float32
	var borderc sdl.Color
	var bgc sdl.Color

	if hovered {
		corner = corner_style.Active
		borderc = border_color.Active
		bgc = bg_color.Active
	} else {
		corner = corner_style.Normal
		borderc = border_color.Normal
		bgc = bg_color.Normal
	}

	border := borderc.A != 0
	bg := bgc.A != 0

	if corner == 0 {
		if bg {
			SetColor(bgc)
			DrawRectFilled(pos.X, pos.Y, size.X, size.Y)
		}

		if border {
			SetColor(borderc)
			DrawRectOutlined(pos.X, pos.Y, size.X, size.Y)
		}
	} else {
		if bg {
			SetColor(bgc)
			DrawRoundRectFilled(pos.X, pos.Y, size.X, size.Y, corner)
		}

		if border {
			SetColor(borderc)
			DrawRoundRectOutlined(pos.X, pos.Y, size.X, size.Y, corner)
		}
	}
}

func defaultRenderFn(n *Node) {
	s := n.GetStyle()
	drawNodeRectBg(n.Pos, n.RealSize, s, CurrentUI.Active == n.UID)
}

func textRenderFn(n *Node) {
	t := uiGet(n, "text", "")
	s := n.GetStyle()

	var c sdl.Color
	var hov = false
	if CurrentUI.Active == n.UID || n.IsChildOfUID(CurrentUI.Active) {
		hov = true
		c = s.Foreground.Active
	} else {
		c = s.Foreground.Normal
	}

	drawNodeRectBg(n.Pos, n.RealSize, s, hov)

	tex := GetTextTexture(Platform.Font, t, c)
	m := TextMetrics(t)

	Platform.Renderer.CopyF(tex, nil, &sdl.FRect{
		float32(math.Round(float64(n.Pos.X + n.Padding.Left))),
		float32(math.Round(float64(n.Pos.Y + n.Padding.Top))),
		m.X, m.Y,
	})
}

func sliderRenderFn(n *Node) {
	s := n.GetStyle()
	hov := n.Focused()
	perc := Animate(uiGet[float32](n, "perc", 0.5), n.UID + "-slider-anim")
	scaled_size := V2{ X: n.RealSize.X * perc, Y: n.RealSize.Y }
	if perc < 1 { drawNodeRectBg(n.Pos, n.RealSize, s, hov) }
	if perc > 0 { drawNodeRect(n.Pos, scaled_size, s.CornerRadius, s.Foreground, StyleVariant[sdl.Color]{}, hov) }
}

type Number interface {
	constraints.Integer|constraints.Float
}

func sliderUpdateFn(n *Node) {
	if CurrentUI.Mode == IM_MOUSE {
		if n.UID == n.UI.Hot {
			new_perc := Clamp((Platform.MousePos.X - n.Pos.X) / n.RealSize.X, 0, 1)
			n.Set("perc", new_perc)
		}

		if n.HasMouse() {
			CurrentUI.SetActive(n, false)
			if MousePressed(sdl.BUTTON_LEFT) { CurrentUI.SetHot(n, false) }
			if MouseReleased(sdl.BUTTON_LEFT) { CurrentUI.SetHot(nil, false) }
		}
	}

	if CurrentUI.Mode == IM_KBD && CurrentUI.Active == n.UID {
		// xkbd := Btof(KeyboardPressed(sdl.SCANCODE_RIGHT)) - Btof(KeyboardPressed(sdl.SCANCODE_LEFT))
		// ykbd := Btof(KeyboardPressed(sdl.SCANCODE_DOWN)) - Btof(KeyboardPressed(sdl.SCANCODE_UP))
		/*
			if xkbd != 0 && ykbd != 0 {
				_UI.MoveActive(n, xkbd, ykbd)
			}
		*/
	}
}

func invisibleRenderFn(n *Node) {
	SetColor(sdl.Color{255, 0, 0, 255})
	DrawRectOutlined(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
}

func WithNode(n *Node, fn func(*Node)) *Node {
	fn(n)
	CurrentUI.Pop(n)
	return n
}

func Row() *Node {
	n := CurrentUI.Push("row")
	n.Layout = LT_HORIZONTAL
	return n
}

func Column() *Node {
	n := CurrentUI.Push("column")
	n.Layout = LT_VERTICAL
	return n
}

// func Margin(dim Dimension, inner *Node) *Node {
// 	margin := GetNode("margin-h", nil)
// 	margin.RenderFn = invisibleRenderFn
// 	margin.Layout = LT_HORIZONTAL
// 	margin.Parent = inner.Parent
// 	idx := inner.Index()
//
// 	cur := UI.Current
// 	defer func(){ UI.Current = cur }()
// 	UI.Current = margin
//
// 	Invisible(dim)
// 	WithNode(UI.Push("margin-v"), func(r *Node) {
// 		r.Layout = LT_VERTICAL
// 		r.RenderFn = invisibleRenderFn
//
// 		Invisible(dim)
// 		r.Children = append(r.Children, inner)
// 		inner.Parent = r
// 		Invisible(dim)
// 	})
// 	Invisible(dim)
//
// 	margin.Parent.Children[idx] = margin
// 	margin.UID = buildNodeUID(margin)
//
// 	old_uid := inner.UID
// 	inner.UID = buildNodeUID(inner)
//
// 	UI_Data[inner.UID] = UI_Data[old_uid]
// 	delete(UI_Data, old_uid)
//
// 	return margin
// }

func Invisible(dim Dimension) *Node {
	n := CurrentUI.Push("invisible")
	defer CurrentUI.Pop(n)
	n.Size.W = dim
	n.Size.H = dim
	n.Padding = Padding{}
	n.RenderFn = invisibleRenderFn
	return n
}

func Text(text string) *Node {
	n := CurrentUI.Push("text")
	defer CurrentUI.Pop(n)

	n.Set("text", text)
	n.Size.W = FitText()
	n.Size.H = FitText()
	n.Padding = Padding{}

	n.RenderFn = textRenderFn
	return n
}

func TextButton(text string) bool {
	n := CurrentUI.Push("button")
	defer CurrentUI.Pop(n)

	n.Flags.Focusable = true
	n.Style = &ButtonStyle
	n.Padding = Padding2(8, 4)

	Text(text)

	return n.Clicked()
}

func Button(fn func(*Node)) *Node {
	n := CurrentUI.Push("button")
	defer CurrentUI.Pop(n)

	n.Flags.Focusable = true
	n.Style = &ButtonStyle
	n.Padding = Padding2(8, 4)

	WithNode(n, fn)

	return n
}

func Slider(val, min, max float32) (float32, *Node) {
	n := CurrentUI.Push("slider")
	defer CurrentUI.Pop(n)

	diff := max - min

	n.Flags.Focusable = true
	n.RenderFn = sliderRenderFn
	n.UpdateFn = sliderUpdateFn
	n.Style = SliderStyle
	n.Size.W = Px(200)


	var perc float32
	if n.UID == CurrentUI.Hot {
		perc = uiGet(n, "perc", perc)
	} else {
		perc = Clamp(val - min, 0, diff) / diff
		n.Set("perc", perc)
	}

	return (perc*diff)+min, n
}
