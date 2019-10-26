package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Salary ...
type Salary struct {
	basic float64
}

// Lesson is a lesson class
type Lesson struct {
	theme string
}

func fun() Salary {
	var data Salary
	data.basic = 123
	return data

}

func main() {

	data := fun()

	jsonString, _ := json.Marshal(data)
	fmt.Println(jsonString)
	fmt.Println(string(jsonString))
	ioutil.WriteFile("test4.json", jsonString, os.ModePerm)
}
