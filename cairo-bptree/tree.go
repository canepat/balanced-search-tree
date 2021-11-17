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
	promoted, _ := upsert(t.root, kvItems, stats)
	log.Tracef("Upsert: promoted=%v\n", promoted)
	ensure(len(promoted) > 0, "nodes length is zero")
	if len(promoted) == 1 {
		t.root = promoted[0]
	} else {
		t.root = promote(promoted)
	}
	log.Debugf("Upsert: t=%p root=%p\n", t, t.root)
	return t
}

func (t *Tree23) DeleteNoStats(keyToDelete []Felt) *Tree23 {
	return t.Delete(keyToDelete, &Stats{})
}

func (t *Tree23) Delete(keyToDelete []Felt, stats *Stats) *Tree23 {
	log.Debugf("Delete: t=%p root=%p keyToDelete=%v\n", t, t.root, keyToDelete)
	newRoot, nextKey := delete(t.root, keyToDelete, stats)
	t.root = newRoot
	if nextKey != nil {
		updateNextKey(t.root, nextKey)
	}
	return t
}

func (t *Tree23) reset() {
	if t.root == nil {
		return
	}
	t.root.reset()
}
