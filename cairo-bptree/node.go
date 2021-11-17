package cairo_bptree

import (
	"fmt"
	"sort"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

type Keys []Felt

func (keys Keys) Len() int { return len(keys) }

func (keys Keys) Less(i, j int) bool { return keys[i] < keys[j] }

func (keys Keys) Swap(i, j int) { keys[i], keys[j] = keys[j], keys[i] }

func (keys Keys) Contains(key Felt) bool {
	return sort.Search(len(keys), func(i int) bool { return keys[i] == key }) != len(keys)
}

type KeyValue struct {
	key	Felt
	value	Felt
}

type KeyValueByKey []KeyValue

func (kv KeyValueByKey) Len() int { return len(kv) }

func (kv KeyValueByKey) Less(i, j int) bool { return kv[i].key < kv[j].key }

func (kv KeyValueByKey) Swap(i, j int) { kv[i], kv[j] = kv[j], kv[i] }

type Node23 struct {
	isLeaf		bool
	children	[]*Node23
	keys		[]*Felt
	values		[]*Felt
	exposed		bool
}

func (n *Node23) String() string {
	s := fmt.Sprintf("{%p isLeaf=%t keys=%v-%v children=[", n, n.isLeaf, ptr2pte(n.keys), n.keys)
	for i, child := range n.children {
		s += fmt.Sprintf("%p", child)
		if i != len(n.children)-1 {
			s += " "
		}
	}
	s += "]}"
	return s
}

func makeInternalNode(children []*Node23) *Node23 {
	internalKeys := internalKeysFromChildren(children)
	n := &Node23{isLeaf: false, children: children, keys: internalKeys, values: make([]*Felt, 0), exposed: true}
	return n
}

func makeLeafNode(keys, values []*Felt) *Node23 {
	ensure(len(keys) > 0, "number of keys is zero")
	ensure(len(keys) == len(values), "keys and values have different cardinality")
	leafKeys := append(make([]*Felt, 0, len(keys)), keys...)
	leafValues := append(make([]*Felt, 0, len(values)), values...)
	n := &Node23{isLeaf: true, children: make([]*Node23, 0), keys: leafKeys, values: leafValues, exposed: true}
	return n
}

func makeEmptyLeafNode() (*Node23) {
	return makeLeafNode(make([]*Felt, 1), make([]*Felt, 1))
}

func internalKeysFromChildren(children []*Node23) []*Felt {
	ensure(len(children) > 1, "number of children is lower than 2")
	internalKeys := make([]*Felt, 0, len(children)-1)
	for _, child := range children[:len(children)-1] {
		ensure(child.nextKey() != nil, "child next key is zero")
		internalKeys = append(internalKeys, child.nextKey())
	}
	log.Tracef("internalKeysFromChildren: children=%v internalKeys=%v\n", children, ptr2pte(internalKeys))
	return internalKeys
}

func promote(nodes []*Node23) *Node23 {
	log.Debugf("promote: #nodes=%d nodes=%v\n", len(nodes), nodes)
	promotedRoot := makeInternalNode(nodes)
	log.Debugf("promote: promotedRoot=%s\n", promotedRoot)
	if promotedRoot.keyCount() > 2 {
		intermediateNodes := make([]*Node23, 0)
		promotedKeys := make([]*Felt, 0)
		for promotedRoot.keyCount() > 2 {
			intermediateNodes = append(intermediateNodes, makeInternalNode(promotedRoot.children[:2]))
			promotedRoot.children = promotedRoot.children[2:]
			promotedKeys = append(promotedKeys, promotedRoot.keys[1])
			promotedRoot.keys = promotedRoot.keys[2:]
		}
		intermediateNodes = append(intermediateNodes, makeInternalNode(promotedRoot.children[:]))
		promotedRoot.children = intermediateNodes
		promotedRoot.keys = promotedKeys
		log.Debugf("promote: #keys>2 promotedRoot=%s\n", promotedRoot)
		return promotedRoot
	} else {
		log.Debugf("promote: #keys<=2 promotedRoot=%s\n", promotedRoot)
		return promotedRoot
	}
}

func (n *Node23) reset() {
	if n.isLeaf {
		n.exposed = false
	} else {
		for _, child := range n.children {
			child.reset()
		}
	}
}

func (n *Node23) isTwoThree() bool {
	if n.isLeaf {
		keyCount := n.keyCount()
		/* Any leaf node can have either 1 or 2 keys (plus next key) */
		return keyCount == 2 || keyCount == 3
	} else {
		// Check that each child subtree is a 2-3 tree
		for _, child := range n.children {
			if !child.isTwoThree() {
				return false
			}
		}
		/* Any internal node can have either 2 or 3 children */
		return n.childrenCount() == 2 || n.childrenCount() == 3
	}
}

func (n *Node23) keyCount() int {
	return len(n.keys)
}

func (n *Node23) childrenCount() int {
	return len(n.children)
}

func (n *Node23) valueCount() int {
	return len(n.values)
}

func (n *Node23) firstKey() *Felt {
	ensure(len(n.keys) > 0, "firstKey: node has no key")
	return n.keys[0]
}

func (n *Node23) lastChild() *Node23 {
	ensure(len(n.children) > 0, "lastChild: node has no children")
	return n.children[len(n.children)-1]
}

func (n *Node23) nextKey() *Felt {
	ensure(len(n.keys) > 0, "nextKey: node has no key")
	return n.keys[len(n.keys)-1]
}

func (n *Node23) nextValue() *Felt {
	ensure(len(n.values) > 0, "nextValue: node has no value")
	return n.values[len(n.values)-1]
}

func (n *Node23) rawPointer() uintptr {
	return uintptr(unsafe.Pointer(n))
}

func (n *Node23) setNextKey(nextKey *Felt) {
	ensure(len(n.keys) > 0, "setNextKey: node has no key")
	n.keys[len(n.keys)-1] = nextKey
}

func (n *Node23) canonicalKeys() []Felt {
	if n.isLeaf {
		ensure(len(n.keys) > 0, "canonicalKeys: node has no key")
		return ptr2pte(n.keys[:len(n.keys)-1])
	} else {
		return ptr2pte(n.keys[:])
	}
}

func (n *Node23) height() int {
	if n.isLeaf {
		return 1
	} else {
		ensure(len(n.children) > 0, "heigth: internal node has zero children")
		return n.children[0].height() + 1
	}
}

func (n *Node23) keysByLevel(level int) []Felt {
	if level == 0 {
		return n.canonicalKeys()
	} else {
		levelKeys := make([]Felt, 0)
		for _, child := range n.children {
			childLevelKeys := child.keysByLevel(level-1)
			levelKeys = append(levelKeys, childLevelKeys...)
		}
		return levelKeys
	}
}

type Walker func(*Node23) interface{}

func (n *Node23) walkPostOrder(w Walker) []interface{} {
	items := make([]interface{}, 0)
	if !n.isLeaf {
		for _, child := range n.children {
			log.Tracef("walkPostOrder: n=%s child=%s\n", n, child)
			child_items := child.walkPostOrder(w)
			items = append(items, child_items...)
		}
	}
	items = append(items, w(n))
	return items
}

func (n *Node23) walkNodesPostOrder() []*Node23 {
	nodeItems := n.walkPostOrder(func(n *Node23) interface{} { return n })
	nodes := make([]*Node23, len(nodeItems))
	for i := range nodeItems {
		nodes[i] = nodeItems[i].(*Node23)
	}
	return nodes
}
func (n *Node23) hashNode() []byte {
	if n.isLeaf {
		return n.hashLeaf()
	} else {
		return n.hashInternal()
	}
}

func (n *Node23) hashLeaf() []byte {
	ensure(n.isLeaf, "hashLeaf: node is not leaf")
	ensure(n.valueCount() == n.keyCount(), "hashLeaf: insufficient number of values")
	switch n.keyCount() {
	case 2:
		k, nextKey, v := *n.keys[0], n.keys[1], *n.values[0]
		h := hash2(k.Binary(), v.Binary())
		if nextKey == nil {
			return h
		} else {
			return hash2(h, (*nextKey).Binary())
		}
	case 3:
		k1, k2, nextKey, v1, v2 := *n.keys[0], *n.keys[1], n.keys[2], *n.values[0], *n.values[1]
		h1 := hash2(k1.Binary(), v1.Binary())
		h2 := hash2(k2.Binary(), v2.Binary())
		h12 := hash2(h1, h2)
		if nextKey == nil {
			return h12
		} else {
			return hash2(h12, (*nextKey).Binary())
		}
	default:
		ensure(false, fmt.Sprintf("hashLeaf: unexpected keyCount=%d\n", n.keyCount()))
		return []byte{}
	}
}

func (n *Node23) hashInternal() []byte {
	ensure(!n.isLeaf, "hashInternal: node is not internal")
	switch n.childrenCount() {
	case 2:
		child1, child2 := n.children[0], n.children[1]
		return hash2(child1.hashNode(), child2.hashNode())
	case 3:
		child1, child2, child3 := n.children[0], n.children[1], n.children[2]
		return hash2(hash2(child1.hashNode(), child2.hashNode()), child3.hashNode())
	default:
		ensure(false, fmt.Sprintf("hashInternal: unexpected childrenCount=%d\n", n.childrenCount()))
		return []byte{}
	}
}
