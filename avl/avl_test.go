//go:build gofuzzbeta
// +build gofuzzbeta

package avl

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var t0 Node
var t1, t2, t3 *Node

func init() {
	t1 = NewNode(big.NewInt(0), nil, nil)
	t2 = NewNode(big.NewInt(18),
		NewNode(big.NewInt(15), nil, nil), NewNode(big.NewInt(21), nil, nil),
	)
	t3 = NewNode(big.NewInt(188),
		NewNode(big.NewInt(155),
			NewNode(big.NewInt(154), nil, nil), NewNode(big.NewInt(156), nil, nil),
		),
		NewNode(big.NewInt(210),
			NewNode(big.NewInt(200),
				NewNode(big.NewInt(199), nil, nil), NewNode(big.NewInt(202), nil, nil),
			),
			NewNode(big.NewInt(300),
				NewNode(big.NewInt(201), nil, nil), NewNode(big.NewInt(1560), nil, nil),
			),
		),
	)
}

func TestWalkKeysInOrder(t *testing.T) {
	assert.Equal(t, t0.WalkKeysInOrder(), []uint64{}, "t0: in-order keys mismatch")
	assert.Equal(t, t1.WalkKeysInOrder(), []uint64{0}, "t1: in-order keys mismatch")
	assert.Equal(t, t2.WalkKeysInOrder(), []uint64{18, 15, 21}, "t2: in-order keys mismatch")
	assert.Equal(t, t3.WalkKeysInOrder(), []uint64{188, 155, 154, 156, 210, 200, 199, 202, 300, 201, 1560}, "t3: in-order keys mismatch")
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

func TestSearch(t *testing.T) {
	n1 := NewNode(big.NewInt(154), nil, nil)
	n2 := NewNode(big.NewInt(1560), nil, nil)
	t1 := NewNode(big.NewInt(188),
		NewNode(big.NewInt(155),
			n1, NewNode(big.NewInt(156), nil, nil),
		),
		NewNode(big.NewInt(210),
			NewNode(big.NewInt(200),
				NewNode(big.NewInt(199), nil, nil), NewNode(big.NewInt(202), nil, nil),
			),
			NewNode(big.NewInt(300),
				NewNode(big.NewInt(201), nil, nil), n2,
			),
		),
	)
	assert.Equal(t, Search(t1, big.NewInt(111)), (*Node)(nil), "search mismatch for unexistent node")
	assert.Equal(t, Search(t1, big.NewInt(154)), n1, "search mismatch for node: ", n1)
	assert.Equal(t, Search(t1, big.NewInt(1560)), n2, "search mismatch for node: ", n2)
}

func TestBasicOperations(t *testing.T) {
	t1 := NewNode(big.NewInt(0), nil, nil)
	t2 := Insert(t1, big.NewInt(1))
	assert.Equal(t, t2.WalkKeysInOrder(), []uint64{1, 0}, "different t2 keys")
	t3 := Insert(t1, big.NewInt(2))
	assert.Equal(t, t3.WalkKeysInOrder(), []uint64{2, 0}, "different t3 keys")
}

func TestBulkOperations(t *testing.T) {
	j1 := Join(t2, big.NewInt(50), t3)
	assert.Equal(t, j1.WalkKeysInOrder(), []uint64{188, 50, 18, 15, 21, 155, 154, 156, 210, 200, 199, 202, 300, 201, 1560}, "different j1 keys")
	GraphAndPicture(j1, "j1")

	t4 := NewNode(big.NewInt(19),
		NewNode(big.NewInt(11), nil, nil), NewNode(big.NewInt(157), nil, nil),
	)
	assert.Equal(t, t4.WalkKeysInOrder(), []uint64{19, 11, 157}, "different t4 keys")

	u1 := Union(j1, t4)
	assert.Equal(t, u1.WalkKeysInOrder(), []uint64{157, 19, 11, 18, 15, 50, 21, 155, 154, 156, 210, 188, 200, 199, 202, 300, 201, 1560}, "different u1 keys")
	GraphAndPicture(u1, "u1")

	t5 := NewNode(big.NewInt(4),
		NewNode(big.NewInt(1), nil, nil), NewNode(big.NewInt(5), nil, nil),
	)
	assert.Equal(t, t5.WalkKeysInOrder(), []uint64{4, 1, 5}, "different t5 keys")
	t6 := NewNode(big.NewInt(3),
		NewNode(big.NewInt(2), nil, nil), NewNode(big.NewInt(7), nil, nil),
	)
	assert.Equal(t, t6.WalkKeysInOrder(), []uint64{3, 2, 7}, "different t6 keys")
	u2 := Union(t5, t6)
	assert.Equal(t, u2.WalkKeysInOrder(), []uint64{3, 2, 1, 7, 4, 5}, "different u2 keys")
	GraphAndPicture(u2, "u2")

	d1 := Difference(u2, t5)
	assert.Equal(t, d1.WalkKeysInOrder(), []uint64{3, 2, 7}, "different d1 keys")
	GraphAndPicture(d1, "d1")
	d2 := Difference(u2, t6)
	assert.Equal(t, d2.WalkKeysInOrder(), []uint64{1, 5, 4}, "different d2 keys")
	GraphAndPicture(d2, "d2")

	t7 := NewNode(big.NewInt(2),
		NewNode(big.NewInt(1), nil, nil), NewNode(big.NewInt(3), nil, nil),
	)
	i1 := Intersect(u2, t7)
	assert.Equal(t, i1.WalkKeysInOrder(), []uint64{2, 1, 3}, "different i1 keys")
	GraphAndPicture(i1, "i1")
}

func TestMarshalingUnmarshaling(t *testing.T) {
	input1 := []byte(`{"key":4,"left":{"key":1},"right":{"key":5}}`)
	input2 := []byte(`{"key":3,"left":{"key":2},"right":{"key":7}}`)
	var t1, t2 Node
	json.Unmarshal(input1, &t1)
	json.Unmarshal(input2, &t2)
	output1, _ := json.Marshal(t1)
	output2, _ := json.Marshal(t2)
	assert.Equal(t, output1, input1, "unmarshal/marshal mismatch: input1=", input1, " output1=", output1)
	assert.Equal(t, output2, input2, "unmarshal/marshal mismatch: input2=", input2, " output2=", output2)
}

func TestUnion(t *testing.T) {
	input1 := []byte{67, 64, 60, 56, 231, 228, 202, 246, 241}
	input2 := []byte{56}
	var T1, T2 *Node
	for _, b1 := range input1 {
		T1 = Insert(T1, big.NewInt(int64(b1)))
	}
	assert.True(t, T1.IsBST(), "BST property failed for tree: ", T1)
	for _, b2 := range input2 {
		T2 = Insert(T2, big.NewInt(int64(b2)))
	}
	assert.True(t, T2.IsBST(), "BST property failed for tree: ", T2)

	Tu := Union(T1, T2)

	// Check BST property holds for Tu
	assert.True(t, Tu.IsBST(), "BST property failed for tree: ", Tu)
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
			T1 = Insert(T1, big.NewInt(int64(b1)))
		}
		assert.True(t, T1.IsBST(), "BST property failed for tree: ", T1)
		for _, b2 := range input2 {
			T2 = Insert(T2, big.NewInt(int64(b2)))
		}
		assert.True(t, T2.IsBST(), "BST property failed for tree: ", T2)

		Tu := Union(T1, T2)

		// Check BST property holds for Tu
		assert.True(t, Tu.IsBST(), "BST property failed for tree: ", Tu)
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
			T1 = Insert(T1, big.NewInt(int64(b1)))
		}
		assert.True(t, T1.IsBST(), "BST property failed for tree: ", T1)
		for _, b2 := range input2 {
			T2 = Insert(T2, big.NewInt(int64(b2)))
		}
		assert.True(t, T2.IsBST(), "BST property failed for tree: ", T2)

		Td := Difference(T1, T2)

		// Check BST property holds for Td
		assert.True(t, Td.IsBST(), "BST property failed for tree: ", Td)
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
}
