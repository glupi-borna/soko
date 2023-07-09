package ui

import (
	. "github.com/glupi-borna/soko/internal/platform"
	. "github.com/glupi-borna/soko/internal/debug"
)

type NodeData map[string] any

var CurrentUI *UI_State

func MakeUI() *UI_State {
	ui := &UI_State{
		Data: make(map[string]NodeData, 1000),
		AnimState: make(map[string]float32, 100),
	}
	return ui
}

// Smoothly animates a value
func Animate(val float32, id string) float32 {
	Assert(CurrentUI != nil, "UI not initialized!")

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
	Assert(CurrentUI != nil, "UI not initialized!")

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
	Assert(CurrentUI != nil, "UI not initialized!")

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

type UI_State struct {
	Mode INPUT_MODE

	Data map[string]NodeData
	AnimState map[string]float32

	Root *Node
	Current *Node
	Last *Node

	Active string
	Hot    string

	ActiveChanged bool
	HotChanged    bool
}

func (ui *UI_State) Reset() {
	ui.ActiveChanged = false
	ui.HotChanged = false
}

// Pushes a node on the UI stack.
func (ui *UI_State) Push(t string) (n *Node) {
	n = GetNode(t, CurrentUI.Current)
	ui.Current = n
	ui.Last = n
	return
}

// Pops a node off the UI stack.
func (ui *UI_State) Pop(n *Node) *Node {
	ui.Current = n.Parent
	return n
}

func (ui *UI_State) SetActive(node *Node, force bool) {
	if ui.ActiveChanged && !force { return }
	if ui.Hot != "" { return }

	if node == nil {
		ui.Active = ""
	} else {
		ui.Active = node.UID
	}

	ui.ActiveChanged = true
}

func (ui *UI_State) SetHot(node *Node, force bool) {
	if ui.HotChanged && !force { return }
	if node == nil {
		ui.Hot = ""
	} else {
		ui.Hot = node.UID
	}
	ui.HotChanged = true
}

func (ui *UI_State) Begin() {
	CurrentUI = ui
	ui.Reset()
	ui.Root = GetNode("root", nil)
	ui.Root.UpdateFn = rootUpdateFn
	ui.Root.RenderFn = rootRenderFn
	ui.Root.Size.W = ChildrenSize()//Px(Platform.WindowWidth())
	ui.Root.Size.H = ChildrenSize()//Px(Platform.WindowHeight())
	ui.Root.Style = &DefaultStyle
	ui.Current = ui.Root
}

func (ui *UI_State) End() {
	if ui.Current != ui.Root {
		panic("Unbalanced UI stack!")
	}

	ui.Root.ResolveStandalone()
	ui.Root.ResolveUpwards()
	ui.Root.ResolveDownwards()
	ui.Root.ResolveViolations()
	ui.Root.ResolvePos()

	rw, rh := int32(ui.Root.RealSize.X), int32(ui.Root.RealSize.Y)
	Platform.ResizeWindow(rw, rh)

	if Platform.MouseDelta.ManhattanLength() > 5 { ui.Mode = IM_MOUSE }
	if Platform.AnyKeyPressed { ui.Mode = IM_KBD }

	ui.Root.UpdateFn(ui.Root)
}

func (ui *UI_State) Render() {
	ui.Root.Render()
}
