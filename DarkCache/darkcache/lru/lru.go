package lru
/*
lru缓存淘汰策略
*/
import (
	"container/list"
)

// 采用LRU策略进行缓存，非并发安全
type Cache struct {
	maxBytes int64  // 允许使用最大内存
	nbytes int64   // 已经使用的内存
	ll *list.List  //标准库双向链表
	cache map[string]*list.Element
	OnEvicted func(key string,value Value)  // 某条记录移除时回调函数
}
// Len方法返回占用空间的大小以字节为单位
type Value interface {
	Len() int
}
// 双向链表中的节点
type entry struct {
	key string
	value Value
}

func New(maxBytes int64,onEvicted func(string,Value)) *Cache {
	return &Cache{
		maxBytes:maxBytes,
		ll:list.New(),
		cache:make(map[string]*list.Element,0),
		OnEvicted:onEvicted,
	}
}
func (c *Cache) Len() int {
	return c.ll.Len()
}
// 查找
func (c *Cache) Get(key string) (value Value,ok bool) {
	if ele,ok := c.cache[key];ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry) // 类型断言
		return kv.value,true
	}
	return
}
// 删除
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)  // 获取的是指针不是具体的值
		delete(c.cache,kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key,kv.value)
		}
	}
}
// 新增和修改
func (c *Cache) Add(key string,value Value) {
	if ele,ok := c.cache[key];ok {
		// 修改
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes+=int64(value.Len()) - int64(kv.value.Len())
	} else {
		// 新增
		ele := c.ll.PushFront(&entry{key:key,value:value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes !=0 && c.maxBytes < c.nbytes  {
		c.RemoveOldest()
	}
}