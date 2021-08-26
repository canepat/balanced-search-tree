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

func TestHasBinarySearchTreeProperty(t *testing.T) {
	trees := make([]*Node, 0)
	trees = append(trees, &t0, t1, t2, t3)
	for _, tree := range trees {
		assert.True(t, tree.HasBinarySearchTreeProperty(), "BST property failed for tree: ", tree)
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
	input1 := []byte(`{"key":4,"left":{"key":1},"right":{"key":5}}`)
	input2 := []byte(`{"key":3,"left":{"key":2},"right":{"key":7}}`)
	var T1, T2 Node
	if err1 := json.Unmarshal(input1, &T1); err1 != nil {
		t.Fatal("JSON unmarshaling failed for ", string(input1))
	}
	if err2 := json.Unmarshal(input2, &T2); err2 != nil {
		t.Fatal("JSON unmarshaling failed for ", string(input2))
	}
	if !T1.HasBinarySearchTreeProperty() {
		t.Fatal("BST property failed for tree: ", string(input1))
	}
	if !T2.HasBinarySearchTreeProperty() {
		t.Fatal("BST property failed for tree: ", string(input2))
	}
	Tu := Union(&T1, &T2)
	assert.True(t, Tu.HasBinarySearchTreeProperty(), "BST property failed for tree: ", Tu)
	for _, n1 := range T1.WalkNodesInOrder() {
		search_result := Search(Tu, n1.Key)
		assert.NotNil(t, search_result, "search failed for n1: ", n1.Key)
		assert.Equal(t, search_result.Key, n1.Key, "search for n1: ", n1.Key, " returned: ", search_result.Key)
	}
	for _, n2 := range T2.WalkNodesInOrder() {
		search_result := Search(Tu, n2.Key)
		assert.NotNil(t, search_result, "search failed for n2: ", n2.Key.Uint64())
		assert.Equal(t, search_result.Key, n2.Key, "search for n2: ", n2.Key.Uint64(), " returned: ", search_result.Key.Uint64())
	}
}

func FuzzUnion(f *testing.F) {
	f.Add([]byte(`{"key":4,"left":{"key":1},"right":{"key":5}}`), []byte(`{"key":3,"left":{"key":2},"right":{"key":7}}`))
	f.Add([]byte(`{"key":4}`), []byte(`{}`))
	f.Add([]byte(`{}`), []byte(`{"key":4}`))
	f.Add([]byte(`{}`), []byte(`{}`))
	f.Add([]byte(`{"key":3}`), []byte(`{"right":{}}`))
	f.Fuzz(func (t *testing.T, input1 []byte, input2 []byte) {
		t.Parallel()
		var T1, T2 Node
		if err1 := json.Unmarshal(input1, &T1); err1 != nil {
			t.Skip()
		}
		if err2 := json.Unmarshal(input2, &T2); err2 != nil {
			t.Skip()
		}
		if !T1.HasBinarySearchTreeProperty() || !T2.HasBinarySearchTreeProperty() {
			t.Skip()
		}
		Tu := Union(&T1, &T2)
		// Check BST property holds for Tu
		assert.True(t, Tu.HasBinarySearchTreeProperty(), "BST property failed for tree: ", Tu)
		// Check that *all* T1 nodes are present in Tu
		for _, n1 := range T1.WalkNodesInOrder() {
			assert.Equal(t, Search(Tu, n1.Key).Key, n1.Key)
		}
		// Check that *all* T2 nodes are present in Tu
		for _, n2 := range T2.WalkNodesInOrder() {
			assert.Equal(t, Search(Tu, n2.Key).Key, n2.Key)
		}
	})
}
