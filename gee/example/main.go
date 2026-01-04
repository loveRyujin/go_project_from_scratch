package main

import (
	"log"
	"net/http"

	"github.com/loveRyujin/gee"
)

func main() {
	r := gee.New()
	r.Get("/", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"msg": "success!",
		})
	})
	r.Get("/hello", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{
			"msg": "hi!",
		})
	})
	r.Get("/hello/:name", func(c *gee.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path())
	})
	r.Get("/assets/*filepath", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
	})
	log.Printf("Server is running on :9999...")
	r.Run(":9999")
}
