//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_bptree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertTwoThreeTree(t *testing.T, tree *Tree23, expectedKeysPostOrder []Felt) {
	assert.True(t, tree.IsTwoThree(), "2-3-tree properties do not hold for tree: %v", tree.WalkKeysPostOrder())
	if expectedKeysPostOrder != nil {
		assert.Equal(t, expectedKeysPostOrder, tree.WalkKeysPostOrder(), "different post-order keys: %v", tree.WalkKeysPostOrder())
	}
}

type IsTree23Test struct {
	initialItems		[]KeyValue
	expectedKeysPostOrder	[]Felt
}

type UpsertTest struct {
	initialItems		[]KeyValue
	initialKeysPostOrder	[]Felt
	deltaItems		[]KeyValue
	finalKeysPostOrder	[]Felt
}

var isTree23Tests = []IsTree23Test {
	{[]KeyValue{},							[]Felt{}},
	{[]KeyValue{{1, 1}},						[]Felt{1}},
	{[]KeyValue{{1, 1}, {2, 2}},					[]Felt{1, 2}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},				[]Felt{1, 2, 3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},			[]Felt{1, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},		[]Felt{1, 2, 3, 4, 5}},
	{[]KeyValue{{0, 0}, {1, 1}, {10, 10}, {100, 100}, {101, 101}},	[]Felt{0, 1, 10, 100, 101}},
}

var insertTests = []UpsertTest {
	{[]KeyValue{},					[]Felt{},		[]KeyValue{{0, 0}},				[]Felt{0}},
	{[]KeyValue{{10, 10}},				[]Felt{10},		[]KeyValue{{0, 0}, {5, 5}},			[]Felt{0, 5, 10}},
	{[]KeyValue{{10, 10}, {20, 20}},		[]Felt{10, 20},		[]KeyValue{{0, 0}, {5, 5}, {15, 15}},		[]Felt{0, 5, 10, 15, 20}},
	{[]KeyValue{{10, 10}, {20, 20}, {30, 30}},	[]Felt{10, 20, 30},	[]KeyValue{{0, 0}, {5, 5}, {15, 15}, {25, 25}},	[]Felt{0, 5, 10, 15, 20, 25, 30}},
}

var updateTests = []UpsertTest {
	{[]KeyValue{{10, 10}},				[]Felt{10},		[]KeyValue{{10, 100}},			[]Felt{10}},
	{[]KeyValue{{10, 10}, {20, 20}},		[]Felt{10, 20},		[]KeyValue{{10, 100}, {20, 200}},	[]Felt{10, 20}},
}

func TestIs23Tree(t *testing.T) {
	for _, data := range isTree23Tests {
		tree := NewTree23().Upsert(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysPostOrder)
	}
}

func TestUpsertInsert(t *testing.T) {
	for _, data := range insertTests {
		tree := NewTree23().Upsert(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysPostOrder)
		tree.Upsert(data.deltaItems)
		assertTwoThreeTree(t, tree, data.finalKeysPostOrder)
	}
}

func TestUpsertUpdate(t *testing.T) {
	for _, data := range updateTests {
		tree := NewTree23().Upsert(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysPostOrder)
		tree.Upsert(data.deltaItems)
		assertTwoThreeTree(t, tree, data.finalKeysPostOrder)
		// TODO: add check for new values
	}
}

func TestUpsertIdempotent(t *testing.T) {
	for _, data := range isTree23Tests {
		tree := NewTree23().Upsert(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysPostOrder)
		tree.Upsert(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysPostOrder)
	}
}

func TestUpsertNextKey(t *testing.T) {
	dataCount := 4
	data := make([]KeyValue, dataCount)
	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2), Felt(i*2)}
	}
	tn := NewTree23().Upsert(data)
	tn.GraphAndPicture("tn1", false)

	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2+1), Felt(i*2+1)}
	}
	tn = tn.Upsert(data)
	tn.GraphAndPicture("tn2", false)
	assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7})
	
	data = []KeyValue{{100, 100}, {101, 101}, {200, 200}, {201, 201}, {202, 202}}
	tn = tn.Upsert(data)
	tn.GraphAndPicture("tn3", false)
	assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7, 100, 101, 200, 201, 202})
	
	data = []KeyValue{{10, 10}, {150, 150}, {250, 250}, {251, 251}, {252, 252}}
	tn = tn.Upsert(data)
	tn.GraphAndPicture("tn4", false)
	assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7, 10, 100, 101, 150, 200, 201, 202, 250, 251, 252})
}

func TestUpsertFirstKey(t *testing.T) {
}
