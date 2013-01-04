package mustache

import (
	"errors"
	"io"
	"reflect"
	_g_tpl "text/template"
)

type ValNode struct {
	name   string
	Escape bool
}

func (node *ValNode) Render(ctx Context, w io.Writer) error {
	val, found := ctx.Get(node.name)
	if !found {
		return errors.New("key=" + node.name + " NOT Found")
	}
	str := val.String()
	if node.Escape {
		str = _g_tpl.HTMLEscapeString(str)
	}
	_, err := w.Write([]byte(str))
	return err
}

func (node *ValNode) String() string {
	if node.Escape {
		return "{{&" + node.name + "}}"
	}
	return "{{" + node.name + "}}"
}

func (node *ValNode) Name() string {
	return node.name
}

//------------------------------------------------------
type SectionNode struct {
	name     string
	Inverted bool
	Clildren []Node
}

func (node *SectionNode) Render(ctx Context, w io.Writer) error {
	//TODO person?to_name
	var err error

	val, found := ctx.Get(node.name)
	if !found {
		return errors.New("key=" + node.name + " NOT Found")
	}

	if node.Inverted {
		if val.Bool() {
			return nil
		}
		for _, _node := range node.Clildren {
			err = _node.Render(ctx, w)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if !val.Bool() {
		return nil
	}

	switch val.Val.Type().Kind() {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		for i := 0; i < val.Val.Len(); i++ {
			_ctx := MakeContext(val.Val.Index(i))
			for _, child := range node.Clildren {
				err := child.Render(_ctx, w)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Func:
		return errors.New("Not support Func yet") // TODO impl Lambdas
	default:
		_ctx := MakeContext(val.Val)
		for _, child := range node.Clildren {
			err := child.Render(_ctx, w)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (node *SectionNode) String() string {
	str := "{{#" + node.name + "}}"
	for _, c := range node.Clildren {
		str += c.String()
	}
	str += "{{/" + node.name + "}}"
	return str
}

func (node *SectionNode) Name() string {
	return node.name
}

//------------------------------------------------------
type PartialNode struct {
	name string
}

func (node *PartialNode) Render(ctx Context, w io.Writer) error {
	return nil
}

func (node *PartialNode) String() string {
	return "{{>" + node.name + "}}"
}

func (node *PartialNode) Name() string {
	return node.name
}

//------------------------------------------------------

type ConstantNode struct {
	Val string
}

func (node *ConstantNode) Render(ctx Context, w io.Writer) error {
	_, err := w.Write([]byte(node.Val))
	return err
}

func (node *ConstantNode) String() string {
	return node.Val
}

func (node *ConstantNode) Name() string {
	return node.Val
}
