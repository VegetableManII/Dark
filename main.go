package main

import (
	"github.com/VegetableManII/myHTTPEngine/dark"
	"net/http"
)

func main() {
	r := dark.New()
	// v0
	r.Get("/", func(context *dark.Context) {
		context.HTML(http.StatusOK, "<h1>Dark Web Engine<h1>\n")
	})
	r.Get("/hello", func(context *dark.Context) {
		// 请求格式 /hello？name=xxx
		context.String(http.StatusOK, "hello %s\nPath %s\n", context.Query("name"), context.Path)
	})
	r.POST("/login", func(c *dark.Context) {
		c.JSON(http.StatusOK, dark.H{
			"username": c.PostFrom("username"),
			"password": c.PostFrom("password"),
		})
	})
	// 增加前缀路由及路由参数
	r.Get("/hello/:name", func(context *dark.Context) {
		// 请求格式 /hello/jack
		context.String(http.StatusOK, "hello %s\nPath %s\n", context.Query("name"), context.Path)
	})
	r.Get("/assets/*file", func(context *dark.Context) {
		// 请求格式 /assets/index.html
		context.JSON(http.StatusOK, dark.H{"file": context.Params["file"]})
	})

	r.Run(":9999")
}
