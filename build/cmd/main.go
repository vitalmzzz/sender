package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"gorm.io/datatypes"
)

var wg sync.WaitGroup

type Search []struct {
	ID          int            `json:"id"`
	Iid         int            `json:"iid"`
	ProjectID   int            `json:"project_id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	State       string         `json:"state"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ClosedAt    interface{}    `json:"closed_at"`
	ClosedBy    interface{}    `json:"closed_by"`
	Labels      datatypes.JSON `json:"labels"`
	Milestone   struct {
		ID          int       `json:"id"`
		Iid         int       `json:"iid"`
		ProjectID   int       `json:"project_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		State       string    `json:"state"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		DueDate     string    `json:"due_date"`
		StartDate   string    `json:"start_date"`
		Expired     bool      `json:"expired"`
		WebURL      string    `json:"web_url"`
	} `json:"milestone"`
	Assignees []struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"assignees"`
	Author struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"author"`
	Type     string `json:"type"`
	Assignee struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Name      string `json:"name"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
		WebURL    string `json:"web_url"`
	} `json:"assignee"`
	UserNotesCount     int         `json:"user_notes_count"`
	MergeRequestsCount int         `json:"merge_requests_count"`
	Upvotes            int         `json:"upvotes"`
	Downvotes          int         `json:"downvotes"`
	DueDate            interface{} `json:"due_date"`
	Confidential       bool        `json:"confidential"`
	DiscussionLocked   interface{} `json:"discussion_locked"`
	IssueType          string      `json:"issue_type"`
	WebURL             string      `json:"web_url"`
	TimeStats          struct {
		TimeEstimate        int         `json:"time_estimate"`
		TotalTimeSpent      int         `json:"total_time_spent"`
		HumanTimeEstimate   interface{} `json:"human_time_estimate"`
		HumanTotalTimeSpent interface{} `json:"human_total_time_spent"`
	} `json:"time_stats"`
	TaskCompletionStatus struct {
		Count          int `json:"count"`
		CompletedCount int `json:"completed_count"`
	} `json:"task_completion_status"`
}

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	r := mux.NewRouter()
	r.HandleFunc("/Send", requestData).Methods("GET")
	err := http.ListenAndServe("0.0.0.0:8080", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func requestData(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Get started: %s\n", time.Now().String())

	gitlab_token := os.Getenv("GITLAB_TOKEN")

	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://gitlab.ru/api/v4/issues?state=opened&scope=all&labels=PRODUCTION%20DEPLOYMENT", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+string(gitlab_token))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	size := len(body)
	if size <= 3 {
		fmt.Println("Response length does not meet the minimum:", int(size))
		return
	}

	var repo Search

	json.Unmarshal(body, &repo)
	SendEmail(repo)

	fmt.Println("We print the data structure for sending to E-MAIL")
	for k := range repo {
		fmt.Println(repo[k].Title, repo[k].WebURL, repo[k].CreatedAt, repo[k].State, repo[k].Assignee.Name, repo[k].Author.Name)
	}

	wg.Wait()
	w.Write([]byte("Completed"))

}

func SendEmail(repo Search) {

	mail_server := os.Getenv("MAIL_SERVER")

	//file, err := os.Create("report.txt")
	//if err != nil {
	//	return
	//}

	//defer file.Close()

	//for k := range repo {
	//	file.WriteString(fmt.Sprintln(repo[k].Title, repo[k].WebURL, repo[k].CreatedAt, repo[k].State, repo[k].Assignee.Name, repo[k].Author.Name))
	//}

	m := gomail.NewMessage()
	m.SetHeader("From", "")
	m.SetHeader("To", "")
	//m.SetHeader("To", "")
	m.SetHeader("Subject", "PRODUCTION DEPLOYMENT")

	//m.SetBody("text/plain", fmt.Sprintln("Formation time:", time.Now().String()))
	for k := range repo {
		m.SetBody("text/plain", fmt.Sprintln("Formation time:", time.Now().String(), "\n", fmt.Sprintln("Title: ", repo[k].Title, "\n", "WebURL: ", repo[k].WebURL, "\n", "Created: ", repo[k].CreatedAt, "\n", "State: ", repo[k].State, "\n", "Assignee: ", repo[k].Assignee.Name, "\n", "Author:", repo[k].Author.Name)))
	}
	//m.Attach("report.txt")

	d := gomail.NewDialer(mail_server, 25, "", "")

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func CreateIssue(repo Search) {

}
