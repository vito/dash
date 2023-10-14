package dash

import (
	"fmt"

	"github.com/chewxy/hm"
)

type Default struct {
	Left  Node
	Right Node
}

var _ Node = Default{}

func (d Default) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	lt, err := d.Left.Infer(env, fresh)
	if err != nil {
		return nil, err
	}
	rt, err := d.Right.Infer(env, fresh)
	if err != nil {
		return nil, err
	}
	lt = NonNullType{lt}
	if !lt.Eq(rt) {
		return nil, fmt.Errorf("Default.Infer: mismatched types: %s != %s", lt, rt)
	}
	return rt, nil
}

func (d Default) Body() hm.Expression { return d }
