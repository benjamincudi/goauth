package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()
	// if true, read live files from local filesystem instead of embedded assets
	useLiveFiles := gin.Mode() != gin.ReleaseMode
	loadTemplates(r, useLiveFiles)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", gin.H{
			"title": "GOAuth Index",
		})
	})
	if err := r.Run(":80"); err != nil {
		log.Fatal(err)
	}
}

//go:embed templates
var htmlTemplates  embed.FS

func loadTemplates(r *gin.Engine, useLiveFiles bool) {
	if useLiveFiles {
		r.LoadHTMLGlob("templates/*")
		return
	}
	t, err := template.ParseFS(htmlTemplates,  "templates/*")
	if err != nil {
		panic(fmt.Sprintf("failed to parse embedded templates: %v", err))
	}
	r.SetHTMLTemplate(t)
}