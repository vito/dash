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
		return nil, fmt.Errorf("Resolve.Infer: expected %T, got %T: %s", nn, lt, lt)
	}
	rec, ok := nn.Type.(*Module)
	if !ok {
		return nil, fmt.Errorf("Resolve.Infer: expected %T, got %T", rec, nn.Type)
	}
	scheme, found := rec.SchemeOf(d.Field)
	if !found {
		return nil, fmt.Errorf("Resolve.Infer: field %q not found in record %s", d.Field, rec)
	}
	t, mono := scheme.Type()
	if !mono {
		return nil, fmt.Errorf("Resolve.Infer: type of field %q is not monomorphic", d.Field)
	}
	switch x := t.(type) {
	case *hm.FunctionType:
		definedType := x.Arg()

		inferredType, err := d.Args.Infer(env, fresh)
		if err != nil {
			return nil, err
		}

		definedRecord, ok := definedType.(*RecordType)
		if !ok {
			return nil, fmt.Errorf("Resolve.Infer: expected %T, got %T", definedRecord, definedType)
		}

		inferredRecord, ok := inferredType.(*RecordType)
		if !ok {
			return nil, fmt.Errorf("Resolve.Infer: expected %T, got %T", definedRecord, definedType)
		}

		definedTypes := map[string]*hm.Scheme{}
		for _, f := range definedRecord.Fields {
			definedTypes[f.Key] = f.Value
		}

		inferredTypes := map[string]*hm.Scheme{}
		for _, f := range inferredRecord.Fields {
			inferredTypes[f.Key] = f.Value
		}

		for k, v := range definedTypes {
			definedT, isMono := v.Type()
			if !isMono {
				// TODO should this just be type?
				return nil, fmt.Errorf("Resolve.Infer: field %q is not monomorphic", k)
			}

			inferred, ok := inferredTypes[k]
			if !ok {
				if _, isRequired := definedT.(NonNullType); isRequired {
					return nil, fmt.Errorf("Resolve.Infer: field %q is required", k)
				} else {
					// optional; skip
					continue
				}
			}

			inferredT, isMono := inferred.Type()
			if !isMono {
				// TODO should this just be type?
				return nil, fmt.Errorf("Resolve.Infer: field %q is not monomorphic", k)
			}

			if !definedT.Eq(inferredT) {
				return nil, fmt.Errorf("Resolve.Infer: mismatched types: %s != %s", definedT, inferredT)
			}

			delete(inferredTypes, k)
		}

		if len(inferredTypes) > 0 {
			return nil, fmt.Errorf("Resolve.Infer: unexpected fields: %v", inferredTypes)
		}

		// resolving always calls the function
		return x.Ret(true), nil
	}
	return t, nil
}
