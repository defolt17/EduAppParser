package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
)

var (
	// If something goes wrong "retry" function will
	// try to repeat action for this amount of times
	attempts = 5
	// Time to wait till page loads all the elemets
	// bound by scripts -- Course page with tasks
	longWait   = time.Millisecond * 5100
	normalWait = time.Millisecond * 4000
	loginWait  = time.Millisecond * 1400
)

// Need to use it somehow
func try(f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}
		if i >= (attempts - 1) {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func loadPage(dt time.Duration, dr selenium.WebDriver, link string) {
	dr.Get(link)
	time.Sleep(dt)
}

func loginMain(dr selenium.WebDriver) {
	loadPage(loginWait, dr, "https://my.informatics.ru/accounts/root_login/#/")

	user, err := dr.FindElement(selenium.ByID, "username")
	_checkBasic(err)
	user.Clear()
	user.SendKeys(os.Getenv("USERNAME"))

	pass, err := dr.FindElement(selenium.ByID, "password")
	_checkBasic(err)
	pass.Clear()
	pass.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton, err := dr.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Войти')]")
	_checkBasic(err)
	loginButton.Click()

	time.Sleep(time.Millisecond * 1000)
}

// TEMP FUNCTION NOT DONE YET
// Get list of upcoming classes (name, date, time, cabinet)
func getUpcommingClasses(dr selenium.WebDriver) {
	loadPage(normalWait, dr, "https://my.informatics.ru/pupil/root")
	upcommingClasses, err := dr.FindElements(selenium.ByXPATH, "//div[@class='clearfix clickable nowrap']")
	_checkBasic(err)
	upcommingClassesList := make([]string, len(upcommingClasses))

	for i, elem := range upcommingClasses {
		date, _ := elem.FindElement(selenium.ByXPATH, "//div[@class='top-line']")
		weekday, _ := elem.FindElement(selenium.ByXPATH, "//span[@class='small text-muted']")
		auditory, _ := elem.FindElement(selenium.ByXPATH, "//span[@class='small']")
		courseName, _ := elem.FindElement(selenium.ByXPATH, "//div[@class='col-xs-12 col-sm-6 col-xs-no-padding']")

		dateText, _ := date.Text()
		weekdayText, _ := weekday.Text()
		auditoryText, _ := auditory.Text()
		courseNameText, _ := courseName.Text()

		upcommingClassesList[i] = courseNameText + dateText + weekdayText + auditoryText
	}
	fmt.Println(upcommingClassesList)
}

func main() {
	err := godotenv.Load()
	_checkBasic(err)

	cb := selenium.Capabilities{"browserName": "firefox"}
	dr, err := selenium.NewRemote(cb, "")
	defer dr.Quit()

	loginMain(dr)
	getUpcommingClasses(dr)

	time.Sleep(time.Second * 100)

}
