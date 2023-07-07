package ui

import (
	"strconv"
	"github.com/veandco/go-sdl2/sdl"
	. "github.com/glupi-borna/soko/internal/utils"
	. "github.com/glupi-borna/soko/internal/platform"
	. "github.com/glupi-borna/soko/internal/debug"
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

func Px(amount float32) Dimension { return Dimension{Type: DT_PX, Amount: amount} }
func Fr(amount float32) Dimension { return Dimension{Type: DT_FR, Amount: amount} }
func ChildrenSize() Dimension { return Dimension{ Type: DT_CHILDREN }}
func FitText() Dimension { return Dimension{ Type: DT_TEXT }}
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

	// The UI State object associated with this node
	UI *UI_State
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
	Assert(CurrentUI != nil, "UI not initialized!")

	n := Node{
		Type: t,
		Children: make([]*Node, 0),
		Parent: parent,
		RenderFn: defaultRenderFn,
		UpdateFn: defaultUpdateFn,
		Padding: Padding1(8),
		UI: CurrentUI,
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

func defaultUpdateFn(n *Node) {
	n.UpdateChildren()

	if n.Flags.Focusable {
		if CurrentUI.Mode == IM_MOUSE {
			if n.HasMouse() {
				CurrentUI.SetActive(n, false)
				if Platform.MousePressed(sdl.BUTTON_LEFT) { CurrentUI.SetHot(n, false) }
				if Platform.MouseReleased(sdl.BUTTON_LEFT) { CurrentUI.SetHot(nil, false) }
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
}

func rootUpdateFn(n *Node) {
	n.UpdateChildren()

	if CurrentUI.Mode == IM_MOUSE {
		if n.HasMouse() {
			CurrentUI.SetActive(nil, false)
			if Platform.MousePressed(sdl.BUTTON_LEFT) { CurrentUI.SetHot(nil, false) }
			if Platform.MouseReleased(sdl.BUTTON_LEFT) { CurrentUI.SetHot(nil, false) }
		}
	}
}

func rootRenderFn(n *Node) {
	s := n.GetStyle()
	radius := s.CornerRadius.Normal
	c := n.Style.Background.Normal

	Platform.SetColor(c)
	Platform.DrawRectFilled(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
	Platform.ReshapeWindow(radius, false)
}

func (n *Node) Styled() *Style {
	if n.Style != nil { return n.Style }
	n.Style = DefaultStyle.Copy()
	return n.Style
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

var autoMap = map[string]Size{
	"text": { FitText(), FitText() },
	"row": { Fr(1), ChildrenSize() },
	"column": { ChildrenSize(), Fr(1) },
}

func ResolveAuto(n *Node) {
	if n.Size.W.Type == DT_AUTO || n.Size.H.Type == DT_AUTO {
		size, ok := autoMap[n.Type]

		if n.Size.W.Type == DT_AUTO {
			if !ok {
				n.Size.W = ChildrenSize()
			} else {
				n.Size.W = size.W
			}
		}

		if n.Size.H.Type == DT_AUTO {
			if !ok {
				n.Size.H = ChildrenSize()
			} else {
				n.Size.H = size.H
			}
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
		n.RealSize.X = Platform.TextWidth(t) + n.Padding.Left + n.Padding.Right
		n.IsWidthResolved = true
	}

	if n.Size.H.Type == DT_PX {
		n.RealSize.Y = n.Size.H.Amount
		n.IsHeightResolved = true

	} else if n.Size.H.Type == DT_TEXT {
		t := uiGet(n, "text", "")
		n.RealSize.Y = Platform.TextHeight(t) + n.Padding.Top + n.Padding.Bottom
		n.IsHeightResolved = true
	}

	for _, child := range n.Children {
		child.ResolveStandalone()
	}
}

// Resolves parent-dependent sizes
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

// Resolves child-dependent sizes
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

	total_width := w + n.Padding.Left + n.Padding.Right
	total_height := h + n.Padding.Top + n.Padding.Bottom

	if total_width > n.RealSize.X {
		fracs := n.xFracs()
		if fracs > 0 {
			fracdec := (total_width - n.RealSize.X) / fracs
			for _, child := range n.Children {
				if child.Size.W.Type == DT_FR {
					child.RealSize.X -= fracdec * child.Size.W.Amount
				}
			}
		}
	}

	if total_height > n.RealSize.Y {
		fracs := n.yFracs()
		if fracs > 0 {
			fracdec := (total_height - n.RealSize.Y) / fracs
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
	if n.Parent == nil { return Platform.WindowWidth() }
	if n.Parent.IsWidthResolved { return n.Parent.RealSize.X }
	return n.Parent.ParentWidth()
}

func (n *Node) ParentHeight() float32 {
	if n.Parent == nil { return Platform.WindowHeight() }
	if n.Parent.IsHeightResolved { return n.Parent.RealSize.X }
	return n.Parent.ParentWidth()
}

func (n *Node) ParentRemainingWidth() float32 {
	if n.Parent == nil { return Platform.WindowWidth() }
	w := n.ParentWidth() - n.Parent.Padding.Left - n.Parent.Padding.Right
	if n.Parent.Layout == LT_HORIZONTAL {
		for _, child := range n.Parent.Children {
			if child.IsWidthResolved {
				w -= child.RealSize.X
			}
		}
	}
	return Max(w, 0)
}

func (n *Node) ParentRemainingHeight() float32 {
	if n.Parent == nil { return Platform.WindowHeight() }
	h := n.ParentHeight() - n.Parent.Padding.Top - n.Parent.Padding.Bottom
	if n.Parent.Layout == LT_VERTICAL {
		for _, child := range n.Parent.Children {
			if child.IsHeightResolved {
				h -= child.RealSize.Y
			}
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

func (n *Node) Clicked() bool {
	return n.Flags.Focusable && n.UID == CurrentUI.Hot && Platform.MouseReleased(sdl.BUTTON_LEFT)
}

func (n *Node) Focused() bool {
	return n.Flags.Focusable && (n.UID == CurrentUI.Hot || n.UID == CurrentUI.Active)
}

func (n *Node) MouseOffset() (float32, float32) {
	return Platform.MousePos.X - n.Pos.X, Platform.MousePos.Y - n.Pos.Y
}

func (n *Node) GetFont() *Font {
	if n.Style == nil || n.Style.Font == "" {
		if n.Parent == nil {
			return GetFont("Sans", 16)
		}
		return n.Parent.GetFont()
	}
	return GetFont(n.Style.Font, n.Style.FontSize)
}
