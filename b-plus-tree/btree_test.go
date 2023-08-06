package bplustree

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

type C struct {
	tree  BTree
	pages map[uint64]*BNode
}

func newC(order uint8) *C {
	pages := map[uint64]*BNode{}
	return &C{
		tree: BTree{
			Root:   0,
			Order:  order,
			MinKey: (order+1)/2 - 1,
			Get: func(ptr uint64) *BNode {
				if node, ok := pages[ptr]; !ok {
					return nil
				} else {
					return node
				}
			},
			New: func(node *BNode) uint64 {
				key := uint64(uintptr(unsafe.Pointer(node)))
				pages[key] = node
				return key
			},
			Del: func(ptr uint64) {
				delete(pages, ptr)
			},
		},
		pages: pages,
	}
}

func (c *C) add(key []byte, val []byte) {
	c.tree.Insert(key, val)
}

func (c *C) del(key []byte) {
	c.tree.Delete(key)
}

// func (c *C) PrintNode(nodePointer uint64) {
// 	node := c.tree.Get(nodePointer)
// 	if node == nil {
// 		return
// 	}
// 	if node.IsLeaf {
// 		fmt.Print("(leaf ", node.NumKeys, " ")
// 		defer fmt.Print(")\t")
// 		for idx := uint8(0); idx < node.NumKeys; idx++ {
// 			fmt.Printf("%d=%d, ", binary.LittleEndian.Uint16(node.Keys[idx]), binary.LittleEndian.Uint16(node.Values[idx]))
// 		}
// 	} else {
// 		fmt.Print("(node ", node.NumKeys, " ")
// 		for idx := uint8(0); idx < node.NumKeys; idx++ {
// 			fmt.Printf("%d, ", binary.LittleEndian.Uint16(node.Keys[idx]))
// 		}
// 		fmt.Print(")\n")
// 		for _, child := range node.Child {
// 			c.PrintNode(child)
// 		}
// 		fmt.Printf("\n")
// 	}
// }

// func (c *C) PrintLeaf(nodePointer uint64) {
// 	node := c.tree.Get(nodePointer)
// 	if node == nil {
// 		return
// 	}
// 	cursor := node
// 	for cursor != nil && !cursor.IsLeaf {
// 		cursor = c.tree.Get(cursor.Child[0])
// 	}
// 	for cursor != nil {
// 		fmt.Print("(leaf ", cursor.NumKeys, " ")
// 		for idx := uint8(0); idx < cursor.NumKeys; idx++ {
// 			fmt.Printf("%d=%d, ", binary.LittleEndian.Uint16(cursor.Keys[idx]), binary.LittleEndian.Uint16(cursor.Values[idx]))
// 		}
// 		fmt.Print(") -> ")
// 		cursor = c.tree.Get(cursor.Next)
// 	}
// 	fmt.Printf("\n%+v\n", c.pages)
// }

// func (c *C) PrintTree() {
// 	fmt.Println("------- PRINT TREE ------------")
// 	c.PrintNode(c.tree.Root)
// 	fmt.Println("------- PRINT LEAFs ------------")
// 	c.PrintLeaf(c.tree.Root)
// 	fmt.Printf("\n\n")
// }

func createData(input uint16) Data {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, input)
	return buf
}

func assertNodeKeys(t *testing.T, c *C, expectedKeys []int) {
	node := c.tree.Get(c.tree.Root)
	assert.NotNil(t, node)
	// find first leaf

	nodeQueue := make([]BNode, 0)
	nodeQueue = append(nodeQueue, *node)
	expectedIdx := 0
	for {
		cursor := nodeQueue[0]
		for idx := uint8(0); idx < cursor.NumKeys; idx++ {
			assert.EqualValues(t, cursor.Keys[idx], createData(uint16(expectedKeys[expectedIdx])))
			expectedIdx += 1
		}
		if cursor.Child != nil {
			for _, child := range cursor.Child {
				if childNode := c.tree.Get(child); childNode != nil {
					nodeQueue = append(nodeQueue, *childNode)
				}
			}
		}
		if len(nodeQueue) == 1 {
			break
		}
		nodeQueue = nodeQueue[1:]
	}
}

func assertLeafs(t *testing.T, c *C, expectedLeafs [][2]int) {
	node := c.tree.Get(c.tree.Root)
	assert.NotNil(t, node)
	// find first leaf
	cursor := node
	for cursor != nil && !cursor.IsLeaf {
		cursor = c.tree.Get(cursor.Child[0])
	}
	expectedIdx := 0
	for cursor != nil {
		for idx := uint8(0); idx < cursor.NumKeys; idx++ {
			assert.EqualValues(t, cursor.Keys[idx], createData(uint16(expectedLeafs[expectedIdx][0])))
			assert.EqualValues(t, cursor.Values[idx], createData(uint16(expectedLeafs[expectedIdx][1])))
			expectedIdx += 1
		}
		cursor = c.tree.Get(cursor.Next)
	}
}

func TestCase1(t *testing.T) {
	c := newC(3)
	insertDatas := [][2]int{
		{5, 5}, {15, 15}, {25, 25}, {35, 35}, {45, 45}, {20, 20}, {30, 30}, {55, 55}, {40, 40},
	}
	for _, insertData := range insertDatas {
		c.add(
			createData(uint16(insertData[0])), createData(uint16(insertData[1])),
		)
	}
	assertNodeKeys(t, c, []int{
		25, 15, 35, 45, 5, 15, 20, 25, 30, 35, 40, 45, 55,
	})
	assertLeafs(t, c, [][2]int{
		{5, 5}, {15, 15}, {20, 20}, {25, 25}, {30, 30}, {35, 35}, {40, 40}, {45, 45}, {55, 55},
	})

	deleteDatas := []int{
		40, 5, 45, 35, 25, 55,
	}
	for _, deleteData := range deleteDatas {
		c.del(createData(uint16(deleteData)))
	}
	assertNodeKeys(t, c, []int{
		20, 30, 15, 20, 30,
	})
	assertLeafs(t, c, [][2]int{
		{15, 15}, {20, 20}, {30, 30},
	})
}

func TestCase2(t *testing.T) {
	c := newC(4)
	insertDatas := [][2]int{
		{20, 20}, {15, 15}, {10, 10}, {15, 151}, {25, 25}, {28, 28}, {18, 18}, {21, 21}, {20, 201}, {28, 281}, {20, 202},
	}
	for _, insertData := range insertDatas {
		c.add(
			createData(uint16(insertData[0])), createData(uint16(insertData[1])),
		)
	}
	assertNodeKeys(t, c, []int{
		20, 15, 20, 25, 10, 15, 15, 18, 20, 20, 20, 21, 25, 28, 28,
	})
	assertLeafs(t, c, [][2]int{
		{10, 10}, {15, 151}, {15, 15}, {18, 18}, {20, 202}, {20, 201}, {20, 20}, {21, 21}, {25, 25}, {28, 281}, {28, 28},
	})

	deleteDatas := []int{
		25, 20, 20, 28, 5, 28, 15, 18,
	}
	for _, deleteData := range deleteDatas {
		c.del(createData(uint16(deleteData)))
	}
	assertNodeKeys(t, c, []int{
		20, 15, 21, 10, 15, 20, 21,
	})
	assertLeafs(t, c, [][2]int{
		{10, 10}, {15, 151}, {20, 201}, {21, 21},
	})

}

func TestCase3(t *testing.T) {

	c := newC(4)

	insertDatas := [][2]int{
		{20, 3456}, {15, 45}, {10, 734},
	}
	for _, insertData := range insertDatas {
		c.add(
			createData(uint16(insertData[0])), createData(uint16(insertData[1])),
		)
	}
	rootNode := c.tree.Get(c.tree.Root)
	assert.NotNil(t, rootNode)
	assert.True(t, rootNode.IsLeaf)
	assert.EqualValues(t, rootNode.NumKeys, 3)
	assert.EqualValues(t, rootNode.Keys[0], createData(uint16(10)))
	assert.EqualValues(t, rootNode.Keys[1], createData(uint16(15)))
	assert.EqualValues(t, rootNode.Keys[2], createData(uint16(20)))

	assert.EqualValues(t, rootNode.Values[0], createData(uint16(734)))
	assert.EqualValues(t, rootNode.Values[1], createData(uint16(45)))
	assert.EqualValues(t, rootNode.Values[2], createData(uint16(3456)))

	deleteDatas := []int{
		15, 10, 20,
	}
	for _, deleteData := range deleteDatas {
		c.del(createData(uint16(deleteData)))
	}

	assert.Nil(t, c.tree.Get(c.tree.Root))
}

func TestSearch(t *testing.T) {
	c := newC(4)
	val, found := c.tree.Search(createData(uint16(15)))
	assert.False(t, found)
	assert.Nil(t, val)

	insertDatas := [][2]int{
		{20, 20}, {15, 15}, {10, 6534}, {3, 3}, {8, 8}, {9, 745},
	}
	for _, insertData := range insertDatas {
		c.add(
			createData(uint16(insertData[0])), createData(uint16(insertData[1])),
		)
	}

	val, found = c.tree.Search(createData(uint16(10)))
	assert.True(t, found)
	assert.EqualValues(t, val, createData(uint16(6534)))

	val, found = c.tree.Search(createData(uint16(9)))
	assert.True(t, found)
	assert.EqualValues(t, val, createData(uint16(745)))

	val, found = c.tree.Search(createData(uint16(6)))
	assert.False(t, found)
	assert.Nil(t, val)

	val, found = c.tree.Search(createData(uint16(100)))
	assert.False(t, found)
	assert.Nil(t, val)

	val, found = c.tree.Search(createData(uint16(1)))
	assert.False(t, found)
	assert.Nil(t, val)
}
