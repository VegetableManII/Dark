package main

import (
	"darkcache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string {
	"Jack":"750",
	"Lucy":"550",
	"Tom":"660",
}

func createGroup() *darkcache.Group {
	// 创建一组缓存保存scores数据，设置回调函数数据从db中读取数据
	return darkcache.NewGroup("scores",2<<10, darkcache.GetterFunc(
		func(key string) ([]byte,error) {
			log.Println("[SlowDB] search key",key)
			if v,ok := db[key];ok {
				return []byte(v),nil
			}
			return nil,fmt.Errorf("%s not exist",key)
		}))
}
// 创建缓存服务器创建HTTPPool，添加节点信息注册到引擎中，启动http服务
func startCacheServer(addr string,addrs []string,dark *darkcache.Group) {
	peers := darkcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	dark.RegisterPeers(peers)
	log.Println("darkcache is running at",addr)
	log.Fatal(http.ListenAndServe(addr[7:],peers))
}
// 启动一个API服务，与用户交互
func startAPIServer(apiAddr string,dark *darkcache.Group) {
	http.Handle("/api",http.HandlerFunc(
		func(w http.ResponseWriter,r *http.Request) {
			key := r.URL.Query().Get("key")
			view,err := dark.Get(key)
			if err != nil {
				http.Error(w,err.Error(),http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type","application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at",apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:],nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port,"port",8001,"DarkCache server port")
	flag.BoolVar(&api,"api",false,"Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string {
		8001:"http://localhost:8001",
		8002:"http://localhost:8002",
		8003:"http://localhost:8003",
	}

	var addrs []string
	for _,v := range addrMap{
		addrs = append(addrs,v)
	}

	dark := createGroup()
	if api {
		go startAPIServer(apiAddr,dark)
	}

	startCacheServer(addrMap[port],[]string(addrs),dark)
}