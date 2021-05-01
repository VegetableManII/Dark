package main

import (
	"github.com/VegetableManII/myHTTPEngine/dark"
	"net/http"
)

func main() {
	r := dark.New()
	r.Get("/", func(context *dark.Context) {
		context.HTML(http.StatusOK,"<h1>Dark Web Engine<h1>")
	})
	r.Get("/hello", func(context *dark.Context) {
		// 请求格式 /hello？name=xxx
		context.String(http.StatusOK,"hello %s\nPath %s\n",context.Query("name"),context.Path)
	})
	r.POST("/login", func(c *dark.Context) {
		c.JSON(http.StatusOK,dark.H{
			"username": c.PostFrom("username"),
			"password":c.PostFrom("password"),
		})
	})
	r.Run(":9999")
}