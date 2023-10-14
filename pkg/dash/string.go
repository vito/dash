package dash

import "github.com/chewxy/hm"

var StringType = NewModule("String")

type String struct {
	Value string
}

var _ Node = String{}

func (s String) Body() hm.Expression { return s }

func (s String) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	return NonNullTypeNode{NamedTypeNode{"String"}}.Infer(env, fresh)
}

type Quoted struct {
	Quoter string
	Raw    string
}
