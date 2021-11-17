//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_bptree

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func assertTwoThreeTree(t *testing.T, tree *Tree23, expectedKeysLevelOrder []Felt) {
	assert.True(t, tree.IsTwoThree(), "2-3-tree properties do not hold for tree: %v", tree.WalkKeysPostOrder())
	if expectedKeysLevelOrder != nil {
		assert.Equal(t, expectedKeysLevelOrder, tree.KeysInLevelOrder(), "different keys by level")
	}
}

type HeightTest struct {
	initialItems	[]KeyValue
	expectedHeight	int
}

type IsTree23Test struct {
	initialItems		[]KeyValue
	expectedKeysLevelOrder	[]Felt
}

type UpsertTest struct {
	initialItems		[]KeyValue
	initialKeysPostOrder	[]Felt
	deltaItems		[]KeyValue
	finalKeysPostOrder	[]Felt
}

var heightTestTable = []HeightTest {
	{[]KeyValue{},									0},
	{[]KeyValue{{1, 1}},								1},
	{[]KeyValue{{1, 1}, {2, 2}},							1},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},						2},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},					2},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},				2},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}},			2},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}},		3},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}},	3},
}

var isTree23TestTable = []IsTree23Test {
	{[]KeyValue{},								[]Felt{}},
	{[]KeyValue{{1, 1}},							[]Felt{1}},
	{[]KeyValue{{1, 1}, {2, 2}},						[]Felt{1, 2}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},					[]Felt{3, 1, 2, 3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},				[]Felt{3, 1, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},			[]Felt{3, 5, 1, 2, 3, 4, 5}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}},		[]Felt{3, 5, 1, 2, 3, 4, 5, 6}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}},	[]Felt{5, 3, 7, 1, 2, 3, 4, 5, 6, 7}},
	{[]KeyValue{{0, 0}, {1, 1}, {10, 10}, {100, 100}, {101, 101}},		[]Felt{10, 101, 0, 1, 10, 100, 101}},
}

var insertTestTable = []UpsertTest {
	{[]KeyValue{},					[]Felt{},		[]KeyValue{{0, 0}},				[]Felt{0}},
	{[]KeyValue{{10, 10}},				[]Felt{10},		[]KeyValue{{0, 0}, {5, 5}},			[]Felt{10, 0, 5, 10}},
	{[]KeyValue{{10, 10}, {20, 20}},		[]Felt{10, 20},		[]KeyValue{{0, 0}, {5, 5}, {15, 15}},		[]Felt{10, 20, 0, 5, 10, 15, 20}},
	{[]KeyValue{{10, 10}, {20, 20}, {30, 30}},	[]Felt{30, 10, 20, 30},	[]KeyValue{{0, 0}, {5, 5}, {15, 15}, {25, 25}},	[]Felt{20, 10, 30, 0, 5, 10, 15, 20, 25, 30}},
}

var updateTestTable = []UpsertTest {
	{[]KeyValue{{10, 10}},				[]Felt{10},		[]KeyValue{{10, 100}},			[]Felt{10}},
	{[]KeyValue{{10, 10}, {20, 20}},		[]Felt{10, 20},		[]KeyValue{{10, 100}, {20, 200}},	[]Felt{10, 20}},
}

func init() {
	log.SetLevel(log.TraceLevel)
}

func TestHeight(t *testing.T) {
	for _, data := range heightTestTable {
		tree := NewTree23(data.initialItems)
		assert.Equal(t, data.expectedHeight, tree.Height(), "different height")
	}
}

func TestIs23Tree(t *testing.T) {
	for _, data := range isTree23TestTable {
		tree := NewTree23(data.initialItems)
		tree.GraphAndPicture("tree")
		assertTwoThreeTree(t, tree, data.expectedKeysLevelOrder)
	}
}

func TestUpsertInsert(t *testing.T) {
	for _, data := range insertTestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysPostOrder)
		tree.UpsertNoStats(data.deltaItems)
		assertTwoThreeTree(t, tree, data.finalKeysPostOrder)
	}
}

func TestUpsertUpdate(t *testing.T) {
	for _, data := range updateTestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysPostOrder)
		tree.UpsertNoStats(data.deltaItems)
		assertTwoThreeTree(t, tree, data.finalKeysPostOrder)
		// TODO: add check for new values
	}
}

func TestUpsertIdempotent(t *testing.T) {
	for _, data := range isTree23TestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysLevelOrder)
		tree.UpsertNoStats(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysLevelOrder)
	}
}

func TestUpsertNextKey(t *testing.T) {
	dataCount := 4
	data := make([]KeyValue, dataCount)
	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2), Felt(i*2)}
	}
	tn := NewTree23(data)
	tn.GraphAndPicture("tn1")

	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2+1), Felt(i*2+1)}
	}
	tn = tn.UpsertNoStats(data)
	tn.GraphAndPicture("tn2")
	assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7})
	
	data = []KeyValue{{100, 100}, {101, 101}, {200, 200}, {201, 201}, {202, 202}}
	tn = tn.UpsertNoStats(data)
	tn.GraphAndPicture("tn3")
	assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7, 100, 101, 200, 201, 202})
	
	data = []KeyValue{{10, 10}, {150, 150}, {250, 250}, {251, 251}, {252, 252}}
	tn = tn.UpsertNoStats(data)
	tn.GraphAndPicture("tn4")
	assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7, 10, 100, 101, 150, 200, 201, 202, 250, 251, 252})

	fmt.Printf("tn rootHash=%x\n", tn.RootHash())
}

func TestUpsertFirstKey(t *testing.T) {
}

func FuzzUpsert(f *testing.F) {
	f.Fuzz(func (t *testing.T, input1, input2 []byte) {
		//t.Parallel()
		fmt.Printf("input1=%v input2=%v\n", input1, input2)
		treeFactory := NewTree23BinaryFactory(1)
		bytesReader := bytes.NewReader(input1)
		kvStatePairs := treeFactory.NewUniqueKeyValues(bufio.NewReader(bytesReader))
		if !sort.IsSorted(KeyValueByKey(kvStatePairs)) {
			t.Skip()
		}
		kvStateChangesPairs := make([]KeyValue, len(input2))
		for i, b := range input2 {
			kvStateChangesPairs[i] = KeyValue{Felt(b), Felt(b)}
		}
		if !sort.IsSorted(KeyValueByKey(kvStateChangesPairs)) {
			t.Skip()
		}
		tree := NewTree23(kvStatePairs)
		tree.GraphAndPicture("tree_step1")
		assertTwoThreeTree(t, tree, nil)
		tree = tree.UpsertNoStats(kvStateChangesPairs)
		tree.GraphAndPicture("tree_step2")
		assertTwoThreeTree(t, tree, nil)
	})
}