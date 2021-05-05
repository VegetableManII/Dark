package lru

import (
	"container/list"
	"reflect"
)

// 采用LRU策略进行缓存
type Cache struct {
	maxBytes int64
	nbytes int64
	ll *list.List
	cache map[string]*list.Element
	OnEvicted func(key string,value reflect.Value)
}
type Value interface {

}
type entry struct {
	key string
	value Value
}