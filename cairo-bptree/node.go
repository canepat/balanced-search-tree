package cairo_bptree

import (
	"fmt"
	"sort"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

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

func (n *Node23) upsert(kvItems []KeyValue, stats *Stats) (promoted []*Node23, newFirstKey *Felt) {
	log.Tracef("upsert: n=%p kvItems=%v\n", n, kvItems)
	ensure(sort.IsSorted(KeyValueByKey(kvItems)), "kvItems are not sorted by key")
	if len(kvItems) == 0 {
		return []*Node23{n}, nil
	}
	if n == nil {
		n = makeEmptyLeafNode()
	}
	log.Tracef("upsert: n=%p\n", n)
	if n.isLeaf {
		return n.upsertLeaf(kvItems, stats)
	} else {
		return n.upsertInternal(kvItems, stats)
	}
}

func (n *Node23) upsertLeaf(kvItems []KeyValue, stats *Stats) (promoted []*Node23, newFirstKey *Felt) {
	ensure(n.isLeaf, "node is not leaf")
	log.Tracef("upsertLeaf: n=%p kvItems=%v\n", n, kvItems)
	if !n.exposed {
		n.exposed = true
		stats.ExposedCount++
	}

	currentFirstKey := n.firstKey()
	n.addOrReplaceKeys(kvItems)
	if n.firstKey() != currentFirstKey {
		newFirstKey = n.firstKey()
	} else {
		newFirstKey = nil
	}

	log.Tracef("upsertLeaf: keyCount=%d firstKey=%d\n", n.keyCount(), *n.firstKey())
	if n.keyCount() > 3 {
		nodes := make([]*Node23, 0)
		for n.keyCount() > 3 {
			newLeaf := makeLeafNode(n.keys[:3], n.values[:3])
			log.Tracef("upsertLeaf: newLeaf=%s\n", newLeaf)
			nodes = append(nodes, newLeaf)
			n.keys, n.values = n.keys[2:], n.values[2:]
			log.Tracef("upsertLeaf: updated n=%s\n", n)
		}
		newLeaf := makeLeafNode(n.keys[:], n.values[:])
		log.Tracef("upsertLeaf: last newLeaf=%s\n", newLeaf)
		nodes = append(nodes, newLeaf)
		return nodes, newFirstKey
	} else {
		return []*Node23{n}, newFirstKey
	}
}

func (n *Node23) upsertInternal(kvItems []KeyValue, stats *Stats) (promoted []*Node23, newFirstKey *Felt) {
	ensure(!n.isLeaf, "node is not internal")
	log.Tracef("upsertInternal: n=%s keyCount=%d\n", n, n.keyCount())
	if !n.exposed {
		n.exposed = true
		stats.ExposedCount++
	}

	itemSubsets := n.splitItems(kvItems)

	log.Tracef("upsertInternal: n=%s itemSubsets=%v\n", n, itemSubsets)
	innerPromotedNodes := make([]*Node23, 0)
	for i := len(n.children)-1; i >= 0; i-- {
		child := n.children[i]
		log.Tracef("upsertInternal: reverse i=%d child=%s itemSubsets[i]=%v\n", i, child, itemSubsets[i])
		childPromotedNodes, childNewFirstKey := child.upsert(itemSubsets[i], stats)
		innerPromotedNodes = append(childPromotedNodes, innerPromotedNodes...)
		log.Tracef("upsertInternal: i=%d innerPromotedNodes=%v childNewFirstKey=%s\n", i, innerPromotedNodes, pointerValue(childNewFirstKey))
		if childNewFirstKey != nil {
			if i > 0 {
				// Handle newFirstKey here
				previousChild := n.children[i-1]
				if previousChild.isLeaf {
					ensure(len(previousChild.keys) > 0, "upsertInternal: previousChild has no keys")
					previousChild.setNextKey(childNewFirstKey)
				} else {
					ensure(len(previousChild.children) > 0, "upsertInternal: previousChild has no children")
					previousChild.lastChild().setNextKey(childNewFirstKey)
				}
				log.Tracef("upsertInternal: i=%d update last next key childNewFirstKey=%s\n", i, pointerValue(childNewFirstKey))
			} else {
				// Propagate newFirstKey up
				newFirstKey = childNewFirstKey
				log.Tracef("upsertInternal: i=%d propagated newFirstKey=%s\n", i, pointerValue(newFirstKey))
			}
		}
	}
	log.Tracef("upsertInternal: n=%s innerPromotedNodes=%v\n", n, innerPromotedNodes)
	n.children = innerPromotedNodes
	n.keys = internalKeysFromChildren(n.children)
	log.Tracef("upsertInternal: n=%s newFirstKey=%s\n", n, pointerValue(newFirstKey))
	if n.childrenCount() > 3 {
		nodes := make([]*Node23, 0)
		promotedKeys := make([]*Felt, 0)
		for n.childrenCount() > 3 {
			nodes = append(nodes, makeInternalNode(n.children[:2]))
			n.children = n.children[2:]
			promotedKeys = append(promotedKeys, n.keys[1])
		}
		nodes = append(nodes, makeInternalNode(n.children[:]))
		n.children = nodes
		n.keys = promotedKeys
		return []*Node23{n}, newFirstKey
	} else {
		return []*Node23{n}, newFirstKey
	}
}

func (n *Node23) addOrReplaceKeys(kvItems []KeyValue) {
	ensure(n.isLeaf, "addOrReplaceKeys: node is not leaf")
	ensure(len(n.keys) > 0 && len(n.values) > 0, "addOrReplaceKeys: node keys/values are not empty")
	ensure(len(kvItems) > 0, "addOrReplaceKeys: kvItems is not empty")
	log.Debugf("addOrReplaceKeys: keys=%v-%v values=%v-%v #kvItems=%d\n", ptr2pte(n.keys), n.keys, ptr2pte(n.values), n.values, len(kvItems))
	
	nextKey, nextValue := n.nextKey(), n.nextValue()
	log.Tracef("addOrReplaceKeys: nextKey=%p nextValue=%p\n", nextKey, nextValue)

	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]
	log.Tracef("addOrReplaceKeys: keys=%v-%v values=%v-%v kvItems=%v\n", ptr2pte(n.keys), n.keys, ptr2pte(n.values), n.values, kvItems)

	// TODO: change algorithm
	// kvItems are ordered by key, search there using n.keys that are 1 or 2 by design, insert n.keys[] if not already present
	// change kvItems to KeyValues struct composed by keys []Felt, values []Felt
	for _, kvPair := range kvItems {
		key, value := kvPair.key, kvPair.value
		keyFound := false
		for i, nKey := range n.keys {
			ensure(nKey != nil, fmt.Sprintf("addOrReplaceKeys: key[%d] is nil in %p", i, n))
			log.Tracef("addOrReplaceKeys: key=%d value=%d nKey=%d\n", key, value, *nKey)
			if *nKey == key {
				keyFound = true
				n.values[i] = &value
				break
			}
		}
		if (!keyFound) {
			n.keys = append(n.keys, &key)
			n.values = append(n.values, &value)
		}
	}
	log.Tracef("addOrReplaceKeys: keys=%v-%v values=%v-%v\n", ptr2pte(n.keys), n.keys, ptr2pte(n.values), n.values)

	sort.Slice(n.keys, func(i, j int) bool { return *n.keys[i] < *n.keys[j] })
	sort.Slice(n.values, func(i, j int) bool { return *n.values[i] < *n.values[j] })
	n.keys = append(n.keys, nextKey)
	n.values = append(n.values, nextValue)
	log.Debugf("addOrReplaceKeys: keys=%v-%v values=%v-%v\n", ptr2pte(n.keys), n.keys, ptr2pte(n.values), n.values)
}

func (n *Node23) splitItems(kvItems []KeyValue) [][]KeyValue {
	ensure(!n.isLeaf, "splitItems: node is not internal")
	ensure(len(n.keys) > 0, fmt.Sprintf("splitItems: internal node %s has no keys", n))
	log.Tracef("splitItems: keys=%v-%v kvItems=%v\n", ptr2pte(n.keys), n.keys, kvItems)
	itemSubsets := make([][]KeyValue, 0)
	for i, key := range n.keys {
		splitIndex := sort.Search(len(kvItems), func(i int) bool { return kvItems[i].key >= *key })
		log.Tracef("splitItems: key=%d-(%p) splitIndex=%d\n", *key, key, splitIndex)
		itemSubsets = append(itemSubsets, kvItems[:splitIndex])
		kvItems = kvItems[splitIndex:]
		if i == len(n.keys)-1 {
			itemSubsets = append(itemSubsets, kvItems)
		}
	}
	ensure(len(itemSubsets) == len(n.children), "item subsets and children have different cardinality")
	return itemSubsets
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