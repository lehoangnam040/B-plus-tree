package bplustree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	node := newNode(ORDER)
	assert.False(t, node.IsLeaf)
	assert.Nil(t, node.Values)
	assert.Equal(t, len(node.Keys), ORDER-1)
	assert.Equal(t, len(node.Child), ORDER)
	assert.Equal(t, node.NumKeys, uint8(0))
	assert.Equal(t, node.Next, uint64(0))
}

func TestNewLeaf(t *testing.T) {
	node := newLeaf(ORDER)
	assert.True(t, node.IsLeaf)
	assert.Nil(t, node.Child)
	assert.Equal(t, len(node.Keys), ORDER-1)
	assert.Equal(t, len(node.Values), ORDER-1)
	assert.Equal(t, node.NumKeys, uint8(0))
	assert.Equal(t, node.Next, uint64(0))
}

func TestInsertToInternalNode(t *testing.T) {
	node := newNode(ORDER)
	node.NumKeys = 1
	node.Keys[0] = []byte{0, 10}
	node.Child[0] = 123
	node.Child[1] = 432

	node.insertToInternalNode([]byte{0, 3}, 0, 943, 342)

	assert.Equal(t, node.NumKeys, uint8(2))
	assert.EqualValues(t, node.Keys[0], []byte{0, 3})
	assert.EqualValues(t, node.Keys[1], []byte{0, 10})

	assert.Equal(t, node.Child[0], uint64(943))
	assert.Equal(t, node.Child[1], uint64(342))
	assert.Equal(t, node.Child[2], uint64(432))
}

func TestInsertToLeafNode(t *testing.T) {
	node := newLeaf(ORDER)
	node.NumKeys = 2
	node.Keys[0] = []byte{0, 10}
	node.Values[0] = []byte{0, 10}
	node.Keys[1] = []byte{0, 20}
	node.Values[1] = []byte{0, 20}

	node.insertToLeafNode([]byte{0, 15}, []byte{0, 245})

	assert.Equal(t, node.NumKeys, uint8(3))
	assert.EqualValues(t, node.Keys[0], []byte{0, 10})
	assert.EqualValues(t, node.Keys[1], []byte{0, 15})
	assert.EqualValues(t, node.Keys[2], []byte{0, 20})

	assert.EqualValues(t, node.Values[0], []byte{0, 10})
	assert.EqualValues(t, node.Values[1], []byte{0, 245})
	assert.EqualValues(t, node.Values[2], []byte{0, 20})
}
