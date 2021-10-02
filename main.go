package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()
	if err := r.Run(":80"); err != nil {
		log.Fatal(err)
	}
}