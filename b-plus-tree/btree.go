package bplustree

import (
	"math"
)

type BTree struct {
	Root   uint64
	Order  uint8
	MinKey uint8
	// callbacks for managing on-disk pages
	Get func(uint64) *BNode // reference pointer to a node
	New func(*BNode) uint64 // allocate node with new pointer
	Del func(uint64)        // deallocate a node
}

// ============================= SEARCH OPERATION ==================================
func (t BTree) Search(key Data) (Data, bool) {
	cursor := t.Get(t.Root)
	for {
		if cursor == nil {
			return nil, false
		}
		if cursor.IsLeaf {
			break
		}
		var pos uint8
		for pos = 0; pos < cursor.NumKeys; pos++ {
			if cursor.Keys[pos].gt(key) {
				break
			}
		}
		cursor = t.Get(cursor.Child[pos])
	}
	// cursor now is leaf
	for pos := uint8(0); pos < cursor.NumKeys; pos++ {
		if cursor.Keys[pos].eq(key) {
			return cursor.Values[pos], true
		}
	}
	return nil, false
}

// ============================= INSERT OPERATION ==================================

// Insert key / value pairs into tree:
//
//	key:
//	value:
func (t *BTree) Insert(key Data, value Data) {
	rootNode := t.Get(t.Root)
	if rootNode == nil { // first insert into tree
		rootNode := newLeaf(t.Order)
		rootNode.Keys[0] = key
		rootNode.Values[0] = value
		rootNode.NumKeys += 1
		t.Root = t.New(rootNode)
		return
	}

	insertedNode := t.recursiveInsert(t.Root, key, value)
	if insertedNode == nil {
		return
	}
	if rootNode != insertedNode { // new root pointer
		inserted := t.New(insertedNode)
		t.Root = inserted
	}
}

// Insert `key` / `value` pair into tree recursively, start at `cursor`
func (t *BTree) recursiveInsert(cursor uint64, key Data, value Data) *BNode {
	node := t.Get(cursor)
	if node.IsLeaf {
		if node.NumKeys < t.Order-1 {
			node.insertToLeafNode(key, value)
			return nil
		}
		return t.splitFullLeafAndInsert(cursor, key, value)
	}

	var i uint8
	for i < node.NumKeys {
		if key.gt(node.Keys[i]) {
			i += 1
		} else {
			break
		}
	}
	insertedNode := t.recursiveInsert(node.Child[i], key, value)
	if insertedNode == nil {
		return nil
	}
	if node.NumKeys < t.Order-1 { // merge with current internal node
		node.insertToInternalNode(insertedNode.Keys[0], i, insertedNode.Child[0], insertedNode.Child[1])
		return nil
	} else { // create a parent internal node for `insertedNode` and `node`
		insertedPtr := t.New(insertedNode)
		return t.mergeWithFullNodeAndSplit(cursor, i, insertedPtr)
	}
}

// Insert a `key` / `value` pair into a full leaf node with pointer = `leafPtr`.
// It will create a parent node, `leafPtr` will be left children, and it will create a right child leaf node.
// Returns:
//
//	*BNode: parent internal node
func (t *BTree) splitFullLeafAndInsert(leafPtr uint64, key Data, value Data) *BNode {
	leafNode := t.Get(leafPtr)
	// determine position to insert `key` into, to make sure ascending order
	var insertPos uint8
	for insertPos < leafNode.NumKeys && leafNode.Keys[insertPos].lt(key) {
		insertPos += 1
	}
	// `tempKeys` is a buffer to store keys in ascending order
	tempKeys := make([]Data, t.Order)
	tempKeys[insertPos] = key
	copy(tempKeys[:insertPos], leafNode.Keys[:insertPos])
	copy(tempKeys[insertPos+1:], leafNode.Keys[insertPos:])
	// `tempValues` is a buffer to store values with keys in ascending order
	tempValues := make([]Data, t.Order)
	tempValues[insertPos] = value
	copy(tempValues[:insertPos], leafNode.Values[:insertPos])
	copy(tempValues[insertPos+1:], leafNode.Values[insertPos:])

	// create and allocate new leaf as right child, so `leafNode` will be left children
	rightNode := newLeaf(t.Order)
	rightPtr := t.New(rightNode)

	// determine position to split `tempKeys` and `tempValues`
	splitPos := uint8(math.Ceil(float64(t.Order-1) / 2.0))
	copy(rightNode.Keys, tempKeys[splitPos:])
	copy(rightNode.Values, tempValues[splitPos:])
	copy(leafNode.Keys, tempKeys[:splitPos])
	copy(leafNode.Values, tempValues[:splitPos])
	for i := splitPos; i < t.Order-1; i++ { // reset current keys and values in left child
		leafNode.Keys[i] = nil
		leafNode.Values[i] = nil
	}

	if leafNode.Next != 0 {
		rightNode.Next = leafNode.Next
	}
	leafNode.Next = rightPtr
	leafNode.NumKeys = splitPos
	rightNode.NumKeys = t.Order - splitPos

	parent := newNode(t.Order)
	parent.Keys[0] = tempKeys[splitPos]
	parent.NumKeys = 1
	parent.Child[0] = leafPtr
	parent.Child[1] = rightPtr
	return parent
}

// merge an internal with a full node and split into 2 internal, with a parent internal node
// Returns:
//
//	*BNode: parent internal node after merge and split
func (t *BTree) mergeWithFullNodeAndSplit(fullNodePtr uint64, insertPos uint8, insertedPtr uint64) *BNode {
	// full node will be left, and inserted node will be right
	leftNode := t.Get(fullNodePtr)
	rightNode := t.Get(insertedPtr)

	// `tempKeys` is a buffer to store keys in ascending order
	tempKeys := make([]Data, t.Order)
	tempKeys[insertPos] = rightNode.Keys[0]
	copy(tempKeys[:insertPos], leftNode.Keys[:insertPos])
	copy(tempKeys[insertPos+1:], leftNode.Keys[insertPos:])
	// `tempChilds` is a buffer to store childrens with keys in ascending order
	tempChilds := make([]uint64, t.Order+1)
	copy(tempChilds[:insertPos], leftNode.Child[:insertPos])
	tempChilds[insertPos] = rightNode.Child[0]
	tempChilds[insertPos+1] = rightNode.Child[1]
	if insertPos+2 < t.Order+1 {
		copy(tempChilds[insertPos+2:], leftNode.Child[insertPos+1:])
	}

	// determine position to split `tempKeys` and `tempChilds`
	splitPos := uint8(math.Floor(float64(t.Order) / 2.0))
	// keys and childrens after `splitPos` will be copied to right
	copy(rightNode.Keys, tempKeys[splitPos+1:])
	copy(rightNode.Child, tempChilds[splitPos+1:])
	// keys and children before `splitPos` will be copied to left
	copy(leftNode.Keys, tempKeys[:splitPos])
	copy(leftNode.Child, tempChilds[:splitPos+1])
	for i := splitPos; i < t.Order; i++ { // reset current keys and childrens in left child
		if i < t.Order-1 {
			leftNode.Keys[i] = nil
		}
		if i > splitPos {
			leftNode.Child[i] = 0
		}
	}

	rightNode.NumKeys = leftNode.NumKeys - splitPos
	leftNode.NumKeys = splitPos
	leftNode.Next = insertedPtr

	parent := newNode(t.Order)
	parent.Keys[0] = tempKeys[splitPos]
	parent.NumKeys = 1
	parent.Child[0] = fullNodePtr
	parent.Child[1] = insertedPtr
	return parent
}

// =========================== DELETE BY https://www.cs.usfca.edu/~galles/visualization/BPlusTree.html ===================

// store info of parent of a BNode
type parentInfo struct {
	parentPtr              uint64
	childIndexInParentNode uint8
}

// Delete a node in tree with `key`
func (t *BTree) Delete(key Data) bool {
	return t.doDelete(t.Root, key, make([]parentInfo, 0))
}

// Logic to delete `key` and rebuild tree after delete:
//
//	cursorPointer: start of a sub-tree to delete
//	key: which to delete
//	ancestorsStack: with higher index is closer to parent of `cursorPointer`, and index 0 is the root of a tree
func (t *BTree) doDelete(cursorPointer uint64, key Data, ancestorsStack []parentInfo) bool {
	cursor := t.Get(cursorPointer)
	if cursor == nil {
		return false
	}
	// find sub-tree `key` belong to with `pos`
	var pos uint8
	var cmp int // 0: equal, 1: greater than, -1: less than
	for pos = 0; pos < cursor.NumKeys; pos++ {
		cmp = cursor.Keys[pos].compareValue(key)
		if cmp >= 0 {
			break
		}
	}
	if pos == cursor.NumKeys {
		// delete key in sub-tree of last child of cursor
		if !cursor.IsLeaf {
			return t.doDelete(cursor.Child[cursor.NumKeys], key, append(ancestorsStack, parentInfo{
				parentPtr:              cursorPointer,
				childIndexInParentNode: cursor.NumKeys,
			}))
		}
	} else if !cursor.IsLeaf {
		// delete `key` in sub-tree
		if cmp == 0 {
			return t.doDelete(cursor.Child[pos+1], key, append(ancestorsStack, parentInfo{
				parentPtr:              cursorPointer,
				childIndexInParentNode: pos + 1,
			}))
		} else {
			return t.doDelete(cursor.Child[pos], key, append(ancestorsStack, parentInfo{
				parentPtr:              cursorPointer,
				childIndexInParentNode: pos,
			}))
		}
	} else if cursor.IsLeaf && cmp == 0 {
		// found a leaf contain `key`, delete `key` here
		for j := pos; j < cursor.NumKeys-1; j++ {
			cursor.Keys[j] = cursor.Keys[j+1]
			cursor.Values[j] = cursor.Values[j+1]
		}
		cursor.Keys[cursor.NumKeys-1] = nil
		cursor.Values[cursor.NumKeys-1] = nil
		cursor.NumKeys -= 1

		totalAncestor := len(ancestorsStack)
		if pos == 0 && totalAncestor > 0 {
			// special case: if delete smallest data in leaf, we need to replace every smallest data in parent stack
			var nextSmallest Data
			ancestorIndex := totalAncestor - 1 // initial with direct parent
			ancestorInfo := ancestorsStack[ancestorIndex]
			ancestorNode := t.Get(ancestorInfo.parentPtr)
			childIndexInParentNode := ancestorInfo.childIndexInParentNode
			if cursor.NumKeys == 0 { // delete `key` means delete whole cursor node
				if childIndexInParentNode == ancestorNode.NumKeys {
					// cursor is the last child -> we've just delete maximum value of `ancestorNode` -> delete it in ancestor by nil value
					nextSmallest = nil
				} else {
					// new smallest is minimum value of next child in parent
					nextSmallest = t.Get(ancestorNode.Child[childIndexInParentNode+1]).Keys[0]
				}
			} else { // cursor still have keys -> easy to assign new smallest
				nextSmallest = cursor.Keys[0]
			}
			// update `nextSmallest` to ancestors
			for {
				if childIndexInParentNode > 0 && ancestorNode.Keys[childIndexInParentNode-1].eq(key) {
					ancestorNode.Keys[childIndexInParentNode-1] = nextSmallest
				}
				ancestorIndex -= 1
				if ancestorIndex < 0 {
					break
				}
				ancestorInfo = ancestorsStack[ancestorIndex] // update in grand, grand parents, and so on...
				childIndexInParentNode = ancestorInfo.childIndexInParentNode
				ancestorNode = t.Get(ancestorInfo.parentPtr)
			}
		}
		t.repairAfterDelete(cursorPointer, ancestorsStack)
		return true
	}
	return false
}

// repair sub-tree after delete in this
//
//	cursorPointer: pointer of sub-tree
//	ancestorsStack: with higher index is closer to parent of `cursorPointer`, and index 0 is the root of a tree
func (t *BTree) repairAfterDelete(cursorPointer uint64, ancestorsStack []parentInfo) {
	cursor := t.Get(cursorPointer)
	if cursor.NumKeys >= t.MinKey {
		return
	}
	totalAncestor := len(ancestorsStack)
	if totalAncestor == 0 {
		if cursor.NumKeys == 0 {
			if len(cursor.Child) > 0 {
				t.Root = cursor.Child[0]
			} else {
				// just delete the last `key` of tree, so delete root
				t.Del(t.Root)
				t.Root = 0
			}
		}
	} else {
		parentInfo := ancestorsStack[totalAncestor-1]
		childIndexInParent := parentInfo.childIndexInParentNode
		parentPointer := parentInfo.parentPtr
		parentNode := t.Get(parentPointer)
		if parentNode == nil {
			return
		}
		var leftIdx, rightIdx uint8
		if childIndexInParent > 0 {
			leftIdx = childIndexInParent - 1
		}
		if childIndexInParent < parentNode.NumKeys {
			rightIdx = childIndexInParent + 1
		}

		if l := t.Get(parentNode.Child[leftIdx]); l != nil && l.NumKeys > t.MinKey {
			// steal from left
			t.stealFromLeft(cursorPointer, parentPointer, childIndexInParent)
		} else if r := t.Get(parentNode.Child[rightIdx]); r != nil && r.NumKeys > t.MinKey {
			// steal from right
			t.stealFromRight(cursorPointer, parentPointer, childIndexInParent)
		} else if childIndexInParent == 0 {
			// merge with right sibling
			t.mergeRight(cursorPointer, parentNode.Child[rightIdx], parentPointer, childIndexInParent)
			t.repairAfterDelete(parentPointer, ancestorsStack[:totalAncestor-1])
		} else {
			// merge with left sibling
			t.mergeRight(parentNode.Child[leftIdx], cursorPointer, parentPointer, childIndexInParent-1)
			t.repairAfterDelete(parentPointer, ancestorsStack[:totalAncestor-1])
		}
	}
}

// Args:
//
//	rightPtr: node want to steal a key from left sibling of it
//	parentPtr: parent of right
//	indexInParent: index of rightPtr in parent childrens
func (t *BTree) stealFromLeft(rightPtr uint64, parentPtr uint64, indexInParent uint8) {
	right := t.Get(rightPtr)
	parent := t.Get(parentPtr)
	right.NumKeys += 1
	for i := right.NumKeys - 1; i > 0; i-- {
		right.Keys[i] = right.Keys[i-1]
	}
	leftPtr := parent.Child[indexInParent-1]
	left := t.Get(leftPtr)
	if right.IsLeaf {
		right.Keys[0] = left.Keys[left.NumKeys-1]
		right.Values[0] = left.Values[left.NumKeys-1]
		parent.Keys[indexInParent-1] = left.Keys[left.NumKeys-1]
	} else {
		right.Keys[0] = parent.Keys[indexInParent-1]
		parent.Keys[indexInParent-1] = left.Keys[left.NumKeys-1]
	}
	if !right.IsLeaf {
		for i := right.NumKeys; i > 0; i-- {
			right.Child[i] = right.Child[i-1]
		}
		right.Child[0] = left.Child[left.NumKeys]
		left.Child[left.NumKeys] = 0
	} else {
		left.Values[left.NumKeys-1] = nil
	}
	left.Keys[left.NumKeys-1] = nil
	left.NumKeys -= 1
}

// Args:
//
//	leftPtr: node want to steal a key from right sibling of it
//	parentPtr: parent of right
//	indexInParent: index of leftPtr in parent childrens
func (t *BTree) stealFromRight(leftPtr uint64, parentPtr uint64, indexInParent uint8) {
	left := t.Get(leftPtr)
	parent := t.Get(parentPtr)
	rightPtr := parent.Child[indexInParent+1]
	right := t.Get(rightPtr)
	left.NumKeys += 1

	if left.IsLeaf {
		left.Keys[left.NumKeys-1] = right.Keys[0]
		left.Values[left.NumKeys-1] = right.Values[0]
		parent.Keys[indexInParent] = right.Keys[1]
	} else {
		left.Keys[left.NumKeys-1] = parent.Keys[indexInParent]
		parent.Keys[indexInParent] = right.Keys[0]
	}

	if !left.IsLeaf {
		left.Child[left.NumKeys] = right.Child[0]
		for i := uint8(1); i < right.NumKeys+1; i++ {
			right.Child[i-1] = right.Child[i]
		}
	}
	for i := uint8(1); i < right.NumKeys; i++ {
		right.Keys[i-1] = right.Keys[i]
		if right.IsLeaf {
			right.Values[i-1] = right.Keys[i]
		}
	}
	right.Keys[right.NumKeys-1] = nil
	if right.IsLeaf {
		right.Values[right.NumKeys-1] = nil
	} else {
		right.Child[right.NumKeys] = 0
	}
	right.NumKeys -= 1
}

// Merge 2 adjacency nodes, both has less than 1/2 keys so it can be merged. After merge, right node will be deleted:
//
//	leftPtr: node want to merge with right sibling
//	rightPtr: right sibling of leftPtr
//	parentPtr: parent of both nodes
//	leftIndexInParent: index of leftPtr in parent childrens
func (t *BTree) mergeRight(leftPtr uint64, rightPtr uint64, parentPtr uint64, leftIndexInParent uint8) {
	left := t.Get(leftPtr)
	parent := t.Get(parentPtr)
	right := t.Get(rightPtr)

	// append keys, values, childrens to left node
	if !left.IsLeaf {
		left.Keys[left.NumKeys] = parent.Keys[leftIndexInParent]
		left.Child[left.NumKeys+1] = right.Child[0]
	}
	for i := uint8(0); i < right.NumKeys; i++ {
		insertIndex := left.NumKeys + i
		if left.IsLeaf {
			left.Values[insertIndex] = right.Values[i] // only append values if it is a leaf
		} else {
			insertIndex += 1 // +1 here because: we steal 1 key from parent above and index need to be after that
			left.Child[insertIndex+1] = right.Child[i+1]
		}
		left.Keys[insertIndex] = right.Keys[i]
	}
	if !left.IsLeaf {
		left.NumKeys += right.NumKeys + 1 // +1 here because: we steal 1 key from parent above
	} else {
		left.NumKeys += right.NumKeys
		left.Next = right.Next
	}
	// remove keys, childrens with index of right node (leftIndexInParent + 1) in parent node
	for i := leftIndexInParent + 1; i < parent.NumKeys; i++ {
		parent.Child[i] = parent.Child[i+1]
		parent.Keys[i-1] = parent.Keys[i]
	}
	parent.Keys[parent.NumKeys-1] = nil
	parent.Child[parent.NumKeys] = 0
	parent.NumKeys -= 1
	t.Del(rightPtr)
}
