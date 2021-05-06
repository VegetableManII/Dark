package darkcache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)
var db = map[string]string {
	"Jack":"700",
	"Lucy":"699",
	"Sam":"500",
}
func TestGetterFunc_Get(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte,error) {
		return []byte(key),nil
	})
	expect := []byte("key")
	if v,_ := f.Get("key");!reflect.DeepEqual(v,expect) {
		t.Errorf("callback failed")
	}
}
func TestGet(t *testing.T) {
	loadCounts := make(map[string]int,len(db))
	dark := NewGroup("scores",2<<10,GetterFunc(
		func(key string) ([]byte,error) {
			log.Println("[SlowDB] search key",key)
			if v, ok := db[key];ok {
				if _,ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v),nil
			}
			return nil,fmt.Errorf("%s not exist",key)
		},
		))
	for k,v := range db {
		if view,err := dark.Get(k);err != nil || view.String() != v{
			t.Fatal("failed to get value of Tom")
		} // 通过回调加载数据
		if _ ,err := dark.Get(k);err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss",k)
		} // 缓存命中
	}

	if view , err := dark.Get("unknown");err == nil {
		t.Fatalf("the value of unknow should be empty,but %s got",view)
	}
}