package cairo_avl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Dict struct {
	key		*Felt
	value		*Felt
	height		*Felt
	left		*Dict
	right		*Dict
	upserts		*Dict
	deletes		*Dict
	path		string
}

func NewDict(k, v *Felt, D_L, D_R, D_U, D_D *Dict) *Dict {
	d := Dict{key: k, value: v, left: D_L, right: D_R, upserts: D_U, deletes: D_D}
	d.height = NewFelt(1)
	d.height.Add(MaxBigInt(heightDict(d.left), heightDict(d.right)), NewFelt(1))
	d.updatePath("M")
	return &d
}

func (d *Dict) updatePath(path string) {
	d.path = path
	if d.left != nil {
		d.left.updatePath(d.path[:len(d.path)-1] + "LM")
	}
	if d.right != nil {
		d.right.updatePath(d.path[:len(d.path)-1] + "RM")
	}
	if d.upserts != nil {
		d.upserts.updatePath(d.path[:len(d.path)-1] + "NUM")
	}
	if d.deletes != nil {
		d.deletes.updatePath(d.path[:len(d.path)-1] + "NDM")
	}
}

func exposeDict(d *Dict) (k, v *Felt, D_L, D_R, D_U, D_D *Dict) {
	if d != nil && d.key != nil {
		return d.key, d.value, d.left, d.right, d.upserts, d.deletes
	}
	return nil, nil, nil, nil, nil, nil
}

func heightDict(d *Dict) *Felt {
	if d != nil && d.height != nil {
		return d.height
	}
	return NewFelt(0)
}

func (d *Dict) nesting() int {
	return strings.Count(d.path, "N")
}

func dictToNode(d *Dict) (t *Node) {
	if d != nil {
		return NewNode(d.key, d.value, dictToNode(d.left), dictToNode(d.right), nil)
	}
	return nil
}

func nodeToDict(t *Node) (d *Dict) {
	if t != nil {
		return NewDict(t.key, t.value, nodeToDict(t.treeLeft), nodeToDict(t.treeRight), nodeToDict(t.treeNested), nil)
	}
	return nil
}

type DictWalker func(*Dict) interface{}

func (d *Dict) WalkInOrder(w DictWalker) []interface{} {
	if d == nil || d.key == nil {
		return make([]interface{}, 0)
	}
	var (
		left_items, upsert_items, deletes_items, right_items []interface{}
	)
	if d.left != nil {
		left_items = d.left.WalkInOrder(w)
	} else {
		left_items = make([]interface{}, 0)
	}
	if d.upserts != nil {
		upsert_items = d.upserts.WalkInOrder(w)
	} else {
		upsert_items = make([]interface{}, 0)
	}
	if d.deletes != nil {
		deletes_items = d.deletes.WalkInOrder(w)
	} else {
		deletes_items = make([]interface{}, 0)
	}
	if d.right != nil {
		right_items = d.right.WalkInOrder(w)
	} else {
		right_items = make([]interface{}, 0)
	}
	items := make([]interface{}, 0)
	items = append(items, w(d))
	items = append(items, left_items...)
	items = append(items, upsert_items...)
	items = append(items, deletes_items...)
	items = append(items, right_items...)
	return items
}

func (d *Dict) WalkNodesInOrder() []*Dict {
	dict_items := d.WalkInOrder(func(d *Dict) interface{} { return d })
	dictionaries := make([]*Dict, len(dict_items))
	for i := range dict_items {
		dictionaries[i] = dict_items[i].(*Dict)
	}
	return dictionaries
}

func (d *Dict) WalkKeysInOrder() []uint64 {
	key_items := d.WalkInOrder(func(d *Dict) interface{} { return d.key })
	keys := make([]uint64, len(key_items))
	for i := range key_items {
		keys[i] = key_items[i].(*Felt).Uint64()
	}
	return keys
}

func StateChangesFromCsv(stateChanges *bufio.Scanner) (d *Dict, err error) {
	dictByPointer := make(map[int64]*Dict)
	for stateChanges.Scan() {
		line := stateChanges.Text()
		fmt.Println("stateChanges line: ", line)
		tokens := strings.Split(line, ",")
		if len(tokens) != 11 {
			log.Fatal("stateChanges invalid line: ", line)
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
		left, err := strconv.ParseInt(tokens[7], 10, 64)
		if err != nil {
			return nil, err
		}
		D_L := dictByPointer[left]
		right, err := strconv.ParseInt(tokens[8], 10, 64)
		if err != nil {
			return nil, err
		}
		D_R := dictByPointer[right]
		nestedUpsert, err := strconv.ParseInt(tokens[9], 10, 64)
		if err != nil {
			return nil, err
		}
		D_U := dictByPointer[nestedUpsert]
		fmt.Println("previous D_U: ", D_U)
		nestedDelete, err := strconv.ParseInt(tokens[10], 10, 64)
		if err != nil {
			return nil, err
		}
		D_D := dictByPointer[nestedDelete]
		fmt.Println("previous D_D: ", D_D)
		fmt.Println("p: ", p, " k: ", k, " v: ", v, " left: ", left, " right: ", right, " nestedUpsert: ", nestedUpsert, " nestedDelete: ", nestedDelete)
		dictByPointer[p] = NewDict(NewFelt(k), v, D_L, D_R, D_U, D_D)
		fmt.Println("dictByPointer[p]: ", dictByPointer[p])
		d = dictByPointer[p]
	}
	if err := stateChanges.Err(); err != nil {
		return nil, err
	}
	return d, nil
}

func StateChangesFromBinary(statesReader *bufio.Reader, keySize int, nested bool) (d *Dict, err error) {
	buffer := make([]byte, BufferSize)
	var t *Node
	for {
		bytes_read, err := statesReader.Read(buffer)
		fmt.Println("BINARY state changes bytes read: ", bytes_read, " err: ", err)
		if err == io.EOF {
			break
		}
		key_bytes_count := keySize * (bytes_read / keySize)
		duplicated_keys := 0
		fmt.Println("BINARY state changes key_bytes_count: ", key_bytes_count)
		for i := 0; i < key_bytes_count; i += keySize {
			key := binary.BigEndian.Uint32(buffer[i:i+keySize])
			fmt.Println("BINARY state changes key: ", key)
			if t.Search(NewFelt(int64(key))) != nil {
				duplicated_keys++
				continue
			}
			var nestedTree *Node
			if nested && i % 10 == 0 {
				fmt.Printf("Inserting nested key: %d\n", i)
				nestedTree = Insert(/*T=*/nil, NewFelt(int64(i)), NewFelt(0), /*N=*/nil)
				fmt.Printf("Inserted nested tree: %+v\n", nestedTree)
			}
			t = Insert(t, NewFelt(int64(key)), nil, nestedTree)
		}
		fmt.Printf("BINARY state changes duplicated_keys: %d\n", duplicated_keys)
	}
	if nested {
		t = Insert(/*T=*/nil, NewFelt(int64(0)), /*v=*/nil, Insert(/*T=*/nil, NewFelt(int64(0)), /*v=*/nil, t))
	}
	t.GraphAndPicture("t")
	var node2Dict func(*Node) *Dict
	node2Dict = func(n *Node) *Dict {
		if n != nil {
			if rand.Intn(2) == 1 {
				return NewDict(n.key, n.value, node2Dict(n.treeLeft), node2Dict(n.treeRight), node2Dict(n.treeNested), nil)
			} else {
				return NewDict(n.key, n.value, node2Dict(n.treeLeft), node2Dict(n.treeRight), nil, node2Dict(n.treeNested))
			}
		}
		return nil
	}
	d = node2Dict(t)
	d.updatePath("M")
	return d, nil
}

func (d *Dict) Graph(filename string) {
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
	for _, d := range d.WalkNodesInOrder() {
		fmt.Println("k: ", d.key, " v: ", d.value, " left: ", d.left, " right: ", d.right, " nestedUpsert: ", d.upserts, " nestedDelete: ", d.deletes)
		var down string
		if d.upserts != nil {
			if d.deletes != nil {
				down = "{<N>Nu|Nd}"
			} else {
				down = "<N>Nu"
			}
		} else {
			if d.deletes != nil {
				down = "<N>Nd"
			} else {
				down = d.value.String()
			}
		}
		left, right := "", ""
		if d.left != nil {
			left = "<L>L"
		}
		if d.right != nil {
			right = "<R>R"
		}
		s := fmt.Sprintln(d.path, "[label=\"", left, "|{<C>", d.key, "|", down, "}|", right, "\" style=filled fillcolor=\"", colors[d.nesting()], "\"];")
		fmt.Println(s)
		if _, err := f.WriteString(s); err != nil {
			log.Fatal(err)
		}
	}
	for _, d := range d.WalkNodesInOrder() {
		if d.upserts != nil {
			if _, err := f.WriteString(fmt.Sprintln(d.path, ":N -> ", d.upserts.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if d.deletes != nil {
			if _, err := f.WriteString(fmt.Sprintln(d.path, ":N -> ", d.deletes.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if d.left != nil {
			if _, err := f.WriteString(fmt.Sprintln(d.path, ":L -> ", d.left.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
		if d.right != nil {
			if _, err := f.WriteString(fmt.Sprintln(d.path, ":R -> ", d.right.path, ":C;")); err != nil {
				log.Fatal(err)
			}
		}
	}
	if _, err := f.WriteString("}\n"); err != nil {
		log.Fatal(err)
	}
}

func (d *Dict) GraphAndPicture(filename string) error {
	graphDir := "testdata/graph/"
	_ = os.MkdirAll(graphDir, os.ModePerm)
	filepath := graphDir + filename
	_ = os.Remove(filepath + ".dot")
	_ = os.Remove(filepath + ".png")
	d.Graph(filepath)
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

func (d *Dict) Size() int {
	if d == nil {
		return 0
	}
	return len(d.WalkKeysInOrder())
}
