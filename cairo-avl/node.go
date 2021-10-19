package cairo_avl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
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
}

func NewNode(k, v *Felt, T_L, T_R, T_N *Node) *Node {
	n := Node{key: k, value: v, treeLeft: T_L, treeRight: T_R, treeNested: T_N}
	n.height = NewFelt(1)
	n.height.Add(MaxBigInt(height(n.treeLeft), height(n.treeRight)), NewFelt(1))
	UpdatePath(&n, "M")
	return &n
}

func makeNode(k, v, h *Felt, T_L, T_R, T_N *Node) *Node {
	n := Node{key: k, value: v, height: h, treeLeft: T_L, treeRight: T_R, treeNested: T_N}
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
		fmt.Printf("state line: %s [%x]\n", line, line)

		tokens := strings.Split(line, ",")
		if len(tokens) != 5 {
			log.Fatal("state invalid line: ", line)
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
		fmt.Printf("%v\n", ck)
	}
	sort.SliceStable(compositeKeys, func(i, j int) bool {
		return compare(compositeKeys[i], compositeKeys[j]) < 0
	})
	fmt.Printf("Ordered composite keys:\n")
	for _, ck := range compositeKeys {
		fmt.Printf("%v\n", ck)
	}
	fmt.Printf("Building tree:\n")
	treeByLevel := make(map[int]map[int64]*Node)
	for _, item := range compositeKeys {
		fmt.Printf("item: %v\n", item)
		numKeys := len(item) - 1
		ckItem := item[:numKeys]
		fmt.Printf("numKeys: %d\n", numKeys)
		for nesting := numKeys-1; nesting >= 0; nesting-- {
			fmt.Printf("nesting=%d\n", nesting)
			key := ckItem[nesting]
			fmt.Printf("key=%d\n", key)

			var v *Felt
			if nesting == numKeys-1 {
				fmt.Printf("value=%d\n", item[numKeys])
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
				fmt.Printf("containerKey=%d container=%p %+v\n", containerKey, container, container)

				if container == nil {
					container = Insert(nil, NewFelt(containerKey), nil, nil) // TODO: try treeByLevel[nesting][key]
					treeByLevel[nesting-1][containerKey] = container
					if treeByLevel[nesting][key] != nil {
						container.treeNested = treeByLevel[nesting][key]
					}
				}
				tree := container.treeNested
				fmt.Printf("tree: p=%p %+v\n", tree, tree)
				k := NewFelt(key)
				if n := tree.Search(k); n != nil {
					continue
				}
				newTree := Insert(tree, k, v, /*N=*/nil)
				fmt.Printf("newTree: p=%p %+v\n", newTree, newTree)
				newNode := newTree.Search(k) // TODO: Insert must return inserted node
				if treeByLevel[nesting][key] != nil {
					newNode.treeNested = treeByLevel[nesting][key].treeNested
				}
				treeByLevel[nesting][key] = newNode
				container.treeNested = newTree
				fmt.Printf("newNode=%p %+v\n", newNode, newNode)
				UpdatePath(container, container.path)
				fmt.Printf("newNode.path=%s newNode.nesting=%d\n", newNode.path, newNode.nesting())
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
	fmt.Printf("treeByLevel[0][0]: p=%p %+v\n", t, t)
	return t, nil
}

func MappedStateFromCsv(state *bufio.Scanner) (t *Node, err error) {
	nodeByPointer := make(map[int64]*Node)
	for state.Scan() {
		line := state.Text()
		fmt.Println("mapped state line: ", line)
		tokens := strings.Split(line, ",")
		if len(tokens) != 11 {
			log.Fatal("mapped state invalid line: ", line)
		}
		p, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			return nil, err
		}
		k, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return nil, err
		}
		value := tokens[4]
		var v *Felt
		if value == "" {
			v = nil
		} else {
			int_value, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			v = NewFelt(int_value)
		}
		if err != nil {
			return nil, err
		}
		var T_L *Node
		leftTreeType := tokens[5]
		leftItem := tokens[6]
		if leftTreeType == "HASH" && strings.HasPrefix(leftItem, "hash") {
			T_L = nil
		} else {
			left, err := strconv.ParseInt(leftItem, 10, 64)
			if err != nil {
				return nil, err
			}
			T_L = nodeByPointer[left]
		}
		var T_R *Node
		rightTreeType := tokens[7]
		rightItem := tokens[8]
		if rightTreeType == "HASH" && strings.HasPrefix(rightItem, "hash") {
			T_R = nil
		} else {
			right, err := strconv.ParseInt(rightItem, 10, 64)
			if err != nil {
				return nil, err
			}
			T_R = nodeByPointer[right]
		}
		var T_N *Node
		nestedTreeType := tokens[9]
		nestedItem := tokens[10]
		if nestedTreeType == "HASH" && strings.HasPrefix(nestedItem, "hash") {
			T_N = nil
		} else {
			nested, err := strconv.ParseInt(nestedItem, 10, 64)
			if err != nil {
				return nil, err
			}
			T_N = nodeByPointer[nested]
		}
		fmt.Println("mapped state: p: ", p, " k: ", k, " v: ", v, " left: ", leftItem, " rightItem: ", rightItem, " nestedItem: ", nestedItem)
		nodeByPointer[p] = NewNode(NewFelt(k), v, T_L, T_R, T_N)
		fmt.Println("mapped state: p: ", p, " k: ", k, " v: ", v, " T_L: ", T_L, " T_R: ", T_R, " T_N: ", T_N, " nesting: ", nodeByPointer[p].nesting())
		fmt.Println("mapped state: nodeByPointer[p]: ", nodeByPointer[p])
		t = nodeByPointer[p]
	}
	if err := state.Err(); err != nil {
		return nil, err
	}
	return t, nil
}

func StateFromBinary(statesReader *bufio.Reader) (t *Node, err error) {
	buffer := make([]byte, BufferSize)
	for {
		bytes_read, err := statesReader.Read(buffer)
		fmt.Println("BINARY state bytes read: ", bytes_read, " err: ", err)
		if err == io.EOF {
			break
		}
		key_bytes_count := 4 * (bytes_read / 4)
		fmt.Println("BINARY state key_bytes_count: ", key_bytes_count)
		for i := 0; i < key_bytes_count; i += 4 {
			key := binary.BigEndian.Uint32(buffer[i:i+4])
			fmt.Println("BINARY state key: ", key)
			var nestedTree *Node
			if i % 10 == 0 {
				fmt.Printf("Inserting nested key: %d\n", i)
				nestedTree = Insert(/*T=*/nil, NewFelt(int64(i)), NewFelt(0), /*N=*/nil)
				fmt.Printf("Inserted nested tree: %+v\n", nestedTree)
			}
			t = Insert(t, NewFelt(int64(key)), NewFelt(0), nestedTree)
		}
	}
	t = Insert(/*T=*/nil, NewFelt(int64(0)), /*v=*/nil, Insert(/*T=*/nil, NewFelt(int64(0)), /*v=*/nil, t))
	UpdatePath(t, "M")
	fmt.Printf("t p=%p %+v\n", t, t)
	return t, nil
}

func (n *Node) Graph(filename string) {
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
		fmt.Printf("p=%p %+v k=%d nesting=%d\n", n, n, n.key, n.nesting())
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
		//str := fmt.Sprintf("k=%d p=%p", n.key, n)
		s := fmt.Sprintln(n.path, " [label=\"", left, "|{<C>", /*str*/n.key, "|", down, "}|", right, "\" style=filled fillcolor=\"", colors[n.nesting()], "\"];")
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

func (n *Node) GraphAndPicture(filename string) error {
	graphDir := "testdata/graph/"
	_ = os.MkdirAll(graphDir, os.ModePerm)
	filepath := graphDir + filename
	_ = os.Remove(filepath + ".dot")
	_ = os.Remove(filepath + ".png")
	n.Graph(filepath)
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

func exposeNode(n *Node) (k, v *Felt, T_L, T_R, T_N *Node) {
	if n != nil && n.key != nil {
		if !n.exposed {
			n.exposed = true
		}
		return n.key, n.value, n.treeLeft, n.treeRight, n.treeNested
	}
	return nil, nil, nil, nil, nil
}

func height(n *Node) *Felt {
	if n != nil && n.height != nil {
		return n.height
	}
	return NewFelt(0)
}

func HeightAsInt(n *Node) int {
	return int(height(n).Uint64())
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

func Update(n *Node) *Node {
	n.height.Add(MaxBigInt(height(n.treeLeft), height(n.treeRight)), NewFelt(1))
	return n
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
	return height(L).Cmp(NewFelt(0).Add(height(R), NewFelt(1))) > 0
}
