package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 设置简单的hash算法方便得知转换后的hash值
	// 产生的虚拟节点为
	// 2——02，12，22
	// 4——04，14，24
	// 6——06，16，26
	hash := New(3, func(key []byte) uint32 {
		i,_ := strconv.Atoi(string(key))
		return uint32(i)
	})
	hash.Add("6","4","2")

	testCases := map[string]string {
		"2":"2",
		"11":"2",
		"23":"4",
		"27":"2",
	}
	for k,v := range testCases{
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s",k,v)
		}
	}
	hash.Add("8")
	// 产生新的虚拟节点
	// 8——08，18，28
	testCases["27"] = "8"

	for k,v := range testCases{
		if hash.Get(k) != v {
			t.Errorf("Asking for %s,should have yielded %s",k,v)
		}
	}
}
