package darkcache

import (
	"darkcache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	pb "darkcache/darkcachepb"
	"github.com/golang/protobuf/proto"
)

/*
提供被其他节点访问的能力(基于http)
*/

const  (
	defaultBasePath  = "/_darkcache/"
	defaultReplicas = 50
	)

type HTTPPool struct {
	// 本地节点的URL地址，例如 http://darkcache:8000
	self string
	basePath string
	mu sync.Mutex
	peers *consistenthash.Map
	httpGetters map[string]*httpGetter // 每一个远程节点都与一个Getter对应
}
// 初始化本节点的HTTP协议的连接
// 节点间访问请求为 http://darkcache:8000/_darkcache/
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:self,
		basePath:defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string,v ...interface{}) {
	log.Printf("[Server %s] %s",p.self,fmt.Sprintf(format,v...))
}
// ServeHTTP 核心方法的实现
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter,r *http.Request) {
	if !strings.HasPrefix(r.URL.Path,p.basePath) {
		panic("HTTPPool serving unexpected path: "+ r.URL.Path)
	}
	p.Log("%s %s",r.Method,r.URL.Path)
	// <basepath>/<groupname>|<key> 字符串截取获的groupname和key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):],"/",2)
	if len(parts) != 2 {
		http.Error(w,"bad request",http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w,"no sunch group: "+groupName,http.StatusBadRequest)
		return
	}

	view,err := group.Get(key)
	if err != nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	body,err := proto.Marshal(&pb.Response{Value:view.ByteSlice()})
	if err !=nil {
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type","application/octet-stream")
	w.Write(body)
}

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) (error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
		)
	res,err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returnned：%v",res.Status)
	}

	bytes,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body：%v",err)
	}
	if err = proto.Unmarshal(bytes,out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

var _ PeerGetter = (*httpGetter)(nil)
// 实例化一个一致性hash算法并添加传入的节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas,nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter,len(peers))
	for _,peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL:peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter,bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key);peer != "" && peer != p.self {
		p.Log("Pick peer %s",peer)
		return p.httpGetters[peer],true
	}
	return nil,false
}