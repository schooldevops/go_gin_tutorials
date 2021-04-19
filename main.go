package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/schooldevops/gin_tutorial/handler"
	"github.com/schooldevops/gin_tutorial/utils"
)

func main() {

	utils.LoadConfig()

	r := gin.Default()

	utils.InitLog(r)

	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/someJSON", func(c *gin.Context) {
		data := map[string]interface{}{
			"lang": "GO테스트",
			"tag":  "<br>",
		}

		c.AsciiJSON(http.StatusOK, data)
	})

	handler.AuthRoutes(r)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	s.ListenAndServe()
}
