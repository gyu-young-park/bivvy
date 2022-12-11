package main

import (
	"fmt"
	"os"
)

func main() {
	dal, _ := newDal("db.db", os.Getpagesize())

	p := dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data")

	_ = dal.writePage(p)
	_, _ = dal.writeFreelist()

	// close the db
	_ = dal.close()

	dal, _ = newDal("db.db", os.Getpagesize())
	p = dal.allocateEmptyPage()
	p.num = dal.getNextPage()
	copy(p.data[:], "data2")
	_ = dal.writePage(p)

	pageNum := dal.getNextPage()
	fmt.Println(pageNum)
	dal.releasedPage(pageNum)

	_, _ = dal.writeFreelist()
}
