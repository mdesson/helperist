package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Task struct {
	CreatorID    string   `json:"creator_id"`
	CreatedAt    string   `json:"created_at"`
	AssigneeID   string   `json:"assignee_id"`
	AssignerID   string   `json:"assigner_id"`
	CommentCount int      `json:"comment_count"`
	IsCompleted  bool     `json:"is_completed"`
	Content      string   `json:"content"`
	Description  string   `json:"description"`
	Due          Due      `json:"due"`
	ID           string   `json:"id"`
	Labels       []string `json:"labels"`
	Order        int      `json:"order"`
	Priority     int      `json:"priority"`
	ProjectID    string   `json:"project_id"`
	SectionID    string   `json:"section_id"`
	ParentID     string   `json:"parent_id"`
	URL          string   `json:"url"`
}

type Due struct {
	Date        string `json:"date"`
	IsRecurring bool   `is_recurring"`
	Datetime    string `json:"datetime"`
	String      string `json:"string"`
	Timezone    string `json:"timezone"`
}

func main() {
	token := os.Getenv("TODOIST_TOKEN")

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.todoist.com/rest/v2/tasks", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error executing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var tasks []Task
	err = json.Unmarshal(body, &tasks)
	if err != nil {
		log.Fatalf("Error unmarshalling response: %v", err)
	}

	for _, task := range tasks {
		fmt.Printf("Task ID: %s, Task Content: %s\n", task.ID, task.Content)
	}
}
