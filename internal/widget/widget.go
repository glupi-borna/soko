package widget

import (
	"golang.org/x/exp/constraints"
	"reflect"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"errors"
	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
	"github.com/glupi-borna/wiggo/internal/ui"
	"github.com/glupi-borna/wiggo/internal/platform"
	. "github.com/glupi-borna/wiggo/internal/utils"
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

type ArgCheck func (val lua.LValue) (any, error)

type Stringable interface { String() string }

func ExpectedErr(expected string, got Stringable) error {
	return errors.New("Expected <" + expected + ">, got <" + got.String() + ">")
}

func NumCheck[V constraints.Float | constraints.Integer](val lua.LValue) (any, error) {
	v, ok := val.(lua.LNumber)
	if ok { return V(v), nil }
	return 0, ExpectedErr("number", val.Type())
}

func StrCheck(val lua.LValue) (any, error) {
	v, ok := val.(lua.LString)
	if ok { return v.String(), nil }
	return "", ExpectedErr("string", val.Type())
}

func BoolCheck(val lua.LValue) (any, error) {
	v, ok := val.(lua.LBool)
	if ok { return bool(v), nil }
	return false, ExpectedErr("bool", val.Type())
}

func StructPtrCheck(val lua.LValue) (any, error) {
	if val == lua.LNil { return nil, nil }
	v, ok := val.(*lua.LUserData)
	if !ok { return nil, ExpectedErr("UserData", val.Type()) }
	vval := reflect.ValueOf(v.Value)
	if vval.Kind() != reflect.Ptr { return nil, ExpectedErr("Pointer", vval) }
	el := vval.Elem()
	if el.IsZero() { return nil, nil }
	if el.Kind() != reflect.Struct { return nil, ExpectedErr("Struct", el)}
	return v.Value, nil
}

func FuncCheck(val lua.LValue) (any, error) {
	v, ok := val.(*lua.LFunction)
	if ok { return v.GFunction, nil }
	return nil, ExpectedErr("function", val.Type())
}

func checkArgs(
	args []lua.LValue,
	variadic bool,
	checks []ArgCheck,
) ([]reflect.Value, error) {
	maxcheck := len(checks) - 1

	if variadic {
		if len(args) < maxcheck {
			return nil, errors.New(
				fmt.Sprintln("Expected at least", maxcheck, "arguments, but got", len(args)),
			)
		}
	} else {
		if len(args)-1 != maxcheck {
			return nil, errors.New(
				fmt.Sprintln("Expected exactly", maxcheck, "arguments, but got", len(args)),
			)
		}
	}

	out := make([]reflect.Value, len(args))

	for i, arg := range args {
		idx := Min(i, maxcheck)
		check := checks[idx]
		val, err := check(arg)
		if err != nil { return nil, err }
		out[i] = reflect.ValueOf(val)
	}

	return out, nil
}

func makeLValue(val any) (lua.LValue, error) {
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
		return &lua.LUserData{Value: v}, nil
	}
}

func (lw *LuaWidget) exposeFn(name string, fn any) error {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func { return errors.New("Attempt to expose non-func") }

	fn_type := val.Type()
	variadic := fn_type.IsVariadic()
	checks := make([]ArgCheck, fn_type.NumIn())

	arg_count := fn_type.NumIn()

	for i := 0 ; i < arg_count ; i++ {
		t := fn_type.In(i)

		if variadic && i == arg_count { t = t.Elem() }

		switch t.Kind() {
		case reflect.Int: checks[i] = NumCheck[int]
		case reflect.Int8: checks[i] = NumCheck[int8]
		case reflect.Int16: checks[i] = NumCheck[int16]
		case reflect.Int32: checks[i] = NumCheck[int32]
		case reflect.Int64: checks[i] = NumCheck[int64]
		case reflect.Uint: checks[i] = NumCheck[uint]
		case reflect.Uint8: checks[i] = NumCheck[uint8]
		case reflect.Uint16: checks[i] = NumCheck[uint16]
		case reflect.Uint32: checks[i] = NumCheck[uint32]
		case reflect.Uint64: checks[i] = NumCheck[uint64]
		case reflect.Float32: checks[i] = NumCheck[float32]
		case reflect.Float64: checks[i] = NumCheck[float64]
		case reflect.String: checks[i] = StrCheck
		case reflect.Bool: checks[i] = BoolCheck
		case reflect.Pointer: checks[i] = StructPtrCheck
		case reflect.Func: checks[i] = FuncCheck
		default:
			return errors.New("No check installed for argument type: " + t.String())
		}
	}

	lw.l.SetGlobal(name, lw.l.NewFunction(func (l *lua.LState) int {
		args, err := checkArgs(GetArgs(l), variadic, checks)
		if err != nil {
			l.Error(lua.LString(err.Error()), 0)
			return 0
		}

		results := val.Call(args)
		for _, result := range results {
			val, err := makeLValue(result.Interface())
			if err != nil {
				l.Error(lua.LString(err.Error()), 0)
				return 0
			}
			l.Push(val)
		}

		if len(results) == 0 {
			l.Push(nil)
			return 1
		}

		return len(results)
	}))

	return nil
}

func ReportError(err error) {
	if err == nil { return }
	println(err.Error())
}

func (lw *LuaWidget) Init() error {
	lw.l = lua.NewState()

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
	lw.l.SetGlobal("Invisible", luar.New(lw.l, ui.Invisible))
	lw.l.SetGlobal("With", luar.New(lw.l, ui.WithNode))
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
