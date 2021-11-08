//go:build gofuzzbeta
// +build gofuzzbeta

package cairo_bptree

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func assertTwoThreeTree(t *testing.T, tree *Tree23, expectedKeysPostOrder []Felt) {
	assert.True(t, tree.IsTwoThree(), "2-3-tree properties do not hold for tree: %v", tree.WalkKeysPostOrder())
	if expectedKeysPostOrder != nil {
		assert.Equal(t, expectedKeysPostOrder, tree.WalkKeysPostOrder(), "different in-order keys: %v", tree.WalkKeysPostOrder())
	}
}

var t0, t1, t2, t3, t4, t5, tn *Tree23

func init() {
	log.SetLevel(log.WarnLevel)

	t0 = NewTree23()
	t0.GraphAndPicture("t0_initial", false)

	kv1 := KeyValue{Felt(1), Felt(1)}
	t1 = NewTree23().Upsert([]KeyValue{kv1})
	t1.GraphAndPicture("t1_initial", false)

	kv2 := KeyValue{Felt(2), Felt(2)}
	t2 = NewTree23().Upsert([]KeyValue{kv1, kv2})
	t2.GraphAndPicture("t2_initial", false)

	kv3 := KeyValue{Felt(3), Felt(3)}
	t3 = NewTree23().Upsert([]KeyValue{kv1, kv2, kv3})
	t3.GraphAndPicture("t3_initial", false)

	kv4 := KeyValue{Felt(4), Felt(4)}
	t4 = NewTree23().Upsert([]KeyValue{kv1, kv2, kv3, kv4})
	t4.GraphAndPicture("t4_initial", false)

	kv5 := KeyValue{Felt(5), Felt(5)}
	t5 = NewTree23().Upsert([]KeyValue{kv1, kv2, kv3, kv4, kv5})
	t5.GraphAndPicture("t5_initial", false)

	dataCount := 4
	data := make([]KeyValue, dataCount)
	for i := 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2), Felt(i*2)}
	}
	tn = NewTree23().Upsert(data)
	tn.GraphAndPicture("tn_initial", false)
}

func TestIs23Tree(t *testing.T) {
	assertTwoThreeTree(t, t0, []Felt{})
	assertTwoThreeTree(t, t1, []Felt{1})
	assertTwoThreeTree(t, t2, []Felt{1, 2})
	assertTwoThreeTree(t, t3, []Felt{1, 2, 3})
	assertTwoThreeTree(t, t4, []Felt{1, 2, 3, 4})
	assertTwoThreeTree(t, t5, []Felt{1, 2, 3, 4, 5})
}

func TestUpsert(t *testing.T) {
	dataCount := 4
	data := make([]KeyValue, dataCount)
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
