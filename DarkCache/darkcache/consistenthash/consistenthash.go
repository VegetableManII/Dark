package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/*
一致性Hash均摊各个节点上的缓存数据量
*/
type Hash func(data []byte) uint32

type Map struct {
	hash Hash
	replicas int  // 虚拟节点个数
	keys []int  // 虚拟节点和真实节点的映射表
	hashMap map[int]string  // 虚拟节点及其名称
}

func New(replicas int,fn Hash) *Map {
	m := &Map{
		replicas:replicas,
		hash:fn,
		hashMap:make(map[int]string),
	}
	if m.hash == nil {
		// 设置默认概要算法
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _,key := range keys {
		for i :=0;i<m.replicas ;i++  {
			hash := int(m.hash([]byte(strconv.Itoa(i)+key)))
			m.keys = append(m.keys,hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 二分查找,顺时针方向最接近的虚拟节点
	// 找不到大于hash的节点的情况下返回数组长度
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}