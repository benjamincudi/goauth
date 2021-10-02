package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"google.golang.org/api/idtoken"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	gAppClientID = os.Getenv("GOOGLE_APP_ID")
	// JWTSecret JWT signing secrets are provided as byte slices
	JWTSecret     = []byte(os.Getenv("JWT_SECRET"))
	jwtCookieName = "gintoken"
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
		tCookie, err := c.Request.Cookie(jwtCookieName)
		if err != nil && err != http.ErrNoCookie {
			log.Println("no token cookie found, redirecting to index")
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		token, err := jwt.Parse(tCookie.Value, func(t *jwt.Token) (interface{}, error) {
			return JWTSecret, nil
		})
		if err != nil {
			log.Printf("invalid jwt: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error":"invalid jwt"})
		}
		c.HTML(http.StatusOK, "login.gohtml", token.Claims)
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
		p, err := idtoken.Validate(c.Request.Context(), form.Credential, gAppClientID)
		if err != nil {
			log.Printf("validation error, go back to index: %s\n", err)
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}
		jc := jwt.MapClaims{
			"email": p.Claims["email"],
			"picture": p.Claims["picture"],
			// Add a minute of leeway for the cookie to expire before the token
			"exp": time.Now().Add(time.Hour).Add(time.Minute).Unix(),
		}
		tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jc).SignedString(JWTSecret)
		if err != nil {
			log.Printf("jwt signing error, go back to index: %s\n", err)
			c.Redirect(http.StatusTemporaryRedirect, "/")
		}
		c.SetCookie(jwtCookieName, tokenString, 3600,"", "", true, true)
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