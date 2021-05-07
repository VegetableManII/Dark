package darkcache
/*
缓存值的抽象与封装
*/
// 缓存对象的视图
type ByteView struct {
	b []byte
}
// 字节视图的长度
func (v ByteView) Len() int {
	return len(v.b)
}
// 获取字节流的字节格式
func (v ByteView) ByteSlice() []byte {
	return cloneByte(v.b)
}
// 获取字节流的字符串格式
func (v ByteView) String() string {
	return string(v.b)
}
func cloneByte(b []byte) []byte {
	c := make([]byte,len(b))
	copy(c,b)
	return c
}
