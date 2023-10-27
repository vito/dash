package dash

import (
	"github.com/chewxy/hm"
)

type Self struct{
	// Slots to override.
	Args Record
}

var _ Node = Self{}

func (Self) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	// TODO(vito): this used to Clone(), not sure if still
	// needed or just garbage
	return NonNullType{env.(*Module)}, nil
}

func (s Self) Body() hm.Expression { return s }

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
