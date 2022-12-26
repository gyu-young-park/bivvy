package main

import "encoding/binary"

type Item struct {
	key   []byte
	value []byte
}

type Node struct {
	*dal
	pageNum    pgnum
	items      []*Item
	childNodes []pgnum
}

func NewEmptyNode() *Node {
	return &Node{}
}

func newItem(key []byte, value []byte) *Item {
	return &Item{
		key:   key,
		value: value,
	}
}

func (n *Node) isLeaf() bool {
	return len(n.childNodes) == 0
}

func (n *Node) serialize(buf []byte) []byte {
	leftPos := 0
	rightPos := len(buf) - 1

	// Add page header: isLeaf, key-value pairs count, node num
	isLeaf := n.isLeaf()
	var bitSetVar uint64
	if isLeaf {
		bitSetVar = 1
	}
	buf[leftPos] = byte(bitSetVar)
	leftPos += 1

	binary.LittleEndian.PutUint16(buf[leftPos:], uint16(len(n.items)))
	leftPos += 2

	// We use slotted pages for storing data in the page. It means the actual keys and values (the cells) are appended
	// to right of the page whereas offsets have a fixed size and are appended from the left.
	// It's easier to preserve the logical order (alphabetical in the case of b-tree) using the metadata and performing
	// pointer arithmetic. Using the data itself is harder as it varies by size.

	// Page structure is:
	// ----------------------------------------------------------------------------------
	// |  Page  | key-value /  child node    key-value 		      |    key-value		 |
	// | Header |   offset /	 pointer	  offset         .... |      data      ..... |
	// ----------------------------------------------------------------------------------
	for i := 0; i < len(n.items); i++ {
		item := n.items[i]
		if !isLeaf {
			childNode := n.childNodes[i]

			// Write the child page as a fixed size of 8 bytes
			binary.LittleEndian.PutUint64(buf[leftPos:], uint64(childNode))
			leftPos += pageNumSize
		}

		klen := len(item.key)
		vlen := len(item.value)

		//write offset
		offset := rightPos - klen - vlen - 2
		binary.LittleEndian.PutUint16(buf[leftPos:], uint16(offset))
		leftPos += 2

		rightPos -= vlen
		copy(buf[rightPos:], item.value)

		rightPos -= 1
		buf[rightPos] = byte(vlen)

		rightPos -= klen
		copy(buf[rightPos:], item.key)

		rightPos -= 1
		buf[rightPos] = byte(klen)
	}

	if !isLeaf {
		lastChildNode := n.childNodes[len(n.childNodes)-1]
		binary.LittleEndian.PutUint64(buf[leftPos:], uint64(lastChildNode))
	}
	return buf
}

func (n *Node) deserialize(buf []byte) {
	leftPos := 0

	// read header
	isLeaf := uint16(buf[0])

	itemsCount := int(binary.LittleEndian.Uint16(buf[1:3]))
	leftPos += 3

	for i := 0; i < itemsCount; i++ {
		if isLeaf == 0 {
			pageNum := binary.LittleEndian.Uint64(buf[leftPos:])
			leftPos += pageNumSize

			n.childNodes = append(n.childNodes, pgnum(pageNum))
		}

		// read offset
		offset := binary.LittleEndian.Uint16(buf[leftPos:])
		leftPos += 2

		klen := uint16(buf[int(offset)])
		offset += 1

		key := buf[offset : offset+klen]
		offset += klen

		vlen := uint16(buf[int(offset)])
		offset += 1

		value := buf[offset : offset+vlen]
		offset += vlen

		n.items = append(n.items, newItem(key, value))
	}

	if isLeaf == 0 {
		pageNum := pgnum(binary.LittleEndian.Uint64(buf[leftPos:]))
		n.childNodes = append(n.childNodes, pageNum)
	}
}
