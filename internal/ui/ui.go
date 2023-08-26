package ui

import (
	"time"
	. "github.com/glupi-borna/soko/internal/platform"
	. "github.com/glupi-borna/soko/internal/debug"
)

var CurrentUI *UI_State

func MakeUI() *UI_State {
	ui := &UI_State{
		Data: make(map[string]any, 1000),
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

// Returns true approximately every time `seconds` passes
func Tick(seconds float64) bool {
	Assert(CurrentUI != nil, "UI not initialized!")
	ns := uint64(seconds * 1000 * 1000 * 1000)
	last_frame_tick := uint64(CurrentUI.LastFrameStart) / ns
	current_frame_tick := uint64(CurrentUI.FrameStart) / ns
	return last_frame_tick != current_frame_tick || CurrentUI.LastFrameStart == 0
}

// Returns true or false, switching the value it returns every time `seconds` passes
func Pulse(seconds float64) bool {
	Assert(CurrentUI != nil, "UI not initialized!")
	ns := uint64(seconds * 1000 * 1000 * 1000)
	current_pulse := uint64(CurrentUI.FrameStart) / ns
	return current_pulse % 2 == 0
}

func NodeState[K any](n *Node) *K {
	Assert(CurrentUI != nil, "UI not initialized!")

	var val K

	data, ok := CurrentUI.Data[n.UID]
	if !ok {
		data = &val
		CurrentUI.Data[n.UID] = data
	}

	ptr, ok := data.(*K)
	if !ok {
		CurrentUI.Data[n.UID] = val
		ptr = &val
	}

	return ptr
}

func NodeStateAny(n *Node) any {
	Assert(CurrentUI != nil, "UI not initialized!")
	data, ok := CurrentUI.Data[n.UID]
	if !ok { return nil }
	return data
}

type INPUT_MODE uint8

const (
	IM_MOUSE INPUT_MODE = iota
	IM_KBD
)

type UI_State struct {
	Mode INPUT_MODE

	Data map[string]any
	AnimState map[string]float32

	Root *Node
	Current *Node
	Last *Node

	Active string
	Hot    string

	ActiveChanged bool
	HotChanged    bool

	LastFrameStart time.Duration
	FrameStart time.Duration
	Delta time.Duration
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

func (ui *UI_State) Begin(millis uint64) {
	CurrentUI = ui
	ui.LastFrameStart = ui.FrameStart
	ui.FrameStart = time.Duration(millis*1000*1000)
	ui.Delta = ui.FrameStart - ui.LastFrameStart
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

	if Platform.MouseDelta.ManhattanLength() > 5 { ui.Mode = IM_MOUSE }
	if Platform.AnyKeyPressed { ui.Mode = IM_KBD }

	ui.Root.UpdateFn(ui.Root)

	// rw, rh := int32(ui.Root.RealSize.X), int32(ui.Root.RealSize.Y)
	// Platform.ResizeWindow(rw, rh)
}

func (ui *UI_State) Render() {
	ui.Root.Render()
}
