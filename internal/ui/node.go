package ui

import (
	. "github.com/glupi-borna/soko/internal/debug"
	. "github.com/glupi-borna/soko/internal/platform"
	. "github.com/glupi-borna/soko/internal/utils"
	"github.com/veandco/go-sdl2/sdl"
	"strconv"
	"fmt"
)

type DIM_TYPE uint8

const (
	DT_AUTO DIM_TYPE = iota
	DT_PX
	DT_EM
	DT_FR
	DT_TEXT
	DT_CHILDREN
	DT_LARGEST_SIBLING
	DT_SKIP
)

type Dimension struct {
	Type   DIM_TYPE
	Amount float32
}

func (d *Dimension) String() string {
	switch d.Type {
	case DT_CHILDREN:
		return "children"
	case DT_TEXT:
		return "text"
	case DT_FR:
		return FloatStr(d.Amount) + "fr"
	case DT_EM:
		return FloatStr(d.Amount) + "em"
	case DT_PX:
		return FloatStr(d.Amount) + "px"
	case DT_AUTO:
		return "auto"
	default:
		panic("Unknown Dimension Type: " + strconv.Itoa(int(d.Type)))
	}
}

func Px(amount float32) Dimension { return Dimension{Type: DT_PX, Amount: amount} }
func Fr(amount float32) Dimension { return Dimension{Type: DT_FR, Amount: amount} }
func Em(amount float32) Dimension { return Dimension{Type: DT_EM, Amount: amount} }
func ChildrenSize() Dimension     { return Dimension{Type: DT_CHILDREN} }
func LargestSibling() Dimension   { return Dimension{Type: DT_LARGEST_SIBLING} }
func FitText() Dimension          { return Dimension{Type: DT_TEXT} }
func Auto() Dimension             { return Dimension{Type: DT_AUTO} }

type Size struct{ W, H Dimension }

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
	UID      string      // Unique ID of this Node
	Type     string      // Equivalent to tag name in HTML
	Layout   LAYOUT_TYPE // Children positioning (horizontal/vertical)
	Flags    NodeFlags
	Parent   *Node // Parent of this node - null if the node is the root node, or is detached.
	Children []*Node
	Style    *Style
	Padding  PaddingType
	Text     string

	// Semantic size
	Size Size
	// Translation after positioning (affects children)
	Translation V2

	// Calculated position
	Pos V2
	// Calculated size
	RealSize V2

	// Is RealSize.X calculated for this frame
	isWidthResolved bool
	// Is RealSize.Y calculated for this frame
	isHeightResolved bool

	// Called before layout is done
	PreLayout func(*Node)

	// Called after the position and size of the Node
	// have been resolved, after the node's parents
	// have been rendered, and before the node's children
	// have been rendered.
	RenderFn func(*Node)

	// Called after the position and size of the Node
	// have been resolved, after the node's parents
	// have been rendered, and after the node's children
	// have been rendered.
	PostRenderFn func(*Node)

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
		Type:     t,
		Children: make([]*Node, 0),
		Parent:   parent,
		RenderFn: defaultRenderFn,
		UpdateFn: defaultUpdateFn,
		Padding:  Padding1(2),
		UI:       CurrentUI,
	}

	if n.Parent != nil {
		n.Parent.Children = append(n.Parent.Children, &n)
	}

	n.UID = buildNodeUID(&n)

	return &n
}

func (n *Node) debug(ntype string, message ...any) {
	if DEBUG && n.Type == ntype {
		fmt.Println(message...)
	}
}

func (n *Node) sdlRect() sdl.Rect {
	return sdl.Rect{
		X: int32(n.Pos.X),
		Y: int32(n.Pos.Y),
		W: int32(n.RealSize.X),
		H: int32(n.RealSize.Y),
	}
}

func (n *Node) preLayout() {
	if n.PreLayout != nil {
		n.PreLayout(n)
	}

	for _, child := range n.Children {
		child.preLayout()
	}
}

func defaultUpdateFn(n *Node) {
	n.UpdateChildren()

	if n.Flags.Focusable {
		if CurrentUI.Mode == IM_MOUSE {
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
}

func rootUpdateFn(n *Node) {
	n.UpdateChildren()

	if CurrentUI.Mode == IM_MOUSE {
		if n.HasMouse() {
			CurrentUI.SetScrollTarget(nil, false)
			CurrentUI.SetActive(nil, false)
			if Platform.MousePressed(sdl.BUTTON_LEFT) {
				CurrentUI.SetHot(nil, false)
			}
			if Platform.MouseReleased(sdl.BUTTON_LEFT) {
				CurrentUI.SetHot(nil, false)
			}
		}
	}
}

func rootRenderFn(n *Node) {
	s := n.GetStyle()
	radius := s.CornerRadius.Normal
	c := n.Style.Background.Normal

	Platform.SetColor(c)
	// Platform.DrawRectFilled(n.Pos.X, n.Pos.Y, n.RealSize.X, n.RealSize.Y)
	Platform.Renderer.Clear()
	Platform.ReshapeWindow(radius, false)
}

func (n *Node) Styled() *Style {
	if n.Style != nil {
		return n.Style
	}
	n.Style = DefaultStyle.Copy()
	return n.Style
}

func (n *Node) CountChildrenOfType(t string) int {
	count := 0
	for _, child := range n.Children {
		if child.Type == t {
			count++
		}
	}
	return count
}

func (n *Node) Render() {
	// println(n.UID)
	// println("  pos ", n.Pos.String())
	// println("  size", n.RealSize.String())
	// println("  sem ", n.Size.String())
	// println("  res ", n.IsWidthResolved, n.IsHeightResolved)

	debug, ok := Platform.Mouse[sdl.BUTTON_LEFT]
	if ok && debug == BS_DOWN && n.HasMouse() {
		pos := n.Pos
		size := n.RealSize
		r, g, b, a, _ := Platform.Renderer.GetDrawColor()
		Platform.Renderer.SetDrawColor(255, 0, 0, 255)
		Platform.DrawText(n.UID, pos.X, pos.Y+size.Y*0.5)
		Platform.DrawRectOutlined(pos.X, pos.Y, size.X, size.Y)
		Platform.Renderer.SetDrawColor(r, g, b, a)
	}

	if n.RenderFn != nil {
		n.RenderFn(n)
	}

	for _, child := range n.Children {
		child.Render()
	}

	if n.PostRenderFn != nil {
		n.PostRenderFn(n)
	}
}

func (n *Node) GetStyle() *Style {
	if n.Style != nil {
		return n.Style
	}
	if n.Parent != nil {
		return n.Parent.GetStyle()
	}
	return &DefaultStyle
}

func (n *Node) xFracs() float32 {
	if n.Layout == LT_VERTICAL {
		return 1
	}

	var count float32 = 0
	for _, child := range n.Children {
		cw := child.Size.W
		if cw.Type == DT_FR {
			count += cw.Amount
		}
	}
	return count
}

func (n *Node) yFracs() float32 {
	if n.Layout == LT_HORIZONTAL {
		return 1
	}

	var count float32 = 0
	for _, child := range n.Children {
		ch := child.Size.H
		if ch.Type == DT_FR {
			count += ch.Amount
		}
	}
	return count
}

var autoMap = map[string]Size{
	"text":   {FitText(), FitText()},
	"row":    {Fr(1), ChildrenSize()},
	"column": {ChildrenSize(), Fr(1)},
}

func resolveAuto(n *Node) {
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
func (n *Node) resolveStandalone() {
	resolveAuto(n)

	switch n.Size.W.Type {
	case DT_PX:
		n.RealSize.X = n.Size.W.Amount + n.Padding.xPadding()
		n.isWidthResolved = true

	case DT_EM:
		n.RealSize.X = n.Size.W.Amount*n.GetFont().Height + n.Padding.xPadding()
		n.isWidthResolved = true

	case DT_TEXT:
		n.RealSize.X = Platform.TextWidth(n.Text) + n.Padding.xPadding()
		n.isWidthResolved = true
	}


	switch n.Size.H.Type {
	case DT_PX:
		n.RealSize.Y = n.Size.H.Amount + n.Padding.yPadding()
		n.isHeightResolved = true

	case DT_EM:
		n.RealSize.Y = n.Size.H.Amount*n.GetFont().Height + n.Padding.yPadding()
		n.isHeightResolved = true

	case DT_TEXT:
		n.RealSize.Y = Platform.TextHeight(n.Text) + n.Padding.yPadding()
		n.isHeightResolved = true
	}

	for _, child := range n.Children {
		child.resolveStandalone()
	}
}

// Resolves parent-dependent sizes
func (n *Node) resolveUpwards() {
	if !n.isWidthResolved && n.Size.W.Type == DT_FR {
		pw := n.parentRemainingWidth()
		fracs := n.Parent.xFracs()
		fracw := pw / fracs

		for _, child := range n.Parent.Children {
			if child.Size.W.Type == DT_FR {
				child.RealSize.X = fracw*child.Size.W.Amount + child.Padding.xPadding()
				child.isWidthResolved = true
			}
		}
	}

	if !n.isHeightResolved && n.Size.H.Type == DT_FR {
		ph := n.parentRemainingHeight()
		fracs := n.Parent.yFracs()
		frach := ph / fracs

		for _, child := range n.Parent.Children {
			if child.Size.H.Type == DT_FR {
				child.RealSize.Y = frach*child.Size.H.Amount + child.Padding.yPadding()
				child.isHeightResolved = true
			}
		}
	}

	for _, child := range n.Children {
		child.resolveUpwards()
	}
}

// Resolves child-dependent sizes
func (n *Node) resolveDownwards() {
	for _, child := range n.Children {
		child.resolveDownwards()
	}

	if n.Size.W.Type == DT_CHILDREN {
		switch n.Layout {
		case LT_HORIZONTAL:
			n.RealSize.X = n.childSum(nodeRealX) + n.Padding.xPadding()
			n.isWidthResolved = true

		case LT_VERTICAL:
			n.RealSize.X = n.childMax(nodeRealX) + n.Padding.xPadding()
			n.isWidthResolved = true
		}
	}

	if n.Size.H.Type == DT_CHILDREN {
		switch n.Layout {
		case LT_VERTICAL:
			n.RealSize.Y = n.childSum(nodeRealY) + n.Padding.yPadding()
			n.isHeightResolved = true

		case LT_HORIZONTAL:
			n.RealSize.Y = n.childMax(nodeRealY) + n.Padding.yPadding()
			n.isHeightResolved = true
		}
	}
}

func realWidth(n *Node) float32  { return n.RealSize.X }
func realHeight(n *Node) float32 { return n.RealSize.Y }

func (n *Node) resolveViolations() {
	if n.Size.W.Type == DT_LARGEST_SIBLING {
		n.RealSize.X = n.Parent.childMax(realWidth) + n.Padding.xPadding()
	}

	if n.Size.H.Type == DT_LARGEST_SIBLING {
		n.RealSize.Y = n.Parent.childMax(realHeight) + n.Padding.yPadding()
	}

	var w, h float32

	switch n.Layout {
	case LT_VERTICAL:
		w = n.childMax(nodeRealX)
		h = n.childSum(nodeRealY)

	case LT_HORIZONTAL:
		w = n.childSum(nodeRealX)
		h = n.childMax(nodeRealY)
	}

	total_width := w + n.Padding.Left + n.Padding.Right
	total_height := h + n.Padding.Top + n.Padding.Bottom

	if n.Layout == LT_HORIZONTAL && total_width > n.RealSize.X {
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

	if n.Layout == LT_VERTICAL && total_height > n.RealSize.Y {
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

	for _, child := range n.Children {
		child.resolveViolations()
	}
}

func nodeRealX(n *Node) float32 {
	if n.isWidthResolved {
		return n.RealSize.X
	} else {
		return 0
	}
}

func nodeRealY(n *Node) float32 {
	if n.isHeightResolved {
		return n.RealSize.Y
	} else {
		return 0
	}
}

func (n *Node) childSum(fn func(*Node) float32) float32 {
	var sum float32 = 0
	for _, child := range n.Children {
		sum += fn(child)
	}
	return sum
}

func (n *Node) childMax(fn func(*Node) float32) (val float32) {
	for _, child := range n.Children {
		cv := fn(child)
		if cv > val {
			val = cv
		}
	}
	return
}

func (n *Node) parentWidth() float32 {
	if n.Parent == nil {
		tdb, err := Platform.TargetDisplayBounds()
		Die(err)
		return float32(tdb.W)
	}
	if n.Parent.isWidthResolved {
		return n.Parent.RealSize.X
	}
	return n.Parent.parentWidth()
}

func (n *Node) parentHeight() float32 {
	if n.Parent == nil {
		tdb, err := Platform.TargetDisplayBounds()
		Die(err)
		return float32(tdb.H)
	}
	if n.Parent.isHeightResolved {
		return n.Parent.RealSize.X
	}
	return n.Parent.parentWidth()
}

func (n *Node) parentRemainingWidth() float32 {
	if n.Parent == nil {
		tdb, err := Platform.TargetDisplayBounds()
		Die(err)
		return float32(tdb.W)
	}
	w := n.parentWidth() - n.Parent.Padding.Left - n.Parent.Padding.Right
	if n.Parent.Layout == LT_HORIZONTAL {
		for _, child := range n.Parent.Children {
			if child.isWidthResolved {
				w -= child.RealSize.X
			}
		}
	}
	return Max(w, 0)
}

func (n *Node) parentRemainingHeight() float32 {
	if n.Parent == nil {
		tdb, err := Platform.TargetDisplayBounds()
		Die(err)
		return float32(tdb.H)
	}
	h := n.parentHeight() - n.Parent.Padding.Top - n.Parent.Padding.Bottom
	if n.Parent.Layout == LT_VERTICAL {
		for _, child := range n.Parent.Children {
			if child.isHeightResolved {
				h -= child.RealSize.Y
			}
		}
	}
	return Max(h, 0)
}

func (n *Node) resolvePos() {
	if n.Parent == nil {
		n.Pos.X = 0
		n.Pos.Y = 0
	} else {
		s := n.GetStyle()
		if s != nil {
			switch s.Align {
			case A_CENTER:
				switch n.Parent.Layout {
				case LT_HORIZONTAL:
					n.Pos.Y = n.Parent.Pos.Y + n.Parent.RealSize.Y*.5 - n.RealSize.Y*.5
				case LT_VERTICAL:
					n.Pos.X = n.Parent.Pos.X + n.Parent.RealSize.X*.5 - n.RealSize.X*.5
				}
			case A_END:
				switch n.Parent.Layout {
				case LT_HORIZONTAL:
					n.Pos.Y = n.Parent.Pos.Y + n.Parent.RealSize.Y - n.RealSize.Y
				case LT_VERTICAL:
					n.Pos.X = n.Parent.Pos.X + n.Parent.RealSize.X - n.RealSize.X
				}
			}
		}
	}

	n.Pos.Y += n.Translation.Y
	n.Pos.X += n.Translation.X

	var offset float32 = 0
	xmul := Btof(n.Layout == LT_HORIZONTAL)
	ymul := Btof(n.Layout == LT_VERTICAL)

	for _, child := range n.Children {
		child.Pos.X = n.Pos.X + n.Padding.Left + offset*xmul
		child.Pos.Y = n.Pos.Y + n.Padding.Top + offset*ymul
		offset += child.RealSize.X * xmul
		offset += child.RealSize.Y * ymul
		child.resolvePos()
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
	x2, y2 := x1+n.RealSize.X, y1+n.RealSize.Y

	return (mx >= x1 &&
		mx <= x2 &&
		my >= y1 &&
		my <= y2)
}

func (n *Node) IsChildOf(t *Node) bool {
	p := n.Parent
	for p != nil {
		if p == t {
			return true
		}
		p = p.Parent
	}
	return false
}

func (n *Node) IsChildOfUID(uid string) bool {
	p := n.Parent
	for p != nil {
		if p.UID == uid {
			return true
		}
		p = p.Parent
	}
	return false
}

func (n *Node) Index() int {
	if n.Parent == nil {
		return -1
	}
	for idx, child := range n.Parent.Children {
		if child == n {
			return idx
		}
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
