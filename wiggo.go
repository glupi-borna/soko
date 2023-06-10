package main

import (
	"runtime"
	"strconv"
	"golang.org/x/exp/constraints"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type DIM_TYPE uint8
const (
	DT_AUTO DIM_TYPE = iota
	DT_PX ; DT_FR
)

func FloatStr[F constraints.Float](f F) string {
	return strconv.FormatFloat(float64(f), 'f', 3, 32)
}

type Dimension struct {
	Type       DIM_TYPE
	Amount     float32
	Real       float32
}

func (d *Dimension) FracPX(fracsize float32) float32 {
	if d.Type != DT_FR { return 0 }
	d.Real = 0
	if d.Amount <= 0 { return 0 }
	d.Real = d.Amount * fracsize
	return d.Real
}

func (d *Dimension) String() string {
	switch d.Type {
		case DT_AUTO: return "auto"
		case DT_FR: return FloatStr(d.Amount) + "fr"
		case DT_PX: return FloatStr(d.Amount) + "px"
		default: panic("Unknown Dimension Type: " + strconv.Itoa(int(d.Type)))
	}
}

func px(amount float32) Dimension { return Dimension{Type: DT_PX, Amount: amount} }
func fr(amount float32) Dimension { return Dimension{Type: DT_FR, Amount: amount} }
func Auto() Dimension { return Dimension{ Type: DT_AUTO } }

type Size struct {
	W, H Dimension
	Padding Padding
	IsResolved bool
	Real V2
}

// The sum of the static parts of this Size:
// - width (if defined in pixels)
// - width of the left padding (if defined in pixels)
// - width of the right padding (if defined in pixels)
func (s *Size) StaticWidth() float32 {
	var w float32 = 0
	if s.W.Type == DT_PX { w += s.W.Amount ; s.W.Real = s.W.Amount }
	if s.Padding.Right.Type == DT_PX { w += s.Padding.Right.Amount ; s.Padding.Right.Real = s.Padding.Right.Amount }
	if s.Padding.Left.Type == DT_PX { w += s.Padding.Left.Amount ; s.Padding.Left.Real = s.Padding.Left.Amount }
	return w
}

// The sum of the static parts of this Size:
// - height (if defined in pixels)
// - height of the top padding (if defined in pixels)
// - height of the bottom padding (if defined in pixels)
func (s *Size) StaticHeight() float32 {
	var h float32 = 0
	if s.H.Type == DT_PX { h += s.H.Amount ; s.H.Real = s.H.Amount }
	if s.Padding.Top.Type == DT_PX { h += s.Padding.Top.Amount ; s.Padding.Top.Real = s.Padding.Top.Amount }
	if s.Padding.Bottom.Type == DT_PX { h += s.Padding.Bottom.Amount ; s.Padding.Bottom.Real = s.Padding.Bottom.Amount }
	return h
}

// The sum of the dynamic parts of a node's size:
// - width of the node (if defined in fractions)
// - width of the left padding (if defined in fractions)
// - width of the right padding (if defined in fractions)
// This requires a Node to be passed in because the width
// of a single fraction depends on the total count of
// fractions defined in the widths of that Node's siblings.
func (s *Size) DynamicWidth(n *Node) float32 {
	fw := n.FracWidth()
	return s.W.FracPX(fw) + s.Padding.Right.FracPX(fw) + s.Padding.Left.FracPX(fw)
}

// The sum of the dynamic parts of a node's size:
// - height of the node (if defined in fractions)
// - height of the top padding (if defined in fractions)
// - height of the bottom padding (if defined in fractions)
// This requires a Node to be passed in because the height
// of a single fraction depends on the total count of
// fractions defined in the heights of that Node's siblings.
func (s *Size) DynamicHeight(n *Node) float32 {
	fh := n.FracHeight()
	return s.H.FracPX(fh) + s.Padding.Top.FracPX(fh) + s.Padding.Bottom.FracPX(fh)
}

// Returns the node's resolved size in pixels
func (s *Size) Resolved(n *Node) V2 {
	if s.IsResolved { return s.Real }

	var w float32 = n.Size.StaticWidth() + n.Size.DynamicWidth(n)
	var h float32 = n.Size.StaticHeight() + n.Size.DynamicHeight(n)

	text := ""
	if n.Type == "text" {
		ok := false
		text, ok = n.Get("text", "").(string)
		if !ok { text = "" }
	}

	if n.Size.W.Type == DT_AUTO {
		if n.Type == "text" {
			w += Min(TextWidth(text), n.ParentWidth())
		} else {
			w += n.ChildrenWidth()
		}
	}

	if n.Size.H.Type == DT_AUTO {
		if n.Type == "text" {
			h += TextHeightWrapped(text, n.ParentHeight())
		} else {
			h += n.ChildrenHeight()
		}
	}

	s.Real.X = w
	s.Real.Y = h
	s.IsResolved = true

	return s.Real
}

func (s *Size) String() string {
	x := s.W.String()
	y := s.H.String()
	return "Size{ W: " + x + ", H: " + y + " }"
}

type Padding struct {
	Top, Right, Bottom, Left Dimension
}

type LAYOUT_TYPE uint8
const (
	LT_VERTICAL LAYOUT_TYPE = iota
	LT_HORIZONTAL
)

type V2 struct { X, Y float32 }

func (v *V2) ManhattanLength() float32 {
	return Abs(v.X) + Abs(v.Y)
}

func (v *V2) String() string {
	return "V2{ " + FloatStr(v.X) + ", " + FloatStr(v.Y) + " }"
}

type Style struct {
	Foreground sdl.Color
	Background sdl.Color
}

type NodeData map[string] any

var UI_Data = make(map[string]NodeData)

func uiDataGet(n *Node, key string, dflt any) any {
	data, ok := UI_Data[n.UID]
	if !ok {
		data = make(NodeData)
		UI_Data[n.UID] = data
		data[key] = dflt
		return dflt
	}
	val, ok := data[key]
	if !ok {
		data[key] = dflt
		return dflt
	}
	return val
}

// Sets data for this node.
// Returns true if the data has changed.
func uiDataSet(n *Node, key string, val any) bool {
	data, ok := UI_Data[n.UID]
	if !ok {
		data = make(NodeData)
		UI_Data[n.UID] = data
		data[key] = val
		return true
	}
	old := data[key]
	data[key] = val
	return old != val
}

type NodeFlags struct {
	Focusable bool
}

type Node struct {
	// Unique ID for this Node
	UID       string
	// Equivalent to tag name in HTML
	Type      string
	Layout    LAYOUT_TYPE
	Flags     NodeFlags
	Parent    *Node
	// Logical size
	Size      Size

	Children  []*Node
	// Real position (in pixels)
	Pos       V2


	// Called after the position and size of the Node
	// have been resolved.
	RenderFn func(*Node)
	UpdateFn func(*Node)
}

// Currently does the same as MakeNode, but
// could be used for object pooling later on.
func GetNode(t string, parent *Node) *Node {
	n := MakeNode(t, parent)
	return n
}

func MakeNode(t string, parent *Node) *Node {
	n := Node{
		Type: t,
		Children: make([]*Node, 0),
		Parent: parent,
		RenderFn: defaultRenderFn,
		UpdateFn: defaultUpdateFn,
	}

	if n.Parent != nil {
		n.Parent.Children = append(n.Parent.Children, &n)
		n.UID = n.Parent.UID + "." + n.Type + strconv.Itoa(n.Parent.CountChildrenOfType(n.Type))
	} else {
		n.UID = n.Type
	}

	return &n
}

func (n *Node) Get(key string, dflt any) any {
	return uiDataGet(n, key, dflt)
}

func (n *Node) Set(key string, val any) bool {
	return uiDataSet(n, key, val)
}

func defaultRenderFn(n *Node) {
	if UI.Active == n.UID {
		DrawRectFilled(n.Pos.X, n.Pos.Y, n.Size.Real.X, n.Size.Real.Y)
	}
	DrawRectOutlined(n.Pos.X, n.Pos.Y, n.Size.Real.X, n.Size.Real.Y)

	// println(n.UID)
	// println("  pos ", n.Pos.String())
	// println("  size", n.Size.Real.String())
	// println("  res ", n.Size.IsResolved)
}

func defaultUpdateFn(n *Node) {
	n.UpdateChildren()

	if n.Flags.Focusable {
		if UI.Mode == IM_MOUSE {
			if n.HasMouse() {
				UI.SetActive(n, false)
				if MousePressed(sdl.BUTTON_LEFT) { UI.SetHot(n, false) }
				if MouseReleased(sdl.BUTTON_LEFT) { UI.SetHot(nil, false) }
			}
		}

		if UI.Mode == IM_KBD && UI.Active == n.UID {
			// xkbd := Btof(KeyboardPressed(sdl.SCANCODE_RIGHT)) - Btof(KeyboardPressed(sdl.SCANCODE_LEFT))
			// ykbd := Btof(KeyboardPressed(sdl.SCANCODE_DOWN)) - Btof(KeyboardPressed(sdl.SCANCODE_UP))
			/*
				if xkbd != 0 && ykbd != 0 {
					UI.MoveActive(n, xkbd, ykbd)
				}
			*/
		}
	}
}

func rootUpdateFn(n *Node) {
	n.UpdateChildren()

	if UI.Mode == IM_MOUSE {
		if n.HasMouse() {
			UI.SetActive(nil, false)
			if MousePressed(sdl.BUTTON_LEFT) { UI.SetHot(nil, false) }
		}
	}
}

func (n *Node) CountChildrenOfType(t string) int {
	count := 0
	for _, child := range n.Children {
		if child.Type == t { count++ }
	}
	return count
}

func (n *Node) Render() {
	n.RenderFn(n)
	for _, child := range n.Children { child.Render() }
}

func (n *Node) StaticWidth() float32 {
	var w float32 = 0

	if n.Layout == LT_HORIZONTAL {
		for _, child := range n.Children {
			w += child.Size.StaticWidth()
		}

	} else if n.Layout == LT_VERTICAL {
		for _, child := range n.Children {
			cw := child.Size.StaticWidth()
			if cw > w { w = cw }
		}
	}

	return w
}

func (n *Node) StaticHeight() float32 {
	var h float32 = 0

	if n.Layout == LT_HORIZONTAL {
		for _, child := range n.Children {
			ch := child.Size.StaticHeight()
			if ch > h { h = ch }
		}

	} else if n.Layout == LT_VERTICAL {
		for _, child := range n.Children {
			h += child.Size.StaticHeight()
		}
	}

	return h
}

func (n *Node) ChildrenWidth() float32 {
	var w float32 = 0

	if n.Layout == LT_HORIZONTAL {
		for _, child := range n.Children {
			w += child.Size.Resolved(child).X
		}

	} else if n.Layout == LT_VERTICAL {
		for _, child := range n.Children {
			cw := child.Size.Resolved(child).X
			if cw > w { w = cw }
		}
	}

	return w
}

func (n *Node) ChildrenHeight() float32 {
	var h float32 = 0

	if n.Layout == LT_HORIZONTAL {
		for _, child := range n.Children {
			ch := child.Size.Resolved(child).Y
			if ch > h { h = ch }
		}

	} else if n.Layout == LT_VERTICAL {
		for _, child := range n.Children {
			h += child.Size.Resolved(child).Y
		}
	}

	return h
}

func (n *Node) ParentWidth() float32 {
	if n.Parent == nil {
		return WindowWidth()
	} else {
		if n.Parent.Size.IsResolved {
			return n.Parent.Size.Real.X
		} else {
			return n.Parent.ParentWidth()
		}
	}
}

func (n *Node) ParentHeight() float32 {
	if n.Parent == nil {
		return WindowHeight()
	} else {
		if n.Parent.Size.IsResolved {
			return n.Parent.Size.Real.Y
		} else {
			return n.Parent.ParentHeight()
		}
	}
}

func (n *Node) xFracs() float32 {
	var fracs float32 = 0
	for _, child := range n.Children {
		if child.Size.W.Type == DT_FR {
			fracs += child.Size.W.Amount
		}

		if child.Size.Padding.Right.Type == DT_FR {
			fracs += child.Size.Padding.Right.Amount
		}

		if child.Size.Padding.Left.Type == DT_FR {
			fracs += child.Size.Padding.Left.Amount
		}
	}
	return fracs
}

func (n *Node) yFracs() float32 {
	var fracs float32 = 0
	for _, child := range n.Children {
		if child.Size.H.Type == DT_FR {
			fracs += child.Size.H.Amount
		}

		if child.Size.Padding.Top.Type == DT_FR {
			fracs += child.Size.Padding.Top.Amount
		}

		if child.Size.Padding.Bottom.Type == DT_FR {
			fracs += child.Size.Padding.Bottom.Amount
		}
	}
	return fracs
}

func (n *Node) FracWidth() float32 {
	var psw float32 = 0
	var xfracs float32 = 0

	if n.Parent != nil {
		psw = n.Parent.StaticWidth()
		xfracs = n.Parent.xFracs()
	}

	usable_width := n.ParentWidth() - psw
	var fracw float32 = 0
	if xfracs > 0 { fracw = usable_width / xfracs }
	return fracw
}

func (n *Node) FracHeight() float32 {
	var psh float32 = 0
	var yfracs float32 = 0

	if n.Parent != nil {
		psh = n.Parent.StaticHeight()
		yfracs = n.Parent.yFracs()
	}

	usable_height := n.ParentHeight() - psh
	var frach float32 = 0
	if yfracs > 0 { frach = usable_height / yfracs }
	return frach
}

func (n *Node) ResolveSize() {
	if !n.Size.IsResolved { n.Size.Resolved(n) }
	for _, child := range n.Children { child.ResolveSize() }
}

func (n *Node) ResolvePos() {
	if (n.Type == "root") {
		n.Pos.X = 0
		n.Pos.Y = 0
	}

	var offset float32 = 0
	xmul := Btof(n.Layout == LT_HORIZONTAL)
	ymul := Btof(n.Layout == LT_VERTICAL)

	for _, child := range n.Children {
		child.Pos.X = n.Pos.X + offset * xmul
		child.Pos.Y = n.Pos.Y + offset * ymul
		offset += child.Size.Real.X * xmul
		offset += child.Size.Real.Y * ymul
		child.ResolvePos()
	}
}

func (n *Node) UpdateChildren() {
	for _, child := range n.Children {
		child.UpdateFn(child)
	}
}

func (n *Node) HasMouse() bool {
	mx, my := Platform.MousePos.X, Platform.MousePos.Y
	x1, y1 := n.Pos.X, n.Pos.Y
	x2, y2 := x1 + n.Size.Real.X, y1 + n.Size.Real.Y

	return (
		mx >= x1 &&
		mx <= x2 &&
		my >= y1 &&
		my <= y2)
}

func (n *Node) DrawPos() (float32, float32) {
	return n.Pos.X + n.Size.Padding.Left.Real,
		n.Pos.Y + n.Size.Padding.Top.Real
}

type INPUT_MODE uint8

const (
	IM_MOUSE INPUT_MODE = iota
	IM_KBD
)

var UI = ui_state{}

type ui_state struct {
	Mode INPUT_MODE

	Root *Node
	Current *Node

	Active string
	Hot    string

	ActiveChanged bool
	HotChanged    bool
}

func (ui *ui_state) Reset() {
	ui.ActiveChanged = false
	ui.HotChanged = false
}

// Pushes a node on the UI stack.
func (UI *ui_state) Push(t string) *Node {
	n := GetNode(t, UI.Current)
	UI.Current = n
	return n
}

// Pops a node off the UI stack.
func (UI *ui_state) Pop(n *Node) *Node {
	if UI.Current == UI.Root {
		panic("Unbalanced UI stack!")
	}
	UI.Current = n.Parent
	return n
}

func (ui *ui_state) SetActive(node *Node, force bool) {
	if ui.ActiveChanged && !force { return }
	if node == nil {
		ui.Active = ""
	} else {
		ui.Active = node.UID
	}
	ui.ActiveChanged = true
}

func (ui *ui_state) SetHot(node *Node, force bool) {
	if ui.HotChanged && !force { return }
	if node == nil {
		ui.Active = ""
	} else {
		ui.Active = node.UID
	}
	ui.ActiveChanged = true
}

func (ui *ui_state) Begin() {
	ui.Reset()
	ui.Root = GetNode("root", nil)
	ui.Root.UpdateFn = rootUpdateFn;
	ui.Current = ui.Root
	ui.Root.Size.W = px(WindowWidth())
	ui.Root.Size.H = px(WindowHeight())
}

func (ui *ui_state) End() {
	if ui.Current != ui.Root {
		panic("Unbalanced UI stack!")
	}
	ui.Root.ResolveSize()
	ui.Root.ResolvePos()

	if Platform.MouseDelta.ManhattanLength() > 5 { ui.Mode = IM_MOUSE }
	if Platform.AnyKeyPressed { ui.Mode = IM_KBD }

	ui.Root.UpdateFn(ui.Root)
}

func (ui *ui_state) Render() {
	ui.Root.Render()
}

func WithNewNode(t string, fn func(*Node)) {
	n := UI.Push(t)
	fn(n)
	UI.Pop(n)
}

func WithNode(n *Node, fn func(*Node)) {
	fn(n)
	UI.Pop(n)
}

func Row() *Node {
	n := UI.Push("row")
	n.Layout = LT_HORIZONTAL
	return n
}

func Column() *Node {
	n := UI.Push("column")
	n.Layout = LT_VERTICAL
	return n
}

func textRenderFn(n *Node) {
	t, ok := n.Get("text", "").(string)
	if !ok { t = "" }

	// r,g,b,a,_ := Platform.Renderer.GetDrawColor()
	// defer Platform.Renderer.SetDrawColor(r, g, b, a)
	// Platform.Renderer.SetDrawColor(0, 0, 0, 255)

	surf, _ := Platform.Font.RenderUTF8Blended(t, sdl.Color{255, 255, 255, 255})
	defer surf.Free()

	tex, _ := Platform.Renderer.CreateTextureFromSurface(surf)
	defer tex.Destroy()

	x, y := n.DrawPos()

	Platform.Renderer.CopyF(tex, nil, &sdl.FRect{
		x, y, float32(surf.W), float32(surf.H),
	})
}

func Text(text string) *Node {
	n := UI.Push("text")
	defer UI.Pop(n)

	n.Set("text", text)
	n.Size.W = Auto()
	n.Size.H = Auto()
	n.Size.Padding = Padding{px(8), px(8), px(8), px(8)}

	n.RenderFn = textRenderFn
	return n
}

type BUTTON_STATE uint8
const (
	BS_UP BUTTON_STATE = iota
	BS_PRESSED
	BS_DOWN
	BS_RELEASED
)

type platform struct {
	Window *sdl.Window
	Renderer *sdl.Renderer
	Font *ttf.Font
	Mouse map[uint8]BUTTON_STATE
	MousePos V2
	MouseDelta V2
	Keyboard map[uint32]BUTTON_STATE
	AnyKeyPressed bool
}

var Platform platform

const (
	fontPath = "assets/test.ttf"
	fontSize = 16
)

func (p *platform) Init() {
	window, renderer, err := sdl.CreateWindowAndRenderer(800, 600, sdl.WINDOW_SHOWN)
	die(err)
	p.Window = window
	p.Renderer = renderer
	font, err := ttf.OpenFont(fontPath, fontSize)
	die(err)
	p.Font = font
	p.Mouse = make(map[uint8]BUTTON_STATE)
	p.Keyboard = make(map[uint32]BUTTON_STATE)
	p.MousePos.X = -1
	p.MousePos.Y = -1
}

func (p *platform) Cleanup() {
	p.Window.Destroy()
	p.Font.Close()
}

func DrawRectOutlined(x, y, w, h float32) {
	Platform.Renderer.DrawRectF(&sdl.FRect{x, y, w, h})
}

func DrawRectFilled(x, y, w, h float32) {
	Platform.Renderer.FillRectF(&sdl.FRect{x, y, w, h})
}

func KeyboardPressed(key uint32) bool {
	state, ok := Platform.Keyboard[key]
	if !ok { return false }
	return state == BS_PRESSED
}

func KeyboardReleased(key uint32) bool {
	state, ok := Platform.Keyboard[key]
	if !ok { return false }
	return state == BS_RELEASED
}

func MousePressed(btn uint8) bool {
	state, ok := Platform.Mouse[btn]
	if !ok { return false }
	return state == BS_PRESSED
}

func MouseReleased(btn uint8) bool {
	state, ok := Platform.Mouse[btn]
	if !ok { return false }
	return state == BS_RELEASED
}

func WindowWidth() float32 {
	w, _ := Platform.Window.GetSize()
	return float32(w)
}

func WindowHeight() float32 {
	_, h := Platform.Window.GetSize()
	return float32(h)
}

func TextWidth(text string) float32 {
	w, _, err := Platform.Font.SizeUTF8(text)
	die(err)
	return float32(w)
}

func TextHeightWrapped(text string, _ float32) float32 {
	_, h, err := Platform.Font.SizeUTF8(text)
	die(err)
	return float32(h)
}

func ButtonMapUpdate[K comparable](m map[K]BUTTON_STATE) {
	for btn, state := range m {
		if state == BS_PRESSED {
			m[btn] = BS_DOWN
		} else if state == BS_RELEASED {
			m[btn] = BS_UP
		}
	}
}

func Max[A constraints.Ordered](a, b A) A {
	if a > b { return a } else { return b }
}

func Min[A constraints.Ordered](a, b A) A {
	if a < b { return a } else { return b }
}

func Abs[A constraints.Float | constraints.Integer](a A) A {
	if a < 0 { return -a } else { return a }
}

func Btof(b bool) float32 {
	if b { return 1 } else { return 0 }
}

func die(err error) {
	if err != nil { panic(err) }
}

func main() {
	runtime.LockOSThread()

	err := sdl.Init(sdl.INIT_EVERYTHING)
	die(err)
	defer sdl.Quit()

	err = ttf.Init()
	die(err)
	defer ttf.Quit()

	Platform.Init()

	// Platform.Surface.FillRect(nil, 0)
	// rect := sdl.Rect{0, 0, 200, 200}
	// color := sdl.Color{255, 0, 255, 255}
	// pixel := sdl.MapRGBA(Platform.Surface.Format, color.R, color.G, color.B, color.A)
	// Platform.Surface.FillRect(&rect, pixel)
	// Platform.Window.UpdateSurface()

	running := true
	for running {
		if KeyboardPressed(sdl.SCANCODE_Q) {
			running = false
		}

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
			WithNode(Row(), func(n *Node) {
				Text("Hello, world!")

				WithNode(Column(), func(n *Node) {
					n.Flags.Focusable = true
					Text("This")
					Text("stacks")
					Text("woohoo")
				})

				Text("Back to row")
				Text("Some more")
			})
		} ; UI.End()

		UI.Render()

		Platform.Renderer.Present()
	}
}
