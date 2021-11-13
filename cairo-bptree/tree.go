package cairo_bptree

import (
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
	tree := new(Tree23).Upsert(kvItems, &Stats{})
	tree.reset()
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

func (t *Tree23) GraphAndPicture(filename string, debug bool) error {
	graph := NewGraph(t.root)
	return graph.saveDotAndPicture(filename, debug)
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
			log.Tracef("WalkKeysInOrder: L n=%p n.keys=%v\n", n, n.keys)
			key_pointers = append(key_pointers, n.keys[:len(n.keys)-1]...)
		}
		return nil
	})
	keys := ptr2pte(key_pointers)
	log.Tracef("WalkKeysInOrder: keys=%v\n", keys)
	return keys
}

func (t *Tree23) UpsertNoStats(kvItems []KeyValue) *Tree23 {
	return t.Upsert(kvItems, &Stats{})
}

func (t *Tree23) Upsert(kvItems []KeyValue, stats *Stats) *Tree23 {
	log.Tracef("Upsert: t=%p root=%p kvItems=%v\n", t, t.root, kvItems)
	nodes, _ := t.root.upsert(kvItems, stats)
	log.Tracef("Upsert: nodes=%v\n", nodes)
	ensure(len(nodes) > 0, "nodes length is zero")
	if len(nodes) == 1 {
		t.root = nodes[0]
	} else {
		t.root = t.promote(nodes)
	}
	log.Tracef("Upsert: t=%p root=%p\n", t, t.root)
	return t
}

func (t *Tree23) promote(nodes []*Node23) *Node23 {
	log.Debugf("promote: nodes=%v\n", nodes)
	numberOfGroups := len(nodes) / 3
	if len(nodes) % 3 > 0 {
		numberOfGroups++
	}
	log.Tracef("promote: numberOfGroups=%d\n", numberOfGroups)
	upperNodes := make([]*Node23, 0)
	for i := 0; i < numberOfGroups; i++ {
		firstChildIndex, secondChildIndex, thirdChildIndex := i*3, i*3+1, i*3+2
		childNodes := make([]*Node23, 0)
		childNodes = append(childNodes, nodes[firstChildIndex])
		if secondChildIndex < len(nodes) {
			childNodes = append(childNodes, nodes[secondChildIndex])
		}
		if thirdChildIndex < len(nodes) {
			childNodes = append(childNodes, nodes[thirdChildIndex])
		}
		upperNodes = append(upperNodes, makeInternalNode(childNodes))
	}
	ensure(len(upperNodes) > 0, "upperNodes length is zero")
	if len(upperNodes) == 1 {
		log.Debugf("promote: root=%s\n", upperNodes[0])
		return upperNodes[0]
	} else {
		log.Debugf("promote: upperNodes=%v\n", upperNodes)
		return t.promote(upperNodes)
	}
}

func (t *Tree23) reset() {
	if t.root == nil {
		return
	}
	t.root.reset()
}
