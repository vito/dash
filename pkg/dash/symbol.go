package dash

import (
	"fmt"

	"github.com/chewxy/hm"
)

type Symbol struct {
	Name string
}

var _ Node = Symbol{}

func (s Symbol) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	scheme, found := env.SchemeOf(s.Name)
	if !found {
		return nil, fmt.Errorf("Symbol.Infer: %q not found in env", s.Name)
	}
	t, isMono := scheme.Type()
	if !isMono {
		return nil, fmt.Errorf("Symbol.Infer: TODO: %q is not monomorphic", s.Name)
	}
	return InferThunkResult(t), nil
}

func (s Symbol) Body() hm.Expression { return s }

func InferThunkResult(t hm.Type) hm.Type {
	switch x := t.(type) {
	case NonNullType:
		return NonNullType{InferThunkResult(x.Type)}
	case *hm.FunctionType:
		args := x.Arg().(*RecordType)

		var hasReq bool
		for _, arg := range args.Fields {
			argT, _ := arg.Value.Type() // TODO: care about isMono?
			_, isReq := argT.(NonNullType)
			if isReq {
				hasReq = true
				break
			}
		}
		if !hasReq {
			return x.Ret(false)
		} else {
			return t
		}
	default:
		return t
	}
}
