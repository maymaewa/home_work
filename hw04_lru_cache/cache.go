package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value any) bool
	Get(key Key) (any, bool)
	Clear()
}

type cacheItem struct {
	key   Key
	value any
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (c *lruCache) Set(key Key, value any) bool {
	if c.capacity == 0 {
		return false
	}

	// Элемент уже есть в кэше
	if item, ok := c.items[key]; ok {
		item.Value = cacheItem{
			key:   key,
			value: value,
		}
		c.queue.MoveToFront(item)

		return true
	}

	// Если кэш заполнен — удаляем самый давно использовавшийся элемент
	if c.queue.Len() >= c.capacity {
		back := c.queue.Back()
		if back != nil {
			data := back.Value.(cacheItem)

			c.queue.Remove(back)
			delete(c.items, data.key)
		}
	}

	item := c.queue.PushFront(cacheItem{
		key:   key,
		value: value,
	})
	c.items[key] = item

	return false
}

func (c *lruCache) Get(key Key) (any, bool) {
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	c.queue.MoveToFront(item)

	return item.Value.(cacheItem).value, true
}

func (c *lruCache) Clear() {
	c.queue = NewList()
	c.items = make(map[Key]*ListItem, c.capacity)
}
