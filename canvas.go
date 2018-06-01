package main

import (
	"bytes"
	"flag"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"regexp"

	"github.com/veer66/mapkha"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	m "github.com/veer66/mapkha"
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
const LAST_POSITION_BOUND_HORIZONTAL_CENTER = -9

//will center to last rectangle at X axis, only works if rectangle has defined height
const LAST_POSITION_BOUND_VERTICAL_CENTER = -10

//text width
const TEXT_WIDTH_TOFIT = 0
const TEXT_HEIGHT_TOFIT = 0

const TEXT_SINGLE_LINE = 9999
const TEXT_HEIGHT_MAX_LINE = -1

//text alignVertical
const TEXT_ALLIGN_HORIZONTAL_LEFT, TEXT_ALLIGN_HORIZONTAL_DEFAULT = gg.AlignLeft, gg.AlignLeft
const TEXT_ALLIGN_HORIZONTAL_CENTER = gg.AlignCenter
const TEXT_ALLIGN_HORIZONTAL_RIGHT = gg.AlignRight

const TEXT_ALLIGN_VERTICAL_TOP, TEXT_ALLIGN_VERTICAL_DEFAULT = 0, 0
const TEXT_ALLIGN_VERTICAL_CENTER = -1
const TEXT_ALLIGN_VERTICAL_BOTTOM = -2

//clip type
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
func (c *Canvas) drawImage(path, fallbackPath string, m Margin, r Rectangle, roundEdge float64) *Canvas {
	c.verifyRectangle(&r)

	if len(path) == 0 {
		path = fallbackPath
	}

	if strings.Index(path, "http") == 0 {
		c.logPrint("is URL : ")
		path = c.cacheURLtoDisk(path)
	}

	if len(path) == 0 && len(fallbackPath) > 0 {
		c.logPrint("path is broken using fallback image : ", path)
		path = fallbackPath
	}

	c.logPrint("is path : ", path)

	image, err := gg.LoadImage(path)

	if err != nil {
		c.logPrint("err ", err)
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
	// c.logPrint("scaleX ", scaleX)
	// c.logPrint("scaleY ", scaleY)
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

func (c *Canvas) toStream(ext string, w io.Writer) {

	if ext == "jpg" {
		jpegOpt := getJpegOpt()
		jpeg.Encode(w, c.context.Image(), &jpegOpt)
	} else {
		png.Encode(w, c.context.Image())
	}
}

func (c *Canvas) saveFile(path string) *Canvas {
	ext := getExtFromPath(path)

	if ext == ".jpg" {
		c.saveJPG(path)
	} else {
		c.context.SavePNG(path)
	}

	return c
}

func getJpegOpt() jpeg.Options {
	qualityenv := os.Getenv("QUALITY")
	quality := int64(75)
	if len(qualityenv) > 1 {
		quality, _ = strconv.ParseInt(qualityenv, 10, 64)
	}
	log.Print("quality ", quality)
	jpegOpt := jpeg.Options{Quality: int(quality)}
	return jpegOpt
}

func (c *Canvas) saveJPG(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	jpegOpt := getJpegOpt()
	return jpeg.Encode(file, c.context.Image(), &jpegOpt)
}

var regExt = regexp.MustCompile(`\.[0-9a-z]+$`)

func getExtFromPath(path string) string {
	path = strings.ToLower(path)

	match := regExt.Find([]byte(path))
	matchString := string(match)

	if len(matchString) > 0 {
		return matchString
	} else {
		return ""
	}
}

func (c *Canvas) drawText(text string, nameFace string, fontSize float64, m Margin, clip TextClipOption, r Rectangle, fontColorHex string, alignHorizontal gg.Align, alignVertical int) *Canvas {
	c.verifyToFit(&r)

	text = strings.TrimSpace(text)
	c.logPrint("rendering Text: ", text)
	font, err := c.getFont(nameFace)

	if err != nil {
		c.logPrint("drawText err : ", err)
		return c
	}

	fontFace := truetype.NewFace(font, &truetype.Options{
		Size: fontSize,
	})

	c.context.SetFontFace(fontFace)
	stringWidth, stringHeight := c.context.MeasureString(text)

	lines := []string{text}

	//calc lines
	var _clipWidth = 0
	if r.Dimension.Width == TEXT_WIDTH_TOFIT && clip.ClipWidth > 0 {
		_clipWidth = clip.ClipWidth
	} else {
		_clipWidth = r.Dimension.Width - m.Left - m.Right
	}
	c.logPrint("---------_clipWidth ", _clipWidth)
	lines = c.context.WordWrap(text, float64(_clipWidth))

	// end calc lines

	//help new line thai text
	if clip.MaxLine > 1 {
		workingLine := 0
		for {
			if workingLine > len(lines)-1 {
				break
			}

			// c.logPrint("workingLine ", workingLine)
			lineString := lines[workingLine]
			w, _ := c.context.MeasureString(lineString)
			if w > float64(_clipWidth) {
				//thai line exceed width, try removing one word to next line
				words := tWordCut(lineString)
				keepWords := words[0 : len(words)-1]
				shiftWords := words[len(words)-1:]

				lines[workingLine] = strings.Join(keepWords, "")

				joinShift := strings.Join(shiftWords, "")

				c.logPrint("workingLine ", workingLine)
				c.logPrint("len(lines) ", len(lines))
				// break
				if workingLine+1 >= len(lines) {
					//new line
					lines = append(lines, joinShift)
				} else {
					//modify line, add shift word in front
					lines[workingLine+1] = strings.Join([]string{joinShift, lines[workingLine+1]}, " ")
				}

				c.logPrint("keep ", keepWords)
				c.logPrint("shift ", shiftWords)
			} else {
				workingLine++
			}
		}
	}

	//set heigh if required
	if r.Dimension.Height == TEXT_HEIGHT_MAX_LINE {
		r.Dimension.Height = int(stringHeight*float64(clip.MaxLine)*nameLineHeight) + m.Bottom + m.Top
	}

	// hasClipLine := false
	if clip.MaxLine > 0 && len(lines) > clip.MaxLine {
		lines = lines[:clip.MaxLine]
		// hasClipLine = true
	}

	hasTrimmed := false
	c.logPrint("lines ", len(lines), " w:", r.Dimension.Width)
	//clip new line
	if clip.ClipWidth > 0 {

		// if clip.ClipWidth > r.Dimension.Width-m.Left-m.Right {
		// 	clip.ClipWidth = r.Dimension.Width - m.Left - m.Right
		// }

		for {
			var _stringWidth = 0.0

			lineText := lines[len(lines)-1]
			c.logPrint("lineText ", lineText)

			if clip.OverFlowOption == TEXT_CLIP_OVERFLOW_ELLIPSIS {
				c.logPrint("_stringWidth TEXT_CLIP_OVERFLOW_ELLIPSIS")
				_stringWidth, _ = c.context.MeasureString(strings.Join([]string{lineText, "..."}, ""))
			} else {
				c.logPrint("_stringWidth norm")

				_stringWidth, _ = c.context.MeasureString(lineText)
			}

			// c.logPrint("ww ", lineText, " ", int(_stringWidth), "--", clip.ClipWidth)
			c.logPrint("_stringWidth ", _stringWidth)
			c.logPrint("ClipWidth ", clip.ClipWidth)
			c.logPrint("r.Dimension.Width ", r.Dimension.Width)
			if _stringWidth > float64(clip.ClipWidth) {
				trimmed := trimLastChar(lineText)
				c.logPrint("trimmed ", trimmed)
				lines[len(lines)-1] = trimmed

				hasTrimmed = true
			} else {
				stringWidth = _stringWidth
				if r.Dimension.Width == TEXT_WIDTH_TOFIT {
					r.Dimension.Width = int(stringWidth)
					c.logPrint("TEXT_TOFIT = ", r.Dimension.Width)
				}

				if hasTrimmed {
					_toJoin := lines[len(lines)-1]

					_toJoin = c.cleanUpThaiText(_toJoin)
					if clip.OverFlowOption == TEXT_CLIP_OVERFLOW_ELLIPSIS {
						lines[len(lines)-1] = strings.Join([]string{_toJoin, "..."}, "")
					} else {
						lines[len(lines)-1] = _toJoin
					}
				}

				break
			}
		}
	}

	// c.logPrint("text ", text)

	var renderedWidth float64
	var isToFit = r.Dimension.Width == TEXT_WIDTH_TOFIT
	if isToFit {
		r.Dimension.Width = int(stringWidth)
	}

	// c.logPrint("p1 ", r)
	c.verifyRectangle(&r)
	// c.logPrint("p2 ", r)

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
	c.logPrint("numLine ", numLine, "  ", lines)
	if numLine > clip.MaxLine {
		numLine = clip.MaxLine
	}

	if r.Dimension.Height == TEXT_WIDTH_TOFIT {
		r.Dimension.Height = int(stringHeight*float64(numLine)*nameLineHeight) + m.Bottom + m.Top
	} else {
		// c.logPrint("find new numline ", int(float64(r.Dimension.Height)/float64((stringHeight*float64(1)*nameLineHeight))))
		numLine = int(float64(r.Dimension.Height) / float64((stringHeight * float64(1) * nameLineHeight)))
	}

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

	//force new line on render
	renderedText := strings.Join(lines, "\n")

	// if hasTrimmed && clip.OverFlowOption == TEXT_CLIP_OVERFLOW_ELLIPSIS {
	// 	text = strings.Join([]string{text, "..."}, "")
	// }

	//shift down if allign vertically
	if (alignVertical == TEXT_ALLIGN_VERTICAL_CENTER ||
		alignVertical == TEXT_ALLIGN_VERTICAL_BOTTOM) &&
		len(lines) < numLine {

		shiftY := int(stringHeight * float64((numLine - len(lines))) * nameLineHeight)
		if alignVertical == TEXT_ALLIGN_VERTICAL_CENTER {
			renderedY += shiftY / 2
			r.Point.y -= shiftY / 2

		} else if alignVertical == TEXT_ALLIGN_VERTICAL_BOTTOM {
			renderedY += shiftY
			// r.Point.y -= shiftY
		}
	}
	c.logPrint("---------renderedWidth ", renderedWidth)

	c.context.DrawStringWrapped(
		renderedText,
		float64(renderedX),
		float64(renderedY),
		float64(0),
		float64(0),
		renderedWidth,
		nameLineHeight, alignHorizontal,
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

	if r.Point.x == LAST_POSITION_BOUND_HORIZONTAL_CENTER {
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

	if r.Point.y == LAST_POSITION_BOUND_VERTICAL_CENTER {
		r.Point.y = c.lastRenderBound.Point.y + c.lastRenderBound.Dimension.Height/2 - r.Dimension.Height/2
	}
	//
}

func (c *Canvas) saveLastBound(r Rectangle) {
	c.lastRenderBound = r
}

func (c *Canvas) saveLastClipText(t string) {
	c.lastClipText = t
}

func (c *Canvas) cacheURLtoDisk(url string) string {

	splits := strings.Split(url, "//")
	c.logPrint("split ", splits[1])

	savePath := strings.Join([]string{"asset", splits[1]}, "/")

	os.MkdirAll(filePathToDirPath(savePath), 0755)
	//request http
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		// path/to/whatever does not exist, load from http

		client := http.Client{
			Timeout: time.Duration(10 * time.Second),
		}
		response, e := client.Get(url)
		if e != nil {
			c.logPrint("Fatal  ", e)
			return ""
		}

		defer response.Body.Close()

		//create file
		c.logPrint("savePath ", savePath)
		file, err := os.Create(savePath)
		_, err = io.Copy(file, response.Body)
		if err != nil {
			c.logPrint("io.Copy err ", err)
			return ""
		}
	} else {
		c.logPrint("File already cached")
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

func (c *Canvas) getFont(fontName string) (*truetype.Font, error) {
	cache, hasCache := ttcache[fontName]
	if hasCache {
		return cache, nil
	}

	fontFlag := flag.String(fontName, strings.Join([]string{"./asset/font/", fontName}, ""), "")
	font, err := c.getFontFile(fontFlag)
	ttcache[fontName] = font

	return font, err
}

func (c *Canvas) getFontFile(fontString *string) (font *truetype.Font, err error) {

	fontbyte, err := ioutil.ReadFile(*fontString)
	if err != nil {
		c.logPrint(err)
		return nil, err
	}

	font, err = truetype.Parse(fontbyte)
	if err != nil {
		c.logPrint(err)
		return nil, err
	}

	return font, err
}

func runeOf(__s string) rune {
	return []rune(__s)[0]
}

func (c *Canvas) cleanUpThaiText(s string) string {

	for {
		lastChar := s[len(s)-1]
		c.logPrint("lastChar ", []rune(string(lastChar))[0])

		_rune := runeOf(string(lastChar))

		if runeOf("�") == _rune ||
			runeOf("ั") == _rune ||
			runeOf("ิ") == _rune ||
			runeOf("็") == _rune ||
			runeOf("ํ") == _rune ||
			runeOf("๊") == _rune ||
			runeOf("ึ") == _rune ||
			runeOf("่") == _rune ||
			runeOf("้") == _rune ||
			runeOf("๋") == _rune ||
			runeOf("ุ") == _rune ||
			runeOf("ู") == _rune {
			s = s[0 : len(s)-1]

		} else {
			c.logPrint("cleaned text")
			break
		}
	}

	return s
}

func trimLastChar(s string) string {
	return s[0 : len(s)-1]
}

func (c *Canvas) logPrint(v ...interface{}) {
	if c.debug {
		log.Print(v)
	}
}

var dict *mapkha.Dict

func tWordCut(s string) []string {

	if dict == nil {
		dict, _ = m.LoadDefaultDict()
	}

	wordcut := m.NewWordcut(dict)

	return wordcut.Segment(s)
}
