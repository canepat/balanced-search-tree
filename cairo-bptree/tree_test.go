//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_bptree

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertTwoThreeTree(t *testing.T, tree *Tree23, expectedKeysLevelOrder []Felt) {
	assert.True(t, tree.IsTwoThree(), "2-3-tree properties do not hold for tree: %v", tree.WalkKeysPostOrder())
	if expectedKeysLevelOrder != nil {
		assert.Equal(t, expectedKeysLevelOrder, tree.KeysInLevelOrder(), "different keys by level")
	}
}

func require23Tree(t *testing.T, tree *Tree23, expectedKeysLevelOrder []Felt) {
	require.True(t, tree.IsTwoThree(), "2-3-tree properties do not hold for tree: %v", tree.WalkKeysPostOrder())
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
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}},			[]Felt{5, 3, 7, 1, 2, 3, 4, 5, 6, 7, 8}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}},		[]Felt{5, 3, 7, 9, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 10}},	[]Felt{5, 3, 7, 9, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 10}, {11, 11}},		[]Felt{5, 9, 3, 7, 11, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 10}, {11, 11}, {12, 12}},	[]Felt{5, 9, 3, 7, 11, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
	{
		[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}, {7, 7}, {8, 8}, {9, 9}, {10, 10}, {11, 11}, {12, 12}, {13, 13}, {14, 14}, {15, 15}, {16, 16}, {17, 17}},
		[]Felt{9, 5, 13, 3, 7, 11, 15, 17, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
	},
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

	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}},		[]Felt{3, 1, 2, 3, 4},		[]KeyValue{{0, 0}},	[]Felt{2, 3, 0, 1, 2, 3, 4}},
	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}, {7, 7}},		[]Felt{5, 1, 3, 5, 7},		[]KeyValue{{0, 0}},	[]Felt{3, 5, 0, 1, 3, 5, 7}},

	{[]KeyValue{{1, 1}, {3, 3}, {5, 5}, {7, 7}, {9, 9}},	[]Felt{5, 9, 1, 3, 5, 7, 9},	[]KeyValue{{0, 0}},	[]Felt{5, 3, 9, 0, 1, 3, 5, 7, 9}},

	// Debug
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {5, 5}, {6, 6}, {7, 7}, {8, 8}},	[]Felt{6, 3, 8, 1, 2, 3, 5, 6, 7, 8},	[]KeyValue{{4, 4}},	[]Felt{6, 3, 5, 8, 1, 2, 3, 4, 5, 6, 7, 8}},
	{
		[]KeyValue{{10, 10}, {15, 15}, {20, 20}},
		[]Felt{20, 10, 15, 20},
		[]KeyValue{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {11, 11}, {13, 13}, {18, 18}, {19, 19}, {30, 30}, {31, 31}},
		[]Felt{15, 5, 20, 3, 11, 19, 31, 1, 2, 3, 4, 5, 10, 11, 13, 15, 18, 19, 20, 30, 31},
	},
	{
		[]KeyValue{{0, 0}, {2, 2}, {4, 4}, {6, 6}, {8, 8}, {10, 10}, {12, 12}, {14, 14}, {16, 16}, {18, 18}, {20, 20}},
		[]Felt{8, 16, 4, 12, 20, 0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20},
		[]KeyValue{{1, 1}, {3, 3}, {5, 5}},
		//[]KeyValue{{1, 1}, {3, 3}, {5, 5}, {7, 7}},
		//[]KeyValue{{1, 1}, {3, 3}, {5, 5}, {7, 7}, {9, 9}, {11, 11}, {13, 13}, {15, 15}},
		[]Felt{8, 4, 16, 2, 6, 12, 20, 0, 1, 2, 3, 4, 5, 6, 8, 10, 12, 14, 16, 18, 20}},
	{
		[]KeyValue{{4, 4}, {10, 10}, {17, 17}, {85, 85}, {104, 104}, {107, 107}, {112, 112}, {115, 115}, {136, 136}, {156, 156}, {191, 191}},
		[]Felt{104, 136, 17, 112, 191, 4, 10, 17, 85, 104, 107, 112, 115, 136, 156, 191},
		[]KeyValue{{0, 0}, {96, 96}, {120, 120}, {129, 129}, {133, 133}, {164, 164}, {187, 187}, {189, 189}},
		nil,
	},
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
	{[]KeyValue{{1, 1}, {2, 2}, {3, 3}},			[]Felt{3, 1, 2, 3},		[]Felt{1, 2, 3},	[]Felt{}},

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
		//tree.GraphAndPicture("tree23")
		assertTwoThreeTree(t, tree, data.expectedKeysLevelOrder)
	}
}

func Test23TreeSeries(t *testing.T) {
	maxNumberOfNodes := 100
	for i := 0; i < maxNumberOfNodes; i++ {
		kvPairs := make([]KeyValue, 0)
		for j := 0; j < i; j++ {
			kvPairs = append(kvPairs, KeyValue{Felt(j), Felt(j)})
		}
		tree := NewTree23(kvPairs)
		require23Tree(t, tree, nil)
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
		//tree.GraphAndPicture("tree_step1")
		tree.UpsertNoStats(data.deltaItems)
		//tree.GraphAndPicture("tree_step2")
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
		//tree.GraphAndPicture("tree_delete1")
		tree.DeleteNoStats(data.keysToDelete)
		//tree.GraphAndPicture("tree_delete2")
		assertTwoThreeTree(t, tree, data.finalKeysLevelOrder)
	}
}

func FuzzUpsert(f *testing.F) {
	f.Fuzz(func (t *testing.T, input1, input2 []byte) {
		//t.Parallel()
		treeFactory := NewTree23BinaryFactory(1)
		bytesReader1 := bytes.NewReader(input1)
		kvStatePairs := treeFactory.NewUniqueKeyValues(bufio.NewReader(bytesReader1))
		require.True(t, sort.IsSorted(KeyValueByKey(kvStatePairs)), "kvStatePairs is not sorted")
		bytesReader2 := bytes.NewReader(input2)
		kvStateChangesPairs := treeFactory.NewUniqueKeyValues(bufio.NewReader(bytesReader2))
		fmt.Printf("kvStatePairs=%v kvStateChangesPairs=%v\n", kvStatePairs, kvStateChangesPairs)
		require.True(t, sort.IsSorted(KeyValueByKey(kvStateChangesPairs)), "kvStateChangesPairs is not sorted")
		tree := NewTree23(kvStatePairs)
		require23Tree(t, tree, nil)
		tree = tree.UpsertNoStats(kvStateChangesPairs)
		require23Tree(t, tree, nil)
	})
}