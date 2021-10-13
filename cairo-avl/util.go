package cairo_avl

import "math/big"

const BufferSize uint = 4096

func MaxBigInt(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return b
	}
	return a
}
