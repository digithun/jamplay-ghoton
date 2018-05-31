package main

import (
	"log"
	"os"

	"github.com/globalsign/mgo"
)

func initMongo() *mgo.Database {
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
