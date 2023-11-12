package dash

import (
	"fmt"

	"github.com/chewxy/hm"
)

type List struct {
	Elements []Node
}

var _ Node = List{}

func (l List) Infer(env hm.Env, f hm.Fresher) (hm.Type, error) {
	if len(l.Elements) == 0 {
		// TODO: is this right?
		return NonNullType{ListType{f.Fresh()}}, nil
	}

	var t hm.Type
	for i, el := range l.Elements {
		et, err := el.Infer(env, f)
		if err != nil {
			return nil, err
		}
		if t == nil {
			t = et
		} else if _, err := hm.Unify(t, et); err != nil {
			// TODO: is this right?
			return nil, fmt.Errorf("unify index %d: %w", i, err)
		}
	}
	return NonNullType{ListType{t}}, nil
}
