package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
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

func startNewRemote(caps selenium.Capabilities, id int) {
	dr, err := selenium.NewRemote(caps, "")
	_checkBasic(err)

	dr.Get("https://my.informatics.ru/accounts/root_login/")
	println("ID: ", id, 1)

	user := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[2]/input[@name='username']")
	user.SendKeys(os.Getenv("USERNAME"))
	println("ID: ", id, 2)
	pass := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[1]/div[3]/input[@name='password']")
	pass.SendKeys(os.Getenv("USERPASSWORD"))
	println("ID: ", id, 3)
	loginButton := FindElementWD(dr, selenium.ByXPATH, "/html/body/div[2]/div/div[3]/div[2]/div/div/form/div[2]/button[contains(text(), 'Войти')]")
	loginButton.Click()
	println("ID: ", id, 4)

	fmt.Println("Worker: ", id, " found name: ")

	dr.Quit()
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

	for i := 0; i < 4; i++ {
		go startNewRemote(caps, i)
	}

	fmt.Scanln()
}
