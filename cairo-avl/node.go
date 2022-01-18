package cairo_avl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Node struct {
	path		string
	key		*Felt
	value		*Felt
	height		*Felt
	treeLeft	*Node
	treeRight	*Node
	treeNested	*Node
	exposed		bool
	heightTaken	bool
}

func NewNode(k, v *Felt, T_L, T_R, T_N *Node) *Node {
	n := Node{key: k, value: v, treeLeft: T_L, treeRight: T_R, treeNested: T_N}
	n.height = NewFelt(1)
	c := &Counters{}
	n.height.Add(MaxBigInt(height(n.treeLeft, c), height(n.treeRight, c)), NewFelt(1))
	UpdatePath(&n, "M")
	return &n
}

func (n *Node) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("path=%s k=%d v=%d h=%d L=%p R=%p, N=%p exposed=%t heightTaken=%t",
		n.path, n.key, n.value, n.height, n.treeLeft, n.treeRight, n.treeNested, n.exposed, n.heightTaken,
	)
}

func makeNode(k, v, h *Felt, T_L, T_R, T_N *Node) *Node {
	n := Node{key: k, value: v, height: h, treeLeft: T_L, treeRight: T_R, treeNested: T_N, exposed: true}
	UpdatePath(&n, "M")
	return &n
}

func (n *Node) nesting() int {
	return strings.Count(n.path, "N")
}

func (T *Node) Search(k *Felt) *Node {
	if T == nil || T.key == nil || k == nil {
		return nil
	}
	if k.Cmp(T.key) == 0 {
		return T
	} else if k.Cmp(T.key) < 0 {
		return T.treeLeft.Search(k)
	} else {
		return T.treeRight.Search(k)
	}
}

type Walker func(*Node) interface{}

func (n *Node) WalkInOrder(w Walker) []interface{} {
	if n == nil || n.key == nil {
		return make([]interface{}, 0)
	}
	var (
		left_items, nested_items, right_items []interface{}
	)
	if n.treeLeft != nil {
		left_items = n.treeLeft.WalkInOrder(w)
	} else {
		left_items = make([]interface{}, 0)
	}
	if n.treeNested != nil {
		nested_items = n.treeNested.WalkInOrder(w)
	} else {
		nested_items = make([]interface{}, 0)
	}
	if n.treeRight != nil {
		right_items = n.treeRight.WalkInOrder(w)
	} else {
		right_items = make([]interface{}, 0)
	}
	items := make([]interface{}, 0)
	items = append(items, w(n))
	items = append(items, left_items...)
	items = append(items, nested_items...)
	items = append(items, right_items...)
	return items
}

func (n *Node) WalkNodesInOrder() []*Node {
	node_items := n.WalkInOrder(func(n *Node) interface{} { return n })
	nodes := make([]*Node, len(node_items))
	for i := range node_items {
		nodes[i] = node_items[i].(*Node)
	}
	return nodes
}

func (n *Node) WalkKeysInOrder() []uint64 {
	key_items := n.WalkInOrder(func(n *Node) interface{} { return n.key })
	keys := make([]uint64, len(key_items))
	for i := range key_items {
		keys[i] = key_items[i].(*Felt).Uint64()
	}
	return keys
}

func (n *Node) WalkPathsInOrder() []string {
	path_items := n.WalkInOrder(func(n *Node) interface{} { return n.path })
	paths := make([]string, len(path_items))
	for i := range path_items {
		paths[i] = path_items[i].(string)
	}
	return paths
}

func compare(s1, s2 []int64) int {
	k := 0
	for k < len(s1) || k < len(s2) {
		if k >= len(s1) {
			if k >= len(s2) {
				return 0
			} else {
				return -1
			}
		}
		if k >= len(s2) {
			return 1
		}
		if s1[k] < s2[k] {
			return -1
		} else if s1[k] > s2[k] {
			return 1
		}
		k++
	}
	return 0
}

func StateFromCsv(state *bufio.Scanner) (t *Node, err error) {
	compositeKeys := make([][]int64, 0)
	for state.Scan() {
		line := state.Text()
		log.Tracef("CSV state line: %s [%x]\n", line, line)

		tokens := strings.Split(line, ",")
		if len(tokens) != 5 {
			log.Fatal("CSV state invalid line: ", line)
		}
		compositeKeyItems := make([]int64, 0)
		for _, token := range tokens {
			if token == "" {
				continue
			}
			subTokens := strings.Split(token, ";");
			if len(subTokens) == 1 {
				token = subTokens[0]
				var i int64
				if i, err = strconv.ParseInt(token, 10, 64); err != nil {
					return nil, err
				}
				compositeKeyItems = append(compositeKeyItems, i)
			} else {
				for _, st := range subTokens {
					var i int64
					if i, err = strconv.ParseInt(st, 10, 64); err != nil {
						return nil, err
					}
					compositeKeyItems = append(compositeKeyItems, i)
				}
			}
		}
		compositeKeys = append(compositeKeys, compositeKeyItems)
	}
	if err := state.Err(); err != nil {
		return nil, err
	}
	for _, ck := range compositeKeys {
		log.Tracef("%v\n", ck)
	}
	sort.SliceStable(compositeKeys, func(i, j int) bool {
		return compare(compositeKeys[i], compositeKeys[j]) < 0
	})
	log.Tracef("Ordered composite keys:\n")
	for _, ck := range compositeKeys {
		log.Tracef("%v\n", ck)
	}
	log.Tracef("Building tree:\n")
	treeByLevel := make(map[int]map[int64]*Node)
	for _, item := range compositeKeys {
		log.Tracef("item: %v\n", item)
		numKeys := len(item) - 1
		ckItem := item[:numKeys]
		log.Tracef("numKeys: %d\n", numKeys)
		for nesting := numKeys-1; nesting >= 0; nesting-- {
			log.Tracef("nesting=%d\n", nesting)
			key := ckItem[nesting]
			log.Tracef("key=%d\n", key)

			var v *Felt
			if nesting == numKeys-1 {
				log.Tracef("value=%d\n", item[numKeys])
				v = NewFelt(item[numKeys])
			}

			if treeByLevel[nesting] == nil {
				treeByLevel[nesting] = make(map[int64]*Node)
			}

			if nesting > 0 {
				if treeByLevel[nesting-1] == nil {
					treeByLevel[nesting-1] = make(map[int64]*Node)
				}
				containerKey := ckItem[nesting-1]
				container := treeByLevel[nesting-1][containerKey]
				log.Tracef("containerKey=%d container=%p %+v\n", containerKey, container, container)

				if container == nil {
					container = Insert(nil, NewFelt(containerKey), nil, nil) // TODO: try treeByLevel[nesting][key]
					treeByLevel[nesting-1][containerKey] = container
					if treeByLevel[nesting][key] != nil {
						container.treeNested = treeByLevel[nesting][key]
					}
				}
				tree := container.treeNested
				log.Tracef("tree: p=%p %+v\n", tree, tree)
				k := NewFelt(key)
				if n := tree.Search(k); n != nil {
					continue
				}
				newTree := Insert(tree, k, v, /*N=*/nil)
				log.Tracef("newTree: p=%p %+v\n", newTree, newTree)
				newNode := newTree.Search(k) // TODO: Insert must return inserted node
				if treeByLevel[nesting][key] != nil {
					newNode.treeNested = treeByLevel[nesting][key].treeNested
				}
				treeByLevel[nesting][key] = newNode
				container.treeNested = newTree
				UpdatePath(container, container.path)
				log.Debugf("newNode=%p %+v nesting=%d\n", newNode, newNode, newNode.nesting())
			} else {
				root := treeByLevel[nesting][key]
				if root == nil {
					root = Insert(/*T=*/nil, NewFelt(key), /*v=*/nil, /*N=*/nil)
					treeByLevel[nesting][key] = root
				}
				UpdatePath(root, root.path)
			}
		}
	}
	t = treeByLevel[0][0]
	t.resetFlags()
	log.Tracef("t p=%p %+v\n", t, t)
	return t, nil
}

func StateFromBinary(statesReader *bufio.Reader, keySize int, nested bool) (t *Node, err error) {
	buffer := make([]byte, BufferSize)
	for {
		bytes_read, err := statesReader.Read(buffer)
		log.Debugf("BINARY state bytes read: %d err: %v\n", bytes_read, err)
		if err == io.EOF {
			break
		}
		key_bytes_count := keySize * (bytes_read / keySize)
		duplicated_keys := 0
		log.Debugf("BINARY state key_bytes_count: %d\n", key_bytes_count)
		for i := 0; i < key_bytes_count; i += keySize {
			key := readKey(buffer, i, keySize)
			log.Debugf("BINARY state key: %d\n", key)
			if t.Search(key) != nil {
				duplicated_keys++
				continue
			}
			var nestedTree *Node
			if nested && i % 10 == 0 {
				log.Debugf("Inserting nested key: %d\n", i)
				nestedTree = Insert(/*T=*/nil, NewFelt(int64(i)), NewFelt(0), /*N=*/nil)
				log.Debugf("Inserted nested tree: %+v\n", nestedTree)
			}
			t = Insert(t, key, NewFelt(0), nestedTree)
		}
		log.Debugf("BINARY state duplicated_keys: %d\n", duplicated_keys)
	}
	if nested {
		t = Insert(/*T=*/nil, NewFelt(int64(0)), /*v=*/nil, Insert(/*T=*/nil, NewFelt(int64(0)), /*v=*/nil, t))
	}
	t.resetFlags()
	UpdatePath(t, "M")
	log.Tracef("t p=%p %+v\n", t, t)
	return t, nil
}

func (n *Node) Graph(filename string, debug bool) {
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
	if _, err := f.WriteString("strict digraph {\nnode [shape=record];\n"); err != nil {
		log.Fatal(err)
	}
	for _, n := range n.WalkNodesInOrder() {
		log.Tracef("p=%p %+v k=%d nesting=%d\n", n, n, n.key, n.nesting())
		left, right := "", ""
		if n.treeLeft != nil {
			left = "<L>L"
		}
		if n.treeRight != nil {
			right = "<R>R"
		}
		var down string
		if n.treeNested != nil {
			down = "<N>N"
		} else if n.value == nil {
			// HASH node type
			down = "<N>H"
		} else {
			down = n.value.String()
		}
		var nodeId string
		if debug {
			nodeId = fmt.Sprintf("k=%d [%t]", n.key, n.exposed)
		} else {
			nodeId = n.key.String()
		}
		var fc string
		if n.heightTaken {
			if n.exposed {
				fc = "red"
			} else {
				fc = "blue"
			}
		} else {
			if (n.exposed) {
				fc = "red"
			} else {
				fc = "black"
			}
		}
		s := fmt.Sprintln(n.path,
			" [label=\"", left, "|{<C>", nodeId, "|", down, "}|", right, "\" style=filled fontcolor=", fc ," fillcolor=\"", colors[n.nesting()], "\"];")
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}
	for _, n := range n.WalkNodesInOrder() {
		if n.treeLeft != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.path, ":L -> ", n.treeLeft.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if n.treeRight != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.path, ":R -> ", n.treeRight.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if n.treeNested != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.path, ":N -> ", n.treeNested.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func (n *Node) GraphAndPicture(filename string, debug bool) error {
	graphDir := "testdata/graph/"
	_ = os.MkdirAll(graphDir, os.ModePerm)
	filepath := graphDir + filename
	_ = os.Remove(filepath + ".dot")
	_ = os.Remove(filepath + ".png")
	n.Graph(filepath, debug)
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

func exposeNode(n *Node, c *Counters) (k, v *Felt, T_L, T_R, T_N *Node) {
	if n != nil && n.key != nil {
		if !n.exposed {
			if n.heightTaken && c.HeightCount > 0 {
				c.HeightCount--
			} else {
				n.heightTaken = true
			}
			c.ExposedCount++
			n.exposed = true
		}
		return n.key, n.value, n.treeLeft, n.treeRight, n.treeNested
	}
	return nil, nil, nil, nil, nil
}

func height(n *Node, c *Counters) *Felt {
	if n != nil && n.height != nil {
		if !n.exposed && !n.heightTaken {
			c.HeightCount++
		}
		n.heightTaken = true
		return n.height
	}
	return NewFelt(0)
}

func HeightAsInt(n *Node) int {
	c := &Counters{}
	return int(height(n, c).Uint64())
}

func UpdatePath(n *Node, path string) {
	n.path = path
	if n.treeLeft != nil {
		UpdatePath(n.treeLeft, n.path[:len(n.path)-1] + "LM")
	}
	if n.treeNested != nil {
		UpdatePath(n.treeNested, n.path[:len(n.path)-1] + "NM")
	}
	if n.treeRight != nil {
		UpdatePath(n.treeRight, n.path[:len(n.path)-1] + "RM")
	}
}

func (n *Node) IsBST() bool {
	if n == nil {
		return true
	}
	if n.treeLeft != nil {
		if n.treeLeft.key.Cmp(n.key) > 0 {
			return false
		}
		if !n.treeLeft.IsBST() {
			return false
		}
	}
	if n.treeRight != nil {
		if n.treeRight.key.Cmp(n.key) < 0 {
			return false
		}
		if !n.treeRight.IsBST() {
			return false
		}
	}
	return true
}

func (n *Node) IsBalanced() bool {
	return n == nil || !(IsLeftHeavy(n.treeLeft, n.treeRight) || IsLeftHeavy(n.treeRight, n.treeLeft))
}

func IsLeftHeavy(L, R *Node) bool {
	c := &Counters{}
	return height(L, c).Cmp(NewFelt(0).Add(height(R, c), NewFelt(1))) > 0
}

func (n *Node) resetFlags() {
	node_items := n.WalkInOrder(func(n *Node) interface{} { return n })
	for i := range node_items {
		node := node_items[i].(*Node)
		node.exposed = false
		node.heightTaken = false
	}
}

func (n *Node) Size() int {
	if n == nil {
		return 0
	}
	return len(n.WalkKeysInOrder())
}

func (n *Node) CountNewHashes() (hashCount uint) {
	node_items := n.WalkInOrder(func(n *Node) interface{} { return n })
	for i := range node_items {
		if node_items[i].(*Node).exposed {
			hashCount++
		}
	}
	return hashCount
}
