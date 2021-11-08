package cairo_bptree

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Node23Graph struct {
	node	*Node23
}

func NewGraph(node *Node23) *Node23Graph {
	return &Node23Graph{node}
}

func (g *Node23Graph) saveDot(filename string, debug bool) {
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
	if g.node == nil {
		if _, err := f.WriteString("strict digraph {\nnode [shape=record];}\n"); err != nil {
			log.Fatal(err)
		}
		return
	}
	if _, err := f.WriteString("strict digraph {\nnode [shape=record];\n"); err != nil {
		log.Fatal(err)
	}
	for _, n := range g.node.walkNodesPostOrder() {
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
		ensure(len(n.keys) > 0 , fmt.Sprintf("node %s has zero keys", n))
		if n.isLeaf {
			next := ""
			if n.nextKey() == nil {
				next = "nil"
			} else {
				next = strconv.FormatUint(*n.nextKey(), 10)
			}
			if debug {
				nodeId = fmt.Sprintf("k=%v %s-%v", ptr2pte(n.keys[:len(n.keys)-1]), next, n.keys)
			} else {
				nodeId = fmt.Sprintf("k=%v %s", ptr2pte(n.keys[:len(n.keys)-1]), next)
			}
		} else {
			if debug {
				nodeId = fmt.Sprintf("k=%v-%v", ptr2pte(n.keys), n.keys)
			} else {
				nodeId = fmt.Sprintf("k=%v", ptr2pte(n.keys))
			}
		}
		//s := fmt.Sprintln(/*n.path*/n.rawPointer(), " [label=\"", left, "|{<C>", nodeId, "|", down, "}|", right, "\" style=filled fillcolor=\"", colors[0], "\"];")
		s := fmt.Sprintf("%d [label=\"%s|{<C>%s|%s}|%s\" style=filled fillcolor=\"%s\"];\n", n.rawPointer(), left, nodeId, down, right, colors[0])
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}
	for _, n := range g.node.walkNodesPostOrder() {
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
			//if _, err := f.WriteString(fmt.Sprintln(n.rawPointer(), ":L -> ", treeLeft.rawPointer(), ":C;")); err != nil {
			if _, err := f.WriteString(fmt.Sprintf("%d:L -> %d:C;\n", n.rawPointer(), treeLeft.rawPointer())); err != nil {
				log.Fatal(err)
			}
		}
		if treeDown != nil {
			//if _, err := f.WriteString(fmt.Sprintln(n.rawPointer(), ":D -> ", treeDown.rawPointer(), ":C;")); err != nil {
			if _, err := f.WriteString(fmt.Sprintf("%d:D -> %d:C;\n", n.rawPointer(), treeDown.rawPointer())); err != nil {
				log.Fatal(err)
			}
		}
		if treeRight != nil {
			//if _, err := f.WriteString(fmt.Sprintln(n.rawPointer(), ":R -> ", treeRight.rawPointer(), ":C;")); err != nil {
			if _, err := f.WriteString(fmt.Sprintf("%d:R -> %d:C;\n", n.rawPointer(), treeRight.rawPointer())); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func (g *Node23Graph) saveDotAndPicture(filename string, debug bool) error {
	graphDir := "testdata/graph/"
	_ = os.MkdirAll(graphDir, os.ModePerm)
	filepath := graphDir + filename
	_ = os.Remove(filepath + ".dot")
	_ = os.Remove(filepath + ".png")
	g.saveDot(filepath, debug)
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
