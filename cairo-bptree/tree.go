package cairo_bptree

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type Stats struct {
	ExposedCount	uint64
}

type Tree23 struct {
	root	*Node23
}

func NewEmptyTree23() *Tree23 {
	return &Tree23{}
}

func NewTree23(kvItems []KeyValue) *Tree23 {
	log.Infof("NewTree23: creating tree with #kvItems=%v\n", len(kvItems))
	tree := new(Tree23).Upsert(kvItems, &Stats{})
	tree.reset()
	log.Infof("NewTree23: created tree root=%s with #kvItems=%v\n", tree.root, len(kvItems))
	return tree
}

func (t *Tree23) Size() int {
	node_items := t.WalkPostOrder(func(n *Node23) interface{} { return n })
	return len(node_items)
}

func (t *Tree23) CountNewHashes() (hashCount uint) {
	node_items := t.WalkPostOrder(func(n *Node23) interface{} { return n })
	for i := range node_items {
		if node_items[i].(*Node23).exposed {
			hashCount++
		}
	}
	return hashCount
}

func (t *Tree23) RootHash() []byte {
	if t.root == nil {
		return []byte{}
	}
	return t.root.hashNode()
}

func (t *Tree23) IsTwoThree() bool {
	if t.root == nil {
		return true
	}
	return t.root.isTwoThree()
}

func (t *Tree23) Graph(filename string, debug bool) {
	graph := NewGraph(t.root)
	graph.saveDot(filename, debug)
}

func (t *Tree23) GraphAndPicture(filename string) error {
	graph := NewGraph(t.root)
	return graph.saveDotAndPicture(filename, false)
}

func (t *Tree23) GraphAndPictureDebug(filename string) error {
	graph := NewGraph(t.root)
	return graph.saveDotAndPicture(filename, true)
}

func (t *Tree23) Height() int {
	if t.root == nil {
		return 0
	}
	return t.root.height()
}

func (t *Tree23) KeysInLevelOrder() []Felt {
	if t.root == nil {
		return []Felt{}
	}
	keysByLevel := make([]Felt, 0)
	for i := 0; i < t.root.height(); i++ {
		keysByLevel = append(keysByLevel, t.root.keysByLevel(i)...)
	}
	return keysByLevel
}

func (t *Tree23) WalkPostOrder(w Walker) []interface{} {
	if t.root == nil {
		return make([]interface{}, 0)
	}
	return t.root.walkPostOrder(w)
}

func (t *Tree23) WalkKeysPostOrder() []Felt {
	key_pointers := make([]*Felt, 0)
	t.WalkPostOrder(func(n *Node23) interface{} {
		if n.isLeaf && n.keyCount() > 0 {
			log.Tracef("WalkKeysPostOrder: L n=%p n.keys=%v\n", n, n.keys)
			key_pointers = append(key_pointers, n.keys[:len(n.keys)-1]...)
		}
		return nil
	})
	keys := ptr2pte(key_pointers)
	log.Tracef("WalkKeysPostOrder: keys=%v\n", keys)
	return keys
}

func (t *Tree23) UpsertNoStats(kvItems []KeyValue) *Tree23 {
	return t.Upsert(kvItems, &Stats{})
}

func (t *Tree23) Upsert(kvItems []KeyValue, stats *Stats) *Tree23 {
	log.Debugf("Upsert: t=%p root=%p kvItems=%v\n", t, t.root, kvItems)
	nodes, _ := t.root.upsert(kvItems, stats)
	log.Tracef("Upsert: nodes=%v\n", nodes)
	ensure(len(nodes) > 0, "nodes length is zero")
	if len(nodes) == 1 {
		t.root = nodes[0]
	} else {
		promoted := t.promote(nodes)
		ensure(len(promoted) == 1, fmt.Sprintf("invalid promoted length: %d", len(promoted)))
		t.root = promoted[0]
	}
	log.Debugf("Upsert: t=%p root=%p\n", t, t.root)
	return t
}

func (t *Tree23) promote(nodes []*Node23) []*Node23 {
	log.Debugf("promote: #nodes=%d nodes=%v\n", len(nodes), nodes)
	upwards := make([]*Node23, 0)
	if len(nodes) > 3 {
		for len(nodes) > 3 {
			upwards = append(upwards, makeInternalNode(nodes[:2]))
			nodes = nodes[2:]
		}
		upwards = append(upwards, makeInternalNode(nodes[:]))
		log.Debugf("promote: >3 #upwards=%d upwards=%v\n", len(upwards), upwards)
		return t.promote(upwards)
	} else {
		upwards = append(upwards, makeInternalNode(nodes))
		log.Debugf("promote: 2/3 #upwards=%d upwards=%v\n", len(upwards), upwards)
		return upwards
	}
}

func (t *Tree23) reset() {
	if t.root == nil {
		return
	}
	t.root.reset()
}
