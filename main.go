package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	//load ENV
	err := godotenv.Load(".env")
	if err != nil {
		log.Print("Error loading .env file.")
	}

	r := gin.Default()
	startTime := time.Now()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"uptime": (time.Since(startTime).Seconds()),
		})
	})

	r.POST("/jamplay/snap", func(c *gin.Context) {

		type format struct {
			Path        string `json:"path"`
			Title       string `json:"title"`
			AuthorName  string `json:"authorName"`
			AuthorImage string `json:"author"`
			CoverImage  string `json:"cover"`
			RenderType  string `json:"type"`
		}

		var f format
		c.BindJSON(&f)
		log.Print("keys ", f.RenderType)

		switch f.RenderType {
		case "share_author":
			log.Print("render share_author")
			renderShareAuthor(f.AuthorName, f.AuthorImage, f.Path, nil)
		case "share_book":
			log.Print("render share_book")
			renderShareBook(f.AuthorName, f.Title, f.CoverImage, f.AuthorImage, "", f.Path, nil)
		}

		c.JSON(200, gin.H{
			"key": f.Path,
		})
	})

	port := os.Getenv("PORT")

	if r.Run(strings.Join([]string{":", port}, "")) != nil {
		log.Print("GIN Run Error: ", err)
	}

	// renderAllBook()
}
