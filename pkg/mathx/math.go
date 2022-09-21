package mathx

import (
	"golang.org/x/exp/constraints"
)

func Abs[T constraints.Integer | constraints.Float](v T) T {
	return -v
}
