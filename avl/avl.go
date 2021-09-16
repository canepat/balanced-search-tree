/*
* Adelson-Velsky and Landis (AVL) self-balancing binary search tree implementation. References:
* [1] https://zhjwpku.com/assets/pdf/AED2-10-avl-paper.pdf
* [2] https://www.cs.cmu.edu/~blelloch/papers/BFS16.pdf
* [3] https://github.com/cmuparlay/PAM
 */

package avl

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
)

func Max(a, b *big.Int) *big.Int {
	if a.Cmp(b) < 0 {
		return b
	}
	return a
}

type Node struct {
	Path	string		`json:"-"`
	Key		*big.Int	`json:"key"`
	Left	*Node		`json:"left,omitempty"`
	Right	*Node		`json:"right,omitempty"`
	Height	*big.Int	`json:"-"`
	Value	*big.Int	`json:"value,omitempty"`
	Exposed	bool		`json:"-"`
}

func NewNode(k, v *big.Int, T_L, T_R *Node) *Node {
	n := Node{Key: k, Value: v, Left: T_L, Right: T_R}
	n.Height = big.NewInt(1)
	UpdatePath(&n, "M")
	UpdateHeight(&n)
	return &n
}

type Walker func(*Node) interface{}

func (n *Node) WalkInOrder(w Walker) []interface{} {
	if n == nil || n.Key == nil {
		return make([]interface{}, 0)
	}
	var (
		left_items, right_items []interface{}
	)
	if n.Left != nil {
		left_items = n.Left.WalkInOrder(w)
	} else {
		left_items = make([]interface{}, 0)
	}
	if n.Right != nil {
		right_items = n.Right.WalkInOrder(w)
	} else {
		right_items = make([]interface{}, 0)
	}
	items := make([]interface{}, 0)
	items = append(items, w(n))
	items = append(items, left_items...)
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
	key_items := n.WalkInOrder(func(n *Node) interface{} { return n.Key })
	keys := make([]uint64, len(key_items))
	for i := range key_items {
		keys[i] = key_items[i].(*big.Int).Uint64()
	}
	return keys
}

func (n *Node) WalkPathsInOrder() []string {
	path_items := n.WalkInOrder(func(n *Node) interface{} { return n.Path })
	paths := make([]string, len(path_items))
	for i := range path_items {
		paths[i] = path_items[i].(string)
	}
	return paths
}

func (n *Node) UnmarshalJSON(data []byte) error {
	type NodeAlias Node
	alias := &NodeAlias{}
	err := json.Unmarshal(data, alias)
	if err != nil {
		return err
	}
	*n = Node(*alias)
	if n.Key == nil {
		return &json.InvalidUnmarshalError{}
	}
	UpdateHeight(n)
	return nil
}

func Graph(n *Node, filename string) {
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
		left, right := "", ""
		if n.Left != nil {
			left = "<L>L"
		}
		if n.Right != nil {
			right = "<R>R"
		}
		s := fmt.Sprintln(n.Key, " [label=\"", left, "|{<C>", n.Key, "|", n.Value, "}|", right, "\" style=filled fillcolor=\"", colors[2], "\"];")
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}
	for _, n := range n.WalkNodesInOrder() {
		if n.Left != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.Key, ":L -> ", n.Left.Key, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if n.Right != nil {
			if _, err := f.WriteString(fmt.Sprintln(n.Key, ":R -> ", n.Right.Key, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func GraphAndPicture(n *Node, filename string) error {
	graphDir := "testdata/graph/"
	_ = os.MkdirAll(graphDir, os.ModePerm)
	filepath := graphDir + filename
	Graph(n, filepath)
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

func Expose(n *Node) (*big.Int, *big.Int, *big.Int, *Node, *Node) {
	if n != nil && n.Key != nil {
		n.Exposed = true
		return n.Key, n.Height, n.Value, n.Left, n.Right
	}
	return nil, nil, nil, nil, nil
}

func Height(n *Node) *big.Int {
	if n != nil && n.Height != nil {
		return n.Height
	}
	return big.NewInt(0)
}

func HeightAsInt(n *Node) int {
	return int(Height(n).Uint64())
}

func UpdatePath(n *Node, path string) {
	n.Path = path
	if n.Left != nil {
		UpdatePath(n.Left, n.Path[:len(n.Path)-1] + "LM")
	}
	if n.Right != nil {
		UpdatePath(n.Right, n.Path[:len(n.Path)-1] + "RM")
	}
}

func UpdateHeight(n *Node) *big.Int {
	if n.Height == nil {
		n.Height = big.NewInt(0)
	}
	if n.Left != nil {
		if n.Right != nil {
			n.Height.Add(Max(UpdateHeight(n.Left), UpdateHeight(n.Right)), big.NewInt(1))
		} else {
			n.Height.Add(UpdateHeight(n.Left), big.NewInt(1))
		}
	} else {
		if n.Right != nil {
			n.Height.Add(UpdateHeight(n.Right), big.NewInt(1))
		} else {
			n.Height = big.NewInt(1)
		}
	}
	return n.Height
}

func Update(n *Node) *Node {
	UpdateHeight(n)
	return n
}

func RotateLeft(x *Node) *Node {
	y := x.Right
	z := y.Left
	y.Left = x
	x.Right = z
	Update(x)
	Update(y)
	return y
}

func RotateRight(x *Node) *Node {
	y := x.Left
	z := y.Right
	y.Right = x
	x.Left = z
	Update(x)
	Update(y)
	return y
}

func DoubleRotateLeft(x *Node) *Node {
	r := x.Right
	root := r.Left
	r.Left = root.Right
	Update(r)
	root.Right = r
	x.Right = root
	return RotateLeft(x)
}

func DoubleRotateRight(x *Node) *Node {
	l := x.Left
	root := l.Right
	l.Right = root.Left
	Update(l)
	root.Left = l
	x.Left = root
	return RotateRight(x)
}

func (n *Node) IsBST() bool {
	if n == nil {
		return true
	}
	if n.Left != nil {
		if n.Left.Key.Cmp(n.Key) > 0 {
			return false
		}
		if !n.Left.IsBST() {
			return false
		}
	}
	if n.Right != nil {
		if n.Right.Key.Cmp(n.Key) < 0 {
			return false
		}
		if !n.Right.IsBST() {
			return false
		}
	}
	return true
}

func (n *Node) IsBalanced() bool {
	return n == nil || !(IsLeftHeavy(n.Left, n.Right) || IsLeftHeavy(n.Right, n.Left))
}

func IsLeftHeavy(left, right *Node) bool {
	return Height(left).Cmp(big.NewInt(0).Add(Height(right), big.NewInt(1))) > 0
}

func IsSingleRotation(n *Node, isLeft bool) bool {
	if n == nil {
		return false
	}
	if isLeft {
		return Height(n.Left).Cmp(Height(n.Right)) > 0
	} else {
		return Height(n.Left).Cmp(Height(n.Right)) <= 0
	}
}

func JoinBalanced(left *Node, k, v *big.Int, right *Node) *Node {
	if k == nil {
		if left == nil {
			return right
		}
		if right == nil {
			return left
		}
		return nil
	} else {
		return NewNode(k, v, left, right)
	}
}

func JoinRight(T_L *Node, k, v *big.Int, T_R *Node) *Node {
	if !IsLeftHeavy(T_L, T_R) {
		return JoinBalanced(T_L, k, v, T_R)
	}
	var n *Node
	if T_L == nil {
		n = nil
	} else {
		n = NewNode(T_L.Key, T_L.Value, T_L.Left, T_L.Right)
	}
	n.Right = JoinRight(n.Right, k, v, T_R)

	if IsLeftHeavy(n.Right, n.Left) {
		if IsSingleRotation(n.Right, false) {
			n = RotateLeft(n)
		} else {
			n = DoubleRotateLeft(n)
		}
	} else {
		Update(n)
	}
	return n
}

func JoinLeft(T_L *Node, k, v *big.Int, T_R *Node) *Node {
	if !IsLeftHeavy(T_R, T_L) {
		return JoinBalanced(T_L, k, v, T_R)
	}
	var n *Node
	if T_R == nil {
		n = nil
	} else {
		n = NewNode(T_R.Key, T_R.Value, T_R.Left, T_R.Right)
	}
	n.Left = JoinLeft(T_L, k, v, n.Left)

	if IsLeftHeavy(n.Left, n.Right) {
		if IsSingleRotation(n.Left, true) {
			n = RotateRight(n)
		} else {
			n = DoubleRotateRight(n)
		}
	} else {
		Update(n)
	}
	return n
}

func Join(T_L *Node, k, v *big.Int, T_R *Node) *Node {
	if IsLeftHeavy(T_L, T_R) {
		return JoinRight(T_L, k, v, T_R)
	}
	if IsLeftHeavy(T_R, T_L) {
		return JoinLeft(T_L, k, v, T_R)
	}
	return JoinBalanced(T_L, k, v, T_R)
}

// Variable are named after SPLIT operation in paper [2]
func Split(T *Node, k *big.Int) (*Node, bool, *Node) {
	if T == nil || T.Key == nil {
		return nil, false, nil
	}
	if k == nil {
		return T, false, nil
	}
	m_k, m_v, _, L, R := Expose(T)
	if kmCmp := k.Cmp(m_k); kmCmp == 0 {
		return L, true, R
	} else if kmCmp < 0 {
		L_l, b, L_r := Split(L, k)
		return L_l, b, Join(L_r, m_k, m_v, R)
	} else {
		R_l, b, R_r := Split(R, k)
		return Join(L, m_k, m_v, R_l), b, R_r
	}
}

// Variable are named after SPLIT operation in paper [2]
func SplitLast(T *Node) (*Node, *big.Int, *big.Int, *big.Int) {
	k, v, h, L, R := Expose(T)
	if R == nil {
		return L, k, v, h
	} else {
		T_dash, k_dash, v_dash, h_dash := SplitLast(R)
		return Join(L, k, v, T_dash), k_dash, v_dash, h_dash
	}
}

func Join2(T_L, T_R *Node) *Node {
	if T_L == nil {
		return T_R
	} else {
		T_L_dash, k, v, _ := SplitLast(T_L)
		return Join(T_L_dash, k, v, T_R)
	}
}

func Search(T *Node, k *big.Int) *Node {
	if T == nil || T.Key == nil || k == nil {
		return nil
	}
	if k.Cmp(T.Key) == 0 {
		return T
	} else if k.Cmp(T.Key) < 0 {
		return Search(T.Left, k)
	} else {
		return Search(T.Right, k)
	}
}

func Insert(T *Node, k, v *big.Int) *Node {
	T_L, _, T_R := Split(T, k)
	return Join(T_L, k, v, T_R)
}

func Delete(T *Node, k *big.Int) *Node {
	T_L, _, T_R := Split(T, k)
	return Join2(T_L, T_R)
}

func Union(T1, T2 *Node) *Node {
	if T1 == nil {
		return T2
	} else if T2 == nil {
		return T1
	} else {
		k2, v2, _, L2, R2 := Expose(T2)
		L1, _, R1 := Split(T1, k2)
		Tl := Union(L1, L2) // parallelizable
		Tr := Union(R1, R2) // parallelizable
		return Join(Tl, k2, v2, Tr)
	}
}

func Intersect(T1, T2 *Node) *Node {
	if T1 == nil {
		return nil
	} else if T2 == nil {
		return nil
	} else {
		k2, v2, _, L2, R2 := Expose(T2)
		L1, b, R1 := Split(T1, k2)
		Tl := Intersect(L1, L2) // parallelizable
		Tr := Intersect(R1, R2) // parallelizable
		if b {
			return Join(Tl, k2, v2, Tr)
		} else {
			return Join2(Tl, Tr)
		}
	}
}

func Difference(T1, T2 *Node) *Node {
	if T1 == nil {
		return nil
	} else if T2 == nil {
		return T1
	} else {
		k2, _, _, L2, R2 := Expose(T2)
		L1, _, R1 := Split(T1, k2)
		Tl := Difference(L1, L2) // parallelizable
		Tr := Difference(R1, R2) // parallelizable
		return Join2(Tl, Tr)
	}
}
