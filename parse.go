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
	longWait   = time.Millisecond * 5000
	normalWait = time.Millisecond * 4000
	loginWait  = time.Millisecond * 1400
	shortWait  = time.Millisecond * 300
)

// Class is a class
type Class struct {
	name, date, weekday, auditory string
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
	user.SendKeys(os.Getenv("USERNAME"))

	pass, err := dr.FindElement(selenium.ByID, "password")
	_checkBasic(err)
	pass.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton, err := dr.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Войти')]")
	_checkBasic(err)
	loginButton.Click()

	time.Sleep(time.Millisecond * 300)
	log.Printf("Logged in as" + os.Getenv("USERNAME"))
}

// Kinda works now
// Get list of upcoming classes (name, date, time, cabinet)
func getUpcommingClasses(dr selenium.WebDriver) {
	loadPage(shortWait, dr, "https://my.informatics.ru/pupil/root")
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
}

func main() {
	err := godotenv.Load()
	_checkBasic(err)

	cb := selenium.Capabilities{"browserName": "firefox"}
	dr, err := selenium.NewRemote(cb, "")
	defer dr.Quit()

	loginMain(dr)
	getUpcommingClasses(dr)

}
