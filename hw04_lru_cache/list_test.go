package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
}

func TestPushFront(t *testing.T) {
	l := NewList()

	first := l.PushFront(1)
	require.Equal(t, 1, l.Len())
	require.Equal(t, first, l.Front())
	require.Equal(t, first, l.Back())
	require.Nil(t, first.Prev)
	require.Nil(t, first.Next)

	second := l.PushFront(2)

	require.Equal(t, 2, l.Len())
	require.Equal(t, second, l.Front())
	require.Equal(t, first, l.Back())

	require.Nil(t, second.Prev)
	require.Equal(t, first, second.Next)

	require.Equal(t, second, first.Prev)
	require.Nil(t, first.Next)
}

func TestPushBack(t *testing.T) {
	l := NewList()

	first := l.PushBack(1)
	second := l.PushBack(2)

	require.Equal(t, 2, l.Len())
	require.Equal(t, first, l.Front())
	require.Equal(t, second, l.Back())

	require.Nil(t, first.Prev)
	require.Equal(t, second, first.Next)

	require.Equal(t, first, second.Prev)
	require.Nil(t, second.Next)
}

func TestRemoveFront(t *testing.T) {
	l := NewList()

	first := l.PushBack(1)
	second := l.PushBack(2)
	third := l.PushBack(3)

	l.Remove(first)

	require.Equal(t, 2, l.Len())
	require.Equal(t, second, l.Front())
	require.Equal(t, third, l.Back())

	require.Nil(t, second.Prev)
	require.Equal(t, third, second.Next)
}

func TestRemoveSingleElement(t *testing.T) {
	l := NewList()

	item := l.PushBack(1)
	l.Remove(item)

	require.Equal(t, 0, l.Len())
	require.Nil(t, l.Front())
	require.Nil(t, l.Back())
}

func TestMoveToFrontMiddle(t *testing.T) {
	l := NewList()

	first := l.PushBack(1)
	second := l.PushBack(2)
	third := l.PushBack(3)

	l.MoveToFront(second)

	require.Equal(t, second, l.Front())
	require.Equal(t, third, l.Back())

	values := []int{}
	for i := l.Front(); i != nil; i = i.Next {
		values = append(values, i.Value.(int))
	}

	require.Equal(t, []int{2, 1, 3}, values)

	require.Nil(t, second.Prev)
	require.Equal(t, first, second.Next)
	require.Equal(t, second, first.Prev)
	require.Equal(t, third, first.Next)
	require.Equal(t, first, third.Prev)
}

func TestMoveToFrontBack(t *testing.T) {
	l := NewList()

	l.PushBack(1)
	l.PushBack(2)
	third := l.PushBack(3)

	l.MoveToFront(third)

	values := []int{}
	for i := l.Front(); i != nil; i = i.Next {
		values = append(values, i.Value.(int))
	}

	require.Equal(t, []int{3, 1, 2}, values)
}

func TestMoveToFrontSingleElement(t *testing.T) {
	l := NewList()

	item := l.PushBack(1)

	l.MoveToFront(item)

	require.Equal(t, 1, l.Len())
	require.Equal(t, item, l.Front())
	require.Equal(t, item, l.Back())
	require.Nil(t, item.Prev)
	require.Nil(t, item.Next)
}

func TestRemoveNil(t *testing.T) {
	l := NewList()

	l.Remove(nil)

	require.Equal(t, 0, l.Len())
	require.Nil(t, l.Front())
	require.Nil(t, l.Back())
}

func TestMoveToFrontNil(t *testing.T) {
	l := NewList()

	l.PushBack(1)
	l.PushBack(2)

	l.MoveToFront(nil)

	values := []int{}
	for i := l.Front(); i != nil; i = i.Next {
		values = append(values, i.Value.(int))
	}

	require.Equal(t, []int{1, 2}, values)
}
