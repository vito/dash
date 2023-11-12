package dash

import (
	"fmt"

	"github.com/chewxy/hm"
)

type Self struct {
	// Slots to override.
	Args Record
}

var _ Node = Self{}

func (s Self) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	mod := env.(*Module)

	argsType, err := s.Args.Infer(env, fresh)
	if err != nil {
		return nil, err
	}

	argsRec := argsType.(*RecordType)

	for _, f := range argsRec.Fields {
		haveT, isMono := f.Value.Type()
		if !isMono {
			return nil, fmt.Errorf("expected monotype for %s", f.Key)
		}
		expectedScheme, found := mod.SchemeOf(f.Key)
		if !found {
			return nil, fmt.Errorf("unknown argument: %s", f.Key)
		}
		expectedT, isMono := expectedScheme.Type()
		if !isMono {
			return nil, fmt.Errorf("expected monotype for %s", f.Key)
		}
		if !haveT.Eq(expectedT) {
			return nil, fmt.Errorf("expected %s to be %s, not %s", f.Key, expectedT, haveT)
		}
	}

	// TODO(vito): this used to Clone(), not sure if still
	// needed or just garbage
	return NonNullType{mod}, nil
}

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
