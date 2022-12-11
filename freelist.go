package main

import "encoding/binary"

const metaPage = 0

type freelist struct {
	maxPage       pgnum   // 할당된 페이지의 가장 max번째를 기록한다. 따라서 maxPage*PageSize = fileSize이다.
	releasedPages []pgnum // 이전에는 할당되었지만 지금은 free된 페이지의 번호를 기록한다.
}

func newFreelist() *freelist {
	return &freelist{
		maxPage:       metaPage,
		releasedPages: []pgnum{},
	}
}

func (fr *freelist) getNextPage() pgnum {
	// 먼저 releasedPages로 부터 페이지를 가져온다.
	// 그렇지 않으면, maxium page을 늘려준다.
	if len(fr.releasedPages) != 0 {
		pageID := fr.releasedPages[len(fr.releasedPages)-1]
		fr.releasedPages = fr.releasedPages[:len(fr.releasedPages)-1]
		return pageID
	}
	fr.maxPage += 1
	return fr.maxPage
}

func (fr *freelist) releasedPage(page pgnum) {
	fr.releasedPages = append(fr.releasedPages, page)
}

func (fr *freelist) serialize(buf []byte) []byte {
	pos := 0
	binary.LittleEndian.PutUint16(buf[pos:], uint16(fr.maxPage))
	pos += 2

	// released page 개수
	binary.LittleEndian.PutUint16(buf[pos:], uint16(len(fr.releasedPages)))
	pos += 2

	for _, page := range fr.releasedPages {
		binary.LittleEndian.PutUint64(buf[pos:], uint64(page))
		pos += pageNumSize
	}
	return buf
}

func (fr *freelist) deserialize(buf []byte) {
	pos := 0
	fr.maxPage = pgnum(binary.LittleEndian.Uint16(buf[pos:]))
	pos += 2

	// released pages count
	releasedPagesCount := int(binary.LittleEndian.Uint16(buf[pos:]))
	pos += 2

	for i := 0; i < releasedPagesCount; i++ {
		fr.releasedPages = append(fr.releasedPages, pgnum(binary.LittleEndian.Uint64(buf[pos:])))
		pos += pageNumSize
	}
}
