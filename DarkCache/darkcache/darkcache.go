package darkcache

import (
	pb "darkcache/darkcachepb"
	"darkcache/singleflight"
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
	peers PeerPicker
	loader *singleflight.Group
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
		loader:&singleflight.Group{},
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
                 |——————> 是否从远程节点获取 ————————> 使用一致性哈希选择节点
                                | 否                   是否是远程节点 ———————————> HTTP客户端访问 ——————> 成功？——————> 服务端返回数据
                                |                           | 否                                       | 否
								|                           | ———————————————————————————————> 回退到本地节点处理
								|
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
	/*
		有当前请求数量决定，每一个key只被获取一次，可以从本地或远程获取
	*/
	viewi,err := g.loader.Do(key, func() (i interface{}, e error) {
		if g.peers != nil {
			if peer,ok := g.peers.PickPeer(key);ok {
				if value,err = g.getFromPeer(peer,key);err == nil {
					return value,nil
				}
				log.Println("[DarkCache] Failed to get from peer",err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return viewi.(ByteView),nil
	}
	return
}
// 从本地获取
func (g *Group) getLocally(key string) (ByteView,error) {
	bytes,err := g.getter.Get(key)
	if err != nil {
		return ByteView{},err
	}
	value := ByteView{b:cloneByte(bytes)}
	g.populateCache(key,value)
	return value,nil
}
func (g *Group) getFromPeer(peer PeerGetter,key string) (ByteView,error) {
	req := &pb.Request{
		Group: g.name,
		Key: key,
	}
	res := &pb.Response{}
	err := peer.Get(req,res)
	if err != nil {
		return ByteView{},err
	}
	return ByteView{b:res.Value},nil
}

func (g *Group) populateCache(key string,value ByteView) {
	g.mainCache.add(key,value)
}
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
