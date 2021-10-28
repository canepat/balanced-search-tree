//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_avl

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertAvl(t *testing.T, n *Node, h int, expectedKeysInOrder []uint64) {
	assert.True(t, n.IsBST(), "BST property failed for tree: %v", n.WalkKeysInOrder())
	assert.True(t, n.IsBalanced(), "AVL balance property failed for tree: %v", n.WalkKeysInOrder())
	assert.Equal(t, h, HeightAsInt(n), "Wrong height %d for tree: %v", HeightAsInt(n), n.WalkKeysInOrder())
	assert.Equal(t, expectedKeysInOrder, n.WalkKeysInOrder(), "different in-order keys: %v", n.WalkKeysInOrder())
}

//var ct0 Node
var ct1, ct2, ct3 *Node
var du, dd *Dict

func init() {
	ct1 = NewNode(big.NewInt(0), big.NewInt(0), nil, nil, nil)
	ct2 = NewNode(big.NewInt(18), big.NewInt(0),
		NewNode(big.NewInt(15), big.NewInt(0), nil, nil, nil), NewNode(big.NewInt(21), big.NewInt(0), nil, nil, nil), nil,
	)
	ct3 = NewNode(big.NewInt(188), big.NewInt(0),
		NewNode(big.NewInt(155), big.NewInt(0),
			NewNode(big.NewInt(154), big.NewInt(0), nil, nil, nil), NewNode(big.NewInt(156), big.NewInt(0), nil, nil, nil), nil,
		),
		NewNode(big.NewInt(210), big.NewInt(0),
			NewNode(big.NewInt(200), big.NewInt(0),
				NewNode(big.NewInt(199), big.NewInt(0), nil, nil, nil), NewNode(big.NewInt(202), big.NewInt(0), nil, nil, nil), nil,
			),
			NewNode(big.NewInt(300), big.NewInt(0),
				NewNode(big.NewInt(211), big.NewInt(0), nil, nil, nil), NewNode(big.NewInt(1560), big.NewInt(0), nil, nil, nil), nil,
			),
			nil,
		),
		nil,
	)
	du = NewDict(NewFelt(212), NewFelt(10), nil, nil, nil, nil)
	dd = NewDict(NewFelt(21), NewFelt(10), nil, nil, nil, nil)
}

func TestComputeHeight(t *testing.T) {
	assert.Equal(t, big.NewInt(1), computeHeight(big.NewInt(0), big.NewInt(0)), "Wrong computed height (expected: 1)")
	assert.Equal(t, big.NewInt(2), computeHeight(big.NewInt(0), big.NewInt(1)), "Wrong computed height (expected: 2)")
	assert.Equal(t, big.NewInt(2), computeHeight(big.NewInt(1), big.NewInt(0)), "Wrong computed height (expected: 2)")
}

func TestBulkOperations(t *testing.T) {
	t7 := NewNode(NewFelt(18), NewFelt(0),
		NewNode(NewFelt(15), NewFelt(0), nil, nil, nil), nil, nil,
	)
	assertAvl(t, t7, 2, []uint64{18, 15})
	t7.GraphAndPicture("t7", /*debug=*/false)

	d8 := NewDict(NewFelt(11), NewFelt(0), nil, nil, nil, nil)

	u3 := Union(t7, d8, &Counters{})
	assertAvl(t, u3, 2, []uint64{15, 11, 18})
	u3.GraphAndPicture("u3", /*debug=*/false)
}

func TestStateTree(t *testing.T) {
	//st := NewNode(NewFelt(0), nil, nil, nil, NewNode(NewFelt(0)))
	//GraphAndPicture(st, "st")
}
