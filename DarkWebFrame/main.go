package main

import (
	"fmt"
	"github.com/VegetableManII/DarkWebFrame/dark"
	"html/template"
	"log"
	"net/http"
	"time"
)

// 基础实现context
func easy(r *dark.Engine) {
	r.Get("/easy", func(context *dark.Context) {
		// 请求格式 /easy?name=xxx
		context.String(http.StatusOK, "hello %s\nPath %s\n", context.Query("name"), context.Path)
	})
	r.POST("/easy", func(c *dark.Context) {
		c.JSON(http.StatusOK, dark.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
}

// 增加前缀路由及路由参数
func addRouteWithParam(r *dark.Engine) {
	r.Get("/route1/:name", func(context *dark.Context) {
		// 请求格式 /route1/jack
		log.Println(context.Query("name"))
		context.String(http.StatusOK, "Hello %s & Path is %s\n", context.Param("name"), context.Path)
	})
	r.Get("/route2/*file", func(context *dark.Context) {
		// 请求格式 /route2/index.tmpl
		context.JSON(http.StatusOK, dark.H{"file": context.Params["file"]})
	})
}

// 增加路由分组 前缀的格式一定要带根目录 "/" 否则创建的路径为 localhost:9999g2
func addRouterGroup(r *dark.Engine) []*dark.RouterGroup {
	groups := make([]*dark.RouterGroup, 0, 4)
	g1 := r.Group("/g1")
	{
		g1.GET("/", func(c *dark.Context) {
			c.String(http.StatusOK, "当前所在路由为 Group-1 %s\n", c.Path)
		})
		g1.GET("/hello", func(c *dark.Context) {
			c.String(http.StatusOK, "你好，%s，当前所在路由分组为 Group-1 %s\n\n",
				c.Query("name"), c.Path)
		})
	}
	g2 := r.Group("/g2")
	{
		g2.GET("/hi/:name", func(c *dark.Context) {
			log.Println(c.Param("name"))
			c.String(http.StatusOK, "你好，%s，当前所在路由分组为 Group-2 %s\n\n",
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
	groups = append(groups, g1)
	groups = append(groups, g2)
	return groups
}

// 中间件
func onlyForG2Middle() dark.HandleFunc {
	return func(c *dark.Context) {
		t := time.Now()
		// 停止g2的所有服务
		c.Fail(500, "Internal Server Error")
		log.Printf("Middleware-onlyForG2Middle:[%d] %s in %v for RouterGroup2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func addMiddleware(r *dark.Engine, group *dark.RouterGroup) {
	r.Use(dark.Logger())         // 全局中间件提供日志打印功能
	group.Use(onlyForG2Middle()) // 添加 g2 分组的中间件功能
}

// 设置模板
// student.tmpl 中使用
type student struct {
	Name string
	Age  int
}

// now.tmpl 中使用
// FuncMap 模板函数
func formatDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}
func addTemplate(r *dark.Engine) {
	r.SetFuncMap(template.FuncMap{
		"formateAsDate": formatDate,
	})
	r.LoadHTMLGlob("templates/*")

	r.GET("/template/", func(c *dark.Context) {
		c.HTML(http.StatusOK, "sayhello.tmpl", nil)
	})
	r.GET("/template/student", func(c *dark.Context) {
		c.HTML(http.StatusOK, "student.tmpl", dark.H{
			"title": "dark frame",
			"stuArr": [2]*student{&student{
				"jack", 18,
			}, &student{
				"lucy", 20,
			}},
		})
	})
	r.GET("/template/date", func(c *dark.Context) {
		c.HTML(http.StatusOK, "now.tmpl", dark.H{
			"title": "dark frame",
			"now":   time.Date(1999, 8, 11, 0, 0, 0, 0, time.UTC),
		})
	})
}

// 设置本地静态文件的路径
func addStaticFilePath(r *dark.Engine) {
	r.Static("/assets", "./static")
}
func main() {
	r := dark.New()

	r.Use(dark.Logger())
	r.LoadHTMLGlob("/Users/jack/Documents/作业/web/实验四/*.tmpl")
	r.Static("/questionnaire","/Users/jack/Documents/作业/web/实验四/")
	r.POST("questionnaire/result", func(c *dark.Context) {
		c.HTML(http.StatusOK,"result.tmpl",dark.H{
			"Q1":c.PostForm("q1"),
			"Q2":c.PostForm("q2"),
			"Q3":c.PostForm("q3"),
			"Q4":c.PostForm("q4"),
			"Q5":c.PostForm("q5"),
		})
	})
	r.Get("/", func(c *dark.Context) {
		c.HTML(http.StatusOK,"index.tmpl",nil)
	})

	//easy(r)
	//addRouteWithParam(r)
	//g := addRouterGroup(r)
	//addMiddleware(r, g[1])
	//addStaticFilePath(r)
	//addTemplate(r)

	r.Run(":9999")
}
