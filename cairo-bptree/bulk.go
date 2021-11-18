package cairo_bptree

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
)

func upsert(n *Node23, kvItems []KeyValue, stats *Stats) (promoted []*Node23, newFirstKey *Felt) {
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
		return upsertLeaf(n, kvItems, stats)
	} else {
		return upsertInternal(n, kvItems, stats)
	}
}

func upsertLeaf(n *Node23, kvItems []KeyValue, stats *Stats) (promoted []*Node23, newFirstKey *Felt) {
	ensure(n.isLeaf, "node is not leaf")
	log.Tracef("upsertLeaf: n=%p kvItems=%v\n", n, kvItems)
	if !n.exposed {
		n.exposed = true
		stats.ExposedCount++
	}

	currentFirstKey := n.firstKey()
	addOrReplaceKeys(n, kvItems)
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

func upsertInternal(n *Node23, kvItems []KeyValue, stats *Stats) (promoted []*Node23, newFirstKey *Felt) {
	ensure(!n.isLeaf, "node is not internal")
	log.Tracef("upsertInternal: n=%s keyCount=%d\n", n, n.keyCount())
	if !n.exposed {
		n.exposed = true
		stats.ExposedCount++
	}

	itemSubsets := splitItems(n, kvItems)

	log.Tracef("upsertInternal: n=%s itemSubsets=%v\n", n, itemSubsets)
	innerPromotedNodes := make([]*Node23, 0)
	for i := len(n.children)-1; i >= 0; i-- {
		child := n.children[i]
		log.Tracef("upsertInternal: reverse i=%d child=%s itemSubsets[i]=%v\n", i, child, itemSubsets[i])
		childPromotedNodes, childNewFirstKey := upsert(child, itemSubsets[i], stats)
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

func addOrReplaceKeys(n *Node23, kvItems []KeyValue) {
	ensure(n.isLeaf, "addOrReplaceKeys: node is not leaf")
	ensure(len(n.keys) > 0 && len(n.values) > 0, "addOrReplaceKeys: node keys/values are not empty")
	ensure(len(kvItems) > 0, "addOrReplaceKeys: kvItems is not empty")
	log.Debugf("addOrReplaceKeys: keys=%v-%v values=%v-%v #kvItems=%d\n", deref(n.keys), n.keys, deref(n.values), n.values, len(kvItems))

	nextKey, nextValue := n.nextKey(), n.nextValue()
	log.Tracef("addOrReplaceKeys: nextKey=%p nextValue=%p\n", nextKey, nextValue)

	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]
	log.Tracef("addOrReplaceKeys: keys=%v-%v values=%v-%v kvItems=%v\n", deref(n.keys), n.keys, deref(n.values), n.values, kvItems)

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
		if !keyFound {
			n.keys = append(n.keys, &key)
			n.values = append(n.values, &value)
		}
	}
	log.Tracef("addOrReplaceKeys: keys=%v-%v values=%v-%v\n", deref(n.keys), n.keys, deref(n.values), n.values)

	sort.Slice(n.keys, func(i, j int) bool { return *n.keys[i] < *n.keys[j] })
	sort.Slice(n.values, func(i, j int) bool { return *n.values[i] < *n.values[j] })
	n.keys = append(n.keys, nextKey)
	n.values = append(n.values, nextValue)
	log.Debugf("addOrReplaceKeys: keys=%v-%v values=%v-%v\n", deref(n.keys), n.keys, deref(n.values), n.values)
}

func splitItems(n *Node23, kvItems []KeyValue) [][]KeyValue {
	ensure(!n.isLeaf, "splitItems: node is not internal")
	ensure(len(n.keys) > 0, fmt.Sprintf("splitItems: internal node %s has no keys", n))
	log.Tracef("splitItems: keys=%v-%v kvItems=%v\n", deref(n.keys), n.keys, kvItems)
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

func delete(n *Node23, keysToDelete []Felt, stats *Stats) (deleted *Node23, nextKey *Felt) {
	log.Tracef("delete: n=%p keysToDelete=%v\n", n, keysToDelete)
	ensure(sort.IsSorted(Keys(keysToDelete)), "keysToDelete are not sorted")
	if n == nil || len(keysToDelete) == 0 {
		return n, nil
	}
	if n.isLeaf {
		deleteKeys(n, keysToDelete)
		if n.keyCount() == 1 {
			return nil, n.nextKey()
		} else {
			return n, nil
		}
	} else {
		keySubsets := splitKeys(n, keysToDelete)
		for i := len(n.children) - 1; i >= 0; i-- {
			child, childNextKey := delete(n.children[i], keySubsets[i], stats)
			log.Tracef("delete: n=%s child=%s childNextKey=%s\n", n, child, pointerValue(childNextKey))
			// TODO: ignore child, childNextKey because handled in update2Node/update3Node?
		}
		switch len(n.children) {
		case 2:
			nextKey = update2Node(n)
		case 3:
			nextKey = update3Node(n)
		default:
			ensure(false, fmt.Sprintf("unexpected number of children in %s", n))
		}
		return demote(n, nextKey)
	}
}

func deleteKeys(n *Node23, keysToDelete []Felt) {
	switch n.keyCount() {
	case 2:
		if Keys(keysToDelete).Contains(*n.keys[0]) {
			n.keys = n.keys[1:]
			n.values = n.values[1:]
		}
	case 3:
		if Keys(keysToDelete).Contains(*n.keys[0]) {
			if Keys(keysToDelete).Contains(*n.keys[1]) {
				n.keys = n.keys[2:]
				n.values = n.values[2:]
			} else {
				n.keys = n.keys[1:]
				n.values = n.values[1:]
			}
		} else {
			if Keys(keysToDelete).Contains(*n.keys[1]) {
				//n.keys = append(n.keys[0:1], n.keys[2:]...)
				n.keys = append(n.keys[:1], n.keys[2])
				//n.values = append(n.values[0:1], n.values[2:]...)
				n.values = append(n.values[:1], n.values[2])
			}
		}
	default:
		ensure(false, fmt.Sprintf("unexpected number of keys in %s", n))
	}
}

func splitKeys(n *Node23, keysToDelete []Felt) [][]Felt {
	ensure(!n.isLeaf, "splitKeys: node is not internal")
	ensure(len(n.keys) > 0, fmt.Sprintf("splitKeys: internal node %s has no keys", n))
	log.Tracef("splitKeys: keys=%v-%v keysToDelete=%v\n", deref(n.keys), n.keys, keysToDelete)
	keySubsets := make([][]Felt, 0)
	for i, key := range n.keys {
		splitIndex := sort.Search(len(keysToDelete), func(i int) bool { return keysToDelete[i] >= *key })
		log.Tracef("splitKeys: key=%d-(%p) splitIndex=%d\n", *key, key, splitIndex)
		keySubsets = append(keySubsets, keysToDelete[:splitIndex])
		keysToDelete = keysToDelete[splitIndex:]
		if i == len(n.keys)-1 {
			keySubsets = append(keySubsets, keysToDelete)
		}
	}
	ensure(len(keySubsets) == len(n.children), "key subsets and children have different cardinality")
	return keySubsets
}

func update2Node(n *Node23) *Felt {
	ensure(len(n.children) == 2, "update2Node: wrong number of children")
	nodeA, nodeC := n.children[0], n.children[1]
	if nodeA.isEmpty() {
		if nodeC.isEmpty() {
			/* A is empty, a_next is the "next key"; C is empty, c_next is the "next key" */
			n.children = n.children[:0]
			n.isLeaf = true
			return nodeC.nextKey()
		} else {
			/* A is empty, a_next is the "next key"; C is not empty */
			n.children = n.children[1:]
			return nodeA.nextKey()
		}
	} else {
		if nodeC.isEmpty() {
			/* A is not empty; C is empty, c_next is the "next key" */
			n.children = n.children[:1]
			nodeA.setNextKey(nodeC.nextKey())
			return nil
		} else {
			/* A is not empty; C is not empty */
			return nil
		}
	}
}

func update3Node(n *Node23) *Felt {
	ensure(len(n.children) == 3, "update3Node: wrong number of children")
	nodeA, nodeB, nodeC := n.children[0], n.children[1], n.children[2]
	if nodeA.isEmpty() {
		if nodeB.isEmpty() {
			if nodeC.isEmpty() {
				/* A is empty, a_next is the "next key"; B is empty, b_next is the "next key"; C is empty, c_next is the "next key" */
				n.children = n.children[:0]
				n.isLeaf = true
				return nodeC.nextKey()
			} else {
				/* A is empty, a_next is the "next key"; B is empty, b_next is the "next key"; C is not empty */
				n.children = n.children[2:]
				return nodeB.nextKey()
			}
		} else {
			if nodeC.isEmpty() {
				/* A is empty, a_next is the "next key"; B is not empty; C is empty, c_next is the "next key" */
				n.children = n.children[1:2]
				nodeB.setNextKey(nodeC.nextKey())
				return nodeA.nextKey()
			} else {
				/* A is empty, a_next is the "next key"; B is not empty; C is not empty */
				n.children = n.children[1:]
				return nodeA.nextKey()
			}
		}
	} else {
		if nodeB.isEmpty() {
			if nodeC.isEmpty() {
				/* A is not empty; B is empty, b_next is the "next key"; C is empty, c_next is the "next key" */
				n.children = n.children[:1]
				nodeA.setNextKey(nodeC.nextKey())
				return nil
			} else {
				/* A is not empty; B is empty, b_next is the "next key"; C is not empty */
				n.children = append(n.children[:1], n.children[2])
				nodeA.setNextKey(nodeB.nextKey())
				return nil
			}
		} else {
			if nodeC.isEmpty() {
				/* A is not empty; B is not empty; C is empty, c_next is the "next key" */
				n.children = n.children[:2]
				nodeB.setNextKey(nodeC.nextKey())
				return nil
			} else {
				/* A is not empty; B is not empty; C is not empty */
				return nil
			}
		}
	}
}

func demote(n *Node23, nextKey *Felt) (*Node23, *Felt) {
	if len(n.children) == 1 {
		return n.children[0], nextKey
	} else if len(n.children) == 2 {
		if n.children[0].keyCount() == 2 && n.children[1].keyCount() == 2 {
			ensure(n.children[0].isLeaf, fmt.Sprintf("unexpected internal node as 1st child: %s", n))
			keys := []*Felt{n.children[0].firstKey(), n.children[1].firstKey(), n.children[1].nextKey()}
			values := []*Felt{n.children[0].firstValue(), n.children[1].firstValue(), n.children[1].nextValue()}
			return makeLeafNode(keys, values), nextKey
		}
	}
	return n, nextKey
}

func updateNextKey(n *Node23, nextKey *Felt) {
}
