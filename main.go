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

	rd, _ := dal.readPage(p.num)
	fmt.Println(string(rd.data))
}
