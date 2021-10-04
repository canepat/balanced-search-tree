package cairo_avl

import "math/big"

func MaxBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return b
	}
	return a
}
