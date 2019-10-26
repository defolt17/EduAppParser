package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
)

const (
	// Time to wait till page loads all the elemets
	// bound by scripts, ex. Course page with tasks
	loginWait            = time.Millisecond * 1000
	mainWait             = time.Millisecond * 300
	zeroWait             = time.Nanosecond * 1
	coursesListWait      = time.Millisecond * 1
	profileWait          = time.Millisecond * 1
	courseWait           = time.Millisecond * 400
	courseChangePartWait = time.Millisecond * 550
)

var (
	currentLink  = ""
	urlPrefix    = "https://my.informatics.ru"
	loginLink    = "/accounts/root_login/"
	mainPageLink = "/pupil/root/"
	coursesLink  = "/pupil/courses/"
	profileLink  = "/accounts/"
)

// Student is a students class
type Student struct {
	Name    string
	Mail    string
	Phone   string
	Courses []Course
}

// Course is a course class
type Course struct {
	Name                              string
	GradeCount                        string
	AvgGrade                          string
	VisitedClasses, NumClassesOverall string
	MainTeacher                       string // This is temporary
	Year                              string // I will add pupil/teaher class
	Link                              string
	Lessons                           []Lesson
	LessonsCount                      string
}

// Lesson is a lesson class
type Lesson struct {
	Theme, Date, Weekday, Link, Part string
}

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}

func loadPage(dt time.Duration, dr selenium.WebDriver, destLink string) {
	if currentLink != destLink {
		dr.Get(urlPrefix + destLink)
		time.Sleep(dt)
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
func FindElementWD(dr selenium.WebDriver, qType string, q string) (selenium.WebElement, error) {
	res, err := dr.FindElement(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 100)
		res, err = FindElementWD(dr, qType, q)
	}
	return res, err
}

// FindElementsWD is basicly the same as FindElemet, but with "s"
// Actually it's quite scary, cause if elements will load at different
// time all of them won't be returned //oAo\\
func FindElementsWD(dr selenium.WebDriver, qType string, q string) ([]selenium.WebElement, error) {
	res, err := dr.FindElements(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res, err = FindElementsWD(dr, qType, q)
	}
	return res, err
}

// FindElementWE ...
func FindElementWE(dr selenium.WebElement, qType string, q string) (selenium.WebElement, error) {
	res, err := dr.FindElement(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res, err = FindElementWE(dr, qType, q)
	}
	return res, err
}

// FindElementsWE ...
func FindElementsWE(dr selenium.WebElement, qType string, q string) ([]selenium.WebElement, error) {
	res, err := dr.FindElements(qType, q)
	if err != nil {
		time.Sleep(time.Millisecond * 50)
		res, err = FindElementsWE(dr, qType, q)
	}
	return res, err
}

func loginMain(dr selenium.WebDriver) {
	err := godotenv.Load()
	_checkBasic(err)

	loadPage(zeroWait, dr, loginLink)

	user, _ := FindElementWD(dr, selenium.ByID, "username")
	user.SendKeys(os.Getenv("USERNAME"))

	pass, _ := FindElementWD(dr, selenium.ByID, "password")
	pass.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton, _ := FindElementWD(dr, selenium.ByXPATH, "//*[contains(text(), 'Войти')]")
	loginButton.Click()

	currentLink = "/pupil/root/"
	log.Printf("Logged in as: " + os.Getenv("USERNAME"))
	time.Sleep(time.Millisecond * 100)
}

/*
Up to refactoring

func getUpcommingClasses(dr selenium.WebDriver) []Lesson {

	loadPage(mainWait, dr, "/pupil/root/")

	upcommingClasses, err := dr.FindElements(selenium.ByXPATH, "/html/body/div[1]/div/div[4]/div[2]/div/div[2]/div[1]/div/div[2]/div[@class='clearfix clickable nowrap']")
	_checkBasic(err)
	upcommingClassesList := make([]Lesson, len(upcommingClasses))

	for i, elem := range upcommingClasses {
		date, _ := elem.FindElement(selenium.ByXPATH, ".//div[@class='top-line']")
		weekday, _ := elem.FindElement(selenium.ByXPATH, ".//span[@class='small text-muted']")
		auditory, _ := elem.FindElement(selenium.ByXPATH, ".//span[@class='small']")
		courseName, _ := elem.FindElement(selenium.ByXPATH, ".//div[@class='col-xs-12 col-sm-6 col-xs-no-padding']")

		upcommingClassesList[i].Name, _ = courseName.Text()
		upcommingClassesList[i].date, _ = date.Text()
		upcommingClassesList[i].weekday, _ = weekday.Text()
		upcommingClassesList[i].auditory, _ = auditory.Text()

	}

	for _, elem := range upcommingClassesList {
		fmt.Println(elem.Name + " " + elem.date + ", " +
			elem.weekday + " " + elem.auditory)
	}
	return upcommingClassesList
}
*/

func getCourses(dr selenium.WebDriver, student *Student) {

	loadPage(coursesListWait, dr, coursesLink)

	exYears, _ := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[2]/ul//a[@href='javascript:void(0)']")

	yearsNum := len(exYears)
	yearsIndex := make([]string, yearsNum)
	years := make([]string, yearsNum)

	var Courses []Course

	for i, elem := range exYears {
		yearsIndex[i], _ = elem.GetAttribute("data-value")
		years[i], _ = elem.Text()
		years[i] = strings.Replace(years[i], "/", ":", 1)
	}

	for i, yearIndex := range yearsIndex {
		loadPage(coursesListWait, dr, coursesLink+"?year_selection="+yearIndex)
		pageCourses, _ := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[contains(@class, 'panel panel-default')]")
		for _, pageCourse := range pageCourses {
			nameEx, err := FindElementWE(pageCourse, selenium.ByXPATH, ".//a[contains(@href, '/pupil/courses/')]")
			gradeCountEx, err := FindElementWE(pageCourse, selenium.ByXPATH, ".//*[contains(@class, 'shp-total-marks') or contains(@class, 'text-muted more-info')]")
			avgGradeEx, err := FindElementWE(pageCourse, selenium.ByXPATH, ".//span[contains(@class, ' shp-average')]")
			classesVisits, err := FindElementWE(pageCourse, selenium.ByXPATH, ".//div[@class='col-lg-4 col-xs-no-padding']")

			mainTeacher, err := pageCourse.FindElement(selenium.ByXPATH, ".//div[@class='media-body lead small']")
			var mainTeacherText string
			if err != nil {
				mainTeacherText = ""
			} else {
				mainTeacherText, _ = mainTeacher.Text()
				mainTeacherText = strings.Replace(mainTeacherText, "Основной преподаватель:", "", 1)
				mainTeacherText = strings.Replace(mainTeacherText, "\n", "", 1)
			}

			classesVisitsText, err := classesVisits.Text()
			var classesVisitsTextList [2]string
			if classesVisitsText != "-" {
				tempClassesList := strings.Split(classesVisitsText, "из")
				classesVisitsTextList[0] = strings.Replace(tempClassesList[0], " ", "", 1)
				classesVisitsTextList[1] = strings.Replace(tempClassesList[1], " ", "", 1)
			} else {
				classesVisitsTextList[0] = ""
				classesVisitsTextList[1] = ""
			}

			linkText, err := nameEx.GetAttribute("href")

			var tempCourse Course
			tempCourse.Name, err = nameEx.Text()
			tempCourse.GradeCount, err = gradeCountEx.Text()
			tempCourse.AvgGrade, err = avgGradeEx.Text()
			tempCourse.VisitedClasses = classesVisitsTextList[0]
			tempCourse.NumClassesOverall = classesVisitsTextList[1]
			tempCourse.MainTeacher = mainTeacherText
			tempCourse.Year = years[i]
			tempCourse.Link = linkText

			Courses = append(Courses, tempCourse)
		}
	}

	for i := range Courses {
		loadPage(courseWait, dr, Courses[i].Link)

		coursePartBlockEx, _ := FindElementWD(dr, selenium.ByXPATH, "//ul[@class='nav nav-pills nav-select-filter inline-block']")
		coursePartsEx, _ := FindElementsWE(coursePartBlockEx, selenium.ByXPATH, ".//a[@href='javascript:void(0)']")

		for partNum := 1; partNum < len(coursePartsEx)+1; partNum++ {
			xpath := "//*[contains(text(), '" + strconv.Itoa(partNum) + " часть')]"
			partButton, _ := FindElementWD(dr, selenium.ByXPATH, xpath)
			partButton.Click()
			time.Sleep(courseChangePartWait)

			lessonsListEx, _ := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[4]/table/tbody/tr[@class='clearfix']")

			if len(lessonsListEx) > 1 {
				for j := 1; j < len(lessonsListEx); j++ {
					lessonDateEx, _ := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//div[@class='top-line']")
					lessonDateText, _ := lessonDateEx.Text()

					lessonWeekdayEx, _ := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//span[@class='small text-muted']")
					lessonWeekdayText, _ := lessonWeekdayEx.Text()

					lessonThemeEx, _ := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//div[@class='col-md-3 small col-xs-6 m-b']")
					lessonThemeText, _ := lessonThemeEx.Text()

					lessonLinkEx, _ := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//a[contains(@href, '/pupil/calendar/')]")
					lessonLinkText, _ := lessonLinkEx.GetAttribute("href")

					var tempLesson Lesson
					tempLesson.Date = lessonDateText
					tempLesson.Weekday = lessonWeekdayText
					tempLesson.Theme = lessonThemeText
					tempLesson.Link = lessonLinkText
					tempLesson.Part = strconv.Itoa(partNum) + " часть"

					Courses[i].Lessons = append(Courses[i].Lessons, tempLesson)
				}
			}
		}
	}

	student.Courses = Courses
	for i := 0; i < len(Courses); i++ {
		student.Courses[i].LessonsCount = strconv.Itoa(len(student.Courses[i].Lessons))
	}

}

func getStudent(dr selenium.WebDriver) Student {
	var student Student
	loadPage(profileWait, dr, profileLink)

	nameEx, _ := FindElementWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[1]/div/div[2]/div[1]/div/div[2]/div/div[1][@class='lead']")
	student.Name, _ = nameEx.Text()

	getCourses(dr, &student)

	return student
}

func main() {

	cb := selenium.Capabilities{"browserName": "firefox"}
	dr, err := selenium.NewRemote(cb, "")
	_checkBasic(err)
	defer dr.Quit()

	loginMain(dr)

	me := getStudent(dr)

	jsonString, err := json.Marshal(me)
	_checkBasic(err)
	fmt.Println(jsonString)
	ioutil.WriteFile("data.json", jsonString, os.ModePerm)
	// Oh shi7 oh flick oh shi7 oh flick
	// Try ~№50 COMPLETE!

}
