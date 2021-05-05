package main

import (
	"fmt"
	"github.com/VegetableManII/DarkWebFrame/dark"
	"html/template"
	"log"
	"net/http"
	"time"
)

func onlyForG2Middle() dark.HandleFunc {
	return func(c *dark.Context) {
		t := time.Now()
		c.Fail(500, "Internal Server Error")
		log.Printf("Middleware-onlyForG2Middle:[%d] %s in %v for RouterGroup2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

type student struct {
	Name string
	Age  int
}

func formatDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}
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
	/*
		r.Get("/hello/:name", func(context *dark.Context) {
			// 请求格式 /hello/jack
			context.String(http.StatusOK, "Hello,%s & Path is %s\n", context.Query("name"), context.Path)
		})
		r.Get("/assets/*file", func(context *dark.Context) {
			// 请求格式 /assets/index.html
			context.JSON(http.StatusOK, dark.H{"file": context.Params["file"]})
		})
	*/
	// 增加路由分组 前缀的格式一定要带根目录 "/"
	g1 := r.Group("/g1")
	{
		g1.GET("/", func(c *dark.Context) {
			c.String(http.StatusOK, "This is RouterGroup1 & Path is %s\n", c.Path)
		})
		g1.GET("/hello", func(c *dark.Context) {
			c.String(http.StatusOK, "Hello, %s. This is RouterGroup1 & Path is %s\n\n",
				c.Query("name"), c.Path)
		})
	}
	g2 := r.Group("/g2")
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
	// 中间件
	r.Use(dark.Logger()) // 全局中间件提供日志打印功能
	// g2.Use(onlyForG2Middle()) // 添加 g2 分组的中间件功能
	r.SetFuncMap(template.FuncMap{
		"formateAsDate": formatDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")
	r.GET("/", func(c *dark.Context) {
		c.HTML(http.StatusOK, "sayhello.tmpl", nil)
	})
	r.GET("/student", func(c *dark.Context) {
		c.HTML(http.StatusOK, "student.tmpl", dark.H{
			"title": "dark frame",
			"stuArr": [2]*student{&student{
				"jack", 18,
			}, &student{
				"lucy", 20,
			}},
		})
	})
	r.GET("/date", func(c *dark.Context) {
		c.HTML(http.StatusOK, "now.tmpl", dark.H{
			"title": "dark frame",
			"now":   time.Date(1999, 8, 11, 0, 0, 0, 0, time.UTC),
		})
	})
	r.Run(":9999")
}
