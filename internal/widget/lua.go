package widget

import (
	"errors"

	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
	"github.com/fsnotify/fsnotify"

	"github.com/glupi-borna/wiggo/internal/ui"
	"github.com/glupi-borna/wiggo/internal/platform"
)

type LuaWidget struct {
	name string
	path string
	l *lua.LState
	initFn *lua.LFunction
	frameFn *lua.LFunction
	cleanUpFn *lua.LFunction
}

func MakeLuaWidget(name, path string) *LuaWidget {
	lw := LuaWidget{name: name, path: path}
	return &lw
}

func getLuaFn(l *lua.LState, name string, required bool) (*lua.LFunction, error) {
	gfn := l.GetGlobal(name)

	if gfn.Type() == lua.LTNil {
		if !required {
			return nil, nil
		} else {
			return nil, errors.New("Missing widget function: '" + name + "'")
		}
	}

	fn, ok := gfn.(*lua.LFunction)
	if !ok {
		return nil, errors.New("Expected '" + name + "' to be a function, got: " + gfn.String())
	}

	return fn, nil
}

func (lw *LuaWidget) Name() string { return lw.name }
func (lw *LuaWidget) Path() string { return lw.path }
func (lw *LuaWidget) Type() string { return "lua" }

func (lw *LuaWidget) init() error {
	if lw.l == nil { lw.l = lua.NewState() }

	fn, err := lw.l.LoadFile(lw.path)
	if err != nil { return err }

	lw.l.Push(fn)
	err = lw.l.PCall(0, lua.MultRet, nil)
	if err != nil { return err }

	lw.l.SetGlobal("UI", luar.New(lw.l, func() *ui.UI_State { return ui.CurrentUI }))
	lw.l.SetGlobal("TextButton", luar.New(lw.l, ui.TextButton))
	lw.l.SetGlobal("Text", luar.New(lw.l, ui.Text))
	lw.l.SetGlobal("Animate", luar.New(lw.l, ui.Animate))
	lw.l.SetGlobal("Column", luar.New(lw.l, ui.Column))
	lw.l.SetGlobal("Row", luar.New(lw.l, ui.Row))
	lw.l.SetGlobal("Button", luar.New(lw.l, ui.Button))
	lw.l.SetGlobal("Slider", luar.New(lw.l, ui.Slider))
	lw.l.SetGlobal("Invisible", luar.New(lw.l, ui.Invisible))

	NodeIter := func (iterable *ui.Node, item *ui.Node) *ui.Node {
		if item == nil { return iterable }
		if item == iterable {
			ui.CurrentUI.Pop(item)
			return nil
		}
		return nil
	}

	lw.l.SetGlobal("With", luar.New(lw.l, func(n *ui.Node) (any, any, any) {
		return NodeIter, n, nil
	}))

	lw.l.SetGlobal("Fr", luar.New(lw.l, ui.Fr))
	lw.l.SetGlobal("Px", luar.New(lw.l, ui.Px))
	lw.l.SetGlobal("Auto", luar.New(lw.l, ui.Auto))
	lw.l.SetGlobal("ChildrenSize", luar.New(lw.l, ui.ChildrenSize))
	lw.l.SetGlobal("FitText", luar.New(lw.l, ui.FitText))
	lw.l.SetGlobal("Close", luar.New(lw.l, platform.Platform.Close))

	initfn, err := getLuaFn(lw.l, "init", false)
	if err != nil { return err }
	framefn, err := getLuaFn(lw.l, "frame", true)
	if err != nil { return err }
	cleanfn, err := getLuaFn(lw.l, "cleanup", false)
	if err != nil { return err }

	lw.initFn = initfn
	lw.frameFn = framefn
	lw.cleanUpFn = cleanfn

	return nil
}

func (lw *LuaWidget) Init() error {
	err := lw.init()
	if err != nil { return err }

	watcher, err := fsnotify.NewWatcher()
	if err != nil { return err }

	go func (lw *LuaWidget) {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok { return }
				if event.Has(fsnotify.Write) {
					err = lw.init()
					if err != nil {
						lw.frameFn = luar.New(lw.l, func() error {
							return err
						}).(*lua.LFunction)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok { return }
				println(err.Error())
			}
		}
	}(lw)

	err = watcher.Add(lw.path)
	if err != nil { return err }

	return nil
}

func (lw *LuaWidget) CallFn(fn *lua.LFunction, args ...lua.LValue) (lua.LValue, error) {
	if fn == nil { return nil, nil }
	err := lw.l.CallByParam(lua.P{
		Fn: fn,
		NRet: 1,
		Protect: false,
	}, args...)
	if err != nil { return nil, err }
	ret := lw.l.Get(-1)
	lw.l.Pop(1)
	return ret, nil
}

func (lw *LuaWidget) Frame() error {
	val, err := lw.CallFn(lw.frameFn)
	if err != nil { return err }
	lud, ok := val.(*lua.LUserData)
	if !ok { return nil }
	err, ok = lud.Value.(error)
	if !ok { return nil }
	return err
}

func (lw *LuaWidget) Cleanup() error {
	_, err := lw.CallFn(lw.cleanUpFn)
	if err != nil { return err }
	lw.l.Close()
	return nil
}
