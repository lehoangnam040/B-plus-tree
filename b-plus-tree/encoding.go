package bplustree

import (
	"encoding/binary"
	"fmt"
)

/*
*
ORDER = 4
Data = x*B
1B = uint8

| IsLeaf | NumKeys | Next   |  Child          | k0len | v0len | k0      | v0      | k1len | v1len | k1      | v1      | k2len | v2len | k2      |  v2
| 1B     | 1B      | 8B     |  ORDER*8B = 32B |  2B   |  2B   | k0len B | v0len B |  2B   |  2B   | k1len B | v1len B |  2B   |  2B   | k2len B | v2len B
*
*/

func EncodeToBytes(node BNode) ([]byte, error) {

	result := make([]byte, BTREE_PAGE_SIZE)
	if node.IsLeaf {
		result[0] = 1
	}
	result[1] = node.NumKeys
	if node.IsLeaf {
		binary.LittleEndian.PutUint64(result[2:10], node.Next)
	}
	if !node.IsLeaf {
		for i := 0; i < ORDER; i++ {
			binary.LittleEndian.PutUint64(result[10+i*8:10+(i+1)*8], node.Child[i])
		}
	}
	// 2+8+32=42
	offset := 42
	for i := 0; i < ORDER-1; i++ {
		klen := len(node.Keys[i])
		if klen > BTREE_MAX_KEY_SIZE {
			return nil, fmt.Errorf("key %d has bytes = %d larger than maximum %d", i, klen, BTREE_MAX_KEY_SIZE)
		}
		vlen := 0
		if node.IsLeaf {
			vlen = len(node.Values[i])
			if vlen > BTREE_MAX_VAL_SIZE {
				return nil, fmt.Errorf("value %d has bytes = %d larger than maximum %d", i, vlen, BTREE_MAX_KEY_SIZE)
			}
		}
		binary.LittleEndian.PutUint16(result[offset:offset+2], uint16(klen))
		if node.IsLeaf {
			binary.LittleEndian.PutUint16(result[offset+2:offset+4], uint16(vlen))
		}
		copy(result[offset+4:offset+4+klen], node.Keys[i])
		if node.IsLeaf {
			copy(result[offset+4+klen:offset+4+klen+vlen], node.Values[i])
		}
		offset += 4 + klen + vlen
	}
	return result, nil
}

func DecodeToBNode(pageData []byte) (*BNode, error) {
	node := BNode{
		Keys:    make([]Data, ORDER-1),
		IsLeaf:  false,
		NumKeys: 0,
		Next:    0,
	}
	node.IsLeaf = (pageData[0] != 0)
	node.NumKeys = pageData[1]
	if node.IsLeaf {
		node.Next = binary.LittleEndian.Uint64(pageData[2:10])
	}
	// child
	if !node.IsLeaf {
		node.Child = make([]uint64, ORDER)
		for i := 0; i < ORDER; i++ {
			node.Child[i] = binary.LittleEndian.Uint64(pageData[10+i*8 : 10+(i+1)*8])
		}
	} else {
		node.Values = make([]Data, ORDER-1)
	}
	// 2+8+32=42
	offset := 42
	for i := 0; i < ORDER-1; i++ {
		klen := int(binary.LittleEndian.Uint16(pageData[offset : offset+2]))
		node.Keys[i] = pageData[offset+4 : offset+4+klen]
		vlen := 0
		if node.IsLeaf {
			vlen = int(binary.LittleEndian.Uint16(pageData[offset+2 : offset+4]))
			node.Values[i] = pageData[offset+4+klen : offset+4+klen+vlen]
		}
		offset += 4 + klen + vlen
	}
	return &node, nil
}
