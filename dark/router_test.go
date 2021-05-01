package dark

import (
	"fmt"
	"reflect"
	"testing"
)

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/:name", nil)
	r.addRoute("GET", "/assets/*file", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/a/b/c"), []string{"p", "a", "b", "c"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/:a/b/:c"), []string{"p", ":a", "b", ":c"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*a/b/*c"), []string{"p", "*a"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/a/*b/*"), []string{"p", "a", "*b"})
	if !ok {
		t.Fatal("parsePattern error")
	}
}
func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/assets/index.html")
	if n == nil {
		t.Fatal("nil returned")
	}
	if n.pattern != "/assets/*file" {
		t.Fatalf("n.pattern is %s", n.pattern)
	}
	if ps["file"] != "index.html" {
		t.Fatalf("ps['file'] is %s", ps["file"])
	}
	fmt.Printf("matched path %s, params['file']: %s\n", n.pattern, ps["file"])
}
func TestGetRoute2(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/jack")
	if n == nil {
		t.Fatal("nil returned")
	}
	if n.pattern != "/hello/:name" {
		t.Fatalf("n.pattern is %s", n.pattern)
	}
	if ps["name"] != "jack" {
		t.Fatalf("ps['name'] is %s", ps["name"])
	}
	fmt.Printf("matched path %s, params['name']: %s\n", n.pattern, ps["name"])
}
