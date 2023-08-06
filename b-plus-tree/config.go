package bplustree

import "bytes"

const (
	ORDER              = 4
	BTREE_PAGE_SIZE    = 4096
	BTREE_MAX_KEY_SIZE = 347
	BTREE_MAX_VAL_SIZE = 1000
)

type Data []byte

func (d Data) compareValue(other Data) int {
	return bytes.Compare(d, other)
}

// Check if `data` equal `other` or not
func (d Data) eq(other Data) bool {
	return bytes.Equal(d, other)
}

// Check if `data` less than `other` or not
func (d Data) lt(other Data) bool {
	return bytes.Compare(d, other) < 0
}

// Check if `data` greater than `other` or not
func (d Data) gt(other Data) bool {
	return bytes.Compare(d, other) > 0
}
