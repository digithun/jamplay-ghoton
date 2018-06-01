package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fanliao/go-promise"
	"github.com/globalsign/mgo/bson"
)

// func imageKeyToStaticURL(s *string) {
// 	if strings.Index(*s, "http") == -1 {
// 		newString := strings.Join([]string{"https://static.jamplay.world/", *s}, "")
// 		*s = newString
// 	}
// }

func renderAllBook() {

	db := initMongo()
	books := db.C("books")
	bookQuery := []bson.M{
		{"$match": bson.M{
			"deleted": false,
			// "_id":     bson.ObjectIdHex("5a999c590ad136000f518fe3"),
		}},
		{"$lookup": bson.M{
			"from":         "authors",
			"localField":   "authorId",
			"foreignField": "_id",
			"as":           "author",
		}},
		{"$unwind": "$author"},
		{"$project": bson.M{
			"_id":         1,
			"title":       1,
			"coverImage":  "$thumbnailImage.url",
			"authorName":  "$author.name",
			"authorImage": "$author.profilePicture.url",
			"category":    bson.M{"$arrayElemAt": []interface{}{"$categories", 0}},
		}},
		// {"$limit": 5},
	}

	p := books.Pipe(bookQuery)
	p = p.Batch(10)
	iter := p.Iter()

	type bookResult struct {
		Id          bson.ObjectId `bson:"_id"`
		Title       string        `bson:"title"`
		CoverImage  string        `bson:"coverImage"`
		AuthorImage string        `bson:"authorImage"`
		AuthorName  string        `bson:"authorName"`
		Category    string        `bson:"category"`
	}
	t1 := time.Now().UnixNano()

	var mongoResult bookResult
	for iter.Next(&mongoResult) {

		// if len(mongoResult.AuthorImage) > 0 {
		// 	imageKeyToStaticURL(&mongoResult.AuthorImage)
		// }

		// if len(mongoResult.CoverImage) > 0 {
		// 	imageKeyToStaticURL(&mongoResult.CoverImage)
		// } else
		if len(mongoResult.CoverImage) == 0 {
			mongoResult.CoverImage = strings.Join([]string{"asset/default-book-cover/", strings.ToLower(mongoResult.Category), ".png"}, "")
		}

		log.Print("mongoResult Title   ", mongoResult.Title)
		log.Print("mongoResult AuthorImage   ", mongoResult.AuthorImage)
		log.Print("mongoResult CoverImage   ", mongoResult.CoverImage)

		log.Print("mongoResult.AuthorName  ", mongoResult.AuthorName)
		log.Print("mongoResult.Category  ", mongoResult.Category)

		savePath := strings.Join([]string{"share/book/", mongoResult.Id.Hex(), ".jpg"}, "")

		renderShareBook(mongoResult.AuthorName, mongoResult.Title, mongoResult.CoverImage, mongoResult.AuthorImage, "", savePath, nil)
	}
	log.Print("Render Whole - Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))

}

func renderAllAuthor() {

	db := initMongo()
	authors := db.C("authors")

	authorQuery := []bson.M{
		// {"$match": bson.M{
		// 	"_id": bson.ObjectIdHex("599ff17af5be162888e5e14f"),
		// }},
		{"$project": bson.M{
			"_id":         1,
			"name":        1,
			"authorImage": "$profilePicture.url",
		}},
		// {"$limit": 5},
	}

	p := authors.Pipe(authorQuery)
	p = p.Batch(10)
	iter := p.Iter()

	type authorResult struct {
		Id          bson.ObjectId `bson:"_id"`
		AuthorName  string        `bson:"name"`
		AuthorImage string        `bson:"authorImage"`
	}
	t1 := time.Now().UnixNano()

	var mongoResult authorResult
	for iter.Next(&mongoResult) {

		log.Print("mongoResult Name   ", mongoResult.AuthorName)
		log.Print("mongoResult AuthorImage   ", mongoResult.AuthorImage)

		savePath := strings.Join([]string{"share/author/", mongoResult.Id.Hex(), ".jpg"}, "")

		renderShareAuthor(mongoResult.AuthorName, mongoResult.AuthorImage, savePath, nil)
	}
	log.Print("Render Whole - Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))

}

func renderShareBook(authorName, title, coverImage, authorImage, category, savePath string, w io.Writer) {

	log.Print("renderShareBook rendering: ", savePath)
	t1 := time.Now().UnixNano()
	// Set your Google Cloud Platform project ID.
	c := NewCanvas(Dimension{1200, 630})
	// c.debug = true

	// authorName = "OResia a b c d e f g h i j k l m n o p q r s t u v w x y z"
	// title = "A Fairy Promise   a b c d e f g h i j k l m n o p q r s t u v w x y z a b c d e f g h i j"

	//default if not known
	if len(savePath) == 0 {
		savePath = strings.Join([]string{
			"result/book/",
			strconv.FormatInt(time.Now().Unix(), 10),
			".png",
		}, "")
	}

	c.drawImage(
		"asset/share_template_book.png", "",
		Margin{}, Rectangle{
			Dimension{1200, 630},
			Point{0, 0},
		}, 0).
		drawImage(
			coverImage, "asset/default-pic-cover-book.png",
			Margin{},
			Rectangle{
				Dimension{405, 590},
				Point{20, 22},
			}, 9).
		drawText(
			title,
			"DBHeaventRoundedMedv3.2.ttf",
			68.0,
			Margin{0, 20, 17, 20},
			TextClipOption{
				MaxLine:   2,
				ClipWidth: c.width - 405 - 20,
				// OverFlowOption: TEXT_CLIP_OVERFLOW_ELLIPSIS,
			},
			Rectangle{
				Dimension{c.width - 405 - 20, TEXT_HEIGHT_MAX_LINE},
				Point{LAST_POSITION_BOUND_NEXT_TO_RIGHT, 69},
			},
			"#fff",
			TEXT_ALLIGN_HORIZONTAL_CENTER, TEXT_ALLIGN_VERTICAL_CENTER).
		drawImage(
			authorImage, "asset/default-pic-editor.png",
			Margin{},
			Rectangle{
				Dimension{220, 220},
				Point{LAST_POSITION_BOUND_HORIZONTAL_CENTER, LAST_POSITION_BOUND_NEXT_TO_BOTTOM}},
			110).
		drawText(
			strings.Join([]string{"__ ", authorName}, ""),
			"DBHeaventRoundedMedv3.2.ttf",
			53.0,
			Margin{5, 0, 0, 0},
			TextClipOption{
				MaxLine:   1,
				ClipWidth: c.width - 405 - 40,
				// OverFlowOption: TEXT_CLIP_OVERFLOW_ELLIPSIS,
			},
			Rectangle{
				Dimension{TEXT_WIDTH_TOFIT, 53},
				Point{LAST_POSITION_BOUND_HORIZONTAL_CENTER, LAST_POSITION_BOUND_NEXT_TO_BOTTOM},
			},
			"#55555500", TEXT_ALLIGN_HORIZONTAL_CENTER, TEXT_ALLIGN_VERTICAL_DEFAULT)

	nameRectangle := c.lastRenderBound
	c.saveLastBound(nameRectangle)
	// log.Print("c.lastClipText ", c.lastClipText)

	nameClip := strings.Replace(c.lastClipText, "__ ", "", -1)
	// log.Print("nameClip ", nameClip)
	c.drawText(
		"โดย",
		"DBHeaventRoundedMedv3.2.ttf",
		40.0,
		Margin{21, 0, 0, 0},
		TextClipOption{
			MaxLine: 1,
			NoClip:  true},
		Rectangle{
			Dimension{45, LAST_DIMENSION_HEIGHT},
			Point{LAST_POSITION_INNNER_BOUND_LEFT, LAST_POSITION_INNER_BOUND_BOTTOM},
		},
		"#fff",
		TEXT_ALLIGN_HORIZONTAL_LEFT, TEXT_ALLIGN_VERTICAL_DEFAULT)

	c.drawText(
		nameClip,
		"DBHeaventRoundedMedv3.2.ttf",
		53.0,
		Margin{8, 0, 0, 13},
		TextClipOption{
			MaxLine:        1,
			ClipWidth:      nameRectangle.Dimension.Width - c.lastRenderBound.Dimension.Width,
			OverFlowOption: TEXT_CLIP_OVERFLOW_ELLIPSIS,
			NoClip:         true,
		},
		Rectangle{
			Dimension{nameRectangle.Dimension.Width - c.lastRenderBound.Dimension.Width, 60},
			Point{LAST_POSITION_BOUND_NEXT_TO_RIGHT, LAST_POSITION_INNER_BOUND_TOP},
		},
		"#ffffff",
		TEXT_ALLIGN_HORIZONTAL_LEFT, TEXT_ALLIGN_VERTICAL_DEFAULT)

	os.MkdirAll(filePathToDirPath(savePath), 0755)

	log.Print("savePath ", savePath)

	if len(savePath) > 0 {
		c.saveFile(savePath)
		handleUpload(savePath)
	}

	if w != nil {
		c.toStream("jpg", w)
	}

	log.Print("Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))
}

func renderShareAuthor(authorName, authorImage, savePath string, w io.Writer) {

	log.Print("renderShareAuthor rendering: ", savePath)
	t1 := time.Now().UnixNano()
	// Set your Google Cloud Platform project ID.
	c := NewCanvas(Dimension{1200, 630})
	// c.debug = true

	//default if not known
	if len(savePath) == 0 {
		savePath = strings.Join([]string{
			"result/book/",
			strconv.FormatInt(time.Now().Unix(), 10),
			".png",
		}, "")
	}

	c.drawImage(
		"asset/share_template_author.png", "",
		Margin{},
		Rectangle{
			Dimension{1200, 630},
			Point{0, 0},
		}, 0).
		drawImage(
			authorImage, "asset/default-pic-editor.png",
			Margin{},
			Rectangle{
				Dimension{336, 336},
				Point{96, LAST_POSITION_BOUND_VERTICAL_CENTER}},
			335/2).
		drawText(
			authorName,
			"DBHeaventRoundedMedv3.2.ttf",
			76.0,
			Margin{0, 0, 0, 0},
			TextClipOption{
				MaxLine:        1,
				ClipWidth:      c.width - (96 + 336 + 20 + 20),
				OverFlowOption: TEXT_CLIP_OVERFLOW_ELLIPSIS,
			},
			Rectangle{
				Dimension{c.width - (96 + 336 + 20 + 20), 88},
				Point{96 + 336 + 20, LAST_POSITION_BOUND_VERTICAL_CENTER},
			},
			"#000", TEXT_ALLIGN_HORIZONTAL_CENTER, 0)

	os.MkdirAll(filePathToDirPath(savePath), 0755)

	log.Print("savePath ", savePath)

	if len(savePath) > 0 {
		c.saveFile(savePath)
		handleUpload(savePath)
	}

	if w != nil {
		c.toStream("jpg", w)
	}

	log.Print("Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))
}

func handleUpload(savePath string) {
	promise.WhenAll(
		func() (r interface{}, err error) {
			uploadFileToS3(savePath, savePath)
			os.Remove(savePath)

			return "ok1", nil
		},
		func() (r interface{}, err error) {
			clearCache(savePath)
			return "ok2", nil
		}).Get()
}
