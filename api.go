package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// Constants
const (
	mshpURL            = "https://my.informatics.ru"
	mshpCoursesPageURL = "/pupil/courses/"
	mshpLoginPageURL   = "/accounts/root_login/"
	mshpLoginAPIURL    = "/api/v1/rest-auth/login/"
)

func main() {
	var token, username, userpassword string

	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	username = os.Getenv("USERNAME")
	userpassword = os.Getenv("USERPASSWORD")

	token = getToken(username, userpassword)

	doc := getCoursesPage(&token)
	getYearLinks(doc)

}

func getToken(username, userpassword string) string {

	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": userpassword,
		"captcha":  "false",
	})

	resp, err := http.Post(mshpURL+mshpLoginAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	s := string(body)

	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		log.Fatal(err)
	}

	return m["token"]
}

func getCoursesPage(token *string) *goquery.Document {
	client := &http.Client{}

	req, err := http.NewRequest("GET", mshpURL+mshpCoursesPageURL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Cookie", "eduapp_jwt="+*token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return document
}

func getYearLinks(doc *goquery.Document) {
	doc.Find("div div div div ul li a").Each(func(index int, item *goquery.Selection) {
		year := item.Text()
		link, exists := item.Attr("data-value")
		if exists {
			fmt.Println(year, link)
		}
	})
}
