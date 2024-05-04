package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccha/envar"
	"github.com/goccha/yubinbango/internal/handlers"
	"net/http"
	"strings"
)

type Option func(r *gin.RouterGroup)

func WithHealthCheck(basePath string) Option {
	return func(r *gin.RouterGroup) {
		if r.BasePath() == basePath {
			r.GET("health", func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})
		}
	}
}

func WithBasicAuth(basePath, baUserPass string) Option {
	userPass := strings.Split(baUserPass, ":")
	bau := ""
	bap := ""
	if len(userPass) != 2 {
		bau = envar.Get("BASIC_AUTH_USER").String("user")
		bap = envar.Get("BASIC_AUTH_PASSWORD").String("pass")
	} else {
		bau = envar.Get("BASIC_AUTH_USER").String(userPass[0])
		bap = envar.Get("BASIC_AUTH_PASSWORD").String(userPass[1])
	}
	fmt.Printf("Basic Authentication Enabled: %s:%s\n", bau, bap)
	basicAuth := gin.BasicAuth(gin.Accounts{
		bau: bap,
	})
	return func(r *gin.RouterGroup) {
		if r.BasePath() == basePath {
			r.Use(basicAuth)
		}
	}
}

func Setup(router *gin.Engine, dirPath string, opt ...Option) error {
	root := router.Group("/")
	for _, o := range opt {
		o(root)
	}
	api := root.Group("api")
	for _, o := range opt {
		o(api)
	}

	api.Group("yubinbango").
		GET(":zip", handlers.Get("", dirPath)).
		GET("jsonp/:zip", handlers.Get("$yubin", dirPath))

	return nil
}

func Shutdown() {

}
