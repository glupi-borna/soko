package widget

import (
	"reflect"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"errors"
	"github.com/yuin/gopher-lua"
	"github.com/glupi-borna/wiggo/internal/ui"
)

type Widget interface {
	Name()    string
	Path()    string
	Type()    string
	Init()    error
	Frame()   error
	Cleanup() error
}

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

func GetArgs(l *lua.LState) []lua.LValue {
	top := l.GetTop()
	args := make([]lua.LValue, top)
	for i := 0; i < top ; i++ {
		v := l.Get(i+1)
		args[i] = v
	}
	return args
}

func LValStr(val lua.LValue) (string, error) {
	v, ok := val.(lua.LString)
	if ok { return v.String(), nil }
	return "", errors.New("Expected a string, got " + val.Type().String())
}

func MakeLValue(val any) (lua.LValue, error) {
	switch v := val.(type) {
	case nil: return lua.LNil, nil
	case string: return lua.LString(v), nil

	case bool: return lua.LBool(v), nil

	case float32: return lua.LNumber(v), nil
	case float64: return lua.LNumber(v), nil
	case int8: return lua.LNumber(v), nil
	case int16: return lua.LNumber(v), nil
	case int32: return lua.LNumber(v), nil
	case int64: return lua.LNumber(v), nil
	case uint8: return lua.LNumber(v), nil
	case uint16: return lua.LNumber(v), nil
	case uint32: return lua.LNumber(v), nil
	case uint64: return lua.LNumber(v), nil

	case error: return lua.LNil, v

	default:
		t := reflect.TypeOf(v)
		println("Unsupported return value in function call: ", t.String())
		return lua.LNil, nil
	}
}

func (lw *LuaWidget) ExposeFn(name string, fn func (args ...lua.LValue) any) {
	lw.l.SetGlobal(name, lw.l.NewFunction(func (l *lua.LState) int {
		args := GetArgs(l)
		lv, err := MakeLValue(fn(args...))
		if err != nil { println("error calling", name, ":", err.Error()) }
		l.Push(lv)
		return 1
	}))
}

func (lw *LuaWidget) Init() error {
	lw.l = lua.NewState()

	fn, err := lw.l.LoadFile(lw.path)
	if err != nil { return err }

	lw.l.Push(fn)
	err = lw.l.PCall(0, lua.MultRet, nil)
	if err != nil { return err }

	lw.ExposeFn("Button", func (args ...lua.LValue) any {
		if len(args) != 1 { return errors.New(fmt.Sprintln("Button: Expected exactly 1 arg, got", len(args))) }
		text, err := LValStr(args[0])
		if err != nil { return errors.New("Button: " + err.Error()) }
		return ui.TextButton(text)
	})

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

func (lw *LuaWidget) CallFn(fn *lua.LFunction, args ...lua.LValue) (lua.LValue, error) {
	if fn == nil { return nil, nil }
	err := lw.l.CallByParam(lua.P{
		Fn: fn,
		NRet: 1,
		Protect: true,
	}, args...)
	if err != nil { return nil, err }
	ret := lw.l.Get(-1)
	lw.l.Pop(1)
	return ret, nil
}

func (lw *LuaWidget) Frame() error {
	_, err := lw.CallFn(lw.frameFn)
	return err
}

func (lw *LuaWidget) Cleanup() error {
	_, err := lw.CallFn(lw.cleanUpFn)
	if err != nil { return err }
	lw.l.Close()
	return nil
}

func RootPath() string {
	return "/home/borna"
}

func ExtSupported(ext string) bool {
	switch ext {
	case ".lua": return true
	default: return false
	}
}

// Finds widget definition files.
// Widget definitions have filenames starting with 'wiggo_', and ending with one
// of the supported extensions (currently only lua).
func FindWidgets() ([]Widget, error) {
	out := []Widget{}

	items, err := ioutil.ReadDir(".")
	if err != nil { return nil, err }

	for _, item := range items {
		if item.IsDir() { continue }

		filename := item.Name()
		if !strings.HasPrefix(filename, "wiggo_") { continue }

		ext := path.Ext(filename)
		if !ExtSupported(ext) { continue }
		name := filename[6:len(filename)-len(ext)]

		switch ext {
		case ".lua":
			out = append(out, MakeLuaWidget(name, filename))
		}
	}

	return out, nil
}

func Load(name string) (Widget, error) {
	all_widgets, err := FindWidgets()
	if err != nil { return nil, err }

	for _, w := range all_widgets {
		if w.Name() != name { continue }
		return w, nil
	}

	return nil, errors.New("Widget '" + name + "' not found!")
}
