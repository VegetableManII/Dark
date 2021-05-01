package main

import (
	"github.com/VegetableManII/myHTTPEngine/dark"
	"net/http"
)

func main() {
	r := dark.New()
	// v0
	/*
		r.Get("/", func(context *dark.Context) {
			context.HTML(http.StatusOK, "<h1>Dark Web Engine<h1>\n")
		})
		r.Get("/hello", func(context *dark.Context) {
			// 请求格式 /hello？name=xxx
			context.String(http.StatusOK, "hello %s\nPath %s\n", context.Query("name"), context.Path)
		})
		r.POST("/login", func(c *dark.Context) {
			c.JSON(http.StatusOK, dark.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	*/
	// 增加前缀路由及路由参数
	r.Get("/hello/:name", func(context *dark.Context) {
		// 请求格式 /hello/jack
		context.String(http.StatusOK, "Hello,%s & Path is %s\n", context.Query("name"), context.Path)
	})
	r.Get("/assets/*file", func(context *dark.Context) {
		// 请求格式 /assets/index.html
		context.JSON(http.StatusOK, dark.H{"file": context.Params["file"]})
	})
	// 增加路由分组
	g1 := r.Group("g1")
	{
		g1.GET("/", func(c *dark.Context) {
			c.String(http.StatusOK, "This is RouterGroup1 & Path is %s\n", c.Path)
		})
		g1.GET("/hello", func(c *dark.Context) {
			c.String(http.StatusOK, "Hello, %s. This is RouterGroup1 & Path is %s\n\n",
				c.Query("name"), c.Path)
		})
	}
	g2 := r.Group("g2")
	{
		g2.GET("/hi/:name", func(c *dark.Context) {
			c.String(http.StatusOK, "Hi，%s. This is RouterGroup2 & Path is %s\n\n",
				c.Param("name"), c.Path)
		})
		g2.POST("/login", func(c *dark.Context) {
			c.JSON(http.StatusOK, dark.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
				"path":     c.Path,
			})
		})
	}
	r.Run(":9999")
}
