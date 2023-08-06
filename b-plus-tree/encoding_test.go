package bplustree

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeToBytesOfLeaf(t *testing.T) {
	expected := make([]byte, BTREE_PAGE_SIZE)

	leaf := newLeaf(ORDER)
	expected[0] = 1

	leaf.insertToLeafNode([]byte{10, 20}, []byte{34, 12, 47})
	leaf.Next = 5678

	expected[1] = 1
	copy(expected[2:10], []byte{46, 22, 0, 0, 0, 0, 0, 0})
	copy(expected[42:44], []byte{2, 0})
	copy(expected[44:46], []byte{3, 0})
	copy(expected[46:48], []byte{10, 20})
	copy(expected[48:51], []byte{34, 12, 47})

	bytesArr, err := EncodeToBytes(*leaf)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, bytesArr)

	// Error``
	leaf.insertToLeafNode([]byte{24, 123}, make(Data, BTREE_MAX_VAL_SIZE+1))
	_, err = EncodeToBytes(*leaf)
	assert.NotNil(t, err)
}

func TestEncodeToBytesOfNode(t *testing.T) {
	expected := make([]byte, BTREE_PAGE_SIZE)

	node := newNode(ORDER)
	expected[0] = 0

	node.insertToInternalNode([]byte{32, 3}, 0, 943, 342)

	expected[1] = 1

	copy(expected[10:18], []byte{175, 3, 0, 0, 0, 0, 0, 0})
	copy(expected[18:26], []byte{86, 1, 0, 0, 0, 0, 0, 0})

	copy(expected[42:44], []byte{2, 0})
	copy(expected[46:48], []byte{32, 3})

	bytesArr, err := EncodeToBytes(*node)
	assert.Nil(t, err)
	assert.EqualValues(t, expected, bytesArr)

	// Error
	node.insertToInternalNode(make(Data, BTREE_MAX_KEY_SIZE+1), 0, 124, 3354)
	_, err = EncodeToBytes(*node)
	assert.NotNil(t, err)
}

func TestDecodeToBNodeOfLeaf(t *testing.T) {

	leaf := newLeaf(ORDER)
	leaf.Next = 53412

	k1 := make(Data, BTREE_MAX_KEY_SIZE)
	rand.Read(k1)
	k1[0] = 0
	v1 := make(Data, BTREE_MAX_VAL_SIZE)
	rand.Read(v1)
	k2 := make(Data, BTREE_MAX_KEY_SIZE)
	rand.Read(k2)
	k2[0] = 1
	v2 := make(Data, BTREE_MAX_VAL_SIZE)
	rand.Read(v2)

	leaf.insertToLeafNode(k1, v1)
	leaf.insertToLeafNode(k2, v2)

	encodedBytes, err := EncodeToBytes(*leaf)
	assert.Nil(t, err)
	assert.LessOrEqual(t, len(encodedBytes), BTREE_PAGE_SIZE)

	decodedNode, err := DecodeToBNode(encodedBytes)
	assert.Nil(t, err)
	assert.EqualValues(t, decodedNode.NumKeys, 2)
	assert.True(t, decodedNode.IsLeaf)
	assert.Equal(t, decodedNode.Next, leaf.Next)
	assert.Nil(t, decodedNode.Child)
	assert.EqualValues(t, len(decodedNode.Keys), ORDER-1)
	assert.EqualValues(t, decodedNode.Keys[0], k1)
	assert.EqualValues(t, decodedNode.Values[0], v1)
	assert.EqualValues(t, decodedNode.Keys[1], k2)
	assert.EqualValues(t, decodedNode.Values[1], v2)
	assert.EqualValues(t, decodedNode.Keys[2], Data{})
	assert.EqualValues(t, decodedNode.Values[2], Data{})

	k3 := make(Data, BTREE_MAX_KEY_SIZE)
	k3[0] = 2
	rand.Read(k3)
	v3 := make(Data, BTREE_MAX_VAL_SIZE)
	rand.Read(v3)
	leaf.insertToLeafNode(k3, v3)
	encodedBytes, err = EncodeToBytes(*leaf)
	assert.Nil(t, err)
	assert.LessOrEqual(t, len(encodedBytes), BTREE_PAGE_SIZE)
	decodedNode, err = DecodeToBNode(encodedBytes)
	assert.Nil(t, err)
	assert.EqualValues(t, decodedNode.NumKeys, 3)
	assert.EqualValues(t, len(decodedNode.Keys), ORDER-1)
	assert.EqualValues(t, decodedNode.Keys[2], k3)
	assert.EqualValues(t, decodedNode.Values[2], v3)
}

func TestDecodeToBNodeOfNode(t *testing.T) {

	node := newNode(ORDER)

	k1 := make(Data, BTREE_MAX_KEY_SIZE)
	rand.Read(k1)
	k2 := make(Data, BTREE_MAX_KEY_SIZE)
	rand.Read(k2)
	k3 := make(Data, BTREE_MAX_KEY_SIZE)
	rand.Read(k3)

	node.insertToInternalNode(k1, 0, 234, 1468)
	node.insertToInternalNode(k2, 0, 654, 41)
	node.insertToInternalNode(k3, 0, 9232, 347)

	encodedBytes, err := EncodeToBytes(*node)
	assert.Nil(t, err)
	assert.LessOrEqual(t, len(encodedBytes), BTREE_PAGE_SIZE)

	decodedNode, err := DecodeToBNode(encodedBytes)
	assert.Nil(t, err)
	assert.EqualValues(t, decodedNode.NumKeys, 3)
	assert.False(t, decodedNode.IsLeaf)
	assert.Nil(t, decodedNode.Values)
	assert.EqualValues(t, len(decodedNode.Keys), ORDER-1)
	assert.EqualValues(t, len(decodedNode.Child), ORDER)
	assert.EqualValues(t, decodedNode.Keys[0], k3)
	assert.EqualValues(t, decodedNode.Keys[1], k2)
	assert.EqualValues(t, decodedNode.Keys[2], k1)
	assert.EqualValues(t, decodedNode.Child[0], 9232)
	assert.EqualValues(t, decodedNode.Child[1], 347)
	assert.EqualValues(t, decodedNode.Child[2], 41)
	assert.EqualValues(t, decodedNode.Child[3], 1468)
}
