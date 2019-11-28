package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

type Course struct {
	Name           string
	Link           string
	GradeCount     int
	GradeAvg       float32
	LessonsVisited int
	LessonsOverall int
	Teacher        string
}

func main() {
	log.Println("Starting Ultimate EduApp parser!")
	var token, username, userpassword string

	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	username = os.Getenv("USERNAME")
	userpassword = os.Getenv("USERPASSWORD")

	token = getToken(username, userpassword)

	doc := loadPage(mshpURL+mshpCoursesPageURL, &token)
	coursesYears := getYearLinks(&doc)

	courses := make([]Course, 0)

	for _, year := range *coursesYears {
		courses = append(courses, getCoursesFromYearPage(&token, &year)...)
	}

	d := 0
	for _, el := range courses {
		d += parseCoursePage(&token, &el.Link)
	}
	fmt.Println(d)

}

func getToken(username, userpassword string) string {

	log.Println("Trying to get user token....")
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

	log.Println("Successfully got user token.")

	return m["token"]
}

func getYearLinks(doc *goquery.Document) *[]Year {
	log.Print("Getting courses years....")
	years := make([]Year, 0, 6)
	doc.Find("div div div div ul li a").Each(func(index int, item *goquery.Selection) {
		year := item.Text()
		link, exists := item.Attr("data-value")
		if exists {
			years = append(years, Year{Year: year, Link: link})
		}
	})
	log.Println("Successfully got courses years.")
	return &years
}

func getCoursesFromYearPage(token *string, year *Year) []Course {

	courses := make([]Course, 0)

	doc := loadPage(mshpURL+mshpCoursesPageURL+"?year_selection="+year.Link, token)

	doc.Find(".panel-top-primary").Each(func(index int, block *goquery.Selection) {
		courses = append(courses, parseCourseBlock(block))
	})

	doc.Find(".panel-top-amazing").Each(func(index int, block *goquery.Selection) {
		courses = append(courses, parseCourseBlock(block))
	})

	return courses

}

func parseCourseBlock(block *goquery.Selection) Course {
	course := Course{}

	courseNameEx := block.Find("div div .panel-title a")
	courseNameText := courseNameEx.Text()

	courseLinkText, linkExists := courseNameEx.Attr("href")
	courseGradeCountEx := block.Find("div div div div div div b")
	courseGradeCountText := courseGradeCountEx.Text()
	courseGradeAvgEx := block.Find(".shp-average")
	courseGradeAvgText := courseGradeAvgEx.Text()
	courseLessonsEx := block.Find(".col-lg-4.col-xs-no-padding")
	courseLessonsVisited := 0
	courseLessonsOverall := 0
	courseTeacherText := "-"

	courseLessonsEx.Each(func(_ int, el *goquery.Selection) {
		courseSelectionText := el.Text()
		if strings.Contains(courseSelectionText, "из") {
			courseLessonsArr := strings.Split(courseSelectionText, "из")
			courseLessonsArr[0] = strings.TrimSpace(courseLessonsArr[0])
			courseLessonsArr[1] = strings.TrimSpace(courseLessonsArr[1])
			courseLessonsVisited, _ = strconv.Atoi(courseLessonsArr[0])
			courseLessonsOverall, _ = strconv.Atoi(courseLessonsArr[1])
		}
	})

	courseTeacherEx := block.Find(".media-body.lead.small")
	courseTeacherEx.Each(func(_ int, el *goquery.Selection) {
		s := el.Text()
		s = strings.Split(s, "  ")[1]
		s = strings.TrimSpace(s)
		courseTeacherText = s
	})

	course.Name = courseNameText

	if linkExists {
		course.Link = courseLinkText
	} else {
		log.Fatalln("Could not find link for", courseNameText)
	}

	courseGradeCountInt, err := strconv.Atoi(courseGradeCountText)
	if err != nil {
		log.Print("Could not convert grade count at ", courseNameText, ". Setting default value 0\n")
		courseGradeCountInt = 0
	}
	course.GradeCount = courseGradeCountInt

	courseGradeAvgFloat, err := strconv.ParseFloat(courseGradeAvgText, 32)
	if err != nil {
		log.Print("Could not convert grade averege at ", courseNameText, ". Setting default value 0.0\n")
		courseGradeAvgFloat = 0.0
	}
	course.GradeAvg = (float32)(courseGradeAvgFloat)
	course.LessonsVisited = courseLessonsVisited
	course.LessonsOverall = courseLessonsOverall
	course.Teacher = courseTeacherText

	return course
}

func parseCoursePage(token *string, link *string) int {

	doc := loadPage(mshpURL+*link, token)

	c := 0
	doc.Find("ul .rounded").Each(func(_ int, el *goquery.Selection) {
		c++
	})

	d := 0
	for i := 1; i < c+1; i++ {

		coursePartPage := loadPage(mshpURL+*link+"?part_selection="+strconv.Itoa(i), token)
		coursePartPage.Find("tbody tr").Each(func(_ int, _ *goquery.Selection) {
			d++
		})
		d--
	}
	return d
}

func loadPage(URL string, token *string) goquery.Document {
	client := &http.Client{}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Loading ", URL, "\n")
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

	return *doc
}
