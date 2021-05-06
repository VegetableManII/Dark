package darkcache

import (
	"fmt"
	"log"
	"sync"
)

/*
负责与外部交互，控制缓存存储和获取的主流程
*/

// 回调函数，当缓存不存在时调用得到源数据
type Getter interface {
	Get(key string) ([]byte,error)
}
// 定义函数类型并实现Get接口，接口型函数
// 调用时既能传入函数作为参数，也能传入实现了该接口的结构体作为参数
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte,error) {
	return f(key)
}
// 一个缓存的命名空间，每一个Group都拥有唯一一个名字
type Group struct {
	name string
	getter Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 并发缓存
}


var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)
// 创建Group实例
func NewGroup(name string,cacheBytes int64,getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:name,
		getter:getter,
		mainCache:cache{cacheBytes:cacheBytes},
	}
	groups[name] = g
	return g
}
// GetGroup 根据名字获得之前创建的Group如果没有返回nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}
// 核心方法Get的实现，两个流程
/*
 接收key ——> 检查是否被缓存 ————> 返回缓存值
                 | 否
                 |——————> 是否从远程节点获取
                                | 否
                                | ————————> 调用回调函数，并取值添加到缓存 ——————> 返回缓存值

*/
func (g *Group) Get(key string) (ByteView,error) {
	if key == "" {
		return ByteView{},fmt.Errorf("key is required")
	}
	if v ,ok := g.mainCache.get(key);ok {
		log.Println("[DarkCache] hit")
		return v,nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView,err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView,error) {
	bytes,err := g.getter.Get(key)
	if err != nil {
		return ByteView{},err
	}
	value := ByteView{b:cloneByte(bytes)}
	g.populateCache(key,value)
	return value,nil
}
func (g *Group) populateCache(key string,value ByteView) {
	g.mainCache.add(key,value)
}
