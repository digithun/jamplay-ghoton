package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func drawImageHandler(c *gin.Context) {
}

func main() {
	r := gin.Default()

	r.POST("/draw", drawImageHandler)

	err := r.Run(":8744")
	if err != nil {
		log.Fatal(err.Error())
	}
}
