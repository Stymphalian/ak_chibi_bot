package misc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRollingArray(t *testing.T) {
	assert := assert.New(t)

	arr := NewRollingArray[int](3)
	assert.Equal(0, arr.GetLength())
	arr.Add(1)
	arr.Add(2)
	arr.Add(3)
	assert.Equal(3, arr.GetLength())

	arr.Add(4)
	// 4, 2, 3
	assert.Equal(3, arr.GetLength())

	// Make sure the values are correct
	assert.Equal(4, arr.Values[0])
	assert.Equal(2, arr.Values[1])
	assert.Equal(3, arr.Values[2])
	assert.Equal(2, arr.Get(0))
	assert.Equal(3, arr.Get(1))
	assert.Equal(4, arr.Get(2))

	arr.RemoveFirst()
	// 4, _, 3
	assert.Equal(2, arr.GetLength())

	arr.RemoveLast()
	// _, _, 3
	assert.Equal(1, arr.GetLength())

	arr.RemoveLast()
	// _, _, _
	assert.Equal(0, arr.GetLength())

	// Test re-adding values
	arr.Add(7)
	arr.Add(8)
	// 8, _, 7  - start is 7
	assert.Equal(2, arr.GetLength())
	assert.Equal(8, arr.Values[0])
	assert.Equal(7, arr.Values[2])
	assert.Equal(7, arr.Get(0))
	assert.Equal(8, arr.Get(1))

	// Test end less than start
	arr = NewRollingArray[int](3)
	arr.Add(1)
	arr.Add(2)
	arr.Add(3)
	arr.Add(4)
	arr.Add(5)
	arr.Add(6)
	arr.Add(7)
	arr.Add(8)
	// 7, 8, 6
	assert.Equal(3, arr.GetLength())
	assert.Equal(7, arr.Values[0])
	assert.Equal(8, arr.Values[1])
	assert.Equal(6, arr.Values[2])
	assert.Equal(6, arr.Get(0))
	assert.Equal(7, arr.Get(1))
	assert.Equal(8, arr.Get(2))
}
