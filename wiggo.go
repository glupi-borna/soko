package main

import (
	"runtime"
	"flag"
	"os"
	"runtime/pprof"
	"strconv"
	"math"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	lru "github.com/glupi-borna/wiggo/internal/lru"
	. "github.com/glupi-borna/wiggo/internal/utils"
)

type DIM_TYPE uint8
const (
	DT_AUTO DIM_TYPE = iota
	DT_PX ; DT_FR ; DT_TEXT ; DT_CHILDREN ; DT_SKIP
)

type Dimension struct {
	Type       DIM_TYPE
	Amount     float32
}

func (d *Dimension) String() string {
	switch d.Type {
		case DT_CHILDREN: return "children"
		case DT_TEXT: return "text"
		case DT_FR: return FloatStr(d.Amount) + "fr"
		case DT_PX: return FloatStr(d.Amount) + "px"
		default: panic("Unknown Dimension Type: " + strconv.Itoa(int(d.Type)))
	}
}

func px(amount float32) Dimension { return Dimension{Type: DT_PX, Amount: amount} }
func fr(amount float32) Dimension { return Dimension{Type: DT_FR, Amount: amount} }
func child_sum() Dimension { return Dimension{ Type: DT_CHILDREN }}
func fit_text() Dimension { return Dimension{ Type: DT_TEXT }}
func Auto() Dimension { return Dimension{ Type: DT_AUTO } }

type Size struct { W, H Dimension }

func (s *Size) String() string {
	x := s.W.String()
	y := s.H.String()
	return "Size{ W: " + x + ", H: " + y + " }"
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
	CornerRadius StyleVariant[float32]
	Padding      Padding
}

var DefaultStyle = Style{
	Foreground: StyleVariant[sdl.Color]{
		Normal: sdl.Color{255, 255, 255, 255},
		Hovered: sdl.Color{255, 0, 0, 255},
	},
	Background: StyleVariant[sdl.Color]{
		Normal: sdl.Color{},
		Hovered: sdl.Color{255, 255, 0, 100},
	},
	CornerRadius: StyleVariant[float32]{
		Normal: 3,
		Hovered: 3,
	},
}

type NodeData map[string] any

var UI_Data = make(map[string]NodeData)

func uiGet[K any](n *Node, key string, dflt K) (out K) {
	data, ok := UI_Data[n.UID]

	// If node data doesn't exist
	if !ok {
		data = make(NodeData)
		UI_Data[n.UID] = data
	}

	// If key does not exist in node data
	val, ok := data[key]
	if !ok {
		data[key] = dflt
		return dflt
	}

	// If key is wrong type
	out, ok = val.(K)
	if !ok {
		data[key] = dflt
		return dflt
	}

	return out
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

var animState = map[string]float32 {}

// Smoothly animates a value
func Animate(val float32, id string) float32 {
	old, ok := animState[id]
	if !ok {
		animState[id] = val
		return val
	}
	new := old + 0.5 * (val-old)
	animState[id] = new
	return new
}

type NodeFlags struct {
	Focusable bool
}

type Node struct {
	UID       string // Unique ID of this Node
	Type      string // Equivalent to tag name in HTML
	Layout    LAYOUT_TYPE // Children positioning (horizontal/vertical)
	Flags     NodeFlags
	Parent    *Node // Parent of this node - null if the node is the root node, or is detached.
	Children  []*Node
	Style     *Style
	Padding   Padding

	// Semantic size
	Size Size

	// Calculated position
	Pos       V2
	// Calculated size
	RealSize  V2

	// Is RealSize.X calculated for this frame
	IsWidthResolved bool
	// Is RealSize.Y calculated for this frame
	IsHeightResolved bool

	// Called after the position and size of the Node
	// have been resolved.
	RenderFn func(*Node)

	// Called after the position and size of the Node
	// have been resolved.
	UpdateFn func(*Node)
}

// Currently does the same as MakeNode, but
// could be used for object pooling later on.
func GetNode(t string, parent *Node) *Node {
	n := MakeNode(t, parent)
	return n
}

func buildNodeUID(n *Node) string {
	if n.Parent != nil {
		return n.Parent.UID + "." + n.Type + strconv.Itoa(n.Parent.CountChildrenOfType(n.Type))
	} else {
		return n.Type
	}
}

func MakeNode(t string, parent *Node) *Node {
	n := Node{
		Type: t,
		Children: make([]*Node, 0),
		Parent: parent,
		RenderFn: defaultRenderFn,
		UpdateFn: defaultUpdateFn,
		Padding: Padding1(8),
	}

	if n.Parent != nil {
		n.Parent.Children = append(n.Parent.Children, &n)
	}

	n.UID = buildNodeUID(&n)

	return &n
}

func (n *Node) Set(key string, val any) bool {
	return uiDataSet(n, key, val)
}

func SetColor(c sdl.Color) {
	Platform.Renderer.SetDrawColor(c.R, c.G, c.B, c.A)
}

func defaultRenderFn(n *Node) {
	s := n.GetStyle()

	if UI.Active == n.UID {
		SetColor(s.Background.Hovered)
		if s.CornerRadius.Hovered == 0 {
			DrawRectFilled(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
		} else {
			DrawRoundRectFilled(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y, s.CornerRadius.Hovered)
		}
	} else {
		SetColor(s.Background.Normal)
		if s.CornerRadius.Hovered == 0 {
			DrawRectOutlined(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
		} else {
			DrawRoundRectOutlined(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y, s.CornerRadius.Normal)
		}
	}
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
	// println(n.UID)
	// println("  pos ", n.Pos.String())
	// println("  size", n.RealSize.String())
	// println("  sem ", n.Size.String())
	// println("  res ", n.IsWidthResolved, n.IsHeightResolved)

	n.RenderFn(n)
	for _, child := range n.Children { child.Render() }
}

func (n *Node) GetStyle() *Style {
	if n.Style != nil { return n.Style }
	if n.Parent != nil { return n.Parent.GetStyle() }
	return &DefaultStyle
}

func (n *Node) xFracs() float32 {
	if n.Layout == LT_VERTICAL { return 1 }

	var count float32 = 0
	for _, child := range n.Children {
		cw := child.Size.W
		if cw.Type == DT_FR { count += cw.Amount }
	}
	return count
}

func (n *Node) yFracs() float32 {
	if n.Layout == LT_HORIZONTAL { return 1 }

	var count float32 = 0
	for _, child := range n.Children {
		ch := child.Size.H
		if ch.Type == DT_FR { count += ch.Amount }
	}
	return count
}

var autoMap = map[string]Dimension{
	"text": fit_text(),
}

func ResolveAuto(n *Node) {
	if n.Size.W.Type == DT_AUTO {
		dim, ok := autoMap[n.Type]
		if !ok {
			n.Size.W = child_sum()
		} else {
			n.Size.W = dim
		}
	}

	if n.Size.H.Type == DT_AUTO {
		dim, ok := autoMap[n.Type]
		if !ok {
			n.Size.H = child_sum()
		} else {
			n.Size.H = dim
		}
	}
}

// Resolves standalone sizes
func (n *Node) ResolveStandalone() {
	ResolveAuto(n)

	if n.Size.W.Type == DT_PX {
		n.RealSize.X = n.Size.W.Amount
		n.IsWidthResolved = true

	} else if n.Size.W.Type == DT_TEXT {
		t := uiGet(n, "text", "")
		n.RealSize.X = TextWidth(t) + n.Padding.Left + n.Padding.Right
		n.IsWidthResolved = true
	}

	if n.Size.H.Type == DT_PX {
		n.RealSize.Y = n.Size.H.Amount
		n.IsHeightResolved = true

	} else if n.Size.H.Type == DT_TEXT {
		t := uiGet(n, "text", "")
		n.RealSize.Y = TextHeight(t) + n.Padding.Top + n.Padding.Bottom
		n.IsHeightResolved = true
	}

	for _, child := range n.Children {
		child.ResolveStandalone()
	}
}

// Resolves upwards-dependent sizes
func (n *Node) ResolveUpwards() {
	if !n.IsWidthResolved && n.Size.W.Type == DT_FR {
		pw := n.ParentRemainingWidth()
		fracs := n.Parent.xFracs()
		fracw := pw / fracs

		for _, child := range n.Parent.Children {
			if child.Size.W.Type == DT_FR {
				child.RealSize.X = fracw * child.Size.W.Amount
				child.IsWidthResolved = true
			}
		}
	}

	if !n.IsHeightResolved && n.Size.H.Type == DT_FR {
		ph := n.ParentRemainingHeight()
		fracs := n.Parent.yFracs()
		frach := ph / fracs

		for _, child := range n.Parent.Children {
			if child.Size.H.Type == DT_FR {
				child.RealSize.Y = frach * child.Size.H.Amount
				child.IsHeightResolved = true
			}
		}
	}

	for _, child := range n.Children {
		child.ResolveUpwards()
	}
}

// Resolves downwards-dependent sizes
func (n *Node) ResolveDownwards() {
	for _, child := range n.Children {
		child.ResolveDownwards()
	}

	if n.Size.W.Type == DT_CHILDREN {
		switch n.Layout {
		case LT_HORIZONTAL:
			n.RealSize.X = n.ChildSum(nodeRealX) + n.Padding.Left + n.Padding.Right
			n.IsWidthResolved = true

		case LT_VERTICAL:
			n.RealSize.X = n.ChildMax(nodeRealX) + n.Padding.Left + n.Padding.Right
			n.IsWidthResolved = true
		}
	}

	if n.Size.H.Type == DT_CHILDREN {
		switch n.Layout {
		case LT_VERTICAL:
			n.RealSize.Y = n.ChildSum(nodeRealY) + n.Padding.Top + n.Padding.Bottom
			n.IsHeightResolved = true

		case LT_HORIZONTAL:
			n.RealSize.Y = n.ChildMax(nodeRealY) + n.Padding.Top + n.Padding.Bottom
			n.IsHeightResolved = true
		}
	}
}

func (n *Node) ResolveViolations() {
	var w, h float32

	switch n.Layout {
	case LT_VERTICAL:
		w = n.ChildMax(nodeRealX)
		h = n.ChildSum(nodeRealY)

	case LT_HORIZONTAL:
		w = n.ChildSum(nodeRealX)
		h = n.ChildMax(nodeRealY)
	}

	if w > n.RealSize.X {
		fracs := n.xFracs()
		if fracs > 0 {
			fracdec := (w - n.RealSize.X) / fracs
			for _, child := range n.Children {
				if child.Size.W.Type == DT_FR {
					child.RealSize.X -= fracdec * child.Size.W.Amount
				}
			}
		}
	}

	if h > n.RealSize.Y {
		fracs := n.yFracs()
		if fracs > 0 {
			fracdec := (h - n.RealSize.Y) / fracs
			for _, child := range n.Children {
				if child.Size.H.Type == DT_FR {
					child.RealSize.Y -= fracdec * child.Size.H.Amount
				}
			}
		}
	}

	for _, child := range n.Children { child.ResolveViolations() }
}

func nodeRealX(n *Node) float32 {
	if n.IsWidthResolved {
		return n.RealSize.X
	} else {
		return 0
	}
}

func nodeRealY(n *Node) float32 {
	if n.IsHeightResolved {
		return n.RealSize.Y
	} else {
		return 0
	}
}

func (n *Node) ChildSum(fn func(*Node)float32) float32 {
	var sum float32 = 0
	for _, child := range n.Children {
		sum += fn(child)
	}
	return sum
}

func (n *Node) ChildMax(fn func(*Node)float32) (val float32) {
	for _, child := range n.Children {
		cv := fn(child)
		if cv > val { val = cv }
	}
	return
}

func (n *Node) ParentWidth() float32 {
	if n.Parent == nil { return WindowWidth() }
	if n.Parent.IsWidthResolved { return n.Parent.RealSize.X }
	return n.Parent.ParentWidth()
}

func (n *Node) ParentHeight() float32 {
	if n.Parent == nil { return WindowHeight() }
	if n.Parent.IsHeightResolved { return n.Parent.RealSize.X }
	return n.Parent.ParentWidth()
}

func (n *Node) ParentRemainingWidth() float32 {
	if n.Parent == nil { return WindowWidth() }
	w := n.ParentWidth() - n.Parent.Padding.Left - n.Parent.Padding.Right
	for _, child := range n.Parent.Children {
		if child.IsWidthResolved {
			w -= child.RealSize.X
		}
	}
	return Max(w, 0)
}

func (n *Node) ParentRemainingHeight() float32 {
	if n.Parent == nil { return WindowHeight() }
	h := n.ParentHeight() - n.Parent.Padding.Top - n.Parent.Padding.Bottom
	for _, child := range n.Parent.Children {
		if child.IsHeightResolved {
			h -= child.RealSize.Y
		}
	}
	return Max(h, 0)
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
		child.Pos.X = n.Pos.X + n.Padding.Left + offset * xmul
		child.Pos.Y = n.Pos.Y + n.Padding.Top + offset * ymul
		offset += child.RealSize.X * xmul
		offset += child.RealSize.Y * ymul
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
	x2, y2 := x1 + n.RealSize.X, y1 + n.RealSize.Y

	return (
		mx >= x1 &&
		mx <= x2 &&
		my >= y1 &&
		my <= y2)
}

func (n *Node) IsChildOfUID(uid string) bool {
	p := n.Parent
	for p != nil {
		if p.UID == uid { return true }
		p = p.Parent
	}
	return false
}

func (n *Node) Index() int {
	if n.Parent == nil { return -1 }
	for idx, child := range n.Parent.Children {
		if child == n { return idx }
	}
	return -1
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
func (UI *ui_state) Push(t string) (n *Node) {
	n = GetNode(t, UI.Current)
	UI.Current = n
	return
}

// Pops a node off the UI stack.
func (UI *ui_state) Pop(n *Node) *Node {
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
		ui.Hot = ""
	} else {
		ui.Hot = node.UID
	}
	ui.HotChanged = true
}

func (ui *ui_state) Begin() {
	ui.Reset()
	ui.Root = GetNode("root", nil)
	ui.Root.UpdateFn = rootUpdateFn
	ui.Root.RenderFn = invisibleRenderFn
	ui.Root.Size.W = px(WindowWidth())
	ui.Root.Size.H = px(WindowHeight())
	ui.Current = ui.Root
}

func (ui *ui_state) End() {
	if ui.Current != ui.Root {
		panic("Unbalanced UI stack!")
	}

	ui.Root.Style = &DefaultStyle
	ui.Root.ResolveStandalone()
	ui.Root.ResolveUpwards()
	ui.Root.ResolveDownwards()
	ui.Root.ResolveViolations()
	ui.Root.ResolvePos()

	if Platform.MouseDelta.ManhattanLength() > 5 { ui.Mode = IM_MOUSE }
	if Platform.AnyKeyPressed { ui.Mode = IM_KBD }

	ui.Root.UpdateFn(ui.Root)
}

func (ui *ui_state) Render() {
	ui.Root.Render()
}

func WithNode(n *Node, fn func(*Node)) *Node {
	fn(n)
	UI.Pop(n)
	return n
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

func textRenderFn(n *Node) {
	t := uiGet(n, "text", "")
	s := n.GetStyle()

	var c sdl.Color
	if UI.Active == n.UID || n.IsChildOfUID(UI.Active) {
		c = s.Foreground.Hovered
	} else {
		c = s.Foreground.Normal
	}

	tex := GetTextTexture(Platform.Font, t, c)
	m := TextMetrics(t)

	Platform.Renderer.CopyF(tex, nil, &sdl.FRect{
		float32(math.Round(float64(n.Pos.X + n.Padding.Left))),
		float32(math.Round(float64(n.Pos.Y + n.Padding.Top))),
		m.X, m.Y,
	})
}

func Margin(dim Dimension, inner *Node) *Node {
	margin := GetNode("margin-h", nil)
	margin.RenderFn = invisibleRenderFn
	margin.Layout = LT_HORIZONTAL
	margin.Parent = inner.Parent
	idx := inner.Index()

	cur := UI.Current
	defer func(){ UI.Current = cur }()
	UI.Current = margin

	Invisible(dim)
	WithNode(UI.Push("margin-v"), func(r *Node) {
		r.Layout = LT_VERTICAL
		r.RenderFn = invisibleRenderFn

		Invisible(dim)
		r.Children = append(r.Children, inner)
		inner.Parent = r
		Invisible(dim)
	})
	Invisible(dim)

	margin.Parent.Children[idx] = margin
	margin.UID = buildNodeUID(margin)

	old_uid := inner.UID
	inner.UID = buildNodeUID(inner)

	UI_Data[inner.UID] = UI_Data[old_uid]
	delete(UI_Data, old_uid)

	return margin
}

func invisibleRenderFn(*Node) {}

func Invisible(dim Dimension) *Node {
	n := UI.Push("invisible")
	defer UI.Pop(n)
	n.Size.W = dim
	n.Size.H = dim
	n.RenderFn = invisibleRenderFn
	return n
}

func Text(text string) *Node {
	n := UI.Push("text")
	defer UI.Pop(n)

	n.Set("text", text)
	n.Size.W = fit_text()
	n.Size.H = fit_text()
	n.Padding = Padding{}

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
}

var Platform platform

const (
	fontPath = "assets/test.ttf"
	fontSize = 16
)

func (p *platform) Init() {
	var window_flags uint32 =
		sdl.WINDOW_SHOWN |
		sdl.WINDOW_BORDERLESS |
		sdl.WINDOW_UTILITY |
		sdl.WINDOW_ALWAYS_ON_TOP

	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4)
	window, err := sdl.CreateWindow(
		"", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, window_flags)
	Die(err)
	p.Window = window

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_PRESENTVSYNC | sdl.RENDERER_ACCELERATED)
	Die(err)
	p.Renderer = renderer

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

func DrawRectOutlined(x, y, w, h float32) {
	Platform.Renderer.DrawRectF(&sdl.FRect{x, y, w, h})
}

func DrawRectFilled(x, y, w, h float32) {
	Platform.Renderer.FillRectF(&sdl.FRect{x, y, w, h})
}

func ArcPoints(x, y, r, startAngle, endAngle float32) []sdl.FPoint {
	CORNER_POINTS := Max(5, int(r)/5)

	points := make([]sdl.FPoint, CORNER_POINTS)
	angleRange := float64(Abs(startAngle - endAngle))
	start := float64(startAngle)
	rad := float64(r)

	imax := float64(CORNER_POINTS-1)

	for i:= 0; i<CORNER_POINTS; i++ {
		frac := float64(i)/imax
		angle := frac*angleRange + start
		xoff := float32(math.Cos(angle) * rad)
		yoff := float32(math.Sin(angle) * rad)
		points[i].X = x+xoff
		points[i].Y = y+yoff
	}

	return points
}

func RoundRectPoints(x, y, w, h, r float32) []sdl.FPoint {
	r = Clamp(r, 0, Min(w/2, h/2))

	CORNER_POINTS := Max(5, int(r)/5)
	TOTAL_POINTS := CORNER_POINTS * 4

	x1 := x+r
	y1 := y+r
	x2 := x+w-r
	y2 := y+h-r

	points := make([]sdl.FPoint, 0, TOTAL_POINTS)
	const pihalf = math.Pi/2
	points = append(points, ArcPoints(x1, y1, r, pihalf*2, pihalf*3)...)
	points = append(points, ArcPoints(x2, y1, r, pihalf*3, pihalf*4)...)
	points = append(points, ArcPoints(x2, y2, r, 0,        pihalf)...)
	points = append(points, ArcPoints(x1, y2, r, pihalf,   pihalf*2)...)

	return points
}

func DrawPoints(pts []sdl.FPoint) {
	Platform.Renderer.DrawPointsF(pts)
}

func DrawRoundRectOutlined(x, y, w, h, r float32) {
	points := RoundRectPoints(x, y, w, h, r)
	Platform.Renderer.DrawLinesF(points)
	l := len(points)-1
	Platform.Renderer.DrawLineF(points[0].X, points[0].Y, points[l].X, points[l].Y)
}

func DrawRoundRectFilled(x, y, w, h, r float32) {
	R,G,B,A,_ := Platform.Renderer.GetDrawColor()
	c := sdl.Color{R,G,B,A}

	points := RoundRectPoints(x, y, w, h, r)
	point_count := len(points)

	verts := make([]sdl.Vertex, point_count*3)
	for i := range verts { verts[i] = sdl.Vertex{Color: c} }

	midx := x+w/2
	midy := y+h/2

	for i:=0 ; i<point_count ; i++ {
		idx := i*3
		verts[idx].Position.X = points[i].X
		verts[idx].Position.Y = points[i].Y

		verts[idx+1].Position.X = midx
		verts[idx+1].Position.Y = midy

		next_i := (i+1)%point_count
		verts[idx+2].Position.X = points[next_i].X
		verts[idx+2].Position.Y = points[next_i].Y
	}

	Platform.Renderer.RenderGeometry(nil, verts, nil)
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
	return TextMetrics(text).X
}

func TextHeight(text string) float32 {
	return TextMetrics(text).Y
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

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var timeout = flag.Uint64("timeout", 0, "stop running after this number of milliseconds (0 = no timeout)")
var vsync = flag.Bool("vsync", true, "enable/disable vsync")
var override_redirect = flag.Bool("override-redirect", true, "enable/disable override-redirect (X11)")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		Die(err)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	runtime.LockOSThread()

	err := sdl.Init(sdl.INIT_EVERYTHING)
	Die(err)
	defer sdl.Quit()

	err = ttf.Init()
	Die(err)
	defer ttf.Quit()

	Platform.Init()

	// Platform.Surface.FillRect(nil, 0)
	// rect := sdl.Rect{0, 0, 200, 200}
	// color := sdl.Color{255, 0, 255, 255}
	// pixel := sdl.MapRGBA(Platform.Surface.Format, color.R, color.G, color.B, color.A)
	// Platform.Surface.FillRect(&rect, pixel)
	// Platform.Window.UpdateSurface()

	running := true

	LastFrameStart := uint64(0)
	FrameTime := uint64(0)

	for running {
		FrameStart := sdl.GetTicks64()
		if *timeout > 0 && FrameStart > *timeout { running = false }
		if KeyboardPressed(sdl.SCANCODE_Q) { running = false }

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

		ButtonStyle := DefaultStyle
		ButtonStyle.CornerRadius.Normal = 3
		ButtonStyle.CornerRadius.Hovered = 3
		ButtonStyle.Background.Normal = sdl.Color{255, 255, 255, 255}
		ButtonStyle.Background.Hovered = sdl.Color{0, 0, 255, 255}

		UI.Begin(); {
			Text(FloatStr(float64(FrameTime)/1000) + "ms")
			WithNode(Row(), func(n *Node) {
				n.Size.W = fr(1)
				Text("Hello, world!")

				WithNode(Column(), func(n *Node) {
					n.Padding = Padding2(8, -8)
					WithNode(Column(), func(n *Node) {
						n.Flags.Focusable = true
						n.Style = &ButtonStyle
						n.Size.W = px(100)
						n.Size.H = child_sum()
						Text("Button")
					})
					Text("Under button").Padding = Padding1(16)
				})

				WithNode(Column(), func(n *Node) {
					n.Padding.Top = 0
					n.Size.H = child_sum()
					Text("Stacked")
					Text("text")
				})

				Seconds := FrameStart / 1000
				t1 := float32(Seconds % 2 + 1)
				t2 := float32(Seconds % 3 + 1)
				Text(FloatStr(t1)).Size.W = fr(Animate(t1, "1"))
				Text(FloatStr(t2)).Size.W = fr(Animate(t2, "2"))
				Text("Some more")
			})
		} ; UI.End()

		UI.Render()
		Platform.Renderer.Present()

		FrameTime = FrameStart - LastFrameStart
		LastFrameStart = FrameStart
	}
}

func PrintTree(n *Node, indent string) {
	child_indent := indent + "  "
	println(indent + n.Type)
	for _, child := range n.Children {
		PrintTree(child, child_indent)
	}
}
