package path

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodePathComponentBinary(t *testing.T) {
	assert.Equal(t, byte(0b00), L.Binary(), "L has unexpected binary value")
	assert.Equal(t, byte(0b01), M.Binary(), "M has unexpected binary value")
	assert.Equal(t, byte(0b10), N.Binary(), "N has unexpected binary value")
	assert.Equal(t, byte(0b11), R.Binary(), "R has unexpected binary value")
}

func TestNodePathComponentString(t *testing.T) {
	assert.Equal(t, "L", L.String(), "L has unexpected string value")
	assert.Equal(t, "M", M.String(), "M has unexpected string value")
	assert.Equal(t, "N", N.String(), "N has unexpected string value")
	assert.Equal(t, "R", R.String(), "R has unexpected string value")
}

func TestNodePathComponentBinaryString(t *testing.T) {
	assert.Equal(t, "00", L.BinaryString(), "L has unexpected binary string value")
	assert.Equal(t, "01", M.BinaryString(), "M has unexpected binary string value")
	assert.Equal(t, "10", N.BinaryString(), "N has unexpected binary string value")
	assert.Equal(t, "11", R.BinaryString(), "R has unexpected binary string value")
}

var b1, b2 *NodePath

func init() {
	b1 = NewNodePath([]byte{0b01})
	b2 = NewNodePath([]byte{0b00, 0b11, 0b01})
}

func TestNodePathValid(t *testing.T) {
	assert.NotNil(t, b1, "b1 is nil")
	assert.NotNil(t, b2, "b2 is nil")
	assert.Nil(t, NewNodePath([]byte{0x11}), "NewPath([]byte{0x11}) is nil")
}

func TestNodePathBinary(t *testing.T) {
	assert.Equal(t, []byte{0b01}, b1.Binary(), "b1 has unexpected binary value")
	assert.Equal(t, []byte{0b00, 0b11, 0b01}, b2.Binary(), "b2 has unexpected binary value")
}

func TestNodePathString(t *testing.T) {
	assert.Equal(t, "M", b1.String(), "b1 has unexpected binary value")
	assert.Equal(t, "LRM", b2.String(), "b2 has unexpected binary value")
}

func TestNodePathBinaryString(t *testing.T) {
	assert.Equal(t, "01", b1.BinaryString(), "b1 has unexpected binary string value")
	assert.Equal(t, "001101", b2.BinaryString(), "b2 has unexpected binary string value")
}
