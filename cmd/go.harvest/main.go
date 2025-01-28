package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"

	"github.com/joho/godotenv"
)

const timeFormat = "2006-01-02"

var (
	harvestAccountsIDs = map[string]string{
		"LuckyGroup": "424855",
		"CoinWithMe": "1274529",
		"IsaevDimka": "616466",
	}
	harvestProjectIDs = map[string]string{
		"LG":  "38921621",
		"LT":  "38921618",
		"CWM": "34544430",
	}
	harvestProjects = []string{"LG", "LT", "CWM"}
	harvestAccounts = []string{"LuckyGroup", "IsaevDimka", "CoinWithMe"}
	projects        = []string{"IGAMING", "INFRA", "LCH", "LPP", "PRIORITY", "LNA"}
)

type Config struct {
	TimeZone         string
	PermalinkPrefix  string
	DefaultProject   string
	HarvestAccountID string
	Token            string
	ProjectID        string
	TaskID           string
	UserID           string
}

type JiraTask struct {
	Project string
	ID      string
	Message string
}

func main() {
	var (
		task JiraTask
		err  error
	)

	err = godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
		return
	}

	cfg := Config{
		TimeZone:         os.Getenv("TIMEZONE"),
		PermalinkPrefix:  os.Getenv("PERMALINK_PREFIX"),
		DefaultProject:   os.Getenv("DEFAULT_PROJECT"),
		HarvestAccountID: os.Getenv("HARVEST_ACCOUNT_ID"),
		Token:            os.Getenv("TOKEN"),
		ProjectID:        os.Getenv("PROJECT_ID"),
		TaskID:           os.Getenv("TASK_ID"),
		UserID:           os.Getenv("USER_ID"),
	}

	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}
	spentDate := time.Now().In(loc).Format(timeFormat)

	reader := bufio.NewReader(os.Stdin)

	var harvestAccount string
	if err = survey.AskOne(&survey.Select{
		Message: "Choose a harvest account:",
		Options: harvestAccounts,
	}, &harvestAccount); err != nil {
		log.Fatalln("error:", err)
	}
	cfg.HarvestAccountID = harvestAccountsIDs[harvestAccount]

	var harvestProject string
	if err = survey.AskOne(&survey.Select{
		Message: "Choose a harvest project:",
		Options: harvestProjects,
	}, &harvestProject); err != nil {
		log.Fatalln("error:", err)
	}
	cfg.ProjectID = harvestProjectIDs[harvestProject]

	if cfg.HarvestAccountID == "424855" {
		if err = survey.AskOne(&survey.Select{
			Message: "Choose a project:",
			Options: projects,
		}, &task.Project); err != nil {
			log.Fatalln("error:", err)
		}

		fmt.Print("Enter a task ID:")
		task.ID, _ = reader.ReadString('\n')
		task.ID = strings.TrimSuffix(task.ID, "\n")
	}

	fmt.Print("Enter a message:")
	task.Message, _ = reader.ReadString('\n')
	task.Message = strings.TrimSuffix(task.Message, "\n")

	u := url.URL{
		Scheme: "https",
		Host:   "api.harvestapp.com",
		Path:   "v2/time_entries",
	}
	q := u.Query()
	q.Set("project_id", cfg.ProjectID)
	q.Set("task_id", cfg.TaskID)
	q.Set("user_id", cfg.UserID)
	q.Set("spent_date", spentDate)

	if cfg.HarvestAccountID == "424855" {
		taskID := fmt.Sprintf("%s-%s", task.Project, task.ID)
		q.Set("notes", fmt.Sprintf("[%s] %s", taskID, task.Message))
		q.Set("external_reference[group_id]", task.Project)
		q.Set("external_reference[id]", taskID)
		q.Set("external_reference[permalink]", fmt.Sprintf("%s/%s", cfg.PermalinkPrefix, taskID))

	} else {
		var permalink string
		fmt.Print("Enter a permalink:")
		permalink, _ = reader.ReadString('\n')
		permalink = strings.TrimSuffix(permalink, "\n")
		q.Set("notes", task.Message)
		q.Set("external_reference[permalink]", permalink)
	}

	u.RawQuery = q.Encode()

	client := &http.Client{}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Harvest-Account-Id", cfg.HarvestAccountID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.Token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %v", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to parse response body: %v", err)
		return
	}
	fmt.Println(string(body))
}
