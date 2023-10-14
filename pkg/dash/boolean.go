package dash

import "github.com/chewxy/hm"

var BooleanType = NewModule("Boolean")

type Boolean bool

var _ Node = Boolean(false)

func (b Boolean) Body() hm.Expression { return b }

func (b Boolean) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	return NonNullTypeNode{NamedTypeNode{"Boolean"}}.Infer(env, fresh)
}
