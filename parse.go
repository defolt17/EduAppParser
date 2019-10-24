package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
)

const (
	// Time to wait till page loads all the elemets
	// bound by scripts, ex. Course page with tasks
	loginWait       = time.Millisecond * 1000
	mainWait        = time.Millisecond * 3
	coursesListWait = time.Millisecond * 5
	profileWait     = time.Millisecond * 2
	courseWait      = time.Millisecond * 1
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
	name    string
	mail    string
	phone   string
	courses []Course
}

// Course is a course class
type Course struct {
	name                              string
	gradeCount                        string
	avgGrade                          string
	visitedClasses, numClassesOverall string
	mainTeacher                       string // This is temporary
	year                              string // I will add pupil/teaher class
	link                              string
	classes                           []Class
}

// Class is a class class
type Class struct {
	name, date, weekday, auditory string
}

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
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

// ### Add succsessfull login verification
func loginMain(dr selenium.WebDriver) {
	err := godotenv.Load()
	_checkBasic(err)

	loadPage(loginWait, dr, loginLink)

	user, err := dr.FindElement(selenium.ByID, "username")
	_checkBasic(err)
	user.SendKeys(os.Getenv("USERNAME"))

	pass, err := dr.FindElement(selenium.ByID, "password")
	_checkBasic(err)
	pass.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton, err := dr.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Войти')]")
	_checkBasic(err)
	loginButton.Click()

	currentLink = "/pupil/root/"
	time.Sleep(time.Millisecond * 700)
	log.Printf("Logged in as: " + os.Getenv("USERNAME"))

}

func getUpcommingClasses(dr selenium.WebDriver) []Class {

	loadPage(mainWait, dr, "/pupil/root/")

	upcommingClasses, err := dr.FindElements(selenium.ByXPATH, "/html/body/div[1]/div/div[4]/div[2]/div/div[2]/div[1]/div/div[2]/div[@class='clearfix clickable nowrap']")
	_checkBasic(err)
	upcommingClassesList := make([]Class, len(upcommingClasses))

	for i, elem := range upcommingClasses {
		date, _ := elem.FindElement(selenium.ByXPATH, ".//div[@class='top-line']")
		weekday, _ := elem.FindElement(selenium.ByXPATH, ".//span[@class='small text-muted']")
		auditory, _ := elem.FindElement(selenium.ByXPATH, ".//span[@class='small']")
		courseName, _ := elem.FindElement(selenium.ByXPATH, ".//div[@class='col-xs-12 col-sm-6 col-xs-no-padding']")

		upcommingClassesList[i].name, _ = courseName.Text()
		upcommingClassesList[i].date, _ = date.Text()
		upcommingClassesList[i].weekday, _ = weekday.Text()
		upcommingClassesList[i].auditory, _ = auditory.Text()

	}

	for _, elem := range upcommingClassesList {
		fmt.Println(elem.name + " " + elem.date + ", " +
			elem.weekday + " " + elem.auditory)
	}
	return upcommingClassesList
}

func getCourses(dr selenium.WebDriver) []Course {

	loadPage(coursesListWait, dr, coursesLink)

	extractedYears, _ := dr.FindElements(selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[2]/ul//a[@href='javascript:void(0)']")
	yearsNum := len(extractedYears)

	yearsIndex := make([]string, yearsNum)
	years := make([]string, yearsNum)
	var Courses []Course

	for i, elem := range extractedYears {
		yearsIndex[i], _ = elem.GetAttribute("data-value")
		years[i], _ = elem.Text()
		years[i] = strings.Replace(years[i], "/", ":", 1)
	}

	for i, yearIndex := range yearsIndex {
		loadPage(coursesListWait, dr, coursesLink+"?year_selection="+yearIndex)
		pageCourses, _ := dr.FindElements(selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[contains(@class, 'panel panel-default')]")
		for _, pageCourse := range pageCourses {
			nameExtracted, err := pageCourse.FindElement(selenium.ByXPATH, ".//a[contains(@href, '/pupil/courses/')]")
			gradeCountExtracted, err := pageCourse.FindElement(selenium.ByXPATH, ".//*[contains(@class, 'shp-total-marks') or contains(@class, 'text-muted more-info')]")
			avgGradeExtracted, err := pageCourse.FindElement(selenium.ByXPATH, ".//span[contains(@class, ' shp-average')]")
			classesVisits, err := pageCourse.FindElement(selenium.ByXPATH, ".//div[@class='col-lg-4 col-xs-no-padding']")

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

			linkText, err := nameExtracted.GetAttribute("href")

			var tempCourse Course
			tempCourse.name, err = nameExtracted.Text()
			tempCourse.gradeCount, err = gradeCountExtracted.Text()
			tempCourse.avgGrade, err = avgGradeExtracted.Text()
			tempCourse.visitedClasses = classesVisitsTextList[0]
			tempCourse.numClassesOverall = classesVisitsTextList[1]
			tempCourse.mainTeacher = mainTeacherText

			tempCourse.year = years[i]
			tempCourse.link = linkText

			Courses = append(Courses, tempCourse)
		}
	}

	for _, course := range Courses {
		loadPage(courseWait, dr, course.link)
		lessonsListExtracted, _ := dr.FindElements(selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[4]/table/tbody/tr")

		fmt.Println(len(lessonsListExtracted))
		//Done for today
	}

	return Courses

}

func getStudent(dr selenium.WebDriver) Student {
	var student Student
	loadPage(profileWait, dr, profileLink)

	nameExtracted, _ := dr.FindElement(selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[1]/div/div[2]/div[1]/div/div[2]/div/div[1][@class='lead']")
	student.name, _ = nameExtracted.Text()

	student.courses = getCourses(dr)

	return student
}

func main() {

	cb := selenium.Capabilities{"browserName": "firefox"}
	dr, err := selenium.NewRemote(cb, "")
	_checkBasic(err)
	defer dr.Quit()

	loginMain(dr)
	me := getStudent(dr)
	fmt.Println(me)

}
