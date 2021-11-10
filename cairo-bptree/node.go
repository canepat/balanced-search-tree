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
}

func makeInternalNode(children []*Node23) *Node23 {
	internalKeys := internalKeysFromChildren(children)
	n := &Node23{isLeaf: false, children: children, keys: internalKeys, values: make([]*Felt, 0)}
	return n
}

func makeLeafNode(keys, values []*Felt) *Node23 {
	ensure(len(keys) > 0, "number of keys is zero")
	ensure(len(keys) == len(values), "keys and values have different cardinality")
	leafKeys := append(make([]*Felt, 0, len(keys)), keys...)
	leafValues := append(make([]*Felt, 0, len(values)), values...)
	n := &Node23{isLeaf: true, children: make([]*Node23, 0), keys: leafKeys, values: leafValues}
	return n
}

func makeEmptyLeafNode() (*Node23) {
	return makeLeafNode(make([]*Felt, 1), make([]*Felt, 1))
}

func internalKeysFromChildren(children []*Node23) []*Felt {
	ensure(len(children) > 0, "number of children is zero")
	internalKeys := make([]*Felt, 0, len(children)-1)
	for _, child := range children[:len(children)-1] {
		ensure(child.nextKey() != nil, "child next key is zero")
		nextKey := *child.nextKey()
		internalKeys = append(internalKeys, &nextKey)
	}
	return internalKeys
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

func (n *Node23) upsert(kvItems []KeyValue) (promoted []*Node23, newFirstKey *Felt) {
	log.Tracef("upsert: n=%p kvItems=%v\n", n, kvItems)
	ensure(sort.IsSorted(KeyValueByKey(kvItems)), "kvItems is not sorted by key")
	if len(kvItems) == 0 {
		return []*Node23{n}, nil
	}
	if n == nil {
		n = makeEmptyLeafNode()
	}
	log.Tracef("upsert: n=%p\n", n)
	if n.isLeaf {
		return n.upsertLeaf(kvItems)
	} else {
		return n.upsertInternal(kvItems)
	}
}

func (n *Node23) upsertLeaf(kvItems []KeyValue) (promoted []*Node23, newFirstKey *Felt) {
	log.Tracef("upsertLeaf: n=%p kvItems=%v\n", n, kvItems)
	ensure(n.isLeaf, "node is not leaf")

	currentFirstKey := n.firstKey()
	n.addOrReplaceKeys(kvItems)
	if n.firstKey() != currentFirstKey {
		newFirstKey = n.firstKey()
	} else {
		newFirstKey = nil
	}

	log.Tracef("upsertLeaf: keyCount=%d firstKey=%d\n", n.keyCount(), newFirstKey)
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

func (n *Node23) upsertInternal(kvItems []KeyValue) (promoted []*Node23, newFirstKey *Felt) {
	ensure(!n.isLeaf, "node is not internal")
	log.Tracef("upsertInternal: n=%s keyCount=%d\n", n, n.keyCount())

	itemSubsets := n.splitItems(kvItems)

	log.Tracef("upsertInternal: n=%s itemSubsets=%v\n", n, itemSubsets)
	innerPromotedNodes := make([]*Node23, 0)
	for i := len(n.children)-1; i >= 0; i-- {
		child := n.children[i]
		log.Tracef("upsertInternal: reverse i=%d child=%s itemSubsets[i]=%v\n", i, child, itemSubsets[i])
		childPromotedNodes, childNewFirstKey := child.upsert(itemSubsets[i])
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
	nodes := make([]*Node23, 0)
	if n.childrenCount() > 3 {
		for n.childrenCount() > 3 {
			nodes = append(nodes, makeInternalNode(n.children[:2]))
			n.children = n.children[2:]
		}
		nodes = append(nodes, makeInternalNode(n.children[:]))
		return nodes, newFirstKey
	} else {
		return []*Node23{n}, newFirstKey
	}
}

func (n *Node23) addOrReplaceKeys(kvItems []KeyValue) {
	ensure(n.isLeaf, "addOrReplaceKeys: node is not leaf")
	ensure(len(n.keys) > 0 && len(n.values) > 0, "addOrReplaceKeys: node keys/values are not empty")
	ensure(len(kvItems) > 0, "addOrReplaceKeys: kvItems is not empty")
	log.Debugf("addOrReplaceKeys: keys=%v-%v values=%v-%v kvItems=%v\n", ptr2pte(n.keys), n.keys, ptr2pte(n.values), n.values, kvItems)
	
	nextKey, nextValue := n.nextKey(), n.nextValue()
	log.Tracef("addOrReplaceKeys: nextKey=%p values=%p\n", nextKey, nextValue)

	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]
	log.Tracef("addOrReplaceKeys: keys=%v-%v values=%v-%v\n", ptr2pte(n.keys), n.keys, ptr2pte(n.values), n.values)

	for _, kvPair := range kvItems {
		key, value := kvPair.key, kvPair.value
		keyFound := false
		for i, nKey := range n.keys {
			ensure(nKey != nil, fmt.Sprintf("addOrReplaceKeys: key[%d] is nil in %s", i, n))
			log.Tracef("addOrReplaceKeys: key=%d values=%d nKey=%d\n", key, value, *nKey)
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
	ensure(len(n.keys) > 0, "splitItems: internal node has no keys")
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
