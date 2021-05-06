package darkrpc

import (
	"darkrpc/codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// RPC Server
type Server struct{}

// NewServer 创建一个Server
func NewServer() *Server {
	return &Server{}
}

// DefaultServer Server的默认实例
var DefaultServer = NewServer()

func (sev *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go sev.ServeConn(conn)
	}
}
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (sev *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
	}
	sev.serveCodec(f(conn))
}

var invalidRequest = struct{}{}

type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
}

func (sev Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read head error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (sev *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := sev.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	// TODO
	// 请求参数类型暂时规定为string
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}
	return req, nil
}

func (sev *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response err:", err)
	}
}

func (sev *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// TODO 调用注册的rpc方法去处理参数正确响应
	// 暂时先打印参数然后返回一个hello
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("darkrpc resp %d", req.h.Seq))
	sev.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func (sev *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // 确保能发送一个完整的响应
	wg := new(sync.WaitGroup)
	for {
		req, err := sev.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			sev.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go sev.handleRequest(cc, req, sending, wg)

	}
	wg.Wait()
	_ = cc.Close()
}
