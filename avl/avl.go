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
	Key		*big.Int	`json:"key"`
	Left	*Node		`json:"left,omitempty"`
	Right	*Node		`json:"right,omitempty"`
	Parent	*Node		`json:"-"`
	Height	*big.Int	`json:"-"`
	Value	*big.Int	`json:"value,omitempty"`
}

func NewNode(Key *big.Int, left, right *Node) *Node {
	n := Node{Key: Key, Left: left, Right: right}
	if n.Left != nil {
		n.Left.Parent = &n
	}
	if n.Right != nil {
		n.Right.Parent = &n
	}
	n.Height = big.NewInt(0)
	UpdateHeight(&n)
	return &n
}

func (n *Node) PathDepth() uint64 {
	depth := uint64(0)
	ancestor := n.Parent
	for ancestor != nil {
		depth += 1
		ancestor = ancestor.Parent
	}
	return depth
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
		if n.Parent != nil {
			var direction string
			if n.Parent.Left == n {
				direction = "L"
			} else {
				direction = "R"
			}
			if _, err := f.WriteString(fmt.Sprintln(n.Parent.Key, ":", direction, " -> ", n.Key, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func GraphAndPicture(n *Node, filename string) error {
	Graph(n, filename)
	dotExecutable, _ := exec.LookPath("dot")
	cmdDot := &exec.Cmd{
		Path: dotExecutable,
		Args: []string{dotExecutable, "-Tpng", filename + ".dot", "-o", filename + ".png"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := cmdDot.Run(); err != nil {
		return err
	}
	return nil
}

func Expose(n *Node) (*Node, *big.Int, *Node) {
	if n != nil && n.Key != nil {
		return n.Left, n.Key, n.Right
	}
	return nil, nil, nil
}

func Height(n *Node) *big.Int {
	if n != nil && n.Height != nil {
		return n.Height
	}
	return big.NewInt(0)
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
	} else if n.Right != nil {
		if n.Right != nil {
			n.Height.Add(UpdateHeight(n.Right), big.NewInt(1))
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

func (n *Node) HasBinarySearchTreeProperty() bool {
	if n == nil {
		return false
	}
	if n.Left != nil {
		if n.Left.Key.Cmp(n.Key) > 0 {
			return false
		}
		if !n.Left.HasBinarySearchTreeProperty() {
			return false
		}
	}
	if n.Right != nil {
		if n.Right.Key.Cmp(n.Key) < 0 {
			return false
		}
		if !n.Right.HasBinarySearchTreeProperty() {
			return false
		}
	}
	return true
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

func JoinBalanced(left *Node, k *big.Int, right *Node) *Node {
	if k == nil {
		if left == nil {
			return right
		}
		if right == nil {
			return left
		}
		return nil
	} else {
		return NewNode(k, left, right)
	}
}

func JoinRight(n1 *Node, k *big.Int, n2 *Node) *Node {
	if !IsLeftHeavy(n1, n2) {
		return JoinBalanced(n1, k, n2)
	}
	n := n1 // TODO: check if clone is needed
	n.Right = JoinRight(n.Right, k, n2)
	n.Right.Parent = n
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

func JoinLeft(n1 *Node, k *big.Int, n2 *Node) *Node {
	if !IsLeftHeavy(n2, n1) {
		return JoinBalanced(n1, k, n2)
	}
	n := n2 // TODO: check if clone is needed
	n.Left = JoinLeft(n1, k, n.Left)
	n.Left.Parent = n
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

func Join(n1 *Node, k *big.Int, n2 *Node) *Node {
	if IsLeftHeavy(n1, n2) {
		return JoinRight(n1, k, n2)
	}
	if IsLeftHeavy(n2, n1) {
		return JoinLeft(n1, k, n2)
	}
	return JoinBalanced(n1, k, n2)
}

// Variable are named after SPLIT operation in paper [2]
func Split(T *Node, k *big.Int) (*Node, bool, *Node) {
	if T == nil || T.Key == nil {
		return nil, false, nil
	}
	if k == nil {
		return T, false, nil
	}
	L, m, R := Expose(T)
	if kmCmp := k.Cmp(m); kmCmp == 0 {
		return L, true, R
	} else if kmCmp < 0 {
		L_l, b, L_r := Split(L, k)
		return L_l, b, Join(L_r, m , R)
	} else {
		R_l, b, R_r := Split(R, k)
		return Join(L, m, R_l), b, R_r
	}
}

// Variable are named after SPLIT operation in paper [2]
func SplitLast(T *Node) (*Node, *big.Int) {
	L, k, R := Expose(T)
	if R == nil {
		return L, k
	} else {
		T_dash, k_dash := SplitLast(R)
		return Join(L, k, T_dash), k_dash
	}
}

func Join2(Tl, Tr *Node) *Node {
	if Tl == nil {
		return Tr
	} else {
		Tl_dash, k := SplitLast(Tl)
		return Join(Tl_dash, k, Tr)
	}
}

func Search(T *Node, k *big.Int) *Node {
	if T == nil || k.Cmp(T.Key) == 0 {
		return T
	}
	if k.Cmp(T.Key) < 0 {
		return Search(T.Left, k)
	} else {
		return Search(T.Right, k)
	}
}

func Insert(T *Node, k *big.Int) *Node {
	Tl, _, Tr := Split(T, k)
	return Join(Tl, k, Tr)
}

func Delete(T *Node, k *big.Int) *Node {
	Tl, _, Tr := Split(T, k)
	return Join2(Tl, Tr)
}

func Union(T1, T2 *Node) *Node {
	if T1 == nil {
		return T2
	} else if T2 == nil {
		return T1
	} else {
		L2, k2, R2 := Expose(T2)
		L1, _, R1 := Split(T1, k2)
		Tl := Union(L1, L2) // parallelizable
		Tr := Union(R1, R2) // parallelizable
		return Join(Tl, k2, Tr)
	}
}

func Intersect(T1, T2 *Node) *Node {
	if T1 == nil {
		return nil
	} else if T2 == nil {
		return nil
	} else {
		L2, k2, R2 := Expose(T2)
		L1, b, R1 := Split(T1, k2)
		Tl := Intersect(L1, L2) // parallelizable
		Tr := Intersect(R1, R2) // parallelizable
		if b {
			return Join(Tl, k2, Tr)
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
		L2, k2, R2 := Expose(T2)
		L1, _, R1 := Split(T1, k2)
		Tl := Difference(L1, L2) // parallelizable
		Tr := Difference(R1, R2) // parallelizable
		return Join2(Tl, Tr)
	}
}
