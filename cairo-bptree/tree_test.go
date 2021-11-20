//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_bptree

import (
	"bufio"
	"bytes"
	"encoding/hex"
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

type RootHashTest struct {
	initialItems	[]KeyValue
	expectedHash	string
}

type UpsertTest struct {
	initialItems		[]KeyValue
	initialKeysLevelOrder	[]Felt
	deltaItems		[]KeyValue
	finalKeysLevelOrder	[]Felt
}

type DeleteTest struct {
	initialItems		[]KeyValue
	initialKeysLevelOrder	[]Felt
	keysToDelete		[]Felt
	finalKeysLevelOrder	[]Felt
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
}

var rootHashTestTable = []RootHashTest {
	{[]KeyValue{},			""},
	{[]KeyValue{{1, 1}},		"532deabf88729cb43995ab5a9cd49bf9b90a079904dc0645ecda9e47ce7345a9"},
	{[]KeyValue{{1, 1}, {2, 2}},	"d3782c59c224da5b6344108ef3431ba4e01d2c30b6570137a91b8b383908c361"},
}

var insertTestTable = []UpsertTest {
	{[]KeyValue{},				[]Felt{},		[]KeyValue{{1, 1}},				[]Felt{1}},
	{[]KeyValue{},				[]Felt{},		[]KeyValue{{1, 1}, {2, 2}},			[]Felt{1, 2}},
	{[]KeyValue{},				[]Felt{},		[]KeyValue{{1, 1}, {2, 2}, {3, 3}},		[]Felt{3, 1, 2, 3}},
	{[]KeyValue{},				[]Felt{},		[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},	[]Felt{3, 1, 2, 3, 4}},

	{[]KeyValue{{1, 1}},			[]Felt{1},		[]KeyValue{{0, 0}},				[]Felt{0, 1}},
	{[]KeyValue{{1, 1}},			[]Felt{1},		[]KeyValue{{2, 2}},				[]Felt{1, 2}},
	{[]KeyValue{{1, 1}},			[]Felt{1},		[]KeyValue{{0, 0}, {2, 2}},			[]Felt{2, 0, 1, 2}},
	{[]KeyValue{{1, 1}},			[]Felt{1},		[]KeyValue{{0, 0}, {2, 2}, {3, 3}},		[]Felt{2, 0, 1, 2, 3}},
	{[]KeyValue{{1, 1}},			[]Felt{1},		[]KeyValue{{0, 0}, {2, 2}, {3, 3}, {4, 4}},	[]Felt{2, 4, 0, 1, 2, 3, 4}},
	{[]KeyValue{{2, 2}},			[]Felt{2},		[]KeyValue{{0, 0}, {1, 1}, {3, 3}, {4, 4}},	[]Felt{2, 4, 0, 1, 2, 3, 4}},
	{[]KeyValue{{3, 3}},			[]Felt{3},		[]KeyValue{{0, 0}, {1, 1}, {2, 2}, {4, 4}},	[]Felt{2, 4, 0, 1, 2, 3, 4}},
	{[]KeyValue{{4, 4}},			[]Felt{4},		[]KeyValue{{0, 0}, {1, 1}, {2, 2}, {3, 3}},	[]Felt{2, 4, 0, 1, 2, 3, 4}},

	{[]KeyValue{{1, 1}, {2, 2}},		[]Felt{1, 2},		[]KeyValue{{0, 0}},				[]Felt{2, 0, 1, 2}},
	{[]KeyValue{{1, 1}, {2, 2}},		[]Felt{1, 2},		[]KeyValue{{0, 0}, {3, 3}},			[]Felt{2, 0, 1, 2, 3}},
	{[]KeyValue{{1, 1}, {2, 2}},		[]Felt{1, 2},		[]KeyValue{{0, 0}, {3, 3}, {4, 4}},		[]Felt{2, 4, 0, 1, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {2, 2}},		[]Felt{1, 2},		[]KeyValue{{0, 0}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{2, 4, 0, 1, 2, 3, 4, 5}},
	{[]KeyValue{{2, 2}, {3, 3}},		[]Felt{2, 3},		[]KeyValue{{0, 0}},				[]Felt{3, 0, 2, 3}},
	{[]KeyValue{{2, 2}, {3, 3}},		[]Felt{2, 3},		[]KeyValue{{0, 0}, {1, 1}},			[]Felt{2, 0, 1, 2, 3}},
	{[]KeyValue{{2, 2}, {3, 3}},		[]Felt{2, 3},		[]KeyValue{{5, 5}},				[]Felt{5, 2, 3, 5}},
	{[]KeyValue{{2, 2}, {3, 3}},		[]Felt{2, 3},		[]KeyValue{{4, 4}, {5, 5}},			[]Felt{4, 2, 3, 4, 5}},
	{[]KeyValue{{2, 2}, {3, 3}},		[]Felt{2, 3},		[]KeyValue{{0, 0}, {4, 4}, {5, 5}},		[]Felt{3, 5, 0, 2, 3, 4, 5}},
	{[]KeyValue{{2, 2}, {3, 3}},		[]Felt{2, 3},		[]KeyValue{{0, 0}, {1, 1}, {4, 4}, {5, 5}},	[]Felt{2, 4, 0, 1, 2, 3, 4, 5}},
	{[]KeyValue{{4, 4}, {5, 5}},		[]Felt{4, 5},		[]KeyValue{{0, 0}},				[]Felt{5, 0, 4, 5}},
	{[]KeyValue{{4, 4}, {5, 5}},		[]Felt{4, 5},		[]KeyValue{{0, 0}, {1, 1}},			[]Felt{4, 0, 1, 4, 5}},
	{[]KeyValue{{4, 4}, {5, 5}},		[]Felt{4, 5},		[]KeyValue{{0, 0}, {1, 1}, {2, 2}},		[]Felt{2, 5, 0, 1, 2, 4, 5}},
	{[]KeyValue{{4, 4}, {5, 5}},		[]Felt{4, 5},		[]KeyValue{{0, 0}, {1, 1}, {2, 2}, {3, 3}},	[]Felt{2, 4, 0, 1, 2, 3, 4, 5}},
	{[]KeyValue{{1, 1}, {4, 4}},		[]Felt{1, 4},		[]KeyValue{{0, 0}},				[]Felt{4, 0, 1, 4}},
	{[]KeyValue{{1, 1}, {4, 4}},		[]Felt{1, 4},		[]KeyValue{{0, 0}, {2, 2}},			[]Felt{2, 0, 1, 2, 4}},
	{[]KeyValue{{1, 1}, {4, 4}},		[]Felt{1, 4},		[]KeyValue{{0, 0}, {2, 2}, {5, 5}},		[]Felt{2, 5, 0, 1, 2, 4, 5}},
	{[]KeyValue{{1, 1}, {4, 4}},		[]Felt{1, 4},		[]KeyValue{{0, 0}, {2, 2}, {3, 3}, {5, 5}},	[]Felt{2, 4, 0, 1, 2, 3, 4, 5}},

	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}},	[]Felt{5, 1, 3, 5},	[]KeyValue{{0, 0}},				[]Felt{3, 5, 0, 1, 3, 5}},
	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}},	[]Felt{5, 1, 3, 5},	[]KeyValue{{0, 0}, {2, 2}, {4, 4}},		[]Felt{4, 2, 5, 0, 1, 2, 3, 4, 5}},
	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}},	[]Felt{5, 1, 3, 5},	[]KeyValue{{6, 6}, {7, 7}, {8, 8}},		[]Felt{5, 7, 1, 3, 5, 6, 7, 8}},
	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}},	[]Felt{5, 1, 3, 5},	[]KeyValue{{6, 6}, {7, 7}, {8, 8}, {9, 9}},	[]Felt{7, 5, 9, 1, 3, 5, 6, 7, 8, 9}},

	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},	[]Felt{3, 1, 2, 3, 4},	[]KeyValue{{0, 0}},			[]Felt{2, 3, 0, 1, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}, {7, 7}},	[]Felt{5, 1, 3, 5, 7},	[]KeyValue{{0, 0}},			[]Felt{3, 5, 0, 1, 3, 5, 7}},
}

var updateTestTable = []UpsertTest {
	{[]KeyValue{{10, 10}},				[]Felt{10},		[]KeyValue{{10, 100}},			[]Felt{10}},
	{[]KeyValue{{10, 10}, {20, 20}},		[]Felt{10, 20},		[]KeyValue{{10, 100}, {20, 200}},	[]Felt{10, 20}},
}

var deleteTestTable = []DeleteTest {
	/* POSITIVE TEST CASES */
	{[]KeyValue{},						[]Felt{},			[]Felt{},		[]Felt{}},

	{[]KeyValue{{1, 1}},					[]Felt{1},			[]Felt{},		[]Felt{1}},
	{[]KeyValue{{1, 1}},					[]Felt{1},			[]Felt{1},		[]Felt{}},

	{[]KeyValue{{1, 1}, {2, 2}},				[]Felt{1, 2},			[]Felt{},		[]Felt{1, 2}},
	{[]KeyValue{{1, 1}, {2, 2}},				[]Felt{1, 2},			[]Felt{1},		[]Felt{2}},
	{[]KeyValue{{1, 1}, {2, 2}},				[]Felt{1, 2},			[]Felt{2},		[]Felt{1}},
	{[]KeyValue{{1, 1}, {2, 2}},				[]Felt{1, 2},			[]Felt{1, 2},		[]Felt{}},

	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{},		[]Felt{3, 1, 2, 3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{1},		[]Felt{2, 3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{2},		[]Felt{1, 3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{3},		[]Felt{1, 2}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{1, 2},		[]Felt{3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{1, 3},		[]Felt{2}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{2, 3},		[]Felt{1}},
	//{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{1, 2, 3},	[]Felt{}},

	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},		[]Felt{3, 1, 2, 3, 4},		[]Felt{1},		[]Felt{3, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},		[]Felt{3, 1, 2, 3, 4},		[]Felt{2},		[]Felt{3, 1, 3, 4}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},		[]Felt{3, 1, 2, 3, 4},		[]Felt{3},		[]Felt{3, 1, 2, 4}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},		[]Felt{3, 1, 2, 3, 4},		[]Felt{4},		[]Felt{3, 1, 2, 3}},

	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{3, 5, 1, 2, 3, 4, 5},	[]Felt{1},		[]Felt{3, 5, 2, 3, 4, 5}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{3, 5, 1, 2, 3, 4, 5},	[]Felt{2},		[]Felt{3, 5, 1, 3, 4, 5}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{3, 5, 1, 2, 3, 4, 5},	[]Felt{3},		[]Felt{3, 5, 1, 2, 4, 5}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{3, 5, 1, 2, 3, 4, 5},	[]Felt{4},		[]Felt{3, 5, 1, 2, 3, 5}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{3, 5, 1, 2, 3, 4, 5},	[]Felt{5},		[]Felt{3, 5, 1, 2, 3, 4}},

	/* NEGATIVE TEST CASES */
	{[]KeyValue{},						[]Felt{},			[]Felt{1},		[]Felt{}},
	{[]KeyValue{{1, 1}},					[]Felt{1},			[]Felt{2},		[]Felt{1}},
	{[]KeyValue{{1, 1}, {2, 2}},				[]Felt{1, 2},			[]Felt{3},		[]Felt{1, 2}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{4},		[]Felt{3, 1, 2, 3}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},		[]Felt{3, 1, 2, 3, 4},		[]Felt{5},		[]Felt{3, 1, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}},	[]Felt{3, 5, 1, 2, 3, 4, 5},	[]Felt{6},		[]Felt{3, 5, 1, 2, 3, 4, 5}},

	/* MIXED TEST CASES */
	// TODO
}

func init() {
	log.SetLevel(log.WarnLevel)
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

func TestRootHash(t *testing.T) {
	for _, data := range rootHashTestTable {
		tree := NewTree23(data.initialItems)
		assert.Equal(t, data.expectedHash, hex.EncodeToString(tree.RootHash()), "different root hash")
	}
}

func TestUpsertInsert(t *testing.T) {
	for _, data := range insertTestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysLevelOrder)
		tree.UpsertNoStats(data.deltaItems)
		assertTwoThreeTree(t, tree, data.finalKeysLevelOrder)
	}
}

func TestUpsertUpdate(t *testing.T) {
	for _, data := range updateTestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysLevelOrder)
		tree.UpsertNoStats(data.deltaItems)
		assertTwoThreeTree(t, tree, data.finalKeysLevelOrder)
		// TODO: add check for new values
	}
}

/*func TestUpsertIdempotent(t *testing.T) {
	for _, data := range isTree23TestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysLevelOrder)
		tree.UpsertNoStats(data.initialItems)
		assertTwoThreeTree(t, tree, data.expectedKeysLevelOrder)
	}
}*/

func TestUpsertNextKey(t *testing.T) {
	dataCount := 4
	data := make([]KeyValue, dataCount)
	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2), Felt(i*2)}
	}
	tn := NewTree23(data)
	//tn.GraphAndPicture("tn1")

	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2+1), Felt(i*2+1)}
	}
	tn = tn.UpsertNoStats(data)
	//tn.GraphAndPicture("tn2")
	assertTwoThreeTree(t, tn, []Felt{4, 2, 6, 0, 1, 2, 3, 4, 5, 6, 7})
	
	data = []KeyValue{{100, 100}, {101, 101}, {200, 200}, {201, 201}, {202, 202}}
	tn = tn.UpsertNoStats(data)
	tn.GraphAndPicture("tn3")
	//assertTwoThreeTree(t, tn, []Felt{2, 2, 100, 0, 1, 2, 3, 4, 5, 6, 7, 100, 101, 200, 201, 202})
	
	//data = []KeyValue{{10, 10}, {150, 150}, {250, 250}, {251, 251}, {252, 252}}
	//tn = tn.UpsertNoStats(data)
	//tn.GraphAndPicture("tn4")
	//assertTwoThreeTree(t, tn, []Felt{0, 1, 2, 3, 4, 5, 6, 7, 10, 100, 101, 150, 200, 201, 202, 250, 251, 252})
}

func TestUpsertFirstKey(t *testing.T) {
}

func TestDelete(t *testing.T) {
	for _, data := range deleteTestTable {
		tree := NewTree23(data.initialItems)
		assertTwoThreeTree(t, tree, data.initialKeysLevelOrder)
		tree.DeleteNoStats(data.keysToDelete)
		assertTwoThreeTree(t, tree, data.finalKeysLevelOrder)
	}
}

func FuzzUpsert(f *testing.F) {
	f.Fuzz(func (t *testing.T, input1, input2 []byte) {
		//t.Parallel()
		//fmt.Printf("input1=%v input2=%v\n", input1, input2)
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
		//tree.GraphAndPicture("tree_step1")
		assertTwoThreeTree(t, tree, nil)
		tree = tree.UpsertNoStats(kvStateChangesPairs)
		//tree.GraphAndPicture("tree_step2")
		assertTwoThreeTree(t, tree, nil)
	})
}