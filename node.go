package main

type Item struct {
	key []byte
	value []byte
}

type Node struct {
	*dal
	pageNum pgnum
	items []*Item
	childNodes []pgnum
}

func NewEmptyNode() *Node {
	return &Node{}
}

func newItem(key []byte, value []byte) *Item {
	return &Item{
		key: key,
		value: value,
	}
}

func (n *Node) isLeaf() bool {
	return len(n.childNodes) == 0
}