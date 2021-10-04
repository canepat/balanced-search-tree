//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_avl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*func assertAvl(t *testing.T, n *Node, h int, expectedKeysInOrder []uint64) {
	assert.True(t, n.IsBST(), "BST property failed for tree: %v", n.WalkKeysInOrder())
	assert.True(t, n.IsBalanced(), "AVL balance property failed for tree: %v", n.WalkKeysInOrder())
	assert.Equal(t, h, HeightAsInt(n), "Wrong height %d for tree: %v", HeightAsInt(n), n.WalkKeysInOrder())
	assert.Equal(t, expectedKeysInOrder, n.WalkKeysInOrder(), "different in-order keys: %v", n.WalkKeysInOrder())
}*/

var t0 Node
var t1, t2, t3 *Node

func init() {
	t1 = NewNode(NewFelt(0), NewFelt(0), nil, nil, nil)
	t2 = NewNode(NewFelt(18), NewFelt(0),
		NewNode(NewFelt(15), NewFelt(0), nil, nil, nil), NewNode(NewFelt(21), NewFelt(0), nil, nil, nil), nil,
	)
	t3 = NewNode(NewFelt(188), NewFelt(0),
		NewNode(NewFelt(155), NewFelt(0),
			NewNode(NewFelt(154), NewFelt(0), nil, nil, nil), NewNode(NewFelt(156), NewFelt(0), nil, nil, nil), nil,
		),
		NewNode(NewFelt(210), NewFelt(0),
			NewNode(NewFelt(200), NewFelt(0),
				NewNode(NewFelt(199), NewFelt(0), nil, nil, nil), NewNode(NewFelt(202), NewFelt(0), nil, nil, nil), nil,
			),
			NewNode(NewFelt(300), NewFelt(0),
				NewNode(NewFelt(211), NewFelt(0), nil, nil, nil), NewNode(NewFelt(1560), NewFelt(0), nil, nil, nil), nil,
			),
			nil,
		),
		nil,
	)
}

func TestWalkKeysInOrder(t *testing.T) {
	assert.Equal(t, t0.WalkKeysInOrder(), []uint64{}, "t0: in-order keys mismatch")
	assert.Equal(t, t1.WalkKeysInOrder(), []uint64{0}, "t1: in-order keys mismatch")
	assert.Equal(t, t2.WalkKeysInOrder(), []uint64{18, 15, 21}, "t2: in-order keys mismatch")
	assert.Equal(t, t3.WalkKeysInOrder(), []uint64{188, 155, 154, 156, 210, 200, 199, 202, 300, 211, 1560}, "t3: in-order keys mismatch")
}

func TestWalkPathsInOrder(t *testing.T) {
	assert.Equal(t, []string{}, t0.WalkPathsInOrder(), "t0: in-order paths mismatch")
	assert.Equal(t, []string{"M"}, t1.WalkPathsInOrder(), "t1: in-order paths mismatch")
	assert.Equal(t, []string{"M", "LM", "RM"}, t2.WalkPathsInOrder(), "t2: in-order paths mismatch")
	assert.Equal(t, []string{"M", "LM", "LLM", "LRM", "RM", "RLM", "RLLM", "RLRM", "RRM", "RRLM", "RRRM"}, t3.WalkPathsInOrder(), "t3: in-order paths mismatch")
}

func TestIsBST(t *testing.T) {
	trees := make([]*Node, 0)
	trees = append(trees, nil, &t0, t1, t2, t3)
	for _, tree := range trees {
		assert.True(t, tree.IsBST(), "BST property failed for tree: ", tree)
	}
}

func TestHeight(t *testing.T) {
	assert.Equal(t, 0, HeightAsInt(&t0), "Wrong height %d (expected: 0) for tree: %v", HeightAsInt(&t0), t0.WalkKeysInOrder())
	assert.Equal(t, 1, HeightAsInt(t1), "Wrong height %d (expected: 1) for tree: %v", HeightAsInt(t1), t1.WalkKeysInOrder())
	assert.Equal(t, 2, HeightAsInt(t2), "Wrong height %d (expected: 2) for tree: %v", HeightAsInt(t2), t2.WalkKeysInOrder())
	assert.Equal(t, 4, HeightAsInt(t3), "Wrong height %d (expected: 4) for tree: %v", HeightAsInt(t2), t2.WalkKeysInOrder())
}

/*func TestSearch(t *testing.T) {
	n1 := NewNode(NewFelt(154), NewFelt(0), nil, nil)
	n2 := NewNode(NewFelt(1560), NewFelt(0), nil, nil)
	bst1 := NewNode(NewFelt(188), NewFelt(0),
		NewNode(NewFelt(155), NewFelt(0),
			n1, NewNode(NewFelt(156), NewFelt(0), nil, nil),
		),
		NewNode(NewFelt(210), NewFelt(0),
			NewNode(NewFelt(200), NewFelt(0),
				NewNode(NewFelt(199), NewFelt(0), nil, nil), NewNode(NewFelt(202), NewFelt(0), nil, nil),
			),
			NewNode(NewFelt(300), NewFelt(0),
				NewNode(NewFelt(201), NewFelt(0), nil, nil), n2,
			),
		),
	)
	assert.Equal(t, Search(bst1, NewFelt(111)), (*Node)(nil), "search mismatch for unexistent node")
	assert.Equal(t, Search(bst1, NewFelt(154)), n1, "search mismatch for node: ", n1)
	assert.Equal(t, Search(bst1, NewFelt(1560)), n2, "search mismatch for node: ", n2)
}

func TestBasicOperations(t *testing.T) {
	bst1 := NewNode(NewFelt(0), NewFelt(0), nil, nil)
	bst2 := Insert(bst1, NewFelt(1), NewFelt(0))
	assert.Equal(t, []uint64{1, 0}, bst2.WalkKeysInOrder(), "different t2 keys")
	bst3 := Insert(bst1, NewFelt(2), NewFelt(0))
	assert.Equal(t, []uint64{2, 0}, bst3.WalkKeysInOrder(), "different t3 keys")
}

func TestSpineInsertion(t *testing.T) {
	var bst1 Node
	assert.Equal(t, bst1.WalkKeysInOrder(), []uint64{}, "different bst1 keys")
	bst2 := Insert(&bst1, NewFelt(1), NewFelt(0), nil, nil)
	assert.Equal(t, bst2.WalkKeysInOrder(), []uint64{1}, "different bst2 keys")
	bst3 := Insert(bst2, NewFelt(2), NewFelt(0), nil, nil)
	assert.Equal(t, []uint64{2, 1}, bst3.WalkKeysInOrder(), "different bst3 keys")
	bst4 := Insert(bst3, NewFelt(3), NewFelt(0), nil, nil)
	assert.Equal(t, []uint64{2, 1, 3}, bst4.WalkKeysInOrder(), "different bst4 keys")
	bst5 := Insert(bst4, NewFelt(4), NewFelt(0), nil, nil)
	assert.Equal(t, []uint64{2, 1, 4, 3}, bst5.WalkKeysInOrder(), "different bst5 keys")
	bst6 := Insert(bst5, NewFelt(5), NewFelt(0), nil, nil)
	assert.Equal(t, []uint64{2, 1, 4, 3, 5}, bst6.WalkKeysInOrder(), "different bst6 keys")
	bst7 := Insert(bst6, NewFelt(6), NewFelt(0), nil, nil)
	assert.Equal(t, []uint64{4, 2, 1, 3, 6, 5}, bst7.WalkKeysInOrder(), "different bst7 keys")
}

func TestBulkOperations(t *testing.T) {
	j1 := Join(t2, NewFelt(50), NewFelt(0), t3)
	assertAvl(t, j1, 4, []uint64{188, 50, 18, 15, 21, 155, 154, 156, 210, 200, 199, 202, 300, 211, 1560})
	GraphAndPicture(j1, "j1")

	t4 := NewNode(NewFelt(19), NewFelt(0),
		NewNode(NewFelt(11), NewFelt(0), nil, nil), NewNode(NewFelt(157), NewFelt(0), nil, nil),
	)
	assertAvl(t, t4, 2, []uint64{19, 11, 157})
	GraphAndPicture(t4, "t4")

	u1 := Union(j1, t4)
	assertAvl(t, u1, 5, []uint64{157, 19, 15, 11, 18, 50, 21, 155, 154, 156, 210, 200, 188, 199, 202, 300, 211, 1560})
	GraphAndPicture(u1, "u1")

	t5 := NewNode(NewFelt(4), NewFelt(0),
		NewNode(NewFelt(1), NewFelt(0), nil, nil), NewNode(NewFelt(5), NewFelt(0), nil, nil),
	)
	assertAvl(t, t5, 2, []uint64{4, 1, 5})
	GraphAndPicture(t5, "t5")

	t6 := NewNode(NewFelt(3), NewFelt(0),
		NewNode(NewFelt(2), NewFelt(0), nil, nil), NewNode(NewFelt(7), NewFelt(0), nil, nil),
	)
	assertAvl(t, t6, 2, []uint64{3, 2, 7})
	GraphAndPicture(t6, "t6")

	u2 := Union(t5, t6)
	assertAvl(t, u2, 3, []uint64{3, 2, 1, 5, 4, 7})
	GraphAndPicture(u2, "u2")

	t7 := NewNode(NewFelt(18), NewFelt(0),
		NewNode(NewFelt(15), NewFelt(0), nil, nil), nil,
	)
	assertAvl(t, t7, 2, []uint64{18, 15})
	GraphAndPicture(t7, "t7")

	t8 := NewNode(NewFelt(11), NewFelt(0), nil, nil)
	assertAvl(t, t8, 1, []uint64{11})
	GraphAndPicture(t8, "t8")

	u3 := Union(t7, t8)
	assertAvl(t, u3, 2, []uint64{15, 11, 18})
	GraphAndPicture(u3, "u3")

	d1 := Difference(u2, t5)
	assertAvl(t, d1, 2, []uint64{3, 2, 7})
	GraphAndPicture(d1, "d1")

	d2 := Difference(u2, t6)
	assertAvl(t, d2, 2, []uint64{4, 1, 5})
	GraphAndPicture(d2, "d2")

	t9 := NewNode(NewFelt(2), NewFelt(0),
		NewNode(NewFelt(1), NewFelt(0), nil, nil), NewNode(NewFelt(3), NewFelt(0), nil, nil),
	)
	assertAvl(t, t9, 2, []uint64{2, 1, 3})
	GraphAndPicture(t9, "t9")

	i1 := Intersect(u2, t9)
	assertAvl(t, i1, 2, []uint64{2, 1, 3})
	GraphAndPicture(i1, "i1")

	i2 := Intersect(u2, t1)
	assertAvl(t, i2, 0, []uint64{})
	GraphAndPicture(i2, "i2")
}

func TestUnion(t *testing.T) {
	input1 := []byte{67, 64, 60, 56, 231, 228, 202, 246, 241}
	input2 := []byte{56}
	var T1, T2 *Node
	for _, b1 := range input1 {
		T1 = Insert(T1, NewFelt(int64(b1)), NewFelt(0))
	}
	assertAvl(t, T1, 4, []uint64{202, 64, 56, 60, 67, 241, 231, 228, 246})
	for _, b2 := range input2 {
		T2 = Insert(T2, NewFelt(int64(b2)), NewFelt(0))
	}
	assertAvl(t, T2, 1, []uint64{56})

	Tu := Union(T1, T2)

	// Check Tu is an AVL tree
	assertAvl(t, Tu, 4, []uint64{202, 64, 56, 60, 67, 241, 231, 228, 246})
	// Check that *all* T1 nodes are present in Tu
	for _, n1 := range T1.WalkNodesInOrder() {
		search_result := Search(Tu, n1.Key)
		assert.NotNil(t, search_result, "search failed for n1: ", n1.Key)
		assert.Equal(t, search_result.Key, n1.Key, "search for n1: ", n1.Key, " returned: ", search_result.Key)
	}
	// Check that *all* T2 nodes are present in Tu
	for _, n2 := range T2.WalkNodesInOrder() {
		search_result := Search(Tu, n2.Key)
		assert.NotNil(t, search_result, "search failed for n2: ", n2.Key.Uint64())
		assert.Equal(t, search_result.Key, n2.Key, "search for n2: ", n2.Key.Uint64(), " returned: ", search_result.Key.Uint64())
	}
	// Check that *all* Tu nodes are present either in T1 or T2
	for _, n := range Tu.WalkNodesInOrder() {
		result1 := Search(T1, n.Key)
		if result1 == nil || result1.Key.Cmp(n.Key) != 0 {
			result2 := Search(T2, n.Key)
			if result2 == nil || result2.Key.Cmp(n.Key) != 0 {
				t.Fatalf("search for n: %d failed both in T1 and T2", n.Key.Uint64())
			}
		}
	}
}

func FuzzUnion(f *testing.F) {
	f.Fuzz(func (t *testing.T, input1 []byte, input2 []byte) {
		t.Parallel()
		var T1, T2 *Node
		for _, b1 := range input1 {
			T1 = Insert(T1, NewFelt(int64(b1)), NewFelt(0))
		}
		assert.True(t, T1.IsBST(), "BST property failed for tree: ", T1)
		for _, b2 := range input2 {
			T2 = Insert(T2, NewFelt(int64(b2)), NewFelt(0))
		}
		assert.True(t, T2.IsBST(), "BST property failed for tree: ", T2)

		Tu := Union(T1, T2)

		// Check BST property holds for Tu
		assert.True(t, Tu.IsBST(), "BST property failed for tree: ", Tu)
		assert.True(t, Tu.IsBalanced(), "AVL balance property failed for tree: %v", Tu.WalkKeysInOrder())
		// Check that *each* T1 node is present in Tu
		for _, n1 := range T1.WalkNodesInOrder() {
			result := Search(Tu, n1.Key)
			assert.NotNil(t, result, "search failed for n1: ", n1.Key)
			assert.Equal(t, result.Key, n1.Key, "search for n1: ", n1.Key, " returned: ", result.Key)
		}
		// Check that *each* T2 node is present in Tu
		for _, n2 := range T2.WalkNodesInOrder() {
			result := Search(Tu, n2.Key)
			assert.NotNil(t, result, "search failed for n2: ", n2.Key.Uint64())
			assert.Equal(t, result.Key, n2.Key, "search for n2: ", n2.Key.Uint64(), " returned: ", result.Key.Uint64())
		}
		// Check that *each* Tu node is present either in T1 or T2
		for _, n := range Tu.WalkNodesInOrder() {
			result1 := Search(T1, n.Key)
			if result1 == nil || result1.Key.Cmp(n.Key) != 0 {
				result2 := Search(T2, n.Key)
				if result2 == nil || result2.Key.Cmp(n.Key) != 0 {
					t.Fatalf("search for n: %d failed both in T1 and T2", n.Key.Uint64())
				}
			}
		}
	})
}

func FuzzDifference(f *testing.F) {
	f.Fuzz(func (t *testing.T, input1 []byte, input2 []byte) {
		t.Parallel()
		var T1, T2 *Node
		for _, b1 := range input1 {
			T1 = Insert(T1, NewFelt(int64(b1)), NewFelt(0))
		}
		assert.True(t, T1.IsBST(), "BST property failed for tree: ", T1)
		for _, b2 := range input2 {
			T2 = Insert(T2, NewFelt(int64(b2)), NewFelt(0))
		}
		assert.True(t, T2.IsBST(), "BST property failed for tree: ", T2)

		Td := Difference(T1, T2)

		// Check BST property holds for Td
		assert.True(t, Td.IsBST(), "BST property failed for tree: ", Td)
		assert.True(t, Td.IsBalanced(), "AVL balance property failed for tree: %v", Td.WalkKeysInOrder())
		// Check that *each* T1 node is present either in Td or in T2
		for _, n1 := range T1.WalkNodesInOrder() {
			result := Search(Td, n1.Key)
			if result == nil || result.Key.Cmp(n1.Key) != 0 {
				result2 := Search(T2, n1.Key)
				if result2 == nil || result2.Key.Cmp(n1.Key) != 0 {
					t.Fatalf("search for n1: %d failed both in Td and T2", n1.Key.Uint64())
				}
			}
		}
		// Check that *each* T2 node is not present in Td
		for _, n2 := range T2.WalkNodesInOrder() {
			result := Search(Td, n2.Key)
			assert.Nil(t, result, "search successful for n2: ", n2.Key.Uint64())
		}
		// Check that *each* Td node is present in T1 but not in T2
		for _, n := range Td.WalkNodesInOrder() {
			result1 := Search(T1, n.Key)
			if result1 == nil || result1.Key.Cmp(n.Key) != 0 {
				t.Fatalf("search for n: %d failed in T1", n.Key.Uint64())
			}
			result2 := Search(T2, n.Key)
			if result2 != nil && result2.Key.Cmp(n.Key) == 0 {
				t.Fatalf("search for n: %d successful in T2", n.Key.Uint64())
			}
		}
	})
}*/
