package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

var (
	currentLink  = ""
	urlPrefix    = "https://my.informatics.ru"
	loginLink    = "/accounts/root_login/"
	mainPageLink = "/pupil/root/"
	coursesLink  = "/pupil/courses/"
	profileLink  = "/accounts/"
)

// Student is a user's data class
type Student struct {
	Name    string
	Mail    string
	Phone   string
	Courses []Course
}

func (student *Student) getStudent(dr selenium.WebDriver) {
	loadPage(dr, profileLink)
	nameEx := FindElementWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[1]/div/div[2]/div[1]/div/div[2]/div/div[1][@class='lead']")
	student.Name, _ = nameEx.Text()
}

// Course is a course class
type Course struct {
	ID                                uint16 //This is convinient for distinguishing courses with same name
	Name                              string
	Part                              uint8
	GradeCount                        uint16
	AvgGrade                          float32
	VisitedClasses, NumClassesOverall uint16
	MainTeacher                       string
	Year                              string
	Link                              string
	Lessons                           []Lesson
	LessonsCount                      uint16
}

func (student *Student) getCourses(dr selenium.WebDriver, caps selenium.Capabilities, cookies *[]selenium.Cookie) {
	loadPage(dr, coursesLink)

	exYears := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[2]/ul//a[@href='javascript:void(0)']")

	yearsNum := len(exYears)
	yearsIndex := make([]string, yearsNum)
	years := make([]string, yearsNum)

	//var ID uint16
	//var Courses []Course

	for i, elem := range exYears {
		yearsIndex[i], _ = elem.GetAttribute("data-value")
		years[i], _ = elem.Text()
		years[i] = strings.Replace(years[i], "/", ":", 1)
	}

	for _, yearIndex := range yearsIndex {
		go func(caps *selenium.Capabilities, yearIndex *string, cookies *[]selenium.Cookie) {
			dr, _ := selenium.NewRemote(*caps, "")
			for _, cookie := range *cookies {
				dr.AddCookie(&cookie)
			}

			println(*yearIndex)
			loadPage(dr, coursesLink+"?year_selection="+*yearIndex)
			//pageCourses := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[contains(@class, 'panel panel-default')]")
			time.Sleep(time.Second * 5)
			dr.Quit()
		}(&caps, &yearIndex, cookies)
	}
	fmt.Scanln()
}

// Lesson is a lesson class
type Lesson struct {
	Theme, Date, Weekday, Link string
	Points                     uint16
	Material                   []Block
}

// Block is class for bubble in Material
// Homework/Classwork etc.
type Block struct {
	Name           string
	ExpirationDate string
	ExpirationType string
	Steps          []Step
}

// Step is a small paragraph inside a Block
type Step struct {
	Name  string
	Link  string
	Items []Item
}

// Item is an under paragrapgh inside a paragraph
// I will add item recognition by icon
type Item struct {
	Name string
	Icon string
	Link string
	Type string
}

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}

func loadPage(dr selenium.WebDriver, destLink string) {
	if currentLink != destLink {
		dr.Get(urlPrefix + destLink)
		log.Println("Loaded: " + destLink)
		return
	}
	log.Println("Staying on: " + currentLink)
}

func refreshPage(dr selenium.WebDriver) {
	dr.Get(urlPrefix + currentLink)
}

// FindElementWD is super-duper smart recurent solution to find element
// if it hasn't been loaded yet
func FindElementWD(dr selenium.WebDriver, qType string, q string) selenium.WebElement {
	res, err := dr.FindElement(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		res = FindElementWD(dr, qType, q)
	}
	return res
}

// FindElementsWD is basicly the same as FindElemet, but with "s"
// Actually it's quite scary, cause if elements will load at different
// time all of them won't be returned //oAo\\
func FindElementsWD(dr selenium.WebDriver, qType string, q string) []selenium.WebElement {
	res, err := dr.FindElements(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		res = FindElementsWD(dr, qType, q)
	}
	return res
}

// FindElementWE ...
func FindElementWE(dr selenium.WebElement, qType string, q string) selenium.WebElement {
	res, err := dr.FindElement(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		res = FindElementWE(dr, qType, q)
	}
	return res
}

// FindElementsWE ...
func FindElementsWE(dr selenium.WebElement, qType string, q string) []selenium.WebElement {
	res, err := dr.FindElements(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		res = FindElementsWE(dr, qType, q)
	}
	return res
}

func loginMain(dr selenium.WebDriver) {
	err := godotenv.Load()
	_checkBasic(err)
	loadPage(dr, loginLink)

	userC := make(chan selenium.WebElement)
	passC := make(chan selenium.WebElement)
	loginButtonC := make(chan selenium.WebElement)
	go func() {
		user := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[2]/input[@name='username']")
		userC <- user
	}()
	go func() {
		pass := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[3]/input[@name='password']")
		passC <- pass
	}()
	go func() {
		loginButton := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[2]/button[contains(text(), 'Войти')]")
		loginButtonC <- loginButton
	}()
	for i := 0; i < 2; i++ {
		select {
		case user := <-userC:
			user.SendKeys(os.Getenv("USERNAME"))
		case pass := <-passC:
			pass.SendKeys(os.Getenv("USERPASSWORD"))
		}
	}
	loginButton := <-loginButtonC
	loginButton.Click()

	time.Sleep(time.Millisecond * 400)
	currentLink = "/pupil/root/"
	log.Printf("Logged in as: " + os.Getenv("USERNAME"))

}

func main() {
	var student Student
	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			//"--headless",
			//"--private",
		},
		Prefs: map[string]interface{}{
			"browser.cache.disk.enable":                false,
			"browser.cache.memory.enable":              false,
			"browser.cache.disk.smart_size.enabled":    false,
			"browser.sessionhistory.max_total_viewers": 0,
			"browser.tabs.animate":                     false,
			"browser.sessionstore.max_concurrent_tabs": 0,
			"browser.cache.memory.capacity":            0,
			"network.prefetch-next":                    false,
			"config.trim_on_minimize":                  true,
			"network.http.pipelining":                  true,
			"network.http.pipelining.maxrequests":      50,
		},
	}
	caps.AddFirefox(firefoxCaps)

	dr, err := selenium.NewRemote(caps, "")
	_checkBasic(err)

	loginMain(dr)
	cookies, _ := dr.GetCookies()
	student.getStudent(dr)
	student.getCourses(dr, caps, &cookies)

	dr.Quit()

}
