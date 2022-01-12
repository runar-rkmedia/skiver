package parser

type InterfaceNode interface {
	TokenLiteral() string
}
type Interpolation interface {
	InterfaceNode
	interpolationNode()
}
type Nesting interface {
	InterfaceNode
	nestingNode()
}

func newNodeLiteral(value string) Node {
	return Node{}
}
