package ui

import (
	"math"
	"github.com/veandco/go-sdl2/sdl"
	. "github.com/glupi-borna/wiggo/internal/platform"
)

func drawNodeRect(n *Node, s *Style, hovered bool) {
	var corner float32
	var borderc sdl.Color
	var bgc sdl.Color

	if hovered {
		corner = s.CornerRadius.Hovered
		borderc = s.Border.Hovered
		bgc = s.Background.Hovered
	} else {
		corner = s.CornerRadius.Normal
		borderc = s.Border.Normal
		bgc = s.Background.Normal
	}

	border := borderc.A != 0
	bg := bgc.A != 0

	if corner == 0 {
		if bg {
			SetColor(bgc)
			DrawRectFilled(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
		}

		if border {
			SetColor(borderc)
			DrawRectOutlined(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
		}
	} else {
		if bg {
			SetColor(bgc)
			DrawRoundRectFilled(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y, corner)
		}

		if border {
			SetColor(borderc)
			DrawRoundRectOutlined(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y, corner)
		}
	}
}

func defaultRenderFn(n *Node) {
	s := n.GetStyle()
	drawNodeRect(n, s, CurrentUI.Active == n.UID)
}

func textRenderFn(n *Node) {
	t := uiGet(n, "text", "")
	s := n.GetStyle()

	var c sdl.Color
	var hov = false
	if CurrentUI.Active == n.UID || n.IsChildOfUID(CurrentUI.Active) {
		hov = true
		c = s.Foreground.Hovered
	} else {
		c = s.Foreground.Normal
	}

	drawNodeRect(n, s, hov)

	tex := GetTextTexture(Platform.Font, t, c)
	m := TextMetrics(t)

	Platform.Renderer.CopyF(tex, nil, &sdl.FRect{
		float32(math.Round(float64(n.Pos.X + n.Padding.Left))),
		float32(math.Round(float64(n.Pos.Y + n.Padding.Top))),
		m.X, m.Y,
	})
}

func invisibleRenderFn(*Node) {}

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
