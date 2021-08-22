/*
* Adelson-Velsky and Landis (AVL) self-balancing binary search tree implementation. References:
* [1] https://zhjwpku.com/assets/pdf/AED2-10-avl-paper.pdf
* [2] https://www.cs.cmu.edu/~blelloch/papers/BFS16.pdf
* [3] https://github.com/cmuparlay/PAM
*/

package main

import (
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
	key    *big.Int
	left   *Node
	right  *Node
	parent *Node
	height *big.Int
	value  *big.Int
}

func NewNode(key *big.Int, left, right *Node) *Node {
	n := Node{key: key, left: left, right: right}
	if n.left != nil {
		n.left.parent = &n
	}
	if n.right != nil {
		n.right.parent = &n
	}
	n.height = big.NewInt(0)
	UpdateHeight(&n)
	return &n
}

func (n *Node) PathDepth() uint64 {
	depth := uint64(0)
	ancestor := n.parent
	for ancestor != nil {
		depth += 1
		ancestor = ancestor.parent
	}
	return depth
}

type Walker func(*Node) interface{}

func (n *Node) WalkInOrder(w Walker) []interface{} {
	var (
		left_items, right_items []interface{}
	)
	if n.left != nil {
		left_items = n.left.WalkInOrder(w)
	} else {
		left_items = make([]interface{}, 0)
	}
	if n.right != nil {
		right_items = n.right.WalkInOrder(w)
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

func (n *Node) WalkKeysInOrder() []*big.Int {
	key_items := n.WalkInOrder(func(n *Node) interface{} { return n.key })
	keys := make([]*big.Int, len(key_items))
	for i := range key_items {
		keys[i] = key_items[i].(*big.Int)
	}
	return keys
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
		if n.left != nil {
			left = "<L>L"
		}
		if n.right != nil {
			right = "<R>R"
		}
		s := fmt.Sprintln(n.key, " [label=\"", left, "|{<C>", n.key, "|", n.value, "}|", right, "\" style=filled fillcolor=\"", colors[2], "\"];")
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
		if n.parent != nil {
			var direction string
			if n.parent.left == n {
				direction = "L"
			} else {
				direction = "R"
			}
			if _, err := f.WriteString(fmt.Sprintln(n.parent.key, ":", direction, " -> ", n.key, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func GraphAndPicture(n *Node, filename string) {
	Graph(n, filename)
	dotExecutable, _ := exec.LookPath("dot")
	cmdDot := &exec.Cmd{
		Path: dotExecutable,
		Args: []string{dotExecutable, "-Tpng", filename + ".dot", "-o", filename + ".png"},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	if err := cmdDot.Run(); err != nil {
		fmt.Println("Error executing graphviz dot: ", err)
	}
}

func Expose(n *Node) (*Node, *big.Int, *Node) {
	if n != nil {
		return n.left, n.key, n.right
	}
	return nil, nil, nil
}

func Height(n *Node) *big.Int {
	if n != nil {
		return n.height
	}
	return big.NewInt(0)
}

func UpdateHeight(n *Node) *big.Int {
	if n.left != nil {
		if n.right != nil {
			n.height.Add(Max(UpdateHeight(n.left), UpdateHeight(n.right)), big.NewInt(1))
		} else {
			n.height.Add(UpdateHeight(n.left), big.NewInt(1))
		}
	} else if n.right != nil {
		if n.right != nil {
			n.height.Add(UpdateHeight(n.right), big.NewInt(1))
		}
	}
	return n.height
}

func Update(n *Node) *Node {
	UpdateHeight(n)
	return n
}

func RotateLeft(x *Node) *Node {
	y := x.right
	z := y.left
	y.left = x
	x.right = z
	Update(x)
	Update(y)
	return y
}

func RotateRight(x *Node) *Node {
	y := x.left
	z := y.right
	y.right = x
	x.left = z
	Update(x)
	Update(y)
	return y
}

func DoubleRotateLeft(x *Node) *Node {
	r := x.right
	root := r.left
	r.left = root.right
	Update(r)
	root.right = r
	x.right = root
	return RotateLeft(x)
}

func DoubleRotateRight(x *Node) *Node {
	l := x.left
	root := l.right
	l.right = root.left
	Update(l)
	root.left = l
	x.left = root
	return RotateRight(x)
}

func IsLeftHeavy(left, right *Node) bool {
	return Height(left).Cmp(big.NewInt(0).Add(Height(right), big.NewInt(1))) > 0
}

func IsSingleRotation(n *Node, isLeft bool) bool {
	if n == nil {
		return false
	}
	if isLeft {
		return Height(n.left).Cmp(Height(n.right)) > 0
	} else {
		return Height(n.left).Cmp(Height(n.right)) <= 0
	}
}

func JoinBalanced(left *Node, k *big.Int, right *Node) *Node {
	return NewNode(k, left, right)
}

func JoinRight(n1 *Node, k *big.Int, n2 *Node) *Node {
	if !IsLeftHeavy(n1, n2) {
		return JoinBalanced(n1, k, n2)
	}
	n := n1 // TODO: check if clone is needed
	n.right = JoinRight(n.right, k, n2)
	n.right.parent = n
	if IsLeftHeavy(n.right, n.left) {
		if IsSingleRotation(n.right, false) {
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
	n.left = JoinLeft(n1, k, n.left)
	n.left.parent = n
	if IsLeftHeavy(n.left, n.right) {
		if IsSingleRotation(n.left, true) {
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
	if T == nil {
		return nil, false, nil
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

func main() {
	t1 := NewNode(big.NewInt(0), nil, nil)
	fmt.Println("t1 keys: ", t1.WalkKeysInOrder())
	t2 := NewNode(big.NewInt(18),
		NewNode(big.NewInt(15), nil, nil), NewNode(big.NewInt(21), nil, nil),
	)
	fmt.Println("t2 keys: ", t2.WalkKeysInOrder())
 	t3 := NewNode(big.NewInt(188),
	 	NewNode(big.NewInt(155),
			NewNode(big.NewInt(154), nil, nil), NewNode(big.NewInt(156), nil, nil),
		),
		NewNode(big.NewInt(210),
			NewNode(big.NewInt(200),
				NewNode(big.NewInt(199), nil, nil), NewNode(big.NewInt(202), nil, nil),
			),
			NewNode(big.NewInt(300),
				NewNode(big.NewInt(201), nil, nil), NewNode(big.NewInt(1560), nil, nil),
			),
		),
	)
	fmt.Println("t3 keys: ", t3.WalkKeysInOrder())

	j1 := Join(t2, big.NewInt(50), t3)
	fmt.Println("j1 keys: ", j1.WalkKeysInOrder())
	GraphAndPicture(j1, "j1")

	t4 := NewNode(big.NewInt(19),
		NewNode(big.NewInt(11), nil, nil), NewNode(big.NewInt(157), nil, nil),
	)
	fmt.Println("t4 keys: ", t4.WalkKeysInOrder())

	u1 := Union(j1, t4)
	fmt.Println("u1 keys: ", u1.WalkKeysInOrder())
	GraphAndPicture(u1, "u1")

	t5 := NewNode(big.NewInt(4),
		NewNode(big.NewInt(1), nil, nil), NewNode(big.NewInt(5), nil, nil),
	)
	fmt.Println("t5 keys: ", t5.WalkKeysInOrder())
	t6 := NewNode(big.NewInt(3),
		NewNode(big.NewInt(2), nil, nil), NewNode(big.NewInt(7), nil, nil),
	)
	fmt.Println("t6 keys: ", t6.WalkKeysInOrder())
	u2 := Union(t5, t6)
	fmt.Println("u2 keys: ", u2.WalkKeysInOrder())
	GraphAndPicture(u2, "u2")

	d1 := Difference(u2, t5)
	fmt.Println("d1 keys: ", d1.WalkKeysInOrder())
	GraphAndPicture(d1, "d1")
	d2 := Difference(u2, t6)
	fmt.Println("d2 keys: ", d2.WalkKeysInOrder())
	GraphAndPicture(d2, "d2")
}