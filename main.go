package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/joho/godotenv"
)

func main() {
	//load ENV
	err := godotenv.Load(".env")
	if err != nil {
		log.Print("Error loading .env file.")
	}

	// tWordWrap("ตัดคำ ตัดคำตัด คำ")

	renderAllBook()
}

func imageKeyToStaticURL(s *string) {
	if strings.Index(*s, "http") == -1 {
		newString := strings.Join([]string{"https://static.jamplay.world/", *s}, "")
		*s = newString
	}
}

func initMongo(collectonName string) *mgo.Database {
	mongoURL := os.Getenv("MONGO_URL")
	// log.Print("mongoURL", mongoURL)

	if len(mongoURL) <= len("mongodb://") {
		log.Print("ENV MONGO_URL is not correctly defined")
		return nil
	}

	session, err := mgo.Dial(mongoURL)
	if err != nil {
		log.Print("mgo.Dial failed : ", err)
		return nil
	}

	mongoDB := os.Getenv("MONGO_DB")

	if len(mongoDB) == 0 {
		log.Print("ENV MONGO_DB is not correctly defined")
		return nil
	}
	session.SetMode(mgo.Monotonic, true)

	return session.DB(mongoDB)
}

func renderAllBook() {

	db := initMongo("books")

	books := db.C("books")

	// log.Print("books ", books)
	// a:=[]interface{}{"$categories", 0}

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
			"coverImage":  "$thumbnailImage.key",
			"authorName":  "$author.name",
			"authorImage": "$author.profilePicture.key",
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

		if len(mongoResult.AuthorImage) > 0 {
			imageKeyToStaticURL(&mongoResult.AuthorImage)
		}

		if len(mongoResult.CoverImage) > 0 {
			imageKeyToStaticURL(&mongoResult.CoverImage)
		} else if len(mongoResult.CoverImage) == 0 {
			mongoResult.CoverImage = strings.Join([]string{"asset/default-book-cover/", strings.ToLower(mongoResult.Category), ".png"}, "")
		}
		log.Print("mongoResult Title   ", mongoResult.Title)
		log.Print("mongoResult AuthorImage   ", mongoResult.AuthorImage)
		log.Print("mongoResult CoverImage   ", mongoResult.CoverImage)

		log.Print("mongoResult.AuthorName  ", mongoResult.AuthorName)
		log.Print("mongoResult.Category  ", mongoResult.Category)

		savePath := strings.Join([]string{"render/book/", mongoResult.Id.Hex(), ".png"}, "")

		renderShareBook(mongoResult.AuthorName, mongoResult.Title, mongoResult.CoverImage, mongoResult.AuthorImage, savePath)
	}
	log.Print("Render Whole - Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))

}

func renderShareBook(authorName, title, coverImage, authorImage, savePath string) {

	log.Print("rendering: ", savePath)
	t1 := time.Now().UnixNano()
	// Set your Google Cloud Platform project ID.
	c := NewCanvas(Dimension{1200, 630})
	// c.debug = true

	// authorName = "OResia a b c d e f g h i j k l m n o p q r s t u v w x y z"
	// title = "A Fairy Promise   a b c d e f g h i j k l m n o p q r s t u v w x y z a b c d e f g h i j"

	if len(coverImage) == 0 {
		coverImage = "asset/default-pic-cover-book.png"
	}
	if len(authorImage) == 0 {
		authorImage = "asset/default-pic-editor.png"
	}
	if len(savePath) == 0 {
		savePath = strings.Join([]string{
			"result/book/",
			strconv.FormatInt(time.Now().Unix(), 10),
			".png",
		}, "")
	}

	c.drawImage("asset/share_template_book.png", Margin{}, Rectangle{
		Dimension{1200, 630},
		Point{0, 0},
	}, 0).
		drawImage(
			coverImage,
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
			authorImage,
			Margin{},
			Rectangle{
				Dimension{220, 220},
				Point{LAST_POSITION_BOUND_VERTICAL_CENTER, LAST_POSITION_BOUND_NEXT_TO_BOTTOM}},
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
				Point{LAST_POSITION_BOUND_VERTICAL_CENTER, LAST_POSITION_BOUND_NEXT_TO_BOTTOM},
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
			MaxLine: 1,
			NoClip:  true,
		},
		Rectangle{
			Dimension{nameRectangle.Dimension.Width - c.lastRenderBound.Dimension.Width + 10, 60},
			Point{LAST_POSITION_BOUND_NEXT_TO_RIGHT, LAST_POSITION_INNER_BOUND_TOP},
		},
		"#ffffff",
		TEXT_ALLIGN_HORIZONTAL_LEFT, TEXT_ALLIGN_VERTICAL_DEFAULT)

	os.MkdirAll(filePathToDirPath(savePath), 0755)

	log.Print("savePath ", savePath)
	c.savePNG(savePath)

	log.Print("Time taken ms: ", (time.Now().UnixNano()-t1)/int64(time.Millisecond))
}
