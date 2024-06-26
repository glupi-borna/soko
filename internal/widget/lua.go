package widget

import (
	"fmt"
	"reflect"
	"errors"

	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
	"github.com/fsnotify/fsnotify"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/glupi-borna/soko/internal/ui"
)

type LuaWidget struct {
	name string
	path string
	l *lua.LState
	initFn *lua.LFunction
	frameFn *lua.LFunction
	cleanUpFn *lua.LFunction
	reloadQueued bool
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

func luaColor(val lua.LValue) (sdl.Color, bool) {
	if val == lua.LNil { return sdl.Color{}, true }
	num, ok := val.(lua.LNumber)
	if ok { return ui.ColHex(uint32(num)), true }

	ud, ok := val.(*lua.LUserData)
	if ok {
		c, ok := ud.Value.(sdl.Color)
		if ok { return c, true }
		println("Failed to parse", reflect.ValueOf(c).String())
	}

	return sdl.Color{}, false
}

func luaInt(val lua.LValue) (int, bool) {
	if val == lua.LNil { return 0, true }
	num, ok := val.(lua.LNumber)
	if ok { return int(num), true }
	return 0, false
}

func luaFloat32(val lua.LValue) (float32, bool) {
	if val == lua.LNil { return 0, true }
	num, ok := val.(lua.LNumber)
	if ok { return float32(num), true }
	return 0, false
}

func luaFloat64(lv lua.LValue) (float64, bool) {
	if lv == lua.LNil { return 0, false }
	lvv, ok := lv.(lua.LNumber)
	if !ok { return 0, false }
	return float64(lvv), true
}

func luaString(lv lua.LValue) (string, bool) {
	if lv == lua.LNil { return "", false }
	lvv, ok := lv.(lua.LString)
	if !ok { return "", false }
	return lvv.String(), true
}

func luaStyleVar[K any](val lua.LValue, convert func(lua.LValue)(K, bool)) ui.StyleVariant[K] {
	var sv ui.StyleVariant[K]
	v, ok := convert(val)
	if ok { return ui.StyleVar(v) }

	tbl, ok := val.(*lua.LTable)
	if ok {
		normal, ok := convert(tbl.RawGetString("Normal"))
		sv.Normal = normal
		if !ok { println("Normal: unsupported value") }

		active, ok := convert(tbl.RawGetString("Active"))
		sv.Active = active
		if !ok {
			println("Active: unsupported value")
			sv.Active = normal
		}

		hot, ok := convert(tbl.RawGetString("Hot"))
		sv.Hot = hot
		if !ok {
			println("Hot: unsupported value")
			sv.Hot = active
		}
	}

	return sv
}

func (lw *LuaWidget) Name() string { return lw.name }
func (lw *LuaWidget) Path() string { return lw.path }
func (lw *LuaWidget) Type() string { return "lua" }

func (lw *LuaWidget) Expose(name string, val any) {
	lw.l.SetGlobal(name, luar.New(lw.l, val))
}

func (lw *LuaWidget) init() error {
	if lw.l == nil { lw.l = lua.NewState() }

	fn, err := lw.l.LoadFile(lw.path)
	if err != nil { return err }

	ExposeEnvironment(lw)

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

	lw.l.SetGlobal("Style", luar.New(lw.l, func(arg *lua.LTable) *ui.Style {
		s := ui.DefaultStyle.Copy()

		bg := arg.RawGetString("Background")
		if bg != lua.LNil { s.Background = luaStyleVar(bg, luaColor) }

		fg := arg.RawGetString("Foreground")
		if fg != lua.LNil { s.Foreground = luaStyleVar(fg, luaColor) }

		bo := arg.RawGetString("Border")
		if bo != lua.LNil { s.Border = luaStyleVar(bo, luaColor) }

		cr := arg.RawGetString("CornerRadius")
		if cr != lua.LNil { s.CornerRadius = luaStyleVar(cr, luaFloat32) }

		f, ok := luaString(arg.RawGetString("Font"))
		if ok { s.Font = f }

		fs, ok := luaInt(arg.RawGetString("FontSize"))
		if ok { s.FontSize = fs }

		pd := arg.RawGetString("Padding")
		if pd != lua.LNil {
			pdud, ok := pd.(*lua.LUserData)
			if ok {
				p, ok := pdud.Value.(ui.PaddingType)
				if ok {
					s.Padding = p
				}
			}
		}

		return s
	}))

	lw.l.Push(fn)
	err = lw.l.PCall(0, lua.MultRet, nil)
	if err != nil { return err }

	initfn, err := getLuaFn(lw.l, "init", false)
	if err != nil { return err }
	framefn, err := getLuaFn(lw.l, "frame", true)
	if err != nil { return err }
	cleanfn, err := getLuaFn(lw.l, "cleanup", false)
	if err != nil { return err }

	lw.initFn = initfn
	lw.frameFn = framefn
	lw.cleanUpFn = cleanfn

	lw.CallFn(lw.initFn)

	return nil
}

func (lw *LuaWidget) reload() {
	fmt.Println("LUA:", lw.Path(), "changed, reloading...")
	err := lw.init()
	if err != nil {
		lw.frameFn = luar.New(lw.l, func() error {
			return err
		}).(*lua.LFunction)
	}
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
					lw.reloadQueued = true
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

func (lw *LuaWidget) CallFn(fn *lua.LFunction, args ...lua.LValue) (ret lua.LValue, err error) {
	defer func() {
		panic_val := recover()
		if panic_val != nil {
			ret = nil
			err = errors.New(fmt.Sprint(panic_val))
		}
	}()

	if fn == nil { return nil, nil }
	err = lw.l.CallByParam(lua.P{
		Fn: fn,
		NRet: 1,
		Protect: true,
	}, args...)
	if err != nil { return nil, err }
	ret = lw.l.Get(-1)
	lw.l.Pop(1)
	return ret, nil
}

func (lw *LuaWidget) Frame() error {
	if lw.reloadQueued {
		lw.reloadQueued = false
		lw.reload()
	}

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
