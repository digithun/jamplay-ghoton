package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
)

func clearCache(key string) {

	cloudflare_clearcache_url := os.Getenv("CLOUDFLARE_CLEARCACHE_URL")
	cloudflare_auth_email := os.Getenv("CLOUDFLARE_AUTH_EMAIL")
	cloudflare_auth_key := os.Getenv("CLOUDFLARE_AUTH_KEY")

	// log.Print("cloudflare_clearcache_url ", cloudflare_clearcache_url)
	// log.Print("cloudflare_auth_email ", cloudflare_auth_email)
	// log.Print("cloudflare_auth_key ", cloudflare_auth_key)

	if len(cloudflare_clearcache_url)+len(cloudflare_auth_email)+len(cloudflare_auth_key) > 3 {

		client := &http.Client{Timeout: time.Duration(5 * time.Second)}

		url := strings.Join([]string{"https://static.jamplay.world/", key}, "")

		jsonData, _ := bson.MarshalJSON(bson.M{"files": []string{url}})
		body := bytes.NewBuffer(jsonData)

		req, _ := http.NewRequest(http.MethodDelete, cloudflare_clearcache_url, body)
		req.Header.Set("X-Auth-Email", cloudflare_auth_email)
		req.Header.Set("X-Auth-Key", cloudflare_auth_key)
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			log.Print("cloudflare client req err", err)
		}

		defer req.Body.Close()

		log.Print("res.StatusCode ", res.StatusCode)
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Print("cloudflare err io ", err)
		}

		bodyString := string(bodyBytes)
		log.Print("cloudflare res ", bodyString)
	}
}
