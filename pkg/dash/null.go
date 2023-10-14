package dash

import "github.com/chewxy/hm"

type Null struct{}

var _ Node = Null{}

func (n Null) Body() hm.Expression { return n }

func (Null) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	// Null does not have a type. Its type is always inferred
	// as a free variable.
	return fresh.Fresh(), nil
}
