package cairo_avl

import "math/big"

type Felt = big.Int

func NewFelt(x int64) *Felt {
	return big.NewInt(x)
}
