package main

import (
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

// ImageMeta info of image to draw
type ImageMeta struct {
	Title              string
	Name               string
	ProfileImageURL    string
	ThumbnailImageURL  string
	BackgroundImageURL string
	FileName           string
}

const (
	canvasHeight     = 630
	canvasWidth      = 1200
	padding      int = 120.0

	bookWidth              float64 = 266.0
	bookHeight             float64 = 389.0
	profileImageWidth      float64 = 75.0
	profileImageHeight     float64 = 75.0
	profileImageMarginLeft float64 = 18

	// description measurement
	descriptionMarginLeft = 20
	descriptionPositionX  = float64(padding) + bookWidth + float64(descriptionMarginLeft)

	titleFontSize   float64 = 50.0
	titleLineHeight float64 = 1.5
	nameMarginTop   float64 = 0
	nameFontSize    float64 = 28.0
	nameLineHeight  float64 = 1.5

	maxTitleWidth = 700 //canvasWidth - (padding * 2) - descriptionMarginLeft
	maxNameWidth  = 700 - profileImageWidth - profileImageMarginLeft - 80
)

var (
	fontfile = flag.String("fontfile", "./assets/heavent.ttf", "filename of the ttf font")
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

func drawBookThumbnailImageToCanvas(thumbnailImageURL string, c *gg.Context) error {

	// Prepare asset
	thumbnailImage, err := getImageFromURL(thumbnailImageURL)
	if err != nil {
		return err
	}
	// calculate measuring
	thumbnailWidth, thumbnailHeight := float64(thumbnailImage.Bounds().Dx()), float64(thumbnailImage.Bounds().Dy())
	scaleX, scaleY := bookWidth/thumbnailWidth, bookHeight/thumbnailHeight
	BookPositionY := ((canvasHeight / 2) - (bookHeight / 2)) / scaleY
	BookPositionX := int(float64(padding) / float64(scaleX))
	// Start drawing
	c.Scale(scaleX, scaleY)
	c.DrawRoundedRectangle(float64(BookPositionX), BookPositionY, float64(thumbnailWidth), float64(thumbnailHeight), 15.0)
	c.Clip()
	c.DrawImage(thumbnailImage, BookPositionX, int(BookPositionY))
	c.ResetClip()
	c.Scale(1/scaleX, 1/scaleY)
	return nil
}

func drawDescription(meta *ImageMeta, c *gg.Context) error {
	// Prepare Asset
	authorProfileImage, err := getImageFromURL(meta.ProfileImageURL)
	if err != nil {
		return err
	}
	thumbnailWidth, thumbnailHeight := float64(authorProfileImage.Bounds().Dx()), float64(authorProfileImage.Bounds().Dy())
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return err
	}
	heaventFont, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return err
	}
	titleFace := truetype.NewFace(heaventFont, &truetype.Options{
		Size: titleFontSize,
	})
	nameFace := truetype.NewFace(heaventFont, &truetype.Options{
		Size: nameFontSize,
	})

	// Measureing constant....
	numberOfTitleLines := len(c.WordWrap(meta.Title, float64(maxTitleWidth)))
	numberOfNameLines := len(c.WordWrap(meta.Name, float64(maxNameWidth)))
	titleHeight := (float64(numberOfTitleLines) * (titleFontSize)) + (float64(numberOfTitleLines) * (titleFontSize * titleLineHeight))
	nameHeight := (float64(numberOfNameLines) * (nameFontSize))
	posY := (canvasHeight / 2) - ((nameHeight + titleHeight + nameMarginTop) / 2)

	// Draw title text
	c.SetRGB(1, 1, 1)
	c.SetFontFace(titleFace)
	c.DrawStringWrapped(meta.Title, float64(descriptionPositionX), float64(posY), float64(0), float64(0.5), float64(maxTitleWidth), titleLineHeight, gg.AlignCenter)

	c.SetFontFace(nameFace)
	posY = posY + titleHeight + nameMarginTop
	nameWidth, _ := c.MeasureString(meta.Name)
	if nameWidth > maxNameWidth {
		nameWidth = maxNameWidth
	}

	c.SetHexColor("#fff")
	profilePositionX := descriptionPositionX + profileImageWidth + (((maxNameWidth) / 2) - ((nameWidth) / 2)) - profileImageMarginLeft

	c.DrawStringWrapped(
		meta.Name,
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
	bgFile, err := gg.LoadImage("./assets/share_template_book.png")
	if err != nil {
		log.Fatal(err.Error())
	}

	canvas := gg.NewContext(canvasWidth, canvasHeight)
	canvas.DrawImage(bgFile, 0, 0)

	drawBookThumbnailImageToCanvas(meta.ThumbnailImageURL, canvas)
	err = drawDescription(meta, canvas)

	canvas.SavePNG(meta.FileName)

}
