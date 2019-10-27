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

// Course is a course class # WOW realy?
type Course struct {
	Name                              string
	Part                              string
	GradeCount                        string
	AvgGrade                          string
	VisitedClasses, NumClassesOverall string
	MainTeacher                       string
	Year                              string
	Link                              string
	Lessons                           []Lesson
	LessonsCount                      string
}

func (student *Student) getStudent(dr selenium.WebDriver) {
	loadPage(dr, profileLink)

	nameEx := FindElementWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[1]/div/div[2]/div[1]/div/div[2]/div/div[1][@class='lead']")

	student.Name, _ = nameEx.Text()
}

// Lesson is a lesson class
type Lesson struct {
	Theme, Date, Weekday, Link string
	Points                     string
	Material                   []Block
}

func (lesson *Lesson) getMaterial(dr selenium.WebDriver) {
	loadPage(dr, lesson.Link)
	//time.Sleep(time.Millisecond * 2000) // #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!# #FIX IT!#
	materialBlock := FindElementWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div/div/div/div[2]/div/div/div[2]/div[2]/div/div[2]/div/div[3]")
	blocksEx := FindElementsWE(materialBlock, selenium.ByXPATH, "./div")
	var blocks []Block
	for _, blockEx := range blocksEx {
		nameEx := FindElementWE(blockEx, selenium.ByXPATH, "./div/div[1]/div[1]/div/a/span[2]")
		nameText, _ := nameEx.Text()
		var tempBlock Block
		tempBlock.Name = nameText
		tempBlock.Steps = "PASS ITS A TEMP DATA"

		blocks = append(blocks, tempBlock)
	}
	lesson.Material = blocks
}

// Block is class for bubble in Material
// Homework/Classwork etc.
type Block struct {
	Name  string
	Steps string //[]Step TEMP TEMP TEMP TEMP
}

// Step is a small paragraph inside a Block
type Step struct {
	Name  string
	Link  string
	items []Item
}

// Item is an under paragrapgh inside a paragraph
type Item struct {
	Name string
	Link string
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

	user := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[2]/input[@name='username']")
	user.SendKeys(os.Getenv("USERNAME"))

	pass := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[3]/input[@name='password']")
	pass.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[2]/button[contains(text(), 'Войти')]")
	loginButton.Click()

	currentLink = "/pupil/root/"
	log.Printf("Logged in as: " + os.Getenv("USERNAME"))
	time.Sleep(time.Millisecond * 300)
}

// "Ex" means extracted
func (student *Student) getCourses(dr selenium.WebDriver) {
	loadPage(dr, coursesLink)

	exYears := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[2]/ul//a[@href='javascript:void(0)']")

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
		loadPage(dr, coursesLink+"?year_selection="+yearIndex)
		pageCourses := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[contains(@class, 'panel panel-default')]")
		for _, pageCourse := range pageCourses {
			nameEx := FindElementWE(pageCourse, selenium.ByXPATH, "./div[1]/div[1]/div/a[contains(@href, '/pupil/courses/')]")
			gradeCountEx := FindElementWE(pageCourse, selenium.ByXPATH, "./div[2]/div[2]/div/div/div[1]/div[2]/*[contains(@class, 'shp-total-marks') or contains(@class, 'text-muted more-info')]")
			avgGradeEx := FindElementWE(pageCourse, selenium.ByXPATH, "./div[2]/div[2]/div/div/div[2]/div[2]/div/span[contains(@class, ' shp-average')]")
			classesVisits, err := pageCourse.FindElement(selenium.ByXPATH, "./div[2]/div[2]/div/div/div[5]/div[@class='col-lg-4 col-xs-no-padding']")
			var classesVisitsTextList [2]string
			if err != nil {
				classesVisitsTextList[0] = ""
				classesVisitsTextList[1] = ""
			} else {
				classesVisitsText, _ := classesVisits.Text()
				var classesVisitsTextList [2]string
				if classesVisitsText != "-" {
					tempClassesList := strings.Split(classesVisitsText, "из")
					classesVisitsTextList[0] = strings.Replace(tempClassesList[0], " ", "", 1)
					classesVisitsTextList[1] = strings.Replace(tempClassesList[1], " ", "", 1)
				} else {
					classesVisitsTextList[0] = ""
					classesVisitsTextList[1] = ""
				}
			}
			mainTeacher, err := pageCourse.FindElement(selenium.ByXPATH, "./div[2]/div[3]/div/div/div/div[2]/div[@class='media-body lead small']")
			var mainTeacherText string
			if err != nil {
				mainTeacherText = ""
			} else {
				mainTeacherText, _ = mainTeacher.Text()
				mainTeacherText = strings.Replace(mainTeacherText, "Основной преподаватель:", "", 1)
				mainTeacherText = strings.Replace(mainTeacherText, "\n", "", 1)
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

	// Gets info about lessons of course and copies course if
	// it has several parts
	var partedCourses []Course
	for i := range Courses {
		loadPage(dr, Courses[i].Link)
		time.Sleep(time.Millisecond * 500)
		coursePartBlockEx := FindElementWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[1]/div/ul[@class='nav nav-pills nav-select-filter inline-block']")
		coursePartsEx := FindElementsWE(coursePartBlockEx, selenium.ByXPATH, "./li/a[@href='javascript:void(0)']")
		for partNum := 1; partNum < len(coursePartsEx)+1; partNum++ {
			xpath := "/html/body/div/div/div[4]/div[2]/div/div[1]/div/ul/li[" + strconv.Itoa(partNum) + "]/a[contains(text(), '" + strconv.Itoa(partNum) + " часть')]"
			partButton := FindElementWD(dr, selenium.ByXPATH, xpath)
			partButton.Click()
			time.Sleep(time.Millisecond * 500)
			lessonsListEx := FindElementsWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[4]/table/tbody/tr[@class='clearfix']")
			var tempPartedCourses Course
			tempPartedCourses.Part = strconv.Itoa(partNum)
			tempPartedCourses.Name = Courses[i].Name
			tempPartedCourses.GradeCount = Courses[i].GradeCount
			tempPartedCourses.AvgGrade = Courses[i].AvgGrade
			tempPartedCourses.VisitedClasses = Courses[i].VisitedClasses
			tempPartedCourses.NumClassesOverall = Courses[i].NumClassesOverall
			tempPartedCourses.MainTeacher = Courses[i].MainTeacher
			tempPartedCourses.Year = Courses[i].Year
			tempPartedCourses.Link = Courses[i].Link
			tempPartedCourses.LessonsCount = Courses[i].LessonsCount

			if len(lessonsListEx) > 1 {
				for j := 1; j < len(lessonsListEx); j++ {
					lessonDateEx := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//div[@class='top-line']")
					lessonDateText, _ := lessonDateEx.Text()

					lessonWeekdayEx := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//span[@class='small text-muted']")
					lessonWeekdayText, _ := lessonWeekdayEx.Text()

					lessonThemeEx := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//div[@class='col-md-3 small col-xs-6 m-b']")
					lessonThemeText, _ := lessonThemeEx.Text()

					lessonLinkEx := FindElementWE(lessonsListEx[j], selenium.ByXPATH, ".//a[contains(@href, '/pupil/calendar/')]")
					lessonLinkText, _ := lessonLinkEx.GetAttribute("href")

					var tempLesson Lesson
					tempLesson.Date = lessonDateText
					tempLesson.Weekday = lessonWeekdayText
					tempLesson.Theme = lessonThemeText
					tempLesson.Link = lessonLinkText

					tempPartedCourses.Lessons = append(tempPartedCourses.Lessons, tempLesson)
				}
			}
			partedCourses = append(partedCourses, tempPartedCourses)
		}
	}
	student.Courses = partedCourses
}

func main() {
	cb := selenium.Capabilities{"browserName": "firefox"}
	dr, err := selenium.NewRemote(cb, "")
	_checkBasic(err)
	defer dr.Quit()

	var student Student

	loginMain(dr)
	//student.getStudent(dr)
	//student.getCourses(dr)

	var tempLesson Lesson

	tempLesson.Link = "/pupil/calendar/309336/"

	tempLesson.getMaterial(dr)

	fmt.Println(tempLesson.Material)

	jsonString, err := json.Marshal(student)
	_checkBasic(err)
	ioutil.WriteFile("data.json", jsonString, os.ModePerm)
}
