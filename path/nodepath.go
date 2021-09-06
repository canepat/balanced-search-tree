package path

import "fmt"

type NodePathComponent uint64

const (
	L NodePathComponent = iota
	M
	N
	R
)

func (c NodePathComponent) Valid() bool {
	return byte(c) <= byte(R)
}

func (c NodePathComponent) Binary() byte {
	return byte(c)
}

func (c NodePathComponent) String() string {
	switch c {
	case L:
		return "L"
	case M:
		return "M"
	case N:
		return "N"
	case R:
		return "R"
	default:
		return fmt.Sprintf("%d", byte(c))
	}
}

func (c NodePathComponent) BinaryString() string {
	return fmt.Sprintf("%02b", byte(c))
}

type NodePath struct {
	Bytes		[]byte	// bits packed into bytes
	BitLength	int		// length in bits
}

func NewNodePath(bytes []byte) *NodePath {
	node_path := &NodePath{Bytes: bytes}
	if !node_path.Valid() {
		return nil
	}
	return node_path
}

func (b *NodePath) Valid() bool {
	for _, b := range b.Bytes {
		if !NodePathComponent(b).Valid() {
			return false
		}
	}
	return true
}

func (b *NodePath) Binary() []byte {
	return b.Bytes
}

func (b *NodePath) String() string {
	var s string
	for i:=0; i < len(b.Bytes); i++ {
		s += NodePathComponent(b.Bytes[i]).String()
	}
	return s
}

func (b *NodePath) BinaryString() string {
	var s string
	for i:=0; i < len(b.Bytes); i++ {
		s += NodePathComponent(b.Bytes[i]).BinaryString()
	}
	return s
}

func (b *NodePath) Add(component NodePathComponent) *NodePath {
	return b
}
