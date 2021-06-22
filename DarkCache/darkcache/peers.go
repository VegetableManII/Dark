package darkcache

import pb "darkcache/darkcachepb"

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter,ok bool)
	// 根据传入的key值选择相应节点 PeerGetter
}

// 相当于客户端，通过Get方法去远程访问获取数据
type PeerGetter interface {
	Get(in *pb.Request,out *pb.Response ) (error)
	// 从对应group查找缓存值
}