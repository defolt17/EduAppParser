package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

const (
	authorizationResponseWait = time.Millisecond * 50
)

var (
	currentLink = ""
)

// URLs
const (
	URLPrefix    = "https://my.informatics.ru"
	loginLink    = "/accounts/root_login/"
	mainPageLink = "/pupil/root/"
	coursesLink  = "/pupil/courses/"
	profileLink  = "/accounts/"
)

// Xpath constants
const (
	loginUsernameFieldXpath          = "./html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[2]/input[@name='username']"
	loginPasswordFieldXpath          = "./html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[3]/input[@name='password']"
	loginSubmitButtonXpath           = "./html/body/div[2]/div/div[3]/div[2]/div/div/form/div[2]/button[contains(text(), 'Войти')]"
	rootRealNameXpath                = "./html/body/div[1]/div/div[2]/nav/ul[3]/li[2]/a/span"
	profileNameXpath                 = "/html/body/div/div/div[4]/div[2]/div/div[1]/div/div[2]/div[1]/div/div[2]/div/div[1][@class='lead']"
	profileAvatarXpath               = "/html/body/div/div/div[4]/div[2]/div/div[1]/div/div[2]/div[1]/div/div[1]/div/img"
	coursesYearButtonXpath           = "/html/body/div/div/div[4]/div[2]/div/div[2]/ul//a[@href='javascript:void(0)']"
	coursesYearCourseNameXpath       = "./div[1]/div[1]/div/a[contains(@href, '/pupil/courses/')]"
	coursesYearCourseGradeCount      = "./div[2]/div[2]/div/div/div[1]/div[2]/*[contains(@class, 'shp-total-marks') or contains(@class, 'text-muted more-info')]"
	coursesYearCourseAvgGrade        = "./div[2]/div[2]/div/div/div[2]/div[2]/div/span[contains(@class, ' shp-average')]"
	coursesYearCourseVisits          = "./div[2]/div[2]/div/div/div[5]/div[@class='col-lg-4 col-xs-no-padding']"
	coursesYearCourseMainTeacherName = "./div[2]/div[3]/div/div/div/div[2]/div[@class='media-body lead small']"
	coursesYearCourseBlock           = "/html/body/div/div/div[4]/div[2]/div/div[contains(@class, 'panel panel-default')]"
)

// Student is a user's data class
type Student struct {
	Name         string
	AvatarURL    string
	Courses      []Course
	CoursesCount uint16
}

// Course is a course class
type Course struct {
	ID                                uint16
	Name                              string
	Part                              uint8
	StartDate                         string
	EndDate                           string
	GradeCount                        uint16
	AvgGrade                          float32
	VisitedClasses, NumClassesOverall uint16
	MainTeacher                       string
	Year                              string
	Link                              string
	Lessons                           []Lesson
	LessonsCount                      uint16
}

// Lesson is a lesson class
type Lesson struct {
	Theme, Date, Weekday, Link string
	// Material                   []Block
}

func main() {
	log.Println("Ultimate EduApp parser started!")
	dr := getFirefoxWebDriver()
	defer dr.Quit()
	var student Student

	loginMain(&dr)
	student.getInfo(&dr)
	student.getCourses(&dr)

	for _, course := range student.Courses {
		fmt.Println(course)
	}

	// To be continued....
}

func loginMain(dr *selenium.WebDriver) {
	err := godotenv.Load()
	_checkBasic(err)

	loadPage(dr, loginLink)

	usernameField := FindElementWD(dr, selenium.ByXPATH, loginUsernameFieldXpath)
	usernameField.SendKeys(os.Getenv("USERNAME"))

	passwordField := FindElementWD(dr, selenium.ByXPATH, loginPasswordFieldXpath)
	passwordField.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton := FindElementWD(dr, selenium.ByXPATH, loginSubmitButtonXpath)
	loginButton.Click()

	// Wait till root page is loaded
	// Actually you have to wait for a bit after
	// clicking "Login" button, otherwise authorization
	// is going to fail
	currentLink, err := (*dr).CurrentURL()
	_checkBasic(err)

	for currentLink != URLPrefix+mainPageLink {
		time.Sleep(authorizationResponseWait)
		currentLink, err = (*dr).CurrentURL()
		_checkBasic(err)
	}

	// realName := FindElementWD(*dr, selenium.ByXPATH, // rootRealNameXpath)
	// realNameText, err := realName.Text()
	//_checkBasic(err)

	currentLink = "/pupil/root/"
	log.Println("Logged in as: ", os.Getenv("USERNAME"))
	// log.Println("Real name: ", realNameText)
	// IDK it doesn't work
}

func (student *Student) getInfo(dr *selenium.WebDriver) {
	loadPage(dr, profileLink)

	nameEx := FindElementWD(dr, selenium.ByXPATH, profileNameXpath)
	nameText, err := nameEx.Text()
	_checkBasic(err)

	avatarURLEx := FindElementWD(dr, selenium.ByXPATH, profileAvatarXpath)
	avatarURLText, err := avatarURLEx.GetAttribute("src")
	_checkBasic(err)

	student.AvatarURL = avatarURLText
	student.Name = nameText
}

func (student *Student) getCourses(dr *selenium.WebDriver) {
	loadPage(dr, coursesLink)

	coursesYearsButtonsEx := FindElementsWD(dr, selenium.ByXPATH, coursesYearButtonXpath)
	coursesYearsLink := make([]string, len(coursesYearsButtonsEx))
	coursesYears := make([]string, len(coursesYearsButtonsEx))

	for i, coursesYearButtonEx := range coursesYearsButtonsEx {
		tempCourseYearLink, err := coursesYearButtonEx.GetAttribute("data-value")
		_checkBasic(err)
		coursesYearsLink[i] = tempCourseYearLink
		tempYear, err := coursesYearButtonEx.Text()
		_checkBasic(err)
		coursesYears[i] = strings.Replace(tempYear, "/", ":", 1)
	}

	for _, coursesYearLink := range coursesYearsLink {
		student.Courses = append(student.Courses, getCoursesFromYearPage(dr, &coursesYearLink)...)
	}

}

func getCoursesFromYearPage(dr *selenium.WebDriver, coursesYearLink *string) []Course {
	loadPage(dr, coursesLink+"?year_selection="+*coursesYearLink)

	coursesBlocks := FindElementsWD(dr, selenium.ByXPATH, coursesYearCourseBlock)
	yearCourses := make([]Course, 0, len(coursesBlocks))
	for _, courseBlock := range coursesBlocks {
		yearCourses = append(yearCourses, getCourseFromYearBlock(&courseBlock))
	}
	return yearCourses
}

func getCourseFromYearBlock(webEl *selenium.WebElement) Course {

	var course Course
	var mainTeacherText string

	courseNameEx := FindElementWE(webEl, selenium.ByXPATH, coursesYearCourseNameXpath)
	courseNameText, err := courseNameEx.Text()
	_checkBasic(err)

	gradeCountEx := FindElementWE(webEl, selenium.ByXPATH, coursesYearCourseGradeCount)
	gradeCountText, err := gradeCountEx.Text()
	_checkBasic(err)

	avgGradeEx := FindElementWE(webEl, selenium.ByXPATH, coursesYearCourseAvgGrade)
	avgGradeText, err := avgGradeEx.Text()
	_checkBasic(err)

	var classesVisitsTextList []string
	var visitedClassesNum, classesOverallNum int
	classesVisits, err := (*webEl).FindElement(selenium.ByXPATH, coursesYearCourseVisits)
	if err != nil {
		visitedClassesNum = 0
		classesOverallNum = 0
	} else {
		classesVisitsText, err := classesVisits.Text()
		_checkBasic(err)

		classesVisitsTextList = strings.Split(classesVisitsText, " из ")
		visitedClassesNum, _ = strconv.Atoi(classesVisitsTextList[0])
		classesOverallNum, _ = strconv.Atoi(classesVisitsTextList[1])
	}

	mainTeacherEx, err := (*webEl).FindElement(selenium.ByXPATH, coursesYearCourseMainTeacherName)
	if err != nil {
		mainTeacherText = ""
	} else {
		mainTeacherText, _ = mainTeacherEx.Text()
		mainTeacherText = strings.Replace(mainTeacherText, "Основной преподаватель:", "", 1)
		mainTeacherText = strings.Replace(mainTeacherText, "\n", "", 1)
	}

	courseLinkText, err := courseNameEx.GetAttribute("href")
	course.Name = courseNameText
	course.Link = courseLinkText
	course.NumClassesOverall = (uint16)(classesOverallNum)
	course.VisitedClasses = (uint16)(visitedClassesNum)
	course.MainTeacher = mainTeacherText

	gradeCountNum, err := strconv.Atoi(gradeCountText)
	if err != nil {
		course.GradeCount = 0
	} else {
		course.GradeCount = (uint16)(gradeCountNum)
	}

	avgGradeNum, err := strconv.ParseFloat(avgGradeText, 32)
	if err != nil {
		course.AvgGrade = 0.0
	} else {
		course.AvgGrade = (float32)(avgGradeNum)
	}

	return course
}

func FindElementWE(webEl *selenium.WebElement, qType string, q string) selenium.WebElement {
	res, err := (*webEl).FindElement(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res = FindElementWE(webEl, qType, q)
	}
	return res
}

func FindElementsWE(webEl *selenium.WebElement, qType string, q string) []selenium.WebElement {
	res, err := (*webEl).FindElements(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res = FindElementsWE(webEl, qType, q)
	}
	return res
}

func FindElementWD(dr *selenium.WebDriver, qType string, q string) selenium.WebElement {
	res, err := (*dr).FindElement(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res = FindElementWD(dr, qType, q)
	}
	return res
}

func FindElementsWD(dr *selenium.WebDriver, qType string, q string) []selenium.WebElement {
	res, err := (*dr).FindElements(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res = FindElementsWD(dr, qType, q)
	}
	return res
}

func loadPage(dr *selenium.WebDriver, destLink string) {
	if currentLink != destLink {
		(*dr).Get(URLPrefix + destLink)
		log.Println("Loaded: " + destLink)
		return
	}
	log.Println("Staying on: " + currentLink)
}

func getFirefoxWebDriver() selenium.WebDriver {
	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{
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
			"network.http.pipelining.maxrequests":      10,
		},
	}
	caps.AddFirefox(firefoxCaps)
	dr, err := selenium.NewRemote(caps, "")
	_checkBasic(err)
	return dr
}

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}
