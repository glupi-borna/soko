package ui

import (
	. "github.com/glupi-borna/soko/internal/debug"
	. "github.com/glupi-borna/soko/internal/platform"
	"math"
	"time"
)

var CurrentUI *UI_State

func MakeUI() *UI_State {
	ui := &UI_State{
		Data:      make(map[string]any, 1000),
		AnimState: make(map[string]float32, 100),
	}
	return ui
}

func Interpolate(old, new float32) float32 {
	dt := (CurrentUI.FrameStart.Seconds() - CurrentUI.LastFrameStart.Seconds())
	amt := 1 - float32(math.Pow(32, -4*dt))
	return old + (new-old)*amt
}

// Smoothly animates a value
func Animate(val float32, id string) float32 {
	Assert(CurrentUI != nil, "UI not initialized!")
	old, ok := CurrentUI.AnimState[id]
	if !ok {
		CurrentUI.AnimState[id] = val
		return val
	}
	new := Interpolate(old, val)
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
	return current_pulse%2 == 1
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
	if !ok {
		return nil
	}
	return data
}

type INPUT_MODE uint8

const (
	IM_MOUSE INPUT_MODE = iota
	IM_KBD
)

type UI_State struct {
	Mode INPUT_MODE

	Data      map[string]any
	AnimState map[string]float32

	Root, Current, Last *Node

	Active, Hot, ScrollTarget                      string
	ActiveChanged, HotChanged, ScrollTargetChanged bool

	LastFrameStart,
	FrameStart, Delta time.Duration

	renderWidth,
	renderHeight float32
}

func (ui *UI_State) Reset() {
	ui.ActiveChanged = false
	ui.HotChanged = false
	ui.ScrollTargetChanged = false
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
	if ui.ActiveChanged && !force {
		return
	}
	if ui.Hot != "" {
		return
	}

	if node == nil {
		ui.Active = ""
	} else {
		ui.Active = node.UID
	}

	ui.ActiveChanged = true
}

func (ui *UI_State) SetHot(node *Node, force bool) {
	if ui.HotChanged && !force {
		return
	}
	if node == nil {
		ui.Hot = ""
	} else {
		ui.Hot = node.UID
	}
	ui.HotChanged = true
}

func (ui *UI_State) SetScrollTarget(node *Node, force bool) {
	if ui.ScrollTargetChanged && !force {
		return
	}
	if node == nil {
		ui.ScrollTarget = ""
	} else {
		ui.ScrollTarget = node.UID
	}
	ui.ScrollTargetChanged = true
}

const TIME_DIV = 1

func (ui *UI_State) Begin(millis uint64) {
	CurrentUI = ui
	ui.renderWidth = Platform.WindowWidth()
	ui.renderHeight = Platform.WindowHeight()
	ui.LastFrameStart = ui.FrameStart
	ui.FrameStart = time.Duration(millis * 1000 * 1000 / TIME_DIV)
	ui.Delta = ui.FrameStart - ui.LastFrameStart
	ui.Reset()
	ui.Root = GetNode("root", nil)
	ui.Root.UpdateFn = rootUpdateFn
	ui.Root.RenderFn = rootRenderFn
	ui.Root.Size.W = ChildrenSize() //Px(Platform.WindowWidth())
	ui.Root.Size.H = ChildrenSize() //Px(Platform.WindowHeight())
	ui.Root.Style = &DefaultStyle
	ui.Current = ui.Root
}

func (ui *UI_State) End() {
	if ui.Current != ui.Root {
		panic("Unbalanced UI stack!")
	}

	ui.Root.preLayout()

	ui.Root.resolveStandalone()
	ui.Root.resolveUpwards()
	ui.Root.resolveDownwards()
	ui.Root.resolveViolations()
	ui.Root.resolvePos()

	if Platform.MouseDelta.ManhattanLength() > 5 {
		ui.Mode = IM_MOUSE
	}
	if Platform.AnyKeyPressed {
		ui.Mode = IM_KBD
	}

	ui.Root.UpdateFn(ui.Root)
}

func (ui *UI_State) Render() {
	ui.Root.Render()
	rw, rh := int32(ui.Root.RealSize.X), int32(ui.Root.RealSize.Y)
	if rw > 0 && rh > 0 {
		Platform.ResizeWindow(rw, rh)
	}
	Platform.EndFrame()
}
