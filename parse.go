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

func (student *Student) getCoursesMaterial(dr selenium.WebDriver) {
	for i := 0; i < len(student.Courses); i++ {
		for j := 0; j < len(student.Courses[i].Lessons); j++ {
			student.Courses[i].Lessons[j].getMaterial(dr)
		}
	}
}

// Lesson is a lesson class
type Lesson struct {
	Theme, Date, Weekday, Link string
	Points                     uint16
	Material                   []Block
}

func (lesson *Lesson) getMaterial(dr selenium.WebDriver) {
	loadPage(dr, lesson.Link)
	time.Sleep(time.Millisecond * 400)
	materialBlock := FindElementWD(dr, selenium.ByXPATH, "/html/body/div/div/div[4]/div/div/div/div[2]/div/div/div[2]/div[2]/div/div[2]/div/div[3]")
	blocksEx := FindElementsWE(materialBlock, selenium.ByXPATH, "./div")
	var blocks []Block

	if len(blocksEx) > 0 {
		for _, blockEx := range blocksEx {
			nameEx := FindElementWE(blockEx, selenium.ByXPATH, "./div/div[1]/div[1]/div/a/span[2]")
			nameText, _ := nameEx.Text()
			var tempBlock Block
			tempBlock.Name = nameText
			tempBlock.getSteps(blockEx)
			blocks = append(blocks, tempBlock)
		}
	}
	lesson.Material = blocks
}

// Block is class for bubble in Material
// Homework/Classwork etc.
type Block struct {
	Name           string
	ExpirationDate string
	ExpirationType string
	Steps          []Step
}

func (block *Block) getSteps(blockEx selenium.WebElement) {
	stepsEx := FindElementsWE(blockEx, selenium.ByXPATH, "./div/div[2]/ul/li")
	var steps []Step
	for _, stepEx := range stepsEx {
		var step Step
		stepNameEx := FindElementWE(stepEx, selenium.ByXPATH, "./a/span")
		stepLinkEx := FindElementWE(stepEx, selenium.ByXPATH, "./a")

		step.Name, _ = stepNameEx.Text()
		step.Link, _ = stepLinkEx.GetAttribute("href")
		step.getItems(stepEx)
		steps = append(steps, step)
	}
	block.Steps = steps
}

// Step is a small paragraph inside a Block
type Step struct {
	Name  string
	Link  string
	Items []Item
}

func (step *Step) getItems(stepEx selenium.WebElement) {
	itemsEx := FindElementsWE(stepEx, selenium.ByXPATH, "./ul/li")
	var items []Item
	for _, itemEx := range itemsEx {
		var item Item
		itemNameEx := FindElementWE(itemEx, selenium.ByXPATH, "./a/div/div/span")
		itemLinkEx := FindElementWE(itemEx, selenium.ByXPATH, "./a")
		itemIconEx := FindElementWE(itemEx, selenium.ByXPATH, "./a/div/div/i")

		tempString, _ := itemIconEx.GetAttribute("class")
		tempStringSplit := strings.Fields(tempString)

		itemNameText, _ := itemNameEx.Text()
		itemLinkText, _ := itemLinkEx.GetAttribute("href")
		itemIconText := tempStringSplit[1]

		item.Name = itemNameText
		item.Link = itemLinkText
		item.Icon = itemIconText

		items = append(items, item)
	}
	step.Items = items
}

// Item is an under paragrapgh inside a paragraph
// I will add item recognition by icon
type Item struct {
	Name string
	Icon string
	Link string
	Type string
}

// ContentText is a text based content
type ContentText struct {
	Text string
}

// ContentTask is a draft for task block
// its values and solution send mechanism
// (PRBLY in future I will do this)
type ContentTask struct {
	CheckType       string
	IOType          string
	TimeLimit       uint16
	MemoryLimit     uint16
	Condition       string
	InputDataFormat string
	IODataExample   string
	Points          uint16
	PointsOverall   uint16
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

	var ID uint16
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
			ID++
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
			gradeCountText, err := gradeCountEx.Text()
			gradeCountNum, err := strconv.Atoi(gradeCountText)
			if err != nil {
				tempCourse.GradeCount = 0
			} else {
				tempCourse.GradeCount = (uint16)(gradeCountNum)
			}
			avgGradeText, err := avgGradeEx.Text()
			avgGradeNum, err := strconv.Atoi(avgGradeText)
			if err != nil {
				tempCourse.AvgGrade = 0.0
			} else {
				tempCourse.AvgGrade = (float32)(avgGradeNum)
			}
			visitedClassesNum, err := strconv.Atoi(classesVisitsTextList[0])
			if err != nil {
				tempCourse.VisitedClasses = 0
			} else {
				tempCourse.VisitedClasses = (uint16)(visitedClassesNum)
			}
			NumClassesOverallNum, err := strconv.Atoi(classesVisitsTextList[1])
			if err != nil {
				tempCourse.NumClassesOverall = 0
			} else {
				tempCourse.NumClassesOverall = (uint16)(NumClassesOverallNum)
			}
			tempCourse.MainTeacher = mainTeacherText
			tempCourse.Year = years[i]
			tempCourse.Link = linkText
			tempCourse.ID = ID

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
			tempPartedCourses.Part = (byte)(partNum)
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

					tempPartedCourses.LessonsCount = (uint16)(len(lessonsListEx) - 1)

					tempPartedCourses.Lessons = append(tempPartedCourses.Lessons, tempLesson)
				}
			}
			partedCourses = append(partedCourses, tempPartedCourses)
		}
	}

	// Check for double (courses with lenght > 1 year)
	// Assuming its just 1 course
	// Does not work!!!!!!!!!!!!!!!!!!!!!
	for i, courseI := range partedCourses {
		for j, courseJ := range partedCourses {
			if courseJ.ID != courseI.ID {
				if courseI.Name == courseJ.Name {
					if courseI.Year != courseJ.Year {
						copy(partedCourses[i:], partedCourses[i+2:])
						partedCourses[len(partedCourses)-2] = Course{}
						partedCourses = partedCourses[:len(partedCourses)-2]

						copy(partedCourses[j+2:], partedCourses[i+2+2:])
						partedCourses[len(partedCourses)-2] = Course{}
						partedCourses = partedCourses[:len(partedCourses)-2]
					}
				}
			}
		}
	}
	student.Courses = partedCourses

}

func main() {

	caps := selenium.Capabilities{"browserName": "firefox"}
	firefoxCaps := firefox.Capabilities{
		Args: []string{
			//"--headless",
			//"-private",
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
			"network.http.pipelining.maxrequests":      10,
		},
	}
	caps.AddFirefox(firefoxCaps)
	dr, err := selenium.NewRemote(caps, "")
	_checkBasic(err)
	/*
		file, _ := ioutil.ReadFile("data.json")
		student := Student{}
		_ = json.Unmarshal([]byte(file), &student)
	*/
	var student Student

	loginMain(dr)
	student.getStudent(dr)
	student.getCourses(dr)
	/*
		c := 1
		for i := 0; i < len(student.Courses); i++ {
			for j := 0; j < len(student.Courses[i].Lessons); j++ {
				if c%35 == 0 {
					dr.Quit()
					dr, _ = selenium.NewRemote(caps, "")
					loginMain(dr)
					time.Sleep(time.Millisecond * 1500)
				}
				student.Courses[i].Lessons[j].getMaterial(dr)
				c++
			}
		}

		fmt.Println(c)

		jsonString, err := json.Marshal(student)
		_checkBasic(err)
		ioutil.WriteFile("DatsHowMafiaWorks.json", jsonString, os.ModePerm)
	*/
	for _, course := range student.Courses {
		fmt.Println(course.Name)
	}
	dr.Quit()
}
