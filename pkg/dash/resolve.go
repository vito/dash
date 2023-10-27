package dash

import (
	"fmt"

	"github.com/chewxy/hm"
)

type Resolve struct {
	Receiver Node
	Field    string
	Args     Record
}

var _ Node = Resolve{}

func (d Resolve) Infer(env hm.Env, fresh hm.Fresher) (hm.Type, error) {
	lt, err := d.Receiver.Infer(env, fresh)
	if err != nil {
		return nil, err
	}
	nn, ok := lt.(NonNullType)
	if !ok {
		return nil, fmt.Errorf("Select.Infer: expected %T, got %T: %s", nn, lt, lt)
	}
	rec, ok := nn.Type.(*Module)
	if !ok {
		return nil, fmt.Errorf("Select.Infer: expected %T, got %T", rec, nn.Type)
	}
	scheme, found := rec.SchemeOf(d.Field)
	if !found {
		return nil, fmt.Errorf("Select.Infer: field %q not found in record %s", d.Field, rec)
	}
	t, mono := scheme.Type()
	if !mono {
		return nil, fmt.Errorf("Select.Infer: type of field %q is not monomorphic", d.Field)
	}
	switch x := t.(type) {
	case *hm.FunctionType:
		// resolving always calls the function
		// TODO: check args (presence + types)
		return x.Ret(true), nil
	}
	return t, nil
}

func (d Resolve) Body() hm.Expression { return d }
