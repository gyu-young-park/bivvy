**# 1000 라인의 코드로 NOSQL database를 바닥부터 만들어보자.

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