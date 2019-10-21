package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/tebeka/selenium"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	caps := selenium.Capabilities{"browserName": "firefox"}
	wd, _ := selenium.NewRemote(caps, "")
	defer wd.Quit()

	wd.Get("https://my.informatics.ru/accounts/root_login/#/")

	time.Sleep(time.Millisecond * 1500)

	user, _ := wd.FindElement(selenium.ByID, "username")
	user.Clear()
	user.SendKeys(os.Getenv("USERNAME"))

	pass, _ := wd.FindElement(selenium.ByID, "password")
	pass.Clear()
	pass.SendKeys(os.Getenv("USERPASSWORD"))

	loginButton, _ := wd.FindElement(selenium.ByXPATH, "//*[contains(text(), 'Войти')]")
	loginButton.Click()
	time.Sleep(time.Second * 10)

}
