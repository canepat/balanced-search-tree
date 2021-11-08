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
	assert.Equal(t, expectedKeysPostOrder, tree.WalkKeysPostOrder(), "different in-order keys: %v", tree.WalkKeysPostOrder())
}

var t0, t1, t2, t3, t4, t5, tn *Tree23

func init() {
	log.SetLevel(log.TraceLevel)

	t0 = NewTree23()
	kv1 := KeyValue{Felt(1), Felt(1)}
	t1 = NewTree23().Upsert([]KeyValue{kv1})
	kv2 := KeyValue{Felt(2), Felt(2)}
	t2 = NewTree23().Upsert([]KeyValue{kv1, kv2})
	t2.GraphAndPicture("t2", false)
	kv3 := KeyValue{Felt(3), Felt(3)}
	t3 = NewTree23().Upsert([]KeyValue{kv1, kv2, kv3})
	t3.GraphAndPicture("t3", false)
	kv4 := KeyValue{Felt(4), Felt(4)}
	t4 = NewTree23().Upsert([]KeyValue{kv1, kv2, kv3, kv4})
	t4.GraphAndPicture("t4", false)
	kv5 := KeyValue{Felt(5), Felt(5)}
	t5 = NewTree23().Upsert([]KeyValue{kv1, kv2, kv3, kv4, kv5})
	t5.GraphAndPicture("t5", false)

	dataCount := uint64(4)
	data := make([]KeyValue, dataCount)
	var i uint64
	for i = 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2), Felt(i*2)}
	}
	tn = NewTree23().Upsert(data)
	tn.GraphAndPicture("tn1", false)
	for i = 0; i < dataCount; i++ {
		data[i] = KeyValue{Felt(i*2+1), Felt(i*2+1)}
	}
	tn = tn.Upsert(data)
	tn.GraphAndPicture("tn2", false)
	data = []KeyValue{{100, 100}, {101, 101}, {200, 200}, {201, 201}, {202, 202}}
	tn = tn.Upsert(data)
	tn.GraphAndPicture("tn3", false)
	data = []KeyValue{{10, 10}, {150, 150}, {250, 250}, {251, 251}, {252, 252}}
	tn = tn.Upsert(data)
	tn.GraphAndPicture("tn4", false)
}

func TestIs23Tree(t *testing.T) {
	assertTwoThreeTree(t, t0, []Felt{})
	assertTwoThreeTree(t, t1, []Felt{Felt(1)})
	assertTwoThreeTree(t, t2, []Felt{Felt(1), Felt(2)})
	assertTwoThreeTree(t, t3, []Felt{Felt(1), Felt(2), Felt(3)})
	assertTwoThreeTree(t, t4, []Felt{Felt(1), Felt(2), Felt(3), Felt(4)})
	assertTwoThreeTree(t, t5, []Felt{Felt(1), Felt(2), Felt(3), Felt(4), Felt(5)})
}

func TestUpsert(t *testing.T) {
	//var tree *Tree23
	//kv1 := KeyValue{Felt(1), Felt(1)}
	//tree.Upsert([]KeyValue{kv1})
}
