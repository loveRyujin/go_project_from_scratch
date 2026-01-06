package main

import (
	"log"
	"net/http"
	"time"

	"github.com/loveRyujin/gee"
)

func logMiddleware() gee.Handler {
	return func(c *gee.Context) {
		now := time.Now()

		c.Next()

		log.Printf("Time costing is %vns", time.Since(now).Nanoseconds())
	}
}

func middlewareV1() gee.Handler {
	return func(c *gee.Context) {
		log.Println("v1 middleware begin")

		c.Next()

		log.Println("v1 middleware end")
	}
}

func middlewareV2() gee.Handler {
	return func(c *gee.Context) {
		log.Println("v2 middleware begin")

		c.Next()

		log.Println("v2 middleware end")
	}
}

func main() {
	r := gee.Default()
	r.Use(logMiddleware())
	r.GET("/", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"msg": "success!",
		})
	})
	r.GET("/hello", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"msg": "hi!",
		})
	})
	r.GET("/hello/:name", func(c *gee.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path())
	})
	r.GET("/assets/*filepath", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
	})
	r.GET("/panic", func(c *gee.Context) {
		s := []string{"1", "2", "3"}
		c.String(http.StatusOK, "get %s", s[3])
	})

	v1 := r.Group("/v1")
	v1.Use(middlewareV1())
	v2 := v1.Group("/v2")
	v2.Use(middlewareV2())
	v2.GET("/hello", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"msg": "This is v1 and v2 group!",
		})
	})

	log.Printf("Server is running on :9999...")
	r.Run(":9999")
}
