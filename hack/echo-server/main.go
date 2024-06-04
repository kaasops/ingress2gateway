package main

import (
	"github.com/gin-gonic/gin"
)

func echo(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": c.Request.URL.Path,
	})
}

func main() {
	endpoints := []string{
		"/ping",
		"/hello",
		"/goodbye",
		"/endpoint1",
		"/endpoint2",
		"/endpoint3",
	}
	router := gin.Default()
	for _, endpoint := range endpoints {
		router.GET(endpoint, echo)
	}
	router.Run(":8080")
}
