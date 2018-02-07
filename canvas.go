package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

// ImageMeta info of image to draw
type ImageMeta struct {
	Cover  string `json:"cover" binding:"required"`
	Author string `json:"author" binding:"required"`

	Title      string `json:"title" binding:"required"`
	AuthorName string `json:"authorName" binding:"required"`

	Path string `json:"path" binding:"required"`
	Type string `json:"type" binding:"required"`
}

const (
	canvasHeight     = 630
	canvasWidth      = 1200
	padding      int = 143.0

	bookWidth              float64 = 266.0
	bookHeight             float64 = 389.0
	profileImageWidth      float64 = 75.0
	profileImageHeight     float64 = 75.0
	profileImageMarginLeft float64 = 13

	// description measurement
	descriptionMarginLeft = 20
	descriptionPositionX  = float64(padding) + bookWidth + float64(descriptionMarginLeft)

	titleFontSize   float64 = 68.0
	titleLineHeight float64 = 1.5
	nameMarginTop   float64 = 0
	nameFontSize    float64 = 28.0
	nameLineHeight  float64 = 1.5

	maxTitleWidth = 700 //canvasWidth - (padding * 2) - descriptionMarginLeft
	maxNameWidth  = 700 - profileImageWidth - profileImageMarginLeft - 80
)

var (
	DBHeaventRoundedMed = flag.String("DBHeaventRoundedMed", "./assets/font/DBHeaventRoundedMedv3.2.ttf", "filename of the DBHeaventRoundedMed ttf font")
	mnlannabdv3         = flag.String("mnlannabdv3", "./assets/font/mn_lanna_bd_v3.2-webfont.ttf", "filename of the mnlannabdv3 ttf font")
)

func getImageFromURL(url string) (image.Image, error) {
	response, responseError := http.Get(url)
	if responseError != nil {
		return nil, responseError
	}

	defer response.Body.Close()

	img, _, err := image.Decode(response.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	return img, err
}

func drawURLImage(imageURL string, posX int, posY int, width int, height int, roundEdge float64, c *gg.Context) error {
	// Prepare asset
	image, err := getImageFromURL(imageURL)
	if err != nil {
		return err
	}
	// calculate measuring
	imageWidth, imageHeight := float64(image.Bounds().Dx()), float64(image.Bounds().Dy())
	scaleX, scaleY := float64(width)/imageWidth, float64(height)/imageHeight
	// Start drawing
	c.Scale(scaleX, scaleY)
	c.DrawRoundedRectangle(float64(posX)/scaleX,float64(posY)/scaleY, float64(imageWidth), float64(imageHeight), roundEdge)
	c.Clip()
	c.DrawImage(image, int(float64(posX)/scaleX), int(float64(posY)/scaleY))
	c.ResetClip()
	c.Scale(1/scaleX, 1/scaleY)
	return nil
}

func getFont(fontString *string) (font *truetype.Font, err error) {

	fontbyte, err := ioutil.ReadFile(*fontString)
	if err != nil {
		log.Println(err)
		return
	}

	font, err = truetype.Parse(fontbyte)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func drawDescription(meta *ImageMeta, c *gg.Context) error {
	// Prepare Asset
	authorProfileImage, err := getImageFromURL(meta.Author)
	if err != nil {
		return err
	}
	thumbnailWidth, thumbnailHeight := float64(authorProfileImage.Bounds().Dx()), float64(authorProfileImage.Bounds().Dy())

	mnlannaFont, err := getFont(mnlannabdv3)
	DBHeaventRoundedMedFont, err := getFont(DBHeaventRoundedMed)
	if err != nil {
		log.Println(err)
		return err
	}
	titleFace := truetype.NewFace(DBHeaventRoundedMedFont, &truetype.Options{
		Size: titleFontSize,
	})
	nameFace := truetype.NewFace(mnlannaFont, &truetype.Options{
		Size: nameFontSize,
	})

	// Measureing constant....
	numberOfTitleLines := len(c.WordWrap(meta.Title, float64(maxTitleWidth)))
	numberOfNameLines := len(c.WordWrap(meta.AuthorName, float64(maxNameWidth)))
	titleHeight := (float64(numberOfTitleLines) * (titleFontSize)) + (float64(numberOfTitleLines) * (titleFontSize * titleLineHeight))
	nameHeight := (float64(numberOfNameLines) * (nameFontSize))
	posY := (canvasHeight / 2) - ((nameHeight + titleHeight + nameMarginTop) / 2)

	// Draw title text
	c.SetRGB(1, 1, 1)
	c.SetFontFace(titleFace)
	c.DrawStringWrapped(meta.Title, float64(descriptionPositionX), float64(posY), float64(0), float64(0.5), float64(maxTitleWidth), titleLineHeight, gg.AlignCenter)

	c.SetFontFace(nameFace)
	posY = posY + titleHeight + nameMarginTop
	nameWidth, _ := c.MeasureString(meta.AuthorName)
	if nameWidth > maxNameWidth {
		nameWidth = maxNameWidth
	}

	c.SetHexColor("#fff")
	profilePositionX := descriptionPositionX + profileImageWidth + (((maxNameWidth) / 2) - ((nameWidth) / 2)) - profileImageMarginLeft

	c.DrawStringWrapped(
		meta.AuthorName,
		float64(profilePositionX+profileImageMarginLeft+profileImageWidth/2),
		float64(posY),
		float64(0),
		float64(0.5),
		float64(maxNameWidth),
		nameLineHeight, gg.AlignLeft,
	)

	scaleX, scaleY := profileImageWidth/thumbnailWidth, profileImageHeight/thumbnailHeight
	fmt.Printf("%f %f", scaleX, scaleY)

	profilePositionX = profilePositionX / scaleX
	profilePositionY := posY / scaleY

	c.Scale(scaleX, scaleY)
	c.DrawCircle(profilePositionX, profilePositionY, (profileImageWidth/scaleX)/2)
	c.Clip()
	c.DrawImageAnchored(authorProfileImage, int(profilePositionX), int(profilePositionY), 0.5, 0.5)

	return nil

}

// DrawImage start create new canvas and do image drawing step
func DrawImage(meta *ImageMeta) {
	fmt.Println("[main] Draw image..")

	// Load BackgroundImage from file
	s := strings.Split(meta.Type, "_")
	template := strings.Join([]string{"./assets/", s[0], "_template_", s[1], ".png"}, "")

	bgFile, err := gg.LoadImage(template)
	if err != nil {
		log.Fatal(err.Error())
	}

	canvas := gg.NewContext(canvasWidth, canvasHeight)
	canvas.DrawImage(bgFile, 0, 0)


	drawURLImage(meta.Cover,143,140,int(bookWidth),int(bookHeight),15,canvas)

	// drawBookThumbnailImageToCanvas(meta.Cover, canvas)
	err = drawDescription(meta, canvas)

	canvas.SavePNG(meta.Path)

}
