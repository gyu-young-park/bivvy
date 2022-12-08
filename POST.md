# 1000 라인의 코드로 NOSQL database를 바닥부터 만들어보자.

## Chapter1
앞으로 go언어로 간단한 NOSQL databawe를 만들 것이다. database의 컨셉을 알려주고, go언어로 NOSQL key/value database를 만드는 데 이 컨셉들을 사용할 수 있는 지 알려줄 것이다. 우리는 다음과 같은 질문에 대해 답을 할 것이다.

1. NOSQL이란 무엇인가?
2. disk에 어떻게 데이터를 저장하는가?
3. disk-based와 in-momory database의 차이는 무엇인가??
4. index들은 어떻게 만들어 지는가??
5. ACID란 무엇이고, transaction은 어떻게 동작하는가?
6. 최적의 성능을 위해 디자인된 database란 무엇인가?

첫번째로 우리의 데이터베이스에 사용할 컨셉들의 overview에 대해서 시작해보고, disk에 쓰는 기본적인 매커니즘을 구현하도록 한자.

### SQL vs NOSQL
database는 다른 카테고리들로 나뉘는데, 일반적으로 많이 사용되는 것은 Relational databases (SQL), key-value store, and document store(이를 NOSQL라고 한다.) 이들의 가장 큰 차이는 database에서 사용하는 data model이 무엇이냐는 것이다.


Relational databases에서 비지니스 로직이 database 전체로 확산될 수 있다. 다른 말로, 한 객체의 부분들이 database에 걸쳐 다른 테이블들로 표현될 수 있다. 우리는 '수입'과 '지출'에 대한 다른 테이블을 만들 수 있다는 것이다. 그래서 database에서 '가게'에 대한 전체 entity를 가져오기 위해서는 두 table에 대한 query를 해야할 수 있다.

key-value와 document store는 같은 NOSQL 계열이지만 다르다. 단일 entity의 모든 정보들은 collections/buckets에 모두 집합적으로 저장된다. 가령, '가게'에 대한 정보인 '수입', '지출' 등을 모두 하나의 '가게' 인스턴스에 포함되고, '가게' collection 안에 있다. 

Document stores는 key-value stores의 **subclass**이다. key-value store안의 데이터들은 본질적으로 database에 불투명한 것으로 간주된다. 반면에 document-oriented system은 document’s 내부 구조에 의존한다. 

예를들어, document store에는 내부적인 필드(가령, '수입', '지출')에 의해 모든 '가게' 정보들을 쿼리할 수 있다. 반면에 key-value는 오직 그들의 id에 의해서만 '가게'에 대한 정보를 fetch할 수 있다.

![사진1](./pic/chapter1/1.png)

이것이 가장 기본적인 차이지만, 실제로는 database의 여러가지 타입들이 있다. 

우리의 database는 **key-value** store로 (Not document store)로 실제 구현이 매우 단순하고 직관적이다.

### Disk-Based Storage
database들은 그들의 data(collections, documents...)를 ₩database pages`에 구성한다. Pages는 database와 disk에 의해 교환되는 가장 작은 데이터의 단위이다. 고정된 사이즈를 갖는 것은 매우 편리한 방법이다. 또한 이는 연관된 데이터를 근처에 두도록 하여(in proximity), 데이터들을 한 번에 fetch할 수 있도록 한다.

`Database pages`는 disk seeks(디스크에서 데이터를 찾는 시간)을 최소화하기 위해서 연속적으로 디스크에 저장된다. 만약 8개의 shop들이 있고, 단일 page에 2개의 shop들이 점유된다고 하자. 디스크는 다음과 같이 생길 것이다. 

![사진2](./pic/chapter1/2.png)

`MYSQL`은 기본적으로 16kb의 페이지 사이즈를 갖고 있고, `Postgres SQL`은 8kb page사이즈를 갖고 있다. 

큰 페이지 사이즈를 갖을 수록 더 좋은 성능을 가지지만, 그러나 `torn pages`를 가질 위험도 갖게 된다. `torn pages`란 단일 쓰기 트랜잭션에서 여러 가지 데이터 베이스의 페이지를 사용하는 동안 시스템이 충돌하면서 특정 페이지가 손상되어 가져올 수 없게 되는 것을 말한다. 

때문에 실제로 page 사이즈를 결정할 때, 이러한 사실들을 가지고 많이 고민을 하게 된다. 그러나 이러한 고민들은 우리의 데이터베이스와는 관계가 없다. 때문에 우리는 데이터베이스 page 사이즈를 임의적으로 4kb로 선택하였다.

### Underlying Data Structure
Database는 다양한 데이터 structure를 사용하여 disk에 page들을 구성하는데 대게 **B/B+ tree**를 사용하거나 **hash buckets**를 사용한다. 각 데이터 structure는 여러 가지 장단점을 가지는데 우리는 **B-tree**를 사용할 것이다. 이유는 구현하기 쉽고, **B-tree***의 원칙(principles)이 현실 세계에서 사용되는 데이터베이스와 흡사하기 때문이다.

### Our Database
우리의 데이터베이스는 key-value store이고, data structure 구조는 **B-tree**이다. 또한, 각 page 사이즈는 4KB이다. 이는 다음의 구조를 갖는다.

![사진3](./pic/chapter1/3.png)

**Database**는 우리의 프로그램을 관리하고 transaction들을 오케스트레이션하는 기능의 책임을 진다. 즉, 일련의 read-write 연산을 관리한다는 것이다. 또한, 프로그래머가 database를 사용하는 인터페이스를 제공하고, 그들의 요청을 처리하도록 한다.

**Data Access Layer(DAL)**은 모든 disk와 관련된 연산들을 처리하고, 어떻게 disk에 데이터가 구성되는 지를 처리한다. **DAL**은 기저에 있는 data structure(pages)를 관리하는 책임을 가지며, disk에 database page들을 써주고, fragmentation을 피하기 위해 page를 사용가능한 page를 회수(reclaiming)한다.

우리는 우리의 코드를 bottom-up 방식으로 개발해나갈 것이다. 따라서 Data Access Layer(DAL) 컴포넌트부터 시작을 해보자. `dal.go` 파일을 만들고 다음의 코드를 넣자

- dal.go
```go
package main

import (
	"fmt"
	"os"
)

type dal struct {
	file *os.File
}

func newDal(path string) (*dal, error) {
	dal := &dal{}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	dal.file = file
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
```
`dal`은 디스크에 있는 file과 긴밀히 데이터를 주고 받아야 하기 때문에 file에 대한 포인터를 갖고, 생애주기까지 담당해준다. `dal`을 만들어주었으니, `dal`에 넣을 데이터가 필요하다. 우리는 데이터를 집어넣기 위해 `page`단위를 도입하기로 하였다. 

database page들을 읽고, 쓰는 일은 `DAL`에 의해 관리되어 진다. 따라서 `page` type을 `dal.go`안에 추가해주도록 하자. `page`는 다음과 같은 구조체 형식을 갖는다.

```go
type pgnum uint64

type page struct {
	num  pgnum
	data []byte
}
```
`page`는 하나의 `num`을 갖는데, 이는 unique한 key로 사용된다. 그러나, 그 보다 더 큰 기능을 하는데, 이 숫자를 사용해서 pointer 연산을하여 특정 page에 접근할 수 있기 때문이다. 가령 각 page의 사이즈(우리는 4kb)를 나타내는 `PageSize`를 이용하여 `PageSize * pageNum`연산과 같이 말이다. 추가적으로 `PageSize`을 `page`의 접근을 관리하는 `dal`에 추가해주도록 하자.

```go
type dal struct {
	file     *os.File
	pageSize int
}

func newDal(path string, pageSize int) (*dal, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	dal := &dal{
		file:     file,
		pageSize: pageSize,
	}
	return dal, nil
}
```
이제 `dal`에 read, write연산을 추가해주도록 하자. 

```go
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
```
`allocateEmptyPage`은 `data`에 `PageSize`만큼의 공간을 할당해준다. 이제 `data`부분에 실제로 데이터를 넣어주면 된다. `readPage`는 file에서 특정 데이터를 가져와 page에 넣어준다. 이 떄 사용되는 `ffset := int(pageNum) * d.pageSize`식을 잘보자, 어렵진 않으나 익숙하지 않으면 힘들다. offset만큼 앞을 건너뛰고, p.data만큼(PageSize) 데이터를 가져오는 것이다. 기본적으로 c언어에서 자주 사용하는 파일읽기 방식이다. `writePage`는 page의 `p.num`과 `PageSize`을 곱해서 offset을 만들고 offset만큼 건너뛰고 file에 p.data을 `PageSize`만큼 써주는 것이다.

### Freelist
page를 관리하는 것은 매우 복잡한 일이다. 우리는 어떤 page가 free되었는 지 알아야하고, 어떤 page가 차지되어 있는 지 알아야 한다. 또한 page들은 훗날 데이터를 해제하여 free될 수 있다. 이러한 경우 우리는 freed된 page를 회수(reclaim)해야하며 이는 `fragmentation`을 피하기 위함이다. 

이러한 모든 로직은 `freelist`에서 담당하기로 한다. 해당 component는 `DAL`의 일부분이다. `freelist`는 `maxPage`라는 counter를 가지고 있는데, 이는 여태까지의 할당된 page 중에 가장 높은 숫자를 말한다. 또한, `releasedPages`라는 리스트를 가지는데, 이는 `released page`를 저장하기 위해 있다.

새로운 페이지가 할당되면 `releasedPage`에서 첫번쨰로 free page에 대해 평가된다. 만약 리스트가 비어있다면 counter는 증가하게되고, 새로운 페이지가 주어지면 file size가 증가하게 된다.

![사진3](./pic/chapter1/4.png)

다음과 같이 가장 맨 끝에 저장된 페이지는 `maxPage`로 7번째 page이다. `releasedPages`에 `[5,6]`이 있다면 5,6번째 page는 비어있다는 것이다. 그래서 먼저 page를 할당해주도록 한다. 5,6이 꽉찼다면 7번째 페이지 다음인 8페이지에 데이터를 넣어주면 된다. 그러기 위해서는 7인 `maxPage`와 `PageSize`을 곱한만큼 건너뛰어야 한다.

`freelist.go`을 만들고, `freelist`타입을 만들도록 하자, 그리고 `newFreeList`라는 생성자를 추가해주도록 하자.

- freelist.go
```go
package main

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
```
당연히 `maxPage`의 시작은 0이다. 이를 위해 0을 가지고 있는 `metaPage`를 할당해주도록 하자. 왜 `metaPage`라고 하냐면, 우리의 데이터베이스는 첫번째 페이지를 `metadata`를 저장하기위해 사용할 것이기 때문이다. 이에 대한 것은 다음 장에 자세히 설명한다.

이제 `getNextPage`와 `releasePage`를 추가하자. `getNextPage`은 `freelist`의 `maxPage` counter를 +1증가시켜주는 메서드이다. 다만, `releasedPages`가 있다면 굳이 `maxPage`를 쓸 필요없으니 +1을 해주지 않는다. `releasePage`는 `freelist`의 `releasedPages`에 데이터가 해제된 page를 넣어주는 것이다.

```go
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
```
`freelist`를 만들었으니, 이를 사용할 `dal`의 맴버 변수로 넣어주도록 하자.

```go
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
```
`freelist`를 `dal`에 넣어주었다. 이제 `main.go`를 만들어서 빈 페이지를 하나 할당한 다음 `dal`을 통해 저장과 불러오기를 해보자.

- main.go
```go
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
```
`go run ./...` 명령어로 실행한 후, 만들어진 `db.db` 파일을 dump를 통해서 분석해보자.

```
hexdump -C db.db

00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
*
00001000  64 61 74 61 00 00 00 00  00 00 00 00 00 00 00 00  |data............|
00001010  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
*
00002000
```
`hexdump`명령어에 `-C` flag를 함께 사용하면 이진 파일의 내용들을 보여준다.

첫번째 줄은 0번째 page로 추후에 있을 metadata 영역이다.

우리의 database가 어느정도 프로그램의 구실을 하게되었는지만 치명적인 문제가 있다. 만약 프로그램을 종료하면 페이지를 어디서부터 할당해야하고, 어떤 페이지들이  free되었는 지 기록할 수 없다는 것이다. 즉, `freelist`는 `disk`에 저장되어야 한다는 것이다. 이것이 바로 다음 챕터에서 다룰 내용이다. 