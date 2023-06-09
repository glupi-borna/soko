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

type Dimension struct {
	Type       DIM_TYPE
	// Logical size (px, fr, etc)
	Amount     float32
	// Calculated size (px)
	Real       float32
	IsResolved bool
}

func px(amount float32) Dimension { return Dimension{Type: DT_PX, Amount: amount} }
func fr(amount float32) Dimension { return Dimension{Type: DT_FR, Amount: amount} }
var Auto = Dimension{ Type: DT_AUTO }

type Size struct { X, Y Dimension }

type LAYOUT_TYPE uint8
const (
	LT_VERTICAL LAYOUT_TYPE = iota
	LT_HORIZONTAL
)

type V2 struct { X, Y float32 }

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
	return &n
}

func MakeNode(t string, parent *Node) Node {
	n := Node{
		Type: t,
		Children: make([]*Node, 0),
		Parent: parent,
		RenderFn: defaultRenderFn,
		UpdateFn: defaultUpdateFn,
	}

	n.UID = buildNodeUID(&n)

	return n
}

func (n *Node) Get(key string, dflt any) any {
	return uiDataGet(n, key, dflt)
}

func (n *Node) Set(key string, val any) bool {
	return uiDataSet(n, key, val)
}

func buildNodeUID(n *Node) string {
	if n.Parent != nil {
		return n.Parent.UID + "." + n.Type + strconv.Itoa(n.CountChildrenOfType(n.Type))
	}
	return n.Type
}

func defaultRenderFn(n *Node) {
	if UI.Active == n.UID {
		DrawRectFilled(n.Pos.X, n.Pos.Y, n.Size.X.Real, n.Size.Y.Real)
	}
	DrawRectOutlined(n.Pos.X, n.Pos.Y, n.Size.X.Real, n.Size.Y.Real)
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

func (n *Node) ParentWidth() float32 {
	if n.Parent != nil {
		return WindowWidth()
	} else {
		if n.Parent.Size.X.IsResolved {
			return n.Parent.Size.X.Real
		} else {
			return -1
		}
	}
}

func (n *Node) ParentHeight() float32 {
	if n.Parent != nil {
		return WindowHeight()
	} else {
		if n.Parent.Size.Y.IsResolved {
			return n.Parent.Size.Y.Real
		} else {
			return -1
		}
	}
}

func (n *Node) ResolveSize() {
	if !n.Size.X.IsResolved { n.ResolveWidth(n.ParentWidth()) }
	if !n.Size.Y.IsResolved { n.ResolveHeight(n.ParentHeight()) }
	for _, child := range n.Children { child.ResolveSize() }
}

func (n *Node) setResolvedWidth(width float32) float32 {
	n.Size.X.IsResolved = true
	n.Size.X.Real = width
	return width
}

func (n *Node) setResolvedHeight(height float32) float32 {
	n.Size.Y.IsResolved = true
	n.Size.Y.Real = height
	return height
}

func (n *Node) ResolveWidth(parent_width float32) float32 {
	if n.Size.X.IsResolved { return n.Size.X.Real }

	if n.Type == "root" { return n.setResolvedWidth(WindowWidth()) }

	switch (n.Size.X.Type) {

	case DT_PX:
		return n.setResolvedWidth(n.Size.X.Amount)

	case DT_FR:
		if parent_width == -1 {
			panic("FR: parent width not resolved yet")
		}

		var used_width float32 = 0
		var total_fracs float32 = 0

		for _, sibling := range n.Parent.Children {
			if sibling.Size.X.Type == DT_FR {
				total_fracs += sibling.Size.X.Amount
			} else {
				if !sibling.Size.X.IsResolved {
					sibling.ResolveWidth(parent_width)
				}
				used_width += sibling.Size.X.Real
			}
		}

		available_width := parent_width - used_width
		frac_width := available_width / total_fracs

		for _, sibling := range n.Parent.Children {
			if sibling.Size.X.Type == DT_FR {
				sibling.setResolvedWidth(sibling.Size.X.Amount * frac_width)
			}
		}

		return n.Size.X.Real

	case DT_AUTO:
		if n.Type == "text" {
			t, ok := n.Get("text", "").(string)
			if !ok { t = "" }
			return n.setResolvedWidth(Min(TextWidth(t), parent_width))

		} else {
			switch n.Layout {
			case LT_HORIZONTAL:
				var wsum float32 = 0
				for _, child := range n.Children {
					wsum += child.ResolveWidth(parent_width)
				}
				return n.setResolvedWidth(wsum)

			case LT_VERTICAL:
				var wmax float32 = 0
				for _, child := range n.Children {
					w := child.ResolveWidth(parent_width)
					if w > wmax { wmax = w }
				}
				return n.setResolvedWidth(wmax)

			default:
				panic("Unknown layout type: " + strconv.Itoa(int(n.Layout)))
			}
		}

	default:
		panic("Unknown size type: " + strconv.Itoa(int(n.Size.X.Type)))

	}
}

func (n *Node) ResolveHeight(parent_height float32) float32 {
	if n.Size.Y.IsResolved { return n.Size.Y.Real }

	if n.Type == "root" { return n.setResolvedHeight(WindowHeight()) }

	switch (n.Size.Y.Type) {

	case DT_PX:
		return n.setResolvedHeight(n.Size.Y.Amount)

	case DT_FR:
		if parent_height == -1 {
			panic("FR: parent width not resolved yet")
		}

		var used_height float32 = 0
		var total_fracs float32 = 0

		for _, sibling := range n.Parent.Children {
			if sibling.Size.Y.Type == DT_FR {
				total_fracs += sibling.Size.Y.Amount
			} else {
				if !sibling.Size.Y.IsResolved {
					sibling.ResolveHeight(parent_height)
				}
				used_height += sibling.Size.Y.Real
			}
		}

		available_width := parent_height - used_height
		frac_height := available_width / total_fracs

		for _, sibling := range n.Parent.Children {
			if sibling.Size.Y.Type == DT_FR {
				sibling.setResolvedHeight(sibling.Size.Y.Amount * frac_height)
			}
		}

		return n.Size.Y.Real

	case DT_AUTO:
		if n.Type == "text" {
			t, ok := n.Get("text", "").(string)
			if !ok { t = "" }
			var wwrap float32 = 0
			if n.Size.X.IsResolved {
				wwrap = n.Size.X.Real
			} else {
				wwrap = WindowWidth()
			}
			return n.setResolvedHeight(TextHeightWrapped(t, wwrap))

		} else {
			switch n.Layout {
			case LT_HORIZONTAL:
				var hmax float32 = 0
				for _, child := range n.Children {
					h := child.ResolveHeight(parent_height)
					if h > hmax { hmax = h }
				}
				return n.setResolvedHeight(hmax)

			case LT_VERTICAL:
				var hsum float32 = 0
				for _, child := range n.Children {
					hsum += child.ResolveHeight(parent_height)
				}
				return n.setResolvedHeight(hsum)

			default:
				panic("Unknown layout type: " + strconv.Itoa(int(n.Layout)))
			}
		}

	default:
		panic("Unknown size type: " + strconv.Itoa(int(n.Size.Y.Type)))

	}
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
		offset += child.Size.X.Real * xmul
		offset += child.Size.Y.Real * ymul
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
	x2, y2 := x1 + n.Size.X.Real, y1 + n.Size.Y.Real

	return (
		mx >= x1 &&
		mx <= x2 &&
		my >= y1 &&
		my <= y2)
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

func (ui *ui_state) Init() {
	root := GetNode("root", nil)
	ui.Current = root
	ui.Root = root
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
	Keyboard map[uint32]BUTTON_STATE
}

var Platform platform

const (
	fontPath = "assets/test.ttf"
	fontSize = 32
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

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					Platform.Keyboard[uint32(e.Keysym.Scancode)] = BS_PRESSED
				} else {
					Platform.Keyboard[uint32(e.Keysym.Scancode)] = BS_RELEASED
				}

			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			}
		}
	}
}
