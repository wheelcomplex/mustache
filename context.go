package mustache

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type Context interface {
	GetValue(key string) (interface{}, bool)
	GetString(key string) (string, bool)
	Root() Context
}

type ComboContext struct {
	Ctxs []Context
	root Context
}

func (me *ComboContext) GetValue(key string) (interface{}, bool) {
	for _, ctx := range me.Ctxs {
		val, found := ctx.GetValue(key)
		if found {
			return val, true
		}
	}
	return nil, false
}

func (me *ComboContext) GetString(key string) (string, bool) {
	for _, ctx := range me.Ctxs {
		val, found := ctx.GetString(key)
		if found {
			return val, true
		}
	}
	return "", false
}

func (c *ComboContext) Root() Context {
	return c.root
}

func MakeContext(objs ...interface{}) Context {
	log.Println("Combo Ctx Len=", len(objs), objs)
	ctxs := make([]Context, len(objs))
	for i := range objs {
		ctxs[i] = _makeContext(objs[i])
	}
	return &ComboContext{ctxs, nil}
}

func _makeContext(obj interface{}) Context {
	t := reflect.TypeOf(obj)
	kind := t.Kind().String()
	log.Println("kind="+kind, obj)
	switch {
	case t.Kind() == reflect.Map:
		return &MapContext{obj, nil}
	case strings.HasPrefix(kind, "int") || strings.HasPrefix(kind, "uint"):
		fallthrough
	case strings.HasPrefix(kind, "float") || strings.HasPrefix(kind, "complex"):
		fallthrough
	case kind == "bool" || kind == "string":
		return &ConstContext{fmt.Sprintf("%v", obj), nil}
	case kind == "struct":
		log.Println("As StructContext", obj)
		return &StructContext{obj, nil}
	case kind == "ptr":
		_obj := reflect.ValueOf(obj).Elem()
		if _obj.Kind() == reflect.Struct {
			return &StructContext{obj, nil}
		}
		return _makeContext(_obj.Interface())
	//case strings.HasPrefix(kind, "chan"):
	//	panic("Not support Chan")
	//case strings.HasPrefix(kind, "uintpr"):
	//	panic("Not support uintpr")
	default:
		panic("Not support " + kind)
	}
	return nil
}

type ConstContext struct {
	Value string
	root  Context
}

func (me *ConstContext) GetValue(key string) (interface{}, bool) {
	if key == "." {
		return me.Value, true
	}
	return nil, false
}

func (me *ConstContext) GetString(key string) (string, bool) {
	return me.Value, true
}

func (me *ConstContext) Root() Context {
	return me.root
}

type MapContext struct {
	_map interface{}
	root Context
}

func (me *MapContext) GetValue(key string) (interface{}, bool) {
	v := reflect.ValueOf(me._map)
	keys := v.MapKeys()
	for _, _key := range keys {
		if key == _key.String() {
			return v.MapIndex(_key).Interface(), true
		}
	}
	return nil, false
}

func (me *MapContext) GetString(key string) (string, bool) {
	v := reflect.ValueOf(me._map)
	keys := v.MapKeys()
	for _, _key := range keys {
		if key == _key.String() {
			return fmt.Sprintf("%v", v.MapIndex(_key).Interface()), true
		}
	}
	return "", false
}

func (me *MapContext) Root() Context {
	return me.root
}

type EmtryContext struct {
	root Context
}

func (ctx *EmtryContext) GetValue(key string) (interface{}, bool) {
	return nil, false
}

func (ctx *EmtryContext) GetString(key string) (string, bool) {
	return "", false
}

func (ctx *EmtryContext) Root() Context {
	return ctx.root
}

type StructContext struct {
	obj  interface{}
	root Context
}

func (ctx *StructContext) GetValue(key string) (interface{}, bool) {
	v := reflect.Indirect(reflect.ValueOf(ctx.obj))
	field := v.FieldByName(key)
	if field.IsValid() {
		log.Println("Return Field Value for key=" + key)
		return field.Interface(), true
	}

	t := reflect.TypeOf(ctx.obj)
	method, ok := t.MethodByName(key)
	if !ok {
		log.Println("Miss field or method for -->" + key)
		return nil, false
	}
	if method.Func.Type().NumIn() != 1 || method.Func.Type().NumOut() == 0 {
		log.Println("Named Func Found , but has-args or void return", method, method.Func.Type().NumIn(), method.Func.Type().NumOut())
		return nil, false
	}

	return method.Func.Call([]reflect.Value{reflect.ValueOf(ctx.obj)})[0].Interface(), true
}

func (ctx *StructContext) GetString(key string) (string, bool) {
	val, ok := ctx.GetValue(key)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%v", val), true
}

func (ctx *StructContext) Root() Context {
	return ctx.root
}
