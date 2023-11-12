package dash

import (
	"errors"

	"github.com/chewxy/hm"
)

type Block struct {
	Forms []Node
}

type Hoister interface {
	Hoist(hm.Env, hm.Fresher, int) error
}

var _ Hoister = Block{}

type Set[T comparable] map[T]struct{}

func (b Block) Hoist(env hm.Env, fresh hm.Fresher, depth int) error {
	var errs []error
	for _, form := range b.Forms {
		if hoister, ok := form.(Hoister); ok {
			if err := hoister.Hoist(env, fresh, depth); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

var _ hm.Inferer = Block{}

func (b Block) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	forms := b.Forms
	if len(forms) == 0 {
		forms = append(forms, Null{})
	}

	var t hm.Type
	for _, form := range forms {
		et, err := form.Infer(env, fresh)
		if err != nil {
			return nil, err
		}
		t = et
	}

	return t, nil
}
