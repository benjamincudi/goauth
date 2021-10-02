package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
	"html/template"
	"log"
	"net/http"
	"os"
)

var (
	gAppClientID = os.Getenv("GOOGLE_APP_ID")
)

type googleForm struct {
	Credential string `form:"credential"`
	GoogleCSRFToken string `form:"g_csrf_token"`
}

func main() {
	r := gin.Default()
	// if true, read live files from local filesystem instead of embedded assets
	useLiveFiles := gin.Mode() != gin.ReleaseMode
	loadTemplates(r, useLiveFiles)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", gin.H{
			"title": "GOAuth Index",
			"gAppClientID": gAppClientID,
		})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.gohtml", nil)
	})
	r.POST("/api/login/google", func(c *gin.Context) {
		gCSRFCookie, err := c.Request.Cookie("g_csrf_token")
		if err != nil {
			log.Printf("err looking for g_csrf_token cookie: %s\n", err)
		}
		var form googleForm
		if err = c.ShouldBind(&form); err != nil {
			log.Printf("error parsing form: %v\n", err)
		}
		if gCSRFCookie.Value != form.GoogleCSRFToken {
			log.Println("CSRF mismatch, go back to index")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		_, err = idtoken.Validate(c.Request.Context(), form.Credential, gAppClientID)
		if err != nil {
			log.Printf("validation error, go back to index: %s\n", err)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		c.Redirect(http.StatusSeeOther, "/login")
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