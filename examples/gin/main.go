package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/CodeLineage/brotli"
	"github.com/gin-gonic/gin"
)

func main() {
	g := gin.Default()

	g.Use(brotli.DefaultHandler().Gin)

	g.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "hello",
			"data": "GET /short and GET /long to have a try!",
		})
	})

	// short response will not be compressed
	g.GET("/short", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "This content is not long enough to be compressed.",
			"data": "short!",
		})
	})

	// long response that will be compressed by brotli
	g.GET("/long", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]interface{}{
			"code": 0,
			"msg":  "This content is compressed",
			"data": fmt.Sprintf("l%sng!", strings.Repeat("o", 1000)),
		})
	})

	const port = 8091

	log.Printf("Service is litsenning on port %d...", port)
	log.Println(g.Run(fmt.Sprintf(":%d", port)))
}
