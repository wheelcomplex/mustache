package mustache

import (
	"html"
	"io"
)

type VariableRenderNode struct {
	Key        string
	Escape     bool
	lineNumber int
}

func (node *VariableRenderNode) Name() string {
	return node.Key
}

func (node *VariableRenderNode) Father() RenderNode {
	return nil
}

func (node *VariableRenderNode) Render(w io.Writer, ctx Context) error {
	str, found := ctx.GetString(node.Key)
	if !found && !IgnoreInvaildKey {
		return &RenderError{node.lineNumber, "Variable NOT found : " + node.Key}
	}
	if node.Escape {
		str = html.EscapeString(str)
	}
	_, err := io.WriteString(w, str)
	return err
}

func (node *VariableRenderNode) AddChildren(clildren RenderNode) {
}

func (node *VariableRenderNode) LineNumber() int {
	return node.lineNumber
}
