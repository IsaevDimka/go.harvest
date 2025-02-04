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

type Config struct {
	TimeZone        string
	PermalinkPrefix string
	Token           string
	TaskID          string
	Projects        []string
	Accounts        []string
	TaskPrefix      []string
	AccountIDs      map[string]string
	UserIDs         map[string]string
	ProjectIDs      map[string]string
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
		TimeZone:        os.Getenv("TIMEZONE"),
		PermalinkPrefix: os.Getenv("PERMALINK_PREFIX"),
		Token:           os.Getenv("TOKEN"),
		TaskID:          os.Getenv("TASK_ID"),
		Projects:        strings.Split(os.Getenv("PROJECTS"), ","),
		Accounts:        strings.Split(os.Getenv("ACCOUNTS"), ","),
		TaskPrefix:      strings.Split(os.Getenv("TASK_PREFIX"), ","),
		AccountIDs:      splitMap(os.Getenv("ACCOUNT_IDS")),
		UserIDs:         splitMap(os.Getenv("USER_IDS")),
		ProjectIDs:      splitMap(os.Getenv("PROJECT_IDS")),
	}

	var (
		projectID string
		accountID string
		userID    string
	)

	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}
	spentDate := time.Now().In(loc).Format(timeFormat)

	reader := bufio.NewReader(os.Stdin)

	var account string
	if err = survey.AskOne(&survey.Select{
		Message: "Choose a harvest account:",
		Options: cfg.Accounts,
	}, &account); err != nil {
		log.Fatalln("error:", err)
	}
	accountID = cfg.AccountIDs[account]

	var project string
	if err = survey.AskOne(&survey.Select{
		Message: "Choose a harvest project:",
		Options: cfg.Projects,
	}, &project); err != nil {
		log.Fatalln("error:", err)
	}
	projectID = cfg.ProjectIDs[project]

	if account == "LuckyGroup" {
		if err = survey.AskOne(&survey.Select{
			Message: "Choose a task prefix:",
			Options: cfg.TaskPrefix,
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

	// todo: for CoinWithMe=
	// todo: for IsaevDimka=
	userID = cfg.UserIDs[account]

	// todo: refactor cfg.TaskID

	u := url.URL{
		Scheme: "https",
		Host:   "api.harvestapp.com",
		Path:   "v2/time_entries",
	}
	q := u.Query()
	q.Set("project_id", projectID)
	q.Set("task_id", cfg.TaskID)
	q.Set("user_id", userID)
	q.Set("spent_date", spentDate)

	if account == "LuckyGroup" {
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
	req.Header.Add("Harvest-Account-Id", accountID)
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

func splitMap(v string) map[string]string {
	m := make(map[string]string)
	pairs := strings.Split(v, ",")
	for _, pair := range pairs {
		keyValue := strings.SplitN(pair, "=", 2)
		if len(keyValue) == 2 {
			key := keyValue[0]
			value := keyValue[1]
			m[key] = value
		}
	}

	return m
}
