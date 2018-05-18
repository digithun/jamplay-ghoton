package main

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

const nameLineHeight float64 = 1.2
const DIMENSION_AUTO_RESIZE_ASPECT_RATIO = -1
const LAST_DIMENSION_TOFIT_WIDTH_RIGHT = -1
const LAST_DIMENSION_TOFIT_WIDTH_LEFT = -2
const LAST_DIMENSION_WIDTH = -3
const LAST_DIMENSION_HEIGHT = -4

////relate to last rectangle position

//will place outside the rectangle, X axis
const LAST_POSITION_BOUND_NEXT_TO_LEFT = -1
const LAST_POSITION_BOUND_NEXT_TO_RIGHT = -2

//will place outside the rectangle, Y axis
const LAST_POSITION_BOUND_NEXT_TO_BOTTOM = -3
const LAST_POSITION_BOUND_NEXT_TO_TOP = -4

//will place inside the rectangle, X axis
const LAST_POSITION_INNNER_BOUND_LEFT = -5
const LAST_POSITION_INNNER_BOUND_RIGHT = -6

//will place inside the rectangle, Y axis
const LAST_POSITION_INNER_BOUND_BOTTOM = -7
const LAST_POSITION_INNER_BOUND_TOP = -8

//will center to last rectangle at Y axis
const LAST_POSITION_BOUND_VERTICAL_CENTER = -9

const TEXT_TOFIT = 0
const TEXT_SINGLE_LINE = 9999

//type
const TEXT_CLIP_OVERFLOW_CLIP = 0
const TEXT_CLIP_OVERFLOW_ELLIPSIS = 1

type TextClipOption struct {
	OverFlowOption int
	NoClip         bool
	ClipWidth      int
	MaxLine        int
}

type Point struct {
	x, y int
}

type Dimension struct {
	Width, Height int
}

type Margin struct {
	Top, Right, Bottom, Left int
}

type Padding struct {
	Top, Right, Bottom, Left int
}

type Rectangle struct {
	Dimension Dimension
	Point     Point
}

func (r Rectangle) left() int {
	return r.Point.x
}

func (r Rectangle) right() int {
	return r.Point.x + r.Dimension.Width
}

func (r Rectangle) top() int {
	return r.Point.y
}

func (r Rectangle) bottom() int {
	return r.Point.y + r.Dimension.Height
}

type Canvas struct {
	context         gg.Context
	debug           bool
	width, height   int
	lastRenderBound Rectangle
	lastClipText    string
}

func NewCanvas(size Dimension) *Canvas {
	c := new(Canvas)
	c.context = *gg.NewContext(size.Width, size.Height)
	c.width = size.Width
	c.height = size.Height
	c.lastRenderBound = Rectangle{}

	return c
}

func (c *Canvas) drawCircle(p Point, radius float64, colorHex string) *Canvas {
	c.context.DrawCircle(float64(p.x), float64(p.y), radius)
	c.context.SetHexColor(colorHex)
	c.context.Fill()
	return c
}

// scaleToPx scale to pixel
// scaleDimension select which side to scale. All scaling will retain aspect ratio
func (c *Canvas) drawImage(path string, m Margin, r Rectangle, roundEdge float64) *Canvas {
	c.verifyRectangle(&r)

	if strings.Index(path, "http") == 0 {
		log.Print("is URL : ")
		path = cacheURLtoDisk(path)
	}
	log.Print("is path : ", path)

	image, err := gg.LoadImage(path)

	if err != nil {
		log.Print("err ", err)
		return c
	}

	// calculate measuring
	imageWidth, imageHeight := float64(image.Bounds().Dx()), float64(image.Bounds().Dy())
	scaleX, scaleY := float64(r.Dimension.Width-m.Right-m.Left)/imageWidth, float64(r.Dimension.Height-m.Bottom-m.Top)/imageHeight
	if scaleY < 0 {
		scaleY = scaleX
	}
	if scaleX < 0 {
		scaleX = scaleY
	}

	//imageWidth -= float64(m.left+m.right) / scaleX
	//imageHeight -= float64(m.top+m.bottom) / scaleY
	// log.Print("scaleX ", scaleX)
	// log.Print("scaleY ", scaleY)
	// Start drawing

	if c.debug {
		c.context.SetHexColor("#55ffff55")
		c.context.DrawRectangle(float64(r.Point.x), float64(r.Point.y), float64(r.Dimension.Width), float64(r.Dimension.Height))
		c.context.Fill()

		c.context.SetHexColor("#ffffff55")
		c.context.DrawRectangle(float64(r.Point.x+m.Left), float64(r.Point.y+m.Top), float64(r.Dimension.Width-m.Right-m.Left), float64(r.Dimension.Height-m.Bottom-m.Top))
		c.context.Fill()
	}

	c.context.Scale(scaleX, scaleY)
	c.context.DrawRoundedRectangle(float64(r.Point.x+m.Left)/scaleX, float64(r.Point.y+m.Top)/scaleY, float64(imageWidth), float64(imageHeight), roundEdge/scaleX)
	c.context.Clip()

	r.Dimension.Width = int(imageWidth * scaleX)
	r.Dimension.Height = int(imageHeight * scaleY)

	drawX := int(float64(r.Point.x+m.Left) / scaleX)
	drawY := int(float64(r.Point.y+m.Top) / scaleY)

	c.context.DrawImage(image, drawX, drawY)
	c.context.ResetClip()
	c.context.Scale(1/scaleX, 1/scaleY)

	c.saveLastBound(r)

	return c
}

func (c *Canvas) savePNG(path string) *Canvas {
	c.context.SavePNG(path)
	return c
}

func trimLastChar(s string) string {
	return s[0 : len(s)-1]
}

func (c *Canvas) drawText(text string, nameFace string, fontSize float64, m Margin, clip TextClipOption, r Rectangle, fontColorHex string, align gg.Align) *Canvas {
	c.verifyToFit(&r)

	text = strings.TrimSpace(text)

	font, err := getFont(nameFace)

	if err != nil {
		log.Print("drawText err : ", err)
		return c
	}

	fontFace := truetype.NewFace(font, &truetype.Options{
		Size: fontSize,
	})

	c.context.SetFontFace(fontFace)
	stringWidth, stringHeight := c.context.MeasureString(text)

	lines := c.context.WordWrap(text, float64(r.Dimension.Width))

	if clip.MaxLine > 0 && len(lines) > clip.MaxLine {
		lines = lines[:clip.MaxLine]
	}

	log.Print("lines ", len(lines), " w:", r.Dimension.Width)
	if clip.ClipWidth > 0 {
		dotWidth, _ := c.context.MeasureString("...")
		for {
			_stringWidth, _ := c.context.MeasureString(lines[len(lines)-1])
			if clip.OverFlowOption == TEXT_CLIP_OVERFLOW_ELLIPSIS {
				_stringWidth += dotWidth
			}

			log.Print("ww ", lines[len(lines)-1], " ", int(_stringWidth), "--", clip.ClipWidth)

			if int(_stringWidth) > clip.ClipWidth {
				trimmed := trimLastChar(lines[len(lines)-1])
				log.Print("trimmed ", trimmed)
				lines[len(lines)-1] = trimmed
			} else {
				stringWidth = _stringWidth

				break
			}
		}
	}

	// log.Print("text ", text)

	var renderedWidth float64
	var isToFit = r.Dimension.Width == TEXT_TOFIT
	if isToFit {
		r.Dimension.Width = int(stringWidth)
	}

	// log.Print("p1 ", r)
	c.verifyRectangle(&r)
	// log.Print("p2 ", r)

	renderedX := r.Point.x
	renderedY := r.Point.y

	if isToFit {
		renderedWidth = stringWidth
		renderedX = renderedX + m.Left - m.Right
		renderedY = renderedY + m.Top - m.Bottom
	} else {
		renderedWidth = float64(r.Dimension.Width - m.Right - m.Left)
		renderedX += m.Left
		renderedY += m.Top
	}

	// if align == gg.AlignCenter {
	// 	renderedX += m.left
	// } else {

	// }

	numLine := len(lines) //int(math.Ceil(stringWidth / float64(renderedWidth)))
	if numLine > clip.MaxLine {
		numLine = clip.MaxLine
	}

	r.Dimension.Height = int(stringHeight*float64(numLine)*nameLineHeight) + m.Bottom + m.Top
	if c.debug {
		//blue normal
		c.context.SetHexColor("#0000ffff")
		c.context.DrawRectangle(float64(r.Point.x), float64(r.Point.y), float64(r.Dimension.Width), float64(r.Dimension.Height))
		c.context.Fill()

		//margin red
		c.context.SetHexColor("#ff0000ff")
		c.context.DrawRectangle(float64(renderedX), float64(renderedY), renderedWidth, float64(r.Dimension.Height-m.Bottom-m.Top))
		c.context.Fill()
	} else {
		//CLIPPING
		c.context.DrawRectangle(float64(r.Point.x), float64(r.Point.y), float64(r.Dimension.Width), float64(r.Dimension.Height))
		if !clip.NoClip {
			c.context.Clip()
		}
	}
	c.context.SetHexColor(fontColorHex)

	renderedText := strings.Join(lines, " ")
	if clip.OverFlowOption == TEXT_CLIP_OVERFLOW_ELLIPSIS {
		text = strings.Join([]string{text, "..."}, "")
	}

	c.context.DrawStringWrapped(
		renderedText,
		float64(renderedX),
		float64(renderedY),
		float64(0),
		float64(0),
		renderedWidth,
		nameLineHeight, align,
	)

	if !clip.NoClip {
		c.context.ResetClip()
	}
	c.saveLastBound(r)
	c.saveLastClipText(text)

	return c
}

func (c *Canvas) verifyToFit(r *Rectangle) {
	//fit remaining.
	if r.Dimension.Width == LAST_DIMENSION_TOFIT_WIDTH_RIGHT {
		r.Dimension.Width = c.width - c.lastRenderBound.right()
	}

	if r.Dimension.Width == LAST_DIMENSION_TOFIT_WIDTH_LEFT {
		r.Dimension.Width = c.width - c.lastRenderBound.Dimension.Width
	}
}

func (c *Canvas) verifyRectangle(r *Rectangle) {

	if r.Dimension.Height == LAST_DIMENSION_HEIGHT {
		r.Dimension.Height = c.lastRenderBound.Dimension.Height
	}

	if r.Dimension.Width == LAST_DIMENSION_WIDTH {
		r.Dimension.Width = c.lastRenderBound.Dimension.Width
	}

	if r.Point.x == LAST_POSITION_BOUND_NEXT_TO_RIGHT {
		right := c.lastRenderBound.right()
		if right > c.width {
			r.Point.x = 0
		} else {
			r.Point.x = right
		}
	}

	if r.Point.x == LAST_POSITION_BOUND_NEXT_TO_LEFT {
		r.Point.x = c.lastRenderBound.left() - r.Dimension.Width
	}

	if r.Point.x == LAST_POSITION_INNNER_BOUND_LEFT {
		r.Point.x = c.lastRenderBound.left()
	}

	if r.Point.x == LAST_POSITION_INNNER_BOUND_RIGHT {
		r.Point.x = c.lastRenderBound.right() - r.Dimension.Width
	}

	if r.Point.x == LAST_POSITION_BOUND_VERTICAL_CENTER {
		r.Point.x = c.lastRenderBound.left() + c.lastRenderBound.Dimension.Width/2 - r.Dimension.Width/2
	}

	if r.Point.y == LAST_POSITION_BOUND_NEXT_TO_TOP {
		r.Point.y = c.lastRenderBound.top() - r.Dimension.Height
	}

	if r.Point.y == LAST_POSITION_BOUND_NEXT_TO_BOTTOM {
		r.Point.y = c.lastRenderBound.bottom()
	}

	if r.Point.y == LAST_POSITION_INNER_BOUND_TOP {
		r.Point.y = c.lastRenderBound.top()
	}

	if r.Point.y == LAST_POSITION_INNER_BOUND_BOTTOM {
		r.Point.y = c.lastRenderBound.bottom() - r.Dimension.Height
	}

	//
}

func (c *Canvas) saveLastBound(r Rectangle) {
	c.lastRenderBound = r
}

func (c *Canvas) saveLastClipText(t string) {
	c.lastClipText = t
}

func cacheURLtoDisk(url string) string {

	splits := strings.Split(url, "//")
	log.Print("split ", splits[1])

	savePath := strings.Join([]string{"asset", splits[1]}, "/")

	os.MkdirAll(filePathToDirPath(savePath), 0755)
	//request http
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		response, e := http.Get(url)
		if e != nil {
			log.Fatal("Fatal  ", e)
		}

		defer response.Body.Close()

		//create file
		log.Print("savePath ", savePath)
		file, err := os.Create(savePath)
		_, err = io.Copy(file, response.Body)
		if err != nil {
			log.Fatal("io.Copy err ", err)
		}
	} else {
		log.Print("File already cached")
	}

	return savePath
}

func filePathToDirPath(p string) string {
	sp := strings.Split(p, "/")
	var buffer bytes.Buffer
	for i := 0; i < len(sp)-1; i++ {
		buffer.WriteString(sp[i])

		if i < len(sp)-2 {
			buffer.WriteString("/")
		}
	}

	return buffer.String()
}

var ttcache = make(map[string]*truetype.Font)

func getFont(fontName string) (*truetype.Font, error) {
	cache, hasCache := ttcache[fontName]
	if hasCache {
		return cache, nil
	}

	fontFlag := flag.String(fontName, strings.Join([]string{"./asset/font/", fontName}, ""), "")
	font, err := getFontFile(fontFlag)
	ttcache[fontName] = font

	return font, err
}

func getFontFile(fontString *string) (font *truetype.Font, err error) {

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
