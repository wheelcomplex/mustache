package mustache

import (
	"io"
	"log"
	"reflect"
)

type SectionRenderNode struct {
	Key        string
	Inverted   bool
	lineNumber int
	father     RenderNode
	Clildren   []RenderNode
}

func (node *SectionRenderNode) Name() string {
	return node.Key
}

func (node *SectionRenderNode) Father() RenderNode {
	return node.father
}

func (node *SectionRenderNode) Render(w io.Writer, ctx Context) error {
	val, found := ctx.GetValue(node.Key)
	v := reflect.ValueOf(val)
	kind := v.Type().Kind()
	log.Println("Section Type Kind = " + kind.String())
	ok := false
	if found {
		if val == nil {
			ok = false
		} else {
			switch kind {
			case reflect.Bool:
				ok = v.Bool()
			case reflect.Map:
				ok = v.Len() != 0
			case reflect.Array:
				ok = v.Len() != 0
			case reflect.Slice:
				ok = v.Len() != 0
			case reflect.Ptr:
				ok = true
			case reflect.Struct:
				ok = true
			default:
				return &RenderError{node.lineNumber, "Not support Kind = " + kind.String()}
			}
		}
	}

	if ok && node.Inverted {
		log.Println("Section=true and Inverted, beark")
		return nil
	}
	if !ok && !node.Inverted {
		log.Println("Section=false, beark")
		return nil
	}

	var _ctx Context
	if kind == reflect.Array || kind == reflect.Slice {
		log.Println("Section Arry/Slice")
		for i := 0; i < v.Len(); i++ {
			_ctx = MakeContext(v.Index(i).Interface())
			for _, child := range node.Clildren {
				err := child.Render(w, _ctx)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	if val == nil {
		_ctx = &EmtryContext{ctx.Root()}
	} else {
		_ctx = MakeContext(val)
	}
	for _, child := range node.Clildren {
		err := child.Render(w, _ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (node *SectionRenderNode) AddChildren(clild RenderNode) {
	node.Clildren = append(node.Clildren, clild)
}

func (node *SectionRenderNode) LineNumber() int {
	return node.lineNumber
}
