package cairo_bptree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertNodeEqual(t *testing.T, expected, actual *Node23) {
	assert.Equal(t, expected.keysInLevelOrder(), actual.keysInLevelOrder(), "different keys by level")
}

type MergeTest struct {
	left	*Node23
	right	*Node23
	final	*Node23
}

func K2K(keys []Felt) []*Felt {
	kv := KV(keys, keys)
	return kv.keys
}

func K2KV(keys []Felt) ([]*Felt, []*Felt) {
	values := make([]Felt, len(keys))
	copy(values, keys)
	kv := KV(keys, values)
	return kv.keys, kv.values
}

var mergeLeft2RightTestTable = []MergeTest {
	{
		makeInternalNode([]*Node23{
			makeLeafNode(K2KV([]Felt{12, 127})),
		}, K2K([]Felt{127})),
		makeInternalNode([]*Node23{
			makeLeafNode(K2KV([]Felt{127, 128})),
			makeLeafNode(K2KV([]Felt{128, 135, 173})),
		}, K2K([]Felt{128})),
		makeInternalNode([]*Node23{
			makeLeafNode(K2KV([]Felt{12, 127})),
			makeLeafNode(K2KV([]Felt{127, 128})),
			makeLeafNode(K2KV([]Felt{128, 135, 173})),
		}, K2K([]Felt{127, 128})),
	},
	{
		makeInternalNode([]*Node23{
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{12, 127})),
			}, K2K([]Felt{127})),
		}, K2K([]Felt{44})),
		makeInternalNode([]*Node23{
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{127, 128})),
				makeLeafNode(K2KV([]Felt{128, 135, 173})),
			}, K2K([]Felt{128})),
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{173, 237})),
				makeLeafNode(K2KV([]Felt{237, 1000})),
			}, K2K([]Felt{237})),
		}, K2K([]Felt{173})),
		makeInternalNode([]*Node23{
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{12, 127})),
				makeLeafNode(K2KV([]Felt{127, 128})),
				makeLeafNode(K2KV([]Felt{128, 135, 173})),
			}, K2K([]Felt{127, 128})),
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{173, 237})),
				makeLeafNode(K2KV([]Felt{237, 1000})),
			}, K2K([]Felt{237})),
		}, K2K([]Felt{173})),
	},
}

var mergeRight2LeftTestTable = []MergeTest {
	{
		makeInternalNode([]*Node23{
			makeLeafNode(K2KV([]Felt{127, 128})),
			makeLeafNode(K2KV([]Felt{128, 135, 173})),
		}, K2K([]Felt{128})),
		makeInternalNode([]*Node23{
			makeLeafNode(K2KV([]Felt{173, 190})),
		}, K2K([]Felt{190})),
		makeInternalNode([]*Node23{
			makeLeafNode(K2KV([]Felt{127, 128})),
			makeLeafNode(K2KV([]Felt{128, 135, 173})),
			makeLeafNode(K2KV([]Felt{173, 190})),
		}, K2K([]Felt{128, 173})),
	},
	{
		makeInternalNode([]*Node23{
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{127, 128})),
				makeLeafNode(K2KV([]Felt{128, 135, 173})),
			}, K2K([]Felt{128})),
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{173, 237})),
				makeLeafNode(K2KV([]Felt{237, 1000})),
			}, K2K([]Felt{237})),
		}, K2K([]Felt{173})),
		makeInternalNode([]*Node23{
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{1000, 1002})),
			}, K2K([]Felt{1002})),
		}, K2K([]Felt{1100})),
		makeInternalNode([]*Node23{
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{127, 128})),
				makeLeafNode(K2KV([]Felt{128, 135, 173})),
			}, K2K([]Felt{128})),
			makeInternalNode([]*Node23{
				makeLeafNode(K2KV([]Felt{173, 237})),
				makeLeafNode(K2KV([]Felt{237, 1000})),
				makeLeafNode(K2KV([]Felt{1000, 1002})),
			}, K2K([]Felt{237, 1000})),
		}, K2K([]Felt{173})),
	},
}

func TestMergeLeft2Right(t *testing.T) {
	for _, data := range mergeLeft2RightTestTable {
		_, merged := mergeLeft2Right(data.left, data.right)
		assertNodeEqual(t, data.final, merged)
	}
}

func TestMergeRight2Left(t *testing.T) {
	for _, data := range mergeRight2LeftTestTable {
		merged, _ := mergeRight2Left(data.left, data.right)
		assertNodeEqual(t, data.final, merged)
	}
}
