package lru

type LRUCache[V any] struct {
	Head *Node[V]
	Tail *Node[V]
	HT map[string]*Node[V]
	onEvict func(V)
	Cap int
}

type Node[V any] struct {
	Key string
	Val V
	Prev *Node[V]
	Next *Node[V]
}

func New[V any](capacity int, onEvict func(V)) LRUCache[V] {
	return LRUCache[V]{
		HT: make(map[string]*Node[V]),
		Cap: capacity,
		onEvict: onEvict,
	}
}

func (c *LRUCache[V]) Get(key string) (V, bool) {
	node, ok := c.HT[key]
	if ok {
		c.Remove(node)
		c.Add(node)
		return node.Val, true
	}
	var zero V
	return zero, false
}

func (c *LRUCache[V]) Set(key string, value V)  {
	node, ok := c.HT[key]
	if ok {
		node.Val = value
		c.Remove(node)
		c.Add(node)
		return
	} else {
		node = &Node[V]{Key: key, Val: value}
		c.HT[key] = node
		c.Add(node)
	}
	if len(c.HT) > c.Cap {
		c.onEvict(c.Tail.Val)
		delete(c.HT, c.Tail.Key)
		c.Remove(c.Tail)
	}
}

func (c *LRUCache[V]) Add(node *Node[V]) {
	node.Prev = nil
	node.Next = c.Head
	if c.Head != nil {
		c.Head.Prev = node
	}
	c.Head = node
	if c.Tail == nil {
		c.Tail = node
	}
}

func (c *LRUCache[V]) Remove(node *Node[V]) {
	if node != c.Head {
		node.Prev.Next = node.Next
	} else {
		c.Head = node.Next
	}
	if node != c.Tail {
		node.Next.Prev = node.Prev
	} else {
		c.Tail = node.Prev
	}
}
