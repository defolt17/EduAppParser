package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// Constants
const (
	mshpURL                = "https://my.informatics.ru"
	mshpCoursesPageURL     = "/pupil/courses/"
	mshpLoginPageURL       = "/accounts/root_login/"
	mshpLoginAPIURL        = "/api/v1/rest-auth/login/"
	mshpGetClassesAPIURL   = "/api/v1/teaching_situation/classes_users/headings/"
	mshpGetClassInfoAPIURL = "api/v1/teaching_situation/classes/extended/"
)

type ClassRequest struct {
	Count   int
	Results []ClassFromRequest
}

type ClassFromRequest struct {
	ID int
}

type ClassInfoFromRequest struct {
	Count    int         `json:"count"`
	Previous interface{} `json:"previous"`
	Results  []struct {
		TeacherID        int         `json:"teacher_id"`
		TotalPupils      int         `json:"total_pupils"`
		DatetimeBegin    time.Time   `json:"datetime_begin"`
		ClassroomID      int         `json:"classroom_id"`
		IsEvaluation     bool        `json:"is_evaluation"`
		Duration         int         `json:"duration"`
		DatetimeBeginMin interface{} `json:"datetime_begin_min"`
		DateOf           string      `json:"date_of"`
		ID               int         `json:"id"`
		CuratorID        interface{} `json:"curator_id"`
		Course           struct {
			IsExam           bool        `json:"is_exam"`
			SchoolSubjectID  int         `json:"school_subject_id"`
			OriginalCourseID interface{} `json:"original_course_id"`
			Name             string      `json:"name"`
			SubjectLineID    int         `json:"subject_line_id"`
			DateFrom         string      `json:"date_from"`
			DateTill         string      `json:"date_till"`
			SchoolID         int         `json:"school_id"`
			FullName         string      `json:"full_name"`
			Usage            string      `json:"usage"`
			SimplestName     string      `json:"simplest_name"`
			ID               int         `json:"id"`
			SimpleName       string      `json:"simple_name"`
		} `json:"course"`
		DatetimeEnd    time.Time `json:"datetime_end"`
		ClassesLessons []struct {
			Lesson struct {
				Priority int    `json:"priority"`
				ID       int    `json:"id"`
				Name     string `json:"name"`
			} `json:"lesson"`
			LessonID        int `json:"lesson_id"`
			ID              int `json:"id"`
			ClassesID       int `json:"classes_id"`
			LessonVariantID int `json:"lesson_variant_id"`
		} `json:"classes_lessons"`
		OriginalClassesID  interface{} `json:"original_classes_id"`
		PassesScore        int         `json:"passes_score"`
		ConsultationReview interface{} `json:"consultation_review"`
		ClassesTeachers    []struct {
			TeacherID   int  `json:"teacher_id"`
			IsOnClasses bool `json:"is_on_classes"`
			ClassesID   int  `json:"classes_id"`
			Teacher     struct {
				Surname string `json:"surname"`
				Name    string `json:"name"`
				User    struct {
					Username                    string      `json:"username"`
					FirstName                   string      `json:"first_name"`
					LastName                    string      `json:"last_name"`
					Patronymic                  string      `json:"patronymic"`
					PreferSchoolID              interface{} `json:"prefer_school_id"`
					SubscribedToEmailDeliveries bool        `json:"subscribed_to_email_deliveries"`
					Provider                    string      `json:"provider"`
					SubscribedToPhoneDeliveries bool        `json:"subscribed_to_phone_deliveries"`
					CreatedBy                   string      `json:"created_by"`
					IsTeacher                   bool        `json:"is_teacher"`
					GradeNum                    string      `json:"grade_num"`
					SubscribedToDeliveries      bool        `json:"subscribed_to_deliveries"`
					LoginLast                   time.Time   `json:"login_last"`
					Avatar                      interface{} `json:"avatar"`
					Email                       string      `json:"email"`
					ContactNumber               interface{} `json:"contact_number"`
					IsClient                    bool        `json:"is_client"`
					ID                          int         `json:"id"`
					IsRepresentative            bool        `json:"is_representative"`
					IsPupil                     bool        `json:"is_pupil"`
				} `json:"user"`
				Patronymic string `json:"patronymic"`
				IsPhantom  bool   `json:"is_phantom"`
				ID         int    `json:"id"`
			} `json:"teacher"`
			Role      string `json:"role"`
			ID        int    `json:"id"`
			IsTeacher bool   `json:"is_teacher"`
		} `json:"classes_teachers"`
		ExpectedPupils int `json:"expected_pupils"`
		TotalScore     int `json:"total_score"`
		CourseID       int `json:"course_id"`
		TimingID       int `json:"timing_id"`
	} `json:"results"`
	Next interface{} `json:"next"`
}

type Year struct {
	Year string
	Link string
}

type Course struct {
	Name          string
	ID            string
	GradeCount    int
	GradeAvg      float32
	ClasssVisited int
	ClasssOverall int
	Teacher       string
	Classs        []Class
}

type Class struct {
	ID string
}

type Date struct {
	Day   int
	Month int
	Year  int
}

func main() {
	log.Println("Starting Ultimate EduApp parser!")
	var token, username, userpassword string

	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	username = os.Getenv("USERNAME")
	userpassword = os.Getenv("USERPASSWORD")

	token = getToken(username, userpassword)

	doc := loadPage(mshpURL+mshpCoursesPageURL, &token)
	coursesYears := getYearLinks(&doc)

	courses := make([]Course, 0)

	for _, year := range *coursesYears {
		courses = append(courses, getCoursesFromYearPage(&token, &year)...)
	}

	// getCourseClasses(&courses[1], &token)

	fmt.Println(courses)
}

func getToken(username, userpassword string) string {

	log.Println("Trying to get user token....")
	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": userpassword,
		"captcha":  "false",
	})

	resp, err := http.Post(mshpURL+mshpLoginAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	s := string(body)
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully got user token.")

	return m["token"]
}

func getYearLinks(doc *goquery.Document) *[]Year {
	log.Print("Getting courses years....")
	years := make([]Year, 0, 6)
	doc.Find("div div div div ul li a").Each(func(index int, item *goquery.Selection) {
		year := item.Text()
		link, exists := item.Attr("data-value")
		if exists {
			years = append(years, Year{Year: year, Link: link})
		}
	})
	log.Println("Successfully got courses years.")
	return &years
}

func getCoursesFromYearPage(token *string, year *Year) []Course {

	courses := make([]Course, 0)

	doc := loadPage(mshpURL+mshpCoursesPageURL+"?year_selection="+year.Link, token)

	doc.Find(".panel-top-primary").Each(func(index int, block *goquery.Selection) {
		courses = append(courses, parseCourseBlock(block))
	})

	doc.Find(".panel-top-amazing").Each(func(index int, block *goquery.Selection) {
		courses = append(courses, parseCourseBlock(block))
	})

	return courses

}

func parseCourseBlock(block *goquery.Selection) Course {
	course := Course{}

	courseNameEx := block.Find("div div .panel-title a")
	courseNameText := courseNameEx.Text()

	courseLinkText, linkExists := courseNameEx.Attr("href")
	courseGradeCountEx := block.Find("div div div div div div b")
	courseGradeCountText := courseGradeCountEx.Text()
	courseGradeAvgEx := block.Find(".shp-average")
	courseGradeAvgText := courseGradeAvgEx.Text()
	courseClasssEx := block.Find(".col-lg-4.col-xs-no-padding")
	courseClasssVisited := 0
	courseClasssOverall := 0
	courseTeacherText := "-"

	courseClasssEx.Each(func(_ int, el *goquery.Selection) {
		courseSelectionText := el.Text()
		if strings.Contains(courseSelectionText, "из") {
			courseClasssArr := strings.Split(courseSelectionText, "из")
			courseClasssArr[0] = strings.TrimSpace(courseClasssArr[0])
			courseClasssArr[1] = strings.TrimSpace(courseClasssArr[1])
			courseClasssVisited, _ = strconv.Atoi(courseClasssArr[0])
			courseClasssOverall, _ = strconv.Atoi(courseClasssArr[1])
		}
	})

	courseTeacherEx := block.Find(".media-body.lead.small")
	courseTeacherEx.Each(func(_ int, el *goquery.Selection) {
		s := el.Text()
		s = strings.Split(s, "  ")[1]
		s = strings.TrimSpace(s)
		courseTeacherText = s
	})

	course.Name = courseNameText

	if linkExists {
		idEx := strings.Split(courseLinkText, "/")
		course.ID = idEx[len(idEx)-2]
	} else {
		log.Fatalln("Could not find link for", courseNameText)
	}

	courseGradeCountInt, err := strconv.Atoi(courseGradeCountText)
	if err != nil {
		log.Print("Could not convert grade count at ", courseNameText, ". Setting default value 0\n")
		courseGradeCountInt = 0
	}
	course.GradeCount = courseGradeCountInt

	courseGradeAvgFloat, err := strconv.ParseFloat(courseGradeAvgText, 32)
	if err != nil {
		log.Print("Could not convert grade averege at ", courseNameText, ". Setting default value 0.0\n")
		courseGradeAvgFloat = 0.0
	}
	course.GradeAvg = (float32)(courseGradeAvgFloat)
	course.ClasssVisited = courseClasssVisited
	course.ClasssOverall = courseClasssOverall
	course.Teacher = courseTeacherText

	return course
}

func getCourseClasses(course *Course, token *string) []Class {
	client := &http.Client{}
	Classes := make([]Class, 0)

	URL := mshpURL + mshpGetClassesAPIURL + "?classes__course__school_subject__id=" + course.ID + "&format=json&orderBy=datetime_begin&page=1&limit=999999999"

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Requesting course classes from api: ", URL, "\n")
	req.Header.Set("Cookie", "eduapp_jwt="+*token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	ClassRequestObj := ClassRequest{}

	err = json.Unmarshal(body, &ClassRequestObj)

	for i := 0; i < ClassRequestObj.Count; i++ {
		IDStr := strconv.Itoa(ClassRequestObj.Results[i].ID)
		getClassInfo(&IDStr, token)
	}

	return Classes
}

func getClassInfo(ID *string, token *string) {
	client := &http.Client{}
	//Classes := make([]Class, 0)

	URL := mshpURL + mshpGetClassInfoAPIURL + "?id=" + *ID + "&page=1&limit=999999&format=json&"

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Requesting course info from api: ", URL, "\n")
	req.Header.Set("Cookie", "eduapp_jwt="+*token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	ClassInfoRequestObj := ClassInfoFromRequest{}

	err = json.Unmarshal(body, &ClassInfoRequestObj)

	fmt.Println(ClassInfoRequestObj)
}

func loadPage(URL string, token *string) goquery.Document {
	client := &http.Client{}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	log.Print("Loading ", URL, "\n")
	req.Header.Set("Cookie", "eduapp_jwt="+*token)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return *doc
}
