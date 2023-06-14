package ui

import (
	. "github.com/glupi-borna/wiggo/internal/platform"
)

type NodeData map[string] any

var CurrentUI *ui_state

func MakeUI() *ui_state {
	ui := &ui_state{
		Data: make(map[string]NodeData, 1000),
		AnimState: make(map[string]float32, 100),
	}
	return ui
}

// Smoothly animates a value
func Animate(val float32, id string) float32 {
	old, ok := CurrentUI.AnimState[id]
	if !ok {
		CurrentUI.AnimState[id] = val
		return val
	}
	new := old + 0.5 * (val-old)
	CurrentUI.AnimState[id] = new
	return new
}

func uiGet[K any](n *Node, key string, dflt K) (out K) {
	data, ok := CurrentUI.Data[n.UID]

	// If node data doesn't exist
	if !ok {
		data = make(NodeData)
		CurrentUI.Data[n.UID] = data
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
	data, ok := CurrentUI.Data[n.UID]
	if !ok {
		data = make(NodeData)
		CurrentUI.Data[n.UID] = data
		data[key] = val
		return true
	}
	old := data[key]
	data[key] = val
	return old != val
}

type INPUT_MODE uint8

const (
	IM_MOUSE INPUT_MODE = iota
	IM_KBD
)

type ui_state struct {
	Mode INPUT_MODE

	Data map[string]NodeData
	AnimState map[string]float32

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
func (ui *ui_state) Push(t string) (n *Node) {
	n = GetNode(t, CurrentUI.Current)
	ui.Current = n
	return
}

// Pops a node off the UI stack.
func (ui *ui_state) Pop(n *Node) *Node {
	ui.Current = n.Parent
	return n
}

func (ui *ui_state) SetActive(node *Node, force bool) {
	if ui.ActiveChanged && !force { return }
	if ui.Hot != "" { return }

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
	CurrentUI = ui
	ui.Reset()
	ui.Root = GetNode("root", nil)
	ui.Root.UpdateFn = rootUpdateFn
	ui.Root.RenderFn = invisibleRenderFn
	ui.Root.Size.W = Px(WindowWidth())
	ui.Root.Size.H = Px(WindowHeight())
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
