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

type Year struct {
	Year string
	Link string
}

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
	coursesYears := getYearLinks(doc)

	for _, year := range *coursesYears {
		getCoursesYearPage(&token, &year)
	}

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

func getYearLinks(doc *goquery.Document) *[]Year {
	years := make([]Year, 0, 6)
	doc.Find("div div div div ul li a").Each(func(index int, item *goquery.Selection) {
		year := item.Text()
		link, exists := item.Attr("data-value")
		if exists {
			years = append(years, Year{Year: year, Link: link})
		}
	})
	return &years
}

func getCoursesYearPage(token *string, year *Year) {
	client := &http.Client{}

	URL := mshpURL + mshpCoursesPageURL + "?year_selection=" + year.Link
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("Cookie", "eduapp_jwt="+*token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	doc.Find(".panel-title").Each(func(index int, el *goquery.Selection) {
		courseNameEx := el.Find("a")
		courseNameText := courseNameEx.Text()
		courseLink, exists := courseNameEx.Attr("href")
		if exists {
			fmt.Println(courseNameText, courseLink)
		}
	})

}
