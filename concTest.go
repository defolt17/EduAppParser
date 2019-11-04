package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/firefox"
)

func _checkBasic(err error) {
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
}

func startNewRemote(caps selenium.Capabilities) {
	dr, err := selenium.NewRemote(caps, "")
	_checkBasic(err)
	fmt.Println(dr.CurrentURL())
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
		go startNewRemote(caps)
	}

	fmt.Scanln()
}
