package main

import (
	"log"
	"net/http"

	"github.com/loveRyujin/gee"
)

func main() {
	r := gee.New()
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

	g := r.Group("/v1")
	g.GET("/hello", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"msg": "hi, v1 group!",
		})
	})

	log.Printf("Server is running on :9999...")
	r.Run(":9999")
}
