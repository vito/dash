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
	t, _ := scheme.Type()
	return t, nil
}

func (s Symbol) Body() hm.Expression { return s }
