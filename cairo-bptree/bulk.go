package cairo_bptree

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
)

func upsert(n *Node23, kvItems KeyValues, stats *Stats) (nodes []*Node23, newFirstKey *Felt, intermediateKeys []*Felt) {
	ensure(sort.IsSorted(kvItems), "kvItems are not sorted by key")

	if kvItems.Len() == 0 && n == nil {
		return []*Node23{n}, nil, []*Felt{}
	}
	if n == nil {
		n = makeEmptyLeafNode()
	}
	if n.isLeaf {
		return upsertLeaf(n, kvItems, stats)
	} else {
		return upsertInternal(n, kvItems, stats)
	}
}

func upsertLeaf(n *Node23, kvItems KeyValues, stats *Stats) (nodes []*Node23, newFirstKey *Felt, intermediateKeys []*Felt) {
	ensure(n.isLeaf, "node is not leaf")

	if kvItems.Len() == 0 {
		if n.nextKey() != nil {
			intermediateKeys = append(intermediateKeys, n.nextKey())
		}
		return []*Node23{n}, nil, intermediateKeys
	}

	if !n.exposed {
		n.exposed = true
		stats.ExposedCount++
	}

	currentFirstKey := n.firstKey()
	addOrReplaceLeaf(n, kvItems)
	if n.firstKey() != currentFirstKey {
		newFirstKey = n.firstKey()
	} else {
		newFirstKey = nil
	}

	if n.keyCount() > 3 {
		for n.keyCount() > 3 {
			newLeaf := makeLeafNode(n.keys[:3], n.values[:3])
			intermediateKeys = append(intermediateKeys, n.keys[2])
			nodes = append(nodes, newLeaf)
			n.keys, n.values = n.keys[2:], n.values[2:]
		}
		newLeaf := makeLeafNode(n.keys[:], n.values[:])
		if n.nextKey() != nil {
			intermediateKeys = append(intermediateKeys, n.nextKey())
		}
		nodes = append(nodes, newLeaf)
		return nodes, newFirstKey, intermediateKeys
	} else {
		if n.nextKey() != nil {
			intermediateKeys = append(intermediateKeys, n.nextKey())
		}
		return []*Node23{n}, newFirstKey, intermediateKeys
	}
}

func upsertInternal(n *Node23, kvItems KeyValues, stats *Stats) (nodes []*Node23, newFirstKey *Felt, intermediateKeys []*Felt) {
	ensure(!n.isLeaf, "node is not internal")

	if kvItems.Len() == 0 {
		if n.lastChild().nextKey() != nil {
			intermediateKeys = append(intermediateKeys, n.lastChild().nextKey())
		}
		return []*Node23{n}, nil, intermediateKeys
	}

	if !n.exposed {
		n.exposed = true
		stats.ExposedCount++
	}

	itemSubsets := splitItems(n, kvItems)

	newChildren := make([]*Node23, 0)
	newKeys := make([]*Felt, 0)
	for i := len(n.children)-1; i >= 0; i-- {
		child := n.children[i]
		childNodes, childNewFirstKey, childIntermediateKeys := upsert(child, itemSubsets[i], stats)
		newChildren = append(childNodes, newChildren...)
		newKeys = append(childIntermediateKeys, newKeys...)
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
			} else {
				// Propagate newFirstKey up
				newFirstKey = childNewFirstKey
			}
		}
	}

	n.children = newChildren
	if n.childrenCount() > 3 {
		ensure(len(newKeys) >= n.childrenCount()-1 || n.childrenCount() % 2 == 0 && n.childrenCount() % len(newKeys) == 0, "upsertInternal: inconsistent #children vs #newKeys")
		var hasIntermediateKeys bool
		if len(newKeys) == n.childrenCount()-1 || len(newKeys) == n.childrenCount() {
			/* Groups are: 2,2...2 or 3 */
			hasIntermediateKeys = true
		} else {
			/* Groups are: 2,2...2 */
			hasIntermediateKeys = false
		}
		for n.childrenCount() > 3 {
			nodes = append(nodes, makeInternalNode(n.children[:2], newKeys[:1]))
			n.children = n.children[2:]
			if hasIntermediateKeys {
				intermediateKeys = append(intermediateKeys, newKeys[1])
				newKeys = newKeys[2:]
			} else {
				newKeys = newKeys[1:]
			}
		}
		ensure(n.childrenCount() > 0 && len(newKeys) > 0, "upsertInternal: inconsistent #children vs #newKeys")
		if n.childrenCount() == 2 {
			ensure(len(newKeys) > 0, "upsertInternal: inconsistent #newKeys")
			nodes = append(nodes, makeInternalNode(n.children[:], newKeys[:1]))
			intermediateKeys = append(intermediateKeys, newKeys[1:]...)
		} else if n.childrenCount() == 3 {
			ensure(len(newKeys) > 1, "upsertInternal: inconsistent #newKeys")
			nodes = append(nodes, makeInternalNode(n.children[:], newKeys[:2]))
			intermediateKeys = append(intermediateKeys, newKeys[2:]...)
		} else {
			ensure(false, fmt.Sprintf("upsertInternal: inconsistent #children=%d #newKeys=%d\n", n.childrenCount(), len(newKeys)))
		}
		return nodes, newFirstKey, intermediateKeys
	} else { // n.childrenCount() is 2 or 3
		ensure(len(newKeys) > 0, "upsertInternal: newKeys count is zero")
		if len(newKeys) == len(n.children) {
			n.keys = newKeys[:len(newKeys)-1]
			intermediateKeys = append(intermediateKeys, newKeys[len(newKeys)-1])
		} else {
			n.keys = newKeys
		}
		return []*Node23{n}, newFirstKey, intermediateKeys
	}
}

func addOrReplaceLeaf(n *Node23, kvItems KeyValues) {
	ensure(n.isLeaf, "addOrReplaceLeaf: node is not leaf")
	ensure(len(n.keys) > 0 && len(n.values) > 0, "addOrReplaceLeaf: node keys/values are empty")
	ensure(len(kvItems.keys) > 0 && len(kvItems.keys) == len(kvItems.values), "addOrReplaceLeaf: invalid kvItems")

	// Temporarily remove next key/value
	nextKey, nextValue := n.nextKey(), n.nextValue()

	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]

	// kvItems are ordered by key: search there using n.keys that here are 1 or 2 by design (0 just for empty tree)
	switch (n.keyCount()) {
	case 0:
		n.keys = append(n.keys, kvItems.keys...)
		n.values = append(n.values, kvItems.values...)
	case 1:
		addOrReplaceLeaf1(n, kvItems)
	case 2:
		addOrReplaceLeaf2(n, kvItems)
	default:
		ensure(false, fmt.Sprintf("addOrReplaceLeaf: invalid key count %d", n.keyCount()))
	}

	//ensure(sort.IsSorted(Keys(deref(n.keys))), "addOrReplaceLeaf: keys not ordered")
	
	// Restore next key/value
	n.keys = append(n.keys, nextKey)
	n.values = append(n.values, nextValue)
}

func addOrReplaceLeaf1(n *Node23, kvItems KeyValues) {
	ensure(n.isLeaf, "addOrReplaceLeaf1: node is not leaf")
	ensure(n.keyCount() == 1, "addOrReplaceLeaf1: leaf has not 1 *canonical* key")

	key0, value0 := n.keys[0], n.values[0]
	index0 := sort.Search(kvItems.Len(), func(i int) bool { return *kvItems.keys[i] >= *key0 })
	if index0 < kvItems.Len() {
		// Insert keys/values concatenating new ones around key0
		n.keys = append(make([]*Felt, 0), kvItems.keys[:index0]...)
		if *kvItems.keys[index0] != *key0 {
			n.keys = append(n.keys, key0)
		}
		n.keys = append(n.keys, kvItems.keys[index0:]...)

		n.values = append(make([]*Felt, 0), kvItems.values[:index0]...)
		if *kvItems.keys[index0] != *key0 {
			n.values = append(n.values, value0)
		}
		n.values = append(n.values, kvItems.values[index0:]...)
	} else {
		// key0 greater than any input key
		n.keys = append(kvItems.keys, key0)
		n.values = append(kvItems.values, value0)
	}
}

func addOrReplaceLeaf2(n *Node23, kvItems KeyValues) {
	ensure(n.isLeaf, "addOrReplaceLeaf2: node is not leaf")
	ensure(n.keyCount() == 2, "addOrReplaceLeaf2: leaf has not 2 *canonical* keys")

	key0, value0, key1, value1 := n.keys[0], n.values[0], n.keys[1], n.values[1]
	index0 := sort.Search(kvItems.Len(), func(i int) bool { return *kvItems.keys[i] >= *key0 })
	index1 := sort.Search(kvItems.Len(), func(i int) bool { return *kvItems.keys[i] >= *key1 })
	ensure(index1 >= index0, "addOrReplaceLeaf2: keys not ordered")
	if index0 < kvItems.Len() {
		if index1 < kvItems.Len() {
			// Insert keys/values concatenating new ones around key0 and key1
			n.keys = append(make([]*Felt, 0), kvItems.keys[:index0]...)
			if *kvItems.keys[index0] != *key0 {
				n.keys = append(n.keys, key0)
			}
			n.keys = append(n.keys, kvItems.keys[index0:index1]...)
			if *kvItems.keys[index1] != *key1 {
				n.keys = append(n.keys, key1)
			}
			n.keys = append(n.keys, kvItems.keys[index1:]...)

			n.values = append(make([]*Felt, 0), kvItems.values[:index0]...)
			if *kvItems.keys[index0] != *key0 {
				n.values = append(n.values, value0)
			}
			n.values = append(n.values, kvItems.values[index0:index1]...)
			if *kvItems.keys[index1] != *key1 {
				n.values = append(n.values, value1)
			}
			n.values = append(n.values, kvItems.values[index1:]...)
		} else {
			// Insert keys/values concatenating new ones around key0, then add key1
			n.keys = append(make([]*Felt, 0), kvItems.keys[:index0]...)
			if *kvItems.keys[index0] != *key0 {
				n.keys = append(n.keys, key0)
			}
			n.keys = append(n.keys, kvItems.keys[index0:]...)
			n.keys = append(n.keys, key1)
	
			n.values = append(make([]*Felt, 0), kvItems.values[:index0]...)
			if *kvItems.keys[index0] != *key0 {
				n.values = append(n.values, value0)
			}
			n.values = append(n.values, kvItems.values[index0:]...)
			n.values = append(n.values, value1)
		}
	} else {
		ensure(index1 == index0, "addOrReplaceLeaf2: keys not ordered")
		// Both key0 and key1 greater than any input key
		n.keys = append(kvItems.keys, key0, key1)
		n.values = append(kvItems.values, value0, value1)
	}
}

func splitItems(n *Node23, kvItems KeyValues) []KeyValues {
	ensure(!n.isLeaf, "splitItems: node is not internal")
	ensure(len(n.keys) > 0, "splitItems: internal node has no keys")

	itemSubsets := make([]KeyValues, 0)
	for i, key := range n.keys {
		splitIndex := sort.Search(kvItems.Len(), func(i int) bool { return *kvItems.keys[i] >= *key })
		itemSubsets = append(itemSubsets, KeyValues{kvItems.keys[:splitIndex], kvItems.values[:splitIndex]})
		kvItems = KeyValues{kvItems.keys[splitIndex:], kvItems.values[splitIndex:]}
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
		return deleteLeaf(n, keysToDelete, stats)
	} else {
		return deleteInternal(n, keysToDelete, stats)
	}
}

func deleteLeaf(n *Node23, keysToDelete []Felt, stats *Stats) (deleted *Node23, nextKey *Felt) {
	ensure(n.isLeaf, fmt.Sprintf("node %s is not leaf", n))
	currentFirstKey := n.firstKey()
	deleteLeafKeys(n, keysToDelete)
	if n.keyCount() == 1 {
		return nil, n.nextKey()
	} else if n.firstKey() != currentFirstKey {
		return n, n.firstKey()
	} else {
		return n, nil
	}
}

func deleteLeafKeys(n *Node23, keysToDelete []Felt) {
	ensure(n.isLeaf, "deleteLeafKeys: node is not leaf")
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
				n.keys = append(n.keys[:1], n.keys[2])
				n.values = append(n.values[:1], n.values[2])
			}
		}
	default:
		ensure(false, fmt.Sprintf("unexpected number of keys in %s", n))
	}
}

func deleteInternal(n *Node23, keysToDelete []Felt, stats *Stats) (deleted *Node23, nextKey *Felt) {
	ensure(!n.isLeaf, fmt.Sprintf("node %s is not internal", n))
	keySubsets := splitKeys(n, keysToDelete)
	for i := len(n.children) - 1; i >= 0; i-- {
		child, childNextKey := delete(n.children[i], keySubsets[i], stats)
		log.Tracef("delete: n=%s child=%s childNextKey=%s\n", n, child, pointerValue(childNextKey))
		if i > 0 {
			previousChild := n.children[i-1]
			if child != nil && !child.isLeaf {
				if child.childrenCount() == 0 {
					child.keys = child.keys[:0]
				} else if child.childrenCount() == 1 {
					previousChild.children = append(previousChild.children, child.firstChild())
					firstKey := child.firstChild().firstKey()
					previousChild.keys = append(previousChild.keys, firstKey)
					child.children = child.children[:0]
					child.keys = child.keys[:0]
				}
			}
			if child == nil {
				ensure(len(n.keys) >= i, "delete: n has insufficient keys")
				n.keys = append(n.keys[:i-1], n.keys[i:]...)
			}
			if childNextKey != nil {
				if previousChild.isLeaf {
					ensure(len(previousChild.keys) > 0, "delete: previousChild has no keys")
					previousChild.setNextKey(childNextKey)
					if n.keyCount() >= i {
						n.keys[i-1] = childNextKey
					}
				} else {
					ensure(len(previousChild.children) > 0, "delete: previousChild has no children")
					previousChild.lastChild().setNextKey(childNextKey)
				}
			}
		} else {
			nextChild := n.children[i+1]
			if child != nil && !child.isLeaf {
				if child.childrenCount() == 0 {
					child.keys = child.keys[:0]
				} else if child.childrenCount() == 1 {
					nextChild.children = append([]*Node23{child.firstChild()}, nextChild.children...)
					nextKey := child.firstChild().nextKey()
					nextChild.keys = append([]*Felt{nextKey}, nextChild.keys...)
					child.children = child.children[:0]
				}
			}
			if child == nil && n.keyCount() > 0 {
				n.keys = n.keys[1:]
			}
			if childNextKey != nil {
				nextKey = childNextKey
			}
		}
	}
	switch len(n.children) {
	case 2:
		nextKey = update2Node(n)
	case 3:
		nextKey = update3Node(n)
	default:
		ensure(false, fmt.Sprintf("unexpected number of children in %s", n))
	}
	return n, nextKey
}

func splitKeys(n *Node23, keysToDelete []Felt) [][]Felt {
	ensure(!n.isLeaf, "splitKeys: node is not internal")
	ensure(len(n.keys) > 0, fmt.Sprintf("splitKeys: internal node %s has no keys", n))

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
			n.keys = n.keys[:0]
			n.isLeaf = true
			if nodeC.isLeaf {
				return nodeC.nextKey()
			}
			return nil
		} else {
			/* A is empty, a_next is the "next key"; C is not empty */
			n.children = n.children[1:]
			if nodeA.isLeaf {
				return nodeA.nextKey()
			}
			return nil
		}
	} else {
		if nodeC.isEmpty() {
			/* A is not empty; C is empty, c_next is the "next key" */
			n.children = n.children[:1]
			if nodeC.isLeaf {
				nodeA.setNextKey(nodeC.nextKey())
			}
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
				if nodeA.isLeaf {
					return nodeA.nextKey()
				}
				return nil
			} else {
				/* A is empty, a_next is the "next key"; B is not empty; C is not empty */
				n.children = n.children[1:]
				if nodeA.isLeaf {
					return nodeA.nextKey()
				}
				return nil
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

func demote(node *Node23, nextKey *Felt) (*Node23, *Felt) {
	if node == nil {
		return nil, nextKey
	} else if len(node.children) == 0{
		if len(node.keys) == 0 {
			return nil, nextKey
		} else {
			return node, nextKey
		}
	} else if len(node.children) == 1 {
		return node.children[0], nextKey
	} else if len(node.children) == 2 {
		if node.children[0].keyCount() == 2 && node.children[1].keyCount() == 2 {
			ensure(node.children[0].isLeaf, fmt.Sprintf("unexpected internal node as 1st child: %s", node))
			keys := []*Felt{node.children[0].firstKey(), node.children[1].firstKey(), node.children[1].nextKey()}
			values := []*Felt{node.children[0].firstValue(), node.children[1].firstValue(), node.children[1].nextValue()}
			return makeLeafNode(keys, values), nextKey
		}
	}
	return node, nextKey
}
