package main

import (
	"fmt"
	"os"
)

type pgnum uint64

type page struct {
	num  pgnum
	data []byte
}

type dal struct {
	file     *os.File
	pageSize int
	*freelist
}

func newDal(path string, pageSize int) (*dal, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	dal := &dal{
		file:     file,
		pageSize: pageSize,
		freelist: newFreelist(),
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
