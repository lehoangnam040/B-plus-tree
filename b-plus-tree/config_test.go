package bplustree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConst(t *testing.T) {
	assert.LessOrEqual(t, 2+8+ORDER*8+(ORDER-1)*(2+2+BTREE_MAX_KEY_SIZE+BTREE_MAX_VAL_SIZE), BTREE_PAGE_SIZE)
}

func TestCompareValue(t *testing.T) {

	d := Data([]byte{0, 10})

	assert.Equal(t, d.compareValue([]byte{0, 10}), 0)
	assert.Equal(t, d.compareValue([]byte{0, 9}), 1)
	assert.Equal(t, d.compareValue([]byte{0, 11}), -1)
}

func TestEq(t *testing.T) {

	d := Data([]byte{0, 10})

	assert.True(t, d.eq([]byte{0, 10}))
	assert.False(t, d.eq([]byte{0, 9}))
}

func TestLt(t *testing.T) {

	d := Data([]byte{0, 10})

	assert.True(t, d.lt([]byte{0, 11}))
	assert.False(t, d.lt([]byte{0, 10}))
	assert.False(t, d.lt([]byte{0, 9}))
}

func TestGt(t *testing.T) {

	d := Data([]byte{0, 10})

	assert.True(t, d.gt([]byte{0, 9}))
	assert.False(t, d.gt([]byte{0, 10}))
	assert.False(t, d.gt([]byte{0, 11}))
}
