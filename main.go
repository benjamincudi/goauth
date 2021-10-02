package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()
	r.LoadHTMLFiles("templates/index.gohtml")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", gin.H{
			"title": "GOAuth Index",
		})
	})
	if err := r.Run(":80"); err != nil {
		log.Fatal(err)
	}
}