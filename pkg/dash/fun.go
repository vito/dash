package dash

import (
	"fmt"

	"github.com/chewxy/hm"
)

type FunDecl struct {
	Named      string
	Args       []SlotDecl
	Form       Node
	Ret        TypeNode
	Visibility Visibility
}

var _ hm.Inferer = FunDecl{}

func (f FunDecl) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	// TODO: Lambda semantics

	var err error

	// closure
	env = env.Clone()

	args := []Keyed[*hm.Scheme]{}
	for _, arg := range f.Args {
		var definedArgType hm.Type

		if arg.Type_ != nil {
			// TODO should this take fresh? seems like maybe not?
			definedArgType, err = arg.Type_.Infer(env, fresh)
			if err != nil {
				return nil, fmt.Errorf("FuncDecl.Infer arg: %w", err)
			}
		}

		var inferredValType hm.Type
		if arg.Value != nil {
			inferredValType, err = arg.Value.Infer(env, fresh)
			if err != nil {
				return nil, fmt.Errorf("FuncDecl.Infer arg: %w", err)
			}
		}

		if definedArgType != nil && inferredValType != nil {
			if !definedArgType.Eq(inferredValType) {
				return nil, fmt.Errorf("FuncDecl.Infer arg: %q mismatch: defined as %s, inferred as %s", arg.Named, definedArgType, inferredValType)
			}
		} else if definedArgType != nil {
			inferredValType = definedArgType
		} else if inferredValType != nil {
			definedArgType = inferredValType
		} else {
			return nil, fmt.Errorf("FuncDecl.Infer arg: %q has no type or value", arg.Named)
		}

		scheme := hm.NewScheme(nil, definedArgType)
		env.Add(arg.Named, scheme)
		args = append(args, Keyed[*hm.Scheme]{arg.Named, scheme})
	}

	var definedRet hm.Type

	if f.Ret != nil {
		definedRet, err = f.Ret.Infer(env, fresh)
		if err != nil {
			return nil, fmt.Errorf("FuncDecl.Infer: Ret: %w", err)
		}
	}

	inferredRet, err := f.Form.Infer(env, fresh)
	if err != nil {
		return nil, fmt.Errorf("FuncDecl.Infer: Form: %w", err)
	}

	if definedRet != nil {
		// TODO: Unify?
		if !definedRet.Eq(inferredRet) {
			return nil, fmt.Errorf("FuncDecl.Infer: %q mismatch: defined as %s, inferred as %s (%[2]T != %[3]T)", f.Named, definedRet, inferredRet)
		}
	}

	return hm.NewFnType(NewRecordType("", args...), inferredRet), nil
}
