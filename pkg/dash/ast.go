package dash

import (
	"github.com/chewxy/hm"
)

type Node interface {
	hm.Expression
	hm.Inferer
}

type Keyed[X any] struct {
	Key   string
	Value X
}

type Visibility int

const (
	PublicVisibility Visibility = iota
	PrivateVisibility
)

// TODO record literals?
