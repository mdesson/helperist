package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Due struct {
	String      string `json:"string"`
	Date        string `json:"date"`
	IsRecurring bool   `json:"is_recurring"`
	DateTime    string `json:"datetime,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

type Task struct {
	CreatorID    string `json:"creator_id"`
	CreatedAt    string `json:"created_at"`
	AssigneeID   string `json:"assignee_id"`
	AssignerID   string `json:"assigner_id"`
	CommentCount int    `json:"comment_count"`
	IsCompleted  bool   `json:"is_completed"`
	Content      string `json:"content"`
	Description  string `json:"description"`
	Due          Due    `json:"due"`
	ID           string `json:"id"`
	Labels       []string
	Order        int    `json:"order"`
	Priority     int    `json:"priority"`
	ProjectID    string `json:"project_id"`
	SectionID    string `json:"section_id"`
	ParentID     string `json:"parent_id"`
	URL          string `json:"url"`
}

type Reminder struct {
	Due          Due    `json:"due"`
	ID           string `json:"id"`
	IsDeleted    int    `json:"is_deleted"`
	ItemID       string `json:"item_id"`
	NotifyUID    string `json:"notify_uid"`
	Type         string `json:"type"`
	MinuteOffset int    `json:"minute_offset"`
}

func getActiveTasks(apiToken string) ([]Task, error) {
	client := &http.Client{}

	data := fmt.Sprintf(`sync_token=*&resource_types=["items"]`)

	req, err := http.NewRequest("POST", "https://api.todoist.com/sync/v9/sync", strings.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	items, ok := response["items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response from Todoist API")
	}

	tasks := make([]Task, len(items))
	for i, item := range items {
		itemJSON, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(itemJSON, &tasks[i])
		if err != nil {
			return nil, err
		}
	}

	filteredTasks := make([]Task, 0)
	for _, task := range tasks {
		if strings.HasPrefix(task.Content, "Test Reminder") {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks, nil
}

func setReminderForTasks(tasks []Task, apiToken string) error {
	for _, task := range tasks {
		has8AMReminder, err := hasReminder(apiToken, task.ID)
		if err != nil {
			return err
		}
		if !has8AMReminder && task.Due.Date != "" && task.Due.DateTime == "" {
			// Set reminder at 8 AM the same day the task is due
			reminderTime := fmt.Sprintf("%s 08:00:00", task.Due.Date)
			err := addReminder(apiToken, task.ID, reminderTime)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addReminder(apiToken string, itemID string, reminderTime string) error {
	client := &http.Client{}

	uuid := fmt.Sprintf("reminder-uuid-%s-%d", itemID, time.Now().Unix())

	data := fmt.Sprintf(`commands=[{"type": "reminder_add", "temp_id": "reminder-%s", "uuid": "%s", "args": {"item_id": %s, "due": {"date": "%s", "timezone": "America/New_York"}}}]`, itemID, uuid, itemID, reminderTime)

	req, err := http.NewRequest("POST", "https://api.todoist.com/sync/v9/sync", strings.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return fmt.Errorf("returned %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

func hasReminder(apiToken string, itemID string) (bool, error) {
	client := &http.Client{}

	data := fmt.Sprintf(`sync_token=*&resource_types=["reminders"]`)

	req, err := http.NewRequest("POST", "https://api.todoist.com/sync/v9/sync", strings.NewReader(data))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		return false, fmt.Errorf("returned %d: %s", resp.StatusCode, string(body))
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}

	reminders, ok := response["reminders"].([]interface{})
	if !ok {
		return false, fmt.Errorf("invalid response from Todoist API")
	}

	for _, reminder := range reminders {
		var r Reminder
		reminderJSON, err := json.Marshal(reminder)
		if err != nil {
			return false, err
		}
		err = json.Unmarshal(reminderJSON, &r)
		if err != nil {
			return false, err
		}

		if r.ItemID == itemID && r.Type == "absolute" && r.IsDeleted == 0 {
			return true, nil
		}
	}

	return false, nil
}

func main() {
	apiToken := os.Getenv("TODOIST_TOKEN")
	tasks, err := getActiveTasks(apiToken)
	if err != nil {
		log.Fatalf("Error fetching tasks: %v", err)
	}

	if err = setReminderForTasks(tasks, apiToken); err != nil {
		log.Fatalf("Error setting reminders for tasks: %v", err)
	}

	fmt.Println("Reminders set for tasks with due dates and no due time.")
}
