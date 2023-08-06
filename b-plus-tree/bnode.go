package bplustree

func newNode(order uint8) *BNode {
	return &BNode{
		Keys:   make([]Data, order-1),
		Values: nil,
		Child:  make([]uint64, order),
		IsLeaf: false,
		Next:   0,
	}
}

func newLeaf(order uint8) *BNode {
	return &BNode{
		Keys:   make([]Data, order-1),
		Values: make([]Data, order-1),
		Child:  nil,
		IsLeaf: true,
		Next:   0,
	}
}

type BNode struct {
	Keys    []Data
	Values  []Data
	Child   []uint64 // pointers to child nodes
	Next    uint64   // pointer to next leaf node, if this is leaf
	NumKeys uint8    // total keys inside this node
	IsLeaf  bool
}

// Insert a `key` to an internal node:
//
//	key: `key` want to insert
//	insertPos: position to insert `key` into
//	left: left pointer at `insertPos`
//	right: right pointer at `insertPos`
func (node *BNode) insertToInternalNode(key Data, insertPos uint8, left uint64, right uint64) {
	// shift current keys/childs 1 to the right
	for i := node.NumKeys; i > insertPos; i-- {
		node.Keys[i] = node.Keys[i-1]
		node.Child[i+1] = node.Child[i]
	}
	// to insert `key` at `insertPos`
	node.Keys[insertPos] = key
	node.NumKeys += 1
	node.Child[insertPos] = left
	node.Child[insertPos+1] = right
}

// Insert a `key`/`value` pair to a leaf node:
//
//	key:
//	value:
func (node *BNode) insertToLeafNode(key Data, value Data) {
	// find a position to insert "key" in, make sure to keep ascending order
	var insertPos uint8
	for insertPos < node.NumKeys && node.Keys[insertPos].lt(key) {
		insertPos += 1
	}

	// shift current keys/childs 1 to the right
	for i := node.NumKeys; i > insertPos; i-- {
		node.Keys[i] = node.Keys[i-1]
		node.Values[i] = node.Values[i-1]
	}
	// to insert `key` at `insertPos`
	node.Keys[insertPos] = key
	node.Values[insertPos] = value
	node.NumKeys += 1
}
