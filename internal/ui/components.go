package ui

import (
	. "github.com/glupi-borna/soko/internal/debug"
	. "github.com/glupi-borna/soko/internal/platform"
	. "github.com/glupi-borna/soko/internal/utils"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/exp/constraints"
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
			Platform.SetColor(bgc)
			Platform.DrawRectFilled(pos.X, pos.Y, size.X, size.Y)
		}

		if border {
			Platform.SetColor(borderc)
			Platform.DrawRectOutlined(pos.X, pos.Y, size.X, size.Y)
		}
	} else {
		if bg {
			Platform.SetColor(bgc)
			Platform.DrawRoundRectFilled(pos.X, pos.Y, size.X, size.Y, corner)
		}

		if border {
			Platform.SetColor(borderc)
			Platform.DrawRoundRectOutlined(pos.X, pos.Y, size.X, size.Y, corner)
		}
	}
}

func defaultRenderFn(n *Node) {
	s := n.GetStyle()
	drawNodeRectBg(n.Pos, n.RealSize, s, CurrentUI.Active == n.UID)
}

func textRenderFn(n *Node) {
	t := n.Text
	s := n.GetStyle()
	f := n.GetFont()

	var c sdl.Color
	var hov = false
	if CurrentUI.Active == n.UID || n.IsChildOfUID(CurrentUI.Active) {
		hov = true
		c = s.Foreground.Active
	} else {
		c = s.Foreground.Normal
	}

	drawNodeRectBg(n.Pos, n.RealSize, s, hov)
	Platform.SetColor(c)
	Platform.RawSetFont(f)
	Platform.DrawText(t, n.Pos.X, n.Pos.Y)
}

type SliderState struct{ Perc float32 }

func hSliderRenderFn(n *Node) {
	state := NodeState[SliderState](n)
	s := n.GetStyle()
	hov := n.Focused()
	perc := Animate(state.Perc, n.UID+"-slider-anim")

	scaled_size := V2{X: n.RealSize.X * perc, Y: n.RealSize.Y}
	if perc < 1 {
		drawNodeRectBg(n.Pos, n.RealSize, s, hov)
	}
	if perc > 0 {
		drawNodeRect(n.Pos, scaled_size, s.CornerRadius, s.Foreground, StyleVariant[sdl.Color]{}, hov)
	}
}

func vSliderRenderFn(n *Node) {
	state := NodeState[SliderState](n)
	s := n.GetStyle()
	hov := n.Focused()
	perc := Animate(state.Perc, n.UID+"-slider-anim")

	scaled_size := V2{X: n.RealSize.X, Y: n.RealSize.Y * perc}

	if perc < 1 {
		drawNodeRectBg(n.Pos, n.RealSize, s, hov)
	}
	if perc > 0 {
		drawNodeRect(
			V2{
				X: n.Pos.X + (n.RealSize.X - scaled_size.X),
				Y: n.Pos.Y + (n.RealSize.Y - scaled_size.Y),
			}, scaled_size,
			s.CornerRadius, s.Foreground,
			StyleVariant[sdl.Color]{}, hov)
	}
}

type Number interface {
	constraints.Integer | constraints.Float
}

func hSliderUpdateFn(n *Node) {
	state := NodeState[SliderState](n)

	if CurrentUI.Mode == IM_MOUSE {
		if n.UID == n.UI.Hot {
			state.Perc = Clamp((Platform.MousePos.X-n.Pos.X)/n.RealSize.X, 0, 1)
		}

		if n.HasMouse() {
			CurrentUI.SetActive(n, false)
			if Platform.MousePressed(sdl.BUTTON_LEFT) {
				CurrentUI.SetHot(n, false)
			}
			if Platform.MouseReleased(sdl.BUTTON_LEFT) {
				CurrentUI.SetHot(nil, false)
			}
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

func vSliderUpdateFn(n *Node) {
	state := NodeState[SliderState](n)

	if CurrentUI.Mode == IM_MOUSE {
		if n.UID == n.UI.Hot {
			state.Perc = 1 - Clamp((Platform.MousePos.Y-n.Pos.Y)/n.RealSize.Y, 0, 1)
		}

		if n.HasMouse() {
			CurrentUI.SetActive(n, false)
			if Platform.MousePressed(sdl.BUTTON_LEFT) {
				CurrentUI.SetHot(n, false)
			}
			if Platform.MouseReleased(sdl.BUTTON_LEFT) {
				CurrentUI.SetHot(nil, false)
			}
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

func Invisible(dim Dimension) *Node {
	n := CurrentUI.Push("invisible")
	defer CurrentUI.Pop(n)
	n.Size.W = dim
	n.Size.H = dim
	n.Padding = PaddingType{}
	n.RenderFn = nil
	return n
}

func Text(text string) *Node {
	n := CurrentUI.Push("text")
	defer CurrentUI.Pop(n)

	n.Text = text
	n.Size.W = FitText()
	n.Size.H = FitText()
	n.Padding = PaddingType{}

	n.RenderFn = textRenderFn
	return n
}

type MarqueeState struct{ Speed float32 }

func Marquee(text string, speed float32) *Node {
	n := CurrentUI.Push("marquee")
	defer CurrentUI.Pop(n)

	state := NodeState[MarqueeState](n)

	n.Text = text
	state.Speed = speed
	n.Size.W = Em(8)
	n.Size.H = Em(1)
	n.Padding = PaddingType{}

	n.RenderFn = marqueeRenderFn
	return n
}

func marqueeRenderFn(n *Node) {
	state := NodeState[MarqueeState](n)
	t := n.Text
	speed := state.Speed
	s := n.GetStyle()
	f := n.GetFont()

	var c sdl.Color
	var hov = false
	if CurrentUI.Active == n.UID || n.IsChildOfUID(CurrentUI.Active) {
		hov = true
		c = s.Foreground.Active
	} else {
		c = s.Foreground.Normal
	}

	drawNodeRectBg(n.Pos, n.RealSize, s, hov)

	x := n.Pos.X
	y := n.Pos.Y

	m := Platform.TextMetrics(t)

	if m.X <= n.RealSize.X {
		Platform.SetColor(c)
		Platform.DrawText(t, x, y)
	} else {
		w := int32(n.RealSize.X)
		tex := Platform.GetTextTexture(f, t, c)
		_, _, qtw, _, _ := tex.Query()
		tw := float32(qtw)
		maxoff := tw - n.RealSize.X
		total_time_pps := maxoff / speed
		total_time_ms := max(uint64(total_time_pps*1000), 2000)

		perc := float64(uint64(CurrentUI.FrameStart.Milliseconds())%total_time_ms) / float64(total_time_ms)
		pperc := Clamp((perc-0.25)*2, 0, 1)
		xoff := int32(pperc * float64(maxoff))

		width := tw - float32(xoff)
		// fmt.Println(xoff, maxoff, FloatStr(perc), FloatStr(pperc))
		// Platform.DrawRectOutlined(x-float32(xoff), y, tw, n.RealSize.Y)

		target := sdl.FRect{
			X: x, Y: y,
			W: min(float32(w), width),
			H: m.Y,
		}

		source := sdl.Rect{
			X: xoff, Y: 0,
			W: min(w, int32(width)), H: int32(m.Y),
		}

		Platform.Renderer.CopyF(tex, &source, &target)
	}
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

func Button() (*Node, bool) {
	n := CurrentUI.Push("button")
	n.Flags.Focusable = true
	n.Style = &ButtonStyle
	n.Padding = Padding2(4, 4)
	return n, n.Clicked()
}

func Slider(val, min, max float32) (float32, *Node) {
	n := CurrentUI.Push("hslider")
	defer CurrentUI.Pop(n)

	state := NodeState[SliderState](n)
	diff := max - min

	n.Flags.Focusable = true
	n.RenderFn = hSliderRenderFn
	n.UpdateFn = hSliderUpdateFn
	n.Style = SliderStyle
	n.Size.W = Px(200)

	var perc float32
	if n.UID == CurrentUI.Hot {
		perc = state.Perc
	} else {
		perc = Clamp(val-min, 0, diff) / diff
		state.Perc = perc
	}

	return (perc * diff) + min, n
}

func VSlider(val, min, max float32) (float32, *Node) {
	n := CurrentUI.Push("vslider")
	defer CurrentUI.Pop(n)

	state := NodeState[SliderState](n)
	diff := max - min

	n.Flags.Focusable = true
	n.RenderFn = vSliderRenderFn
	n.UpdateFn = vSliderUpdateFn
	n.Style = SliderStyle
	n.Size.H = Px(200)

	var perc float32
	if n.UID == CurrentUI.Hot {
		perc = state.Perc
	} else {
		perc = Clamp(val-min, 0, diff) / diff
		state.Perc = perc
	}

	return (perc * diff) + min, n
}

func imgRenderFn(n *Node) {
	url := n.Text
	drawNodeRectBg(n.Pos, n.RealSize, n.GetStyle(), CurrentUI.Active == n.UID)
	Platform.DrawImage(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y, url)
}

func Image(url string) *Node {
	n := CurrentUI.Push("img")
	defer CurrentUI.Pop(n)

	n.Text = url
	n.RenderFn = imgRenderFn
	w, h := Platform.ImageSize(url)
	n.Size.W = Px(float32(w))
	n.Size.H = Px(float32(h))

	return n
}

type Scroller struct {
	Window, View *Node
}

type ScrollerState struct {
	Offset V2
	ParentClip sdl.Rect
}

func scrollWindowRenderFn(n *Node) {
	state := NodeState[ScrollerState](n)
	state.ParentClip = Platform.Renderer.GetClipRect()
	rect := n.sdlRect()
	Platform.Renderer.SetClipRect(&rect)
}

func scrollWindowPostRenderFn(n *Node) {
	state := NodeState[ScrollerState](n)
	Platform.Renderer.SetClipRect(&state.ParentClip)
}

func scrollWindowPreLayout(n *Node) {
	state := NodeState[ScrollerState](n)
	Assert(
		len(n.Children) == 1,
		"scroll_window must have exactly 1 child node!")
	Assert(
		n.Children[0].Type == "scroll_view",
		"scroll_window must have a scroll_view child!")
	n.Children[0].Translation = state.Offset
}

func scrollWindowUpdateFn(n *Node) {
	n.UpdateChildren()

	state := NodeState[ScrollerState](n)

	if CurrentUI.Mode == IM_MOUSE {
		if n.UID == n.UI.ScrollTarget {
			state.Offset.X += Platform.WheelDelta.X * 10
			state.Offset.Y += Platform.WheelDelta.Y * 10
		}

		if n.HasMouse() {
			CurrentUI.SetScrollTarget(n, false)
		}
	}

	if CurrentUI.Mode == IM_KBD && CurrentUI.Active == n.UID {
		// @TODO: Keyboard interaction
	}
}

func ScrollBegin() Scroller {
	var s Scroller
	s.Window = CurrentUI.Push("scroll_window")
	s.View = CurrentUI.Push("scroll_view")

	s.Window.Size.W = Px(320)
	s.Window.Size.H = Px(240)
	s.Window.RenderFn = scrollWindowRenderFn
	s.Window.PostRenderFn = scrollWindowPostRenderFn
	s.Window.UpdateFn = scrollWindowUpdateFn
	s.Window.PreLayout = scrollWindowPreLayout

	s.View.Size.W = ChildrenSize()
	s.View.Size.H = ChildrenSize()

	return s
}

func ScrollEnd() {
	Assert(CurrentUI.Current.Type == "scroll_view", "Scroll")
	Assert(CurrentUI.Current.Parent.Type == "scroll_window", "Scroll")
	CurrentUI.Pop(CurrentUI.Current.Parent)
}
