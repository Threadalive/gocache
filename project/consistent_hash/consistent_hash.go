package consistent_hash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//映射到2^32的环上
type Hash func(data []byte) uint32

type Map struct {
	hash    Hash
	replica int
	//存节点keys的哈希值（排序好的），即哈希环
	keys []int
	//存储虚拟节点与其对应的真实节点名称
	hashMap map[int]string
}

//新建一个哈希环结构实例
func New(replica int, f Hash) *Map {
	m := &Map{
		replica: replica,
		hash:    f,
		hashMap: make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) AddNodes(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replica; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

//选取节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	//所查询key对应的hash值
	hash := int(m.hash([]byte(key)))
	//二分查找
	index := sort.Search(len(m.keys), func(i int) bool {
		if m.keys[i] >= hash {
			return true
		} else {
			return false
		}
	})
	//环状，索引取余
	return m.hashMap[m.keys[index%len(m.keys)]]
}
