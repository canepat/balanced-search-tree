package cairo_bptree

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

type Tree23 struct {
	root	*Node23
}

func NewTree23() *Tree23 {
	return &Tree23{}
}

func (t *Tree23) IsTwoThree() bool {
	if t.root == nil {
		return true
	}
	return t.root.isTwoThree()
}

func (t *Tree23) Graph(filename string, debug bool) {
	t.root.graph(filename, debug)
}

func (t *Tree23) GraphAndPicture(filename string, debug bool) error {
	return t.root.graphAndPicture(filename, debug)
}

type Walker func(*Node23) interface{}

func (t *Tree23) WalkPostOrder(w Walker) []interface{} {
	if t.root == nil {
		return make([]interface{}, 0)
	}
	return t.root.walkPostOrder(w)
}

func (t *Tree23) WalkKeysPostOrder() ([]Felt) {
	key_pointers := make([]*Felt, 0)
	t.WalkPostOrder(func(n *Node23) interface{} {
		if n.isLeaf && n.keyCount() > 0 {
			log.Tracef("WalkKeysInOrder: L n=%p n.keys=%v\n", n, n.keys)
			key_pointers = append(key_pointers, n.keys[:len(n.keys)-1]...)
		}
		return nil
	})
	keys := pointer2pointee(key_pointers)
	log.Tracef("WalkKeysInOrder: keys=%v\n", keys)
	return keys
}

func (t *Tree23) Upsert(kvItems []KeyValue) (*Tree23) {
	log.Tracef("Upsert: t=%p root=%p kvItems=%v\n", t, t.root, kvItems)
	nodes := t.root.upsert(kvItems)
	ensure(len(nodes) > 0, "nodes length is zero")
	if len(nodes) == 1 {
		t.root = nodes[0]
	} else {
		t.root = t.promote(nodes)
	}
	log.Tracef("Upsert: t=%p root=%p\n", t, t.root)
	return t
}

func (t *Tree23) promote(nodes []*Node23) (*Node23) {
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

type Node23 struct {
	isLeaf		bool
	children	[]*Node23
	keys		[]*Felt
	values		[]*Felt
}

type KeyValue struct {
	key	Felt
	value	Felt
}

func (n *Node23) String() string {
	s := fmt.Sprintf("{%p isLeaf=%t keys=%v-%v children=[", n, n.isLeaf, pointer2pointee(n.keys), n.keys)
	for i, child := range n.children {
		s += fmt.Sprintf("%p", child)
		if i != len(n.children)-1 {
			s += " "
		}
	}
	s += "]}"
	return s
}

func (node *Node23) graph(filename string, debug bool) {
	colors := []string{"#FDF3D0", "#DCE8FA", "#D9E7D6", "#F1CFCD", "#F5F5F5", "#E1D5E7", "#FFE6CC", "white"}
	f, err := os.OpenFile(filename + ".dot", os.O_RDWR | os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer func () {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	if node == nil {
		if _, err := f.WriteString("strict digraph {\nnode [shape=record];}\n"); err != nil {
			log.Fatal(err)
		}
		return
	}
	if _, err := f.WriteString("strict digraph {\nnode [shape=record];\n"); err != nil {
		log.Fatal(err)
	}
	for _, n := range node.walkNodesPostOrder() {
		log.Tracef("graph: %+v nesting=%d\n", n, 0)
		left, down, right := "", "", ""
		switch n.childrenCount() {
		case 1:
			left = "<L>L"
		case 2:
			left = "<L>L"
			right = "<R>R"
		case 3:
			left = "<L>L"
			down = "<D>D"
			right = "<R>R"
		}
		var nodeId string
		if debug {
			nodeId = fmt.Sprintf("k=%v-%v", pointer2pointee(n.keys), n.keys)
		} else {
			nodeId = fmt.Sprintf("k=%v", pointer2pointee(n.keys))
		}
		s := fmt.Sprintln(/*n.path*/n.rawPointer(), " [label=\"", left, "|{<C>", nodeId, "|", down, "}|", right, "\" style=filled fillcolor=\"", colors[0], "\"];")
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}
	for _, n := range node.walkNodesPostOrder() {
		var treeLeft, treeDown, treeRight *Node23
		switch n.childrenCount() {
		case 1:
			treeLeft = n.children[0]
		case 2:
			treeLeft = n.children[0]
			treeRight = n.children[1]
		case 3:
			treeLeft = n.children[0]
			treeDown = n.children[1]
			treeRight = n.children[2]
		}
		if treeLeft != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.rawPointer(), ":L -> ", treeLeft.rawPointer(), ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if treeDown != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.rawPointer(), ":D -> ", treeDown.rawPointer(), ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if treeRight != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.rawPointer(), ":R -> ", treeRight.rawPointer(), ":C;")); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func (n *Node23) graphAndPicture(filename string, debug bool) error {
	graphDir := "testdata/graph/"
	_ = os.MkdirAll(graphDir, os.ModePerm)
	filepath := graphDir + filename
	_ = os.Remove(filepath + ".dot")
	_ = os.Remove(filepath + ".png")
	n.graph(filepath, debug)
	dotExecutable, _ := exec.LookPath("dot")
	cmdDot := &exec.Cmd{
		Path: dotExecutable,
		Args: []string{dotExecutable, "-Tpng", filepath + ".dot", "-o", filepath + ".png"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := cmdDot.Run(); err != nil {
		return err
	}
	return nil
}

func makeInternalNode(children []*Node23) (*Node23) {
	ensure(len(children) > 0, "number of children is zero")
	internalKeys := make([]*Felt, 0, len(children)-1)
	for _, child := range children[:len(children)-1] {
		ensure(child.nextKey() != nil, "child next key is zero")
		nextKey := *child.nextKey()
		internalKeys = append(internalKeys, &nextKey)
	}
	n := &Node23{isLeaf: false, children: children, keys: internalKeys, values: make([]*Felt, 0)}
	return n
}

func makeLeafNode(keys, values []*Felt) (*Node23) {
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

func (n *Node23) walkPostOrder(w Walker) []interface{} {
	items := make([]interface{}, 0)
	if !n.isLeaf {
		for _, child := range n.children {
			//log.Tracef("walkPostOrder: n=%s child=%s\n", n, child)
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

func (n *Node23) upsert(kvItems []KeyValue) ([]*Node23) {
	log.Tracef("upsert: n=%p kvItems=%v\n", n, kvItems)
	if len(kvItems) == 0 {
		return []*Node23{n}
	}
	// TODO: ensure(kvItems == sort.Sort(kvItems), "kvItems is not sorted by key")
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

func (n *Node23) upsertLeaf(kvItems []KeyValue) ([]*Node23) {
	log.Tracef("upsertLeaf: n=%p kvItems=%v\n", n, kvItems)
	ensure(n.isLeaf, "node is not leaf")
	n.addOrReplace(kvItems)
	log.Tracef("upsertLeaf: keyCount=%d\n", n.keyCount())
	if n.keyCount() > 3 {
		nodes := make([]*Node23, 0)
		for n.keyCount() > 3 {
			newLeaf := makeLeafNode(n.keys[:3], n.values[:3])
			log.Tracef("upsertLeaf: newLeaf=%s\n", newLeaf)
			nodes = append(nodes, newLeaf)
			n.keys, n.values = n.keys[2:], n.values[2:]
			log.Tracef("upsertLeaf: updated n=%s\n", n)
		}
		// TODO: nodes = append(nodes, n) can we save one allocation?
		newLeaf := makeLeafNode(n.keys[:], n.values[:])
		log.Tracef("upsertLeaf: last newLeaf=%s\n", newLeaf)
		nodes = append(nodes, newLeaf)
		//nodes = append(nodes, n)
		return nodes
	} else {
		return []*Node23{n}
	}
}

func (n *Node23) upsertInternal(kvItems []KeyValue) ([]*Node23) {
	ensure(!n.isLeaf, "node is not internal")
	log.Tracef("upsertInternal: n=%s keyCount=%d\n", n, n.keyCount())

	itemSubsets := n.splitItems(kvItems)

	promoted_nodes := make([]*Node23, 0)
	for i, child := range n.children {
		log.Tracef("upsertInternal: i=%d child=%s itemSubsets[i]=%v\n", i, child, itemSubsets[i])
		child_promoted_nodes := child.upsert(itemSubsets[i])
		promoted_nodes = append(promoted_nodes, child_promoted_nodes...)
		log.Tracef("upsertInternal: i=%d promoted_nodes=%v \n", i, promoted_nodes)
	}
	log.Tracef("upsertInternal: n=%s\n", n)
	n.children = promoted_nodes
	log.Tracef("upsertInternal: n=%s\n", n)
	nodes := make([]*Node23, 0)
	if n.childrenCount() > 3 {
		for n.childrenCount() > 3 {
			nodes = append(nodes, makeInternalNode(n.children[:2]))
			n.children = n.children[2:]
		}
		//TODO: nodes = append(nodes, n) can we save one allocation?
		nodes = append(nodes, makeInternalNode(n.children[:]))
		//nodes = append(nodes, n)
		return nodes
	} else {
		return []*Node23{n}
	}
}

func (n *Node23) splitItems(kvItems []KeyValue) [][]KeyValue {
	ensure(!n.isLeaf, "splitItems: node is not internal")
	ensure(len(n.keys) > 0, "splitItems: internal node has no keys")
	log.Tracef("splitItems: keys=%v-%v kvItems=%v\n", pointer2pointee(n.keys), n.keys, kvItems)
	itemSubsets := make([][]KeyValue, 0)
	for i, key := range n.keys {
		splitIndex := sort.Search(len(kvItems), func(i int) bool { return kvItems[i].key > *key })
		log.Tracef("splitItems: key=%d-(%p) splitIndex=%d\n", *key, key, splitIndex)
		if splitIndex < len(kvItems) {
			itemSubsets = append(itemSubsets, kvItems[:splitIndex])
			kvItems = kvItems[splitIndex:]
		}
		if i == len(n.keys)-1 {
			itemSubsets = append(itemSubsets, kvItems)
		}
	}
	return itemSubsets
}

func (n *Node23) addOrReplace(kvItems []KeyValue) {
	ensure(n.isLeaf, "node is not leaf")
	ensure(len(n.keys) > 0 && len(n.values) > 0, "node keys/values are not empty")
	ensure(len(kvItems) > 0, "kvItems is not empty")
	log.Debugf("addOrReplace: keys=%v-%v values=%v-%v kvItems=%v\n", pointer2pointee(n.keys), n.keys, pointer2pointee(n.values), n.values, kvItems)
	ensure(n.keyRangeContains(kvItems), "upsert keys out of node range")
	
	nextKey, nextValue := n.nextKey(), n.nextValue()
	log.Debugf("addOrReplace: nextKey=%p values=%p\n", nextKey, nextValue)
	n.keys = n.keys[:len(n.keys)-1]
	n.values = n.values[:len(n.values)-1]
	log.Debugf("addOrReplace: keys=%v-%v values=%v-%v\n", pointer2pointee(n.keys), n.keys, pointer2pointee(n.values), n.values)
	for _, kvPair := range kvItems {
		key, value := kvPair.key, kvPair.value
		n.keys = append(n.keys, &key)
		n.values = append(n.values, &value)
	}
	log.Debugf("addOrReplace: keys=%v-%v values=%v-%v\n", pointer2pointee(n.keys), n.keys, pointer2pointee(n.values), n.values)
	sort.Slice(n.keys, func(i, j int) bool { return *n.keys[i] < *n.keys[j] })
	sort.Slice(n.values, func(i, j int) bool { return *n.values[i] < *n.values[j] })
	n.keys = append(n.keys, nextKey)
	n.values = append(n.values, nextValue)
	log.Debugf("addOrReplace: keys=%v-%v values=%v-%v\n", pointer2pointee(n.keys), n.keys, pointer2pointee(n.values), n.values)
	log.Debugf("addOrReplace: keys=%v-%v values=%v-%v\n", pointer2pointee(n.keys), n.keys, pointer2pointee(n.values), n.values)
}

func (n *Node23) keyRangeContains(kvItems []KeyValue) bool {
	return n.isSentinelOrLowerThanKey(0, kvItems[0].key) && n.isSentinelOrGreaterThanKey(len(n.keys)-1, kvItems[len(kvItems)-1].key)
}

func (n *Node23) isSentinelOrGreaterThanKey(keyIndex int, otherKey Felt) bool {
	return n.isSentinelKey(keyIndex) || n.isGreaterThanKey(keyIndex, otherKey)
}

func (n *Node23) isSentinelOrLowerThanKey(keyIndex int, otherKey Felt) bool {
	return n.isSentinelKey(keyIndex) || n.isLowerThanKey(keyIndex, otherKey)
}

func (n *Node23) isSentinelKey(keyIndex int) bool {
	return n.keys[keyIndex] == nil
}

func (n *Node23) isGreaterThanKey(i int, otherKey Felt) bool {
	return *n.keys[i] > otherKey;
}

func (n *Node23) isLowerThanKey(i int, otherKey Felt) bool {
	return *n.keys[i] < otherKey;
}

func (n *Node23) keyCount() int {
	return len(n.keys)
}

func (n *Node23) childrenCount() int {
	return len(n.children)
}

func pointer2pointee(pointers []*Felt) ([]Felt) {
	pointees := make([]Felt, 0)
	for _, ptr := range pointers {
		if ptr != nil {
			pointees = append(pointees, *ptr)
		} else {
			break
		}
	}
	return pointees
}
