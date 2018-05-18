package main

import (
	"log"
	"strings"
	"time"

	"github.com/fogleman/gg"

	"github.com/joho/godotenv"
)

func main() {
	//load ENV
	err := godotenv.Load(".env")
	if err != nil {
		log.Print("Error loading .env file.")
	}

	t1 := time.Now().UnixNano()
	// Set your Google Cloud Platform project ID.
	c := NewCanvas(Dimension{1200, 630})
	c.debug = true

	name := "OResia a b c d e f g h i j k l m n o p q r s t u v w x y z"
	title := "A Fairy Promise   a b c d e f g h i j k l m n o p q r s t u v w x y z a b c d e f g h i j"
	c.drawImage("asset/share_template_book.png", Margin{}, Rectangle{
		Dimension{1200, 630},
		Point{0, 0},
	}, 0).
		drawImage(
			"https://static.jamplay.world/book/5ad6f2ca8b7c1b000fac2177/b1556f69-d505-4100-bcc5-b98fba989e9c.blob.jpg",
			Margin{},
			Rectangle{
				Dimension{405, DIMENSION_AUTO_RESIZE_ASPECT_RATIO},
				Point{20, 22},
			}, 9).
		drawText(
			title,
			"DBHeaventRoundedMedv3.2.ttf",
			68.0,
			Margin{0, 10, 17, 10},
			TextClipOption{
				MaxLine:        2,
				ClipWidth:      c.width - 405 - 40,
				OverFlowOption: TEXT_CLIP_OVERFLOW_ELLIPSIS,
			},
			Rectangle{
				Dimension{LAST_DIMENSION_TOFIT_WIDTH_RIGHT, 0},
				Point{LAST_POSITION_BOUND_NEXT_TO_RIGHT, 69},
			},
			"#fff", gg.AlignCenter).
		drawImage(
			"asset/default-pic-editor.png",
			Margin{},
			Rectangle{
				Dimension{220, 220},
				Point{LAST_POSITION_BOUND_VERTICAL_CENTER, LAST_POSITION_BOUND_NEXT_TO_BOTTOM}},
			110).
		drawText(
			strings.Join([]string{"__ ", name}, ""),
			"DBHeaventRoundedMedv3.2.ttf",
			53.0,
			Margin{5, 0, 0, 0},
			TextClipOption{
				MaxLine:        1,
				ClipWidth:      c.width - 405 - 40,
				OverFlowOption: TEXT_CLIP_OVERFLOW_ELLIPSIS,
			},
			Rectangle{
				Dimension{c.width - 405 - 40 - 40, 53},
				Point{LAST_POSITION_BOUND_VERTICAL_CENTER, LAST_POSITION_BOUND_NEXT_TO_BOTTOM},
			},
			"#55555500", gg.AlignCenter)

	nameRectangle := c.lastRenderBound
	c.saveLastBound(nameRectangle)
	nameClip := strings.Replace(c.lastClipText, "__ ", "", -1)
	log.Print("nameClip ", nameClip)
	c.drawText(
		"โดย",
		"DBHeaventRoundedMedv3.2.ttf",
		40.0,
		Margin{18, 0, 0, 0},
		TextClipOption{
			MaxLine: 1,
			NoClip:  true},
		Rectangle{
			Dimension{40, LAST_DIMENSION_HEIGHT},
			Point{LAST_POSITION_INNNER_BOUND_LEFT, LAST_POSITION_INNER_BOUND_TOP},
		},
		"#fff", gg.AlignLeft)
	c.drawText(
		nameClip,
		"DBHeaventRoundedMedv3.2.ttf",
		53.0,
		Margin{0, 0, 8, 13},
		TextClipOption{
			MaxLine: 1,
			NoClip:  true},
		Rectangle{
			Dimension{nameRectangle.Dimension.Width - 40, 53},
			Point{LAST_POSITION_BOUND_NEXT_TO_RIGHT, LAST_POSITION_INNER_BOUND_BOTTOM},
		},
		"#ffffff", gg.AlignLeft)

	c.savePNG("test.png")

	log.Print("Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))
}
