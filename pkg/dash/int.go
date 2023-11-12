package dash

import "github.com/chewxy/hm"

var IntType = NewModule("Int")

type Int int

var _ Node = Int(0)

func (i Int) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	return NonNullTypeNode{NamedTypeNode{"Int"}}.Infer(env, fresh)
}
