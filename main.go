package main

import (
	"log"
	"github.com/gin-gonic/gin"
)

func drawImageHandler(c *gin.Context) {
	var meta1 *ImageMeta
	c.BindJSON(&meta1)


	DrawImage(meta1)

	c.JSON(200, gin.H{
		"Path": meta1.Path,
	})
}

func main() {
	r := gin.Default()

	r.POST("/draw", drawImageHandler)

	err := r.Run(":8744")
	if err != nil {
		log.Fatal(err.Error())
	}
}
