package main

import (
	"errors"
	"fmt"
	"os"
)

type pgnum uint64

type page struct {
	num  pgnum
	data []byte
}

type Options struct {
	pageSize       int
	MinFillPercent float32
	MaxFillPercent float32
}

var DefaultOptions = &Options{
	MinFillPercent: 0.5,
	MaxFillPercent: 0.95,
}

type dal struct {
	file           *os.File
	pageSize       int
	minFillPercent float32
	maxFillPercent float32
	*freelist
	*meta
}

func newDal(path string, options *Options) (*dal, error) {
	dal := &dal{
		meta:           newEmptyMeta(),
		pageSize:       options.pageSize,
		minFillPercent: options.MinFillPercent,
		maxFillPercent: options.MaxFillPercent,
	}
	// exist
	if _, err := os.Stat(path); err == nil {
		dal.file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			_ = dal.close()
			return nil, err
		}

		meta, err := dal.readMeta()
		if err != nil {
			return nil, err
		}
		dal.meta = meta

		freelist, err := dal.readFreelist()
		if err != nil {
			return nil, err
		}
		dal.freelist = freelist
		// dosen't exist
	} else if errors.Is(err, os.ErrNotExist) {
		// init freelist
		dal.file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			_ = dal.close()
			return nil, err
		}
		dal.freelist = newFreelist()
		dal.freelistPage = dal.getNextPage()
		_, err := dal.writeFreelist()
		if err != nil {
			return nil, err
		}

		_, err = dal.writeMeta(dal.meta)
	} else {
		return nil, err
	}
	return dal, nil
}

func (d *dal) close() error {
	if d.file != nil {
		err := d.file.Close()
		if err != nil {
			return fmt.Errorf("could not close file: %s", err)
		}
		d.file = nil
	}
	return nil
}

func (d *dal) allocateEmptyPage() *page {
	return &page{
		data: make([]byte, d.pageSize),
	}
}

func (d *dal) readPage(pageNum pgnum) (*page, error) {
	p := d.allocateEmptyPage()

	// page number와 page size를 이용하여 정확하게 해당 페이지의 데이터를 가져온다.
	offset := int(pageNum) * d.pageSize

	//d.file의 데이터에서 p.data의 사이즈만큼, offset에서부터 데이터를 가져온다.
	_, err := d.file.ReadAt(p.data, int64(offset))
	if err != nil {
		return nil, err
	}
	return p, err
}

func (d *dal) writePage(p *page) error {
	// offset만큼 앞을 건너뛴다.
	offset := int64(p.num) * int64(d.pageSize)
	// offset+p.data의 사이즈만큼, 데이터를 써준다.
	_, err := d.file.WriteAt(p.data, offset)
	return err
}

func (d *dal) writeMeta(meta *meta) (*page, error) {
	p := d.allocateEmptyPage()
	p.num = metaPageNum
	meta.serialize(p.data)

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (d *dal) readMeta() (*meta, error) {
	p, err := d.readPage(metaPageNum)
	if err != nil {
		return nil, err
	}
	meta := newEmptyMeta()
	meta.deserialize(p.data)
	return meta, nil
}

func (d *dal) readFreelist() (*freelist, error) {
	p, err := d.readPage(d.freelistPage)
	if err != nil {
		return nil, err
	}

	freelist := newFreelist()
	freelist.deserialize(p.data)
	return freelist, nil
}

func (d *dal) writeFreelist() (*page, error) {
	p := d.allocateEmptyPage()
	p.num = d.freelistPage
	d.freelist.serialize(p.data)

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}

	d.freelistPage = p.num
	return p, nil
}

func (d *dal) getNode(pageNum pgnum) (*Node, error) {
	p, err := d.readPage(pageNum)
	if err != nil {
		return nil, err
	}

	node := NewEmptyNode()
	node.deserialize(p.data)
	node.pageNum = pageNum
	return node, nil
}

func (d *dal) writeNode(n *Node) (*Node, error) {
	p := d.allocateEmptyPage()
	if n.pageNum == 0 {
		p.num = d.getNextPage()
		n.pageNum = p.num
	} else {
		p.num = n.pageNum
	}

	p.data = n.serialize(p.data)

	err := d.writePage(p)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (d *dal) deleteNode(pageNum pgnum) {
	d.releasedPage(pageNum)
}
