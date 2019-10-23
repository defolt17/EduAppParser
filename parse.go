package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
)

const (
	// Time to wait till page loads all the elemets
	// bound by scripts, ex. Course page with tasks
	loginWait       = time.Millisecond * 1100
	mainWait        = time.Millisecond * 200
	coursesListWait = time.Millisecond * 500
)

var (
	currentLink  = ""
	urlPrefix    = "https://my.informatics.ru"
	loginLink    = "/accounts/root_login/"
	mainPageLink = "/pupil/root/"
	coursesLink  = "/pupil/courses/"
)

// Course is a course class
type Course struct {
	name                              string
	gradeCount                        uint32
	avgGrade                          float32
	visitedClasses, numClassesOverall uint16
	mainTeacher                       string // This is temporary
	stoopidImageLink                  string // I will add pupil/teaher class
	year                              string
	link                              string
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

func getCourses(dr selenium.WebDriver) {

	loadPage(coursesListWait, dr, coursesLink)

	extractedYears, _ := dr.FindElements(selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[2]/ul//a[@href='javascript:void(0)']")
	yearsNum := len(extractedYears)

	yearsIndex := make([]string, yearsNum)
	years := make([]string, yearsNum)
	var Courses []Course

	for i, elem := range extractedYears {
		yearsIndex[i], _ = elem.GetAttribute("data-value")
		years[i], _ = elem.Text()
	}

	for _, yearIndex := range yearsIndex {
		loadPage(coursesListWait, dr, coursesLink+"?year_selection="+yearIndex)
		pageCourses, _ := dr.FindElements(selenium.ByXPATH, "/html/body/div/div/div[4]/div[2]/div/div[contains(@class, 'panel panel-default')]")
		fmt.Println(len(pageCourses))
		for _, pageCourse := range pageCourses {
			nameExtracted, _ := pageCourse.FindElement(selenium.ByXPATH, ".//a[contains(@href, '/pupil/courses/')]")
			var tempCourse Course
			tempCourse.name, _ = nameExtracted.Text()
			Courses = append(Courses, tempCourse)
		}
	}
	for _, course := range Courses {
		fmt.Println(course.name)
		// Done for today
	}

}

func main() {

	cb := selenium.Capabilities{"browserName": "firefox"}
	dr, err := selenium.NewRemote(cb, "")
	_checkBasic(err)
	defer dr.Quit()

	loginMain(dr)
	getCourses(dr)

}
