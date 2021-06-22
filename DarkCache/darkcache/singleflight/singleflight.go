package singleflight

import "sync"
/*
防止缓存击穿策略
首次次请求模式
*/
type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m map[string]*call
}


func (g *Group) Do(key string,fn func()(interface{},error)) (interface{},error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	/*
		除了首次请求，请他请求都会被wg.Wait阻塞等待
		首次请求的返回结果
	*/
	if c,ok := g.m[key];ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val,c.err
	}
	/*
		首次请求会初始化相应的结构
		通过wg来实现并发控制
		首次请求去执行查询数据库其他请求等待结果
	*/
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val,c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m,key)
	g.mu.Unlock()

	return c.val,c.err
}
