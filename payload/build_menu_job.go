package payload

import (
	"encoding/json"
	"fmt"
)

// Map<String, Object>
type json_node map[string]any

type hydrated_content_json struct {
	Id             string                 `json:"id"`
	CollectionName string                 `json:"collection_name"`
	Data           json_node              `json:"data"`
	Resolved       map[string][]json_node `json:"resolved"`
}

type BuildMenuJob struct {
	ThemeLocationURL string                             `json:"theme_location_url"`
	SiteName         string                             `json:"site_name"`
	Contents         map[string][]hydrated_content_json `json:"contents"`
	StatusURL        string                             `json:"status_url"`
	Timestamp        int64                              `json:"timestamp"`
}

func NewBuildMenuJob(payload string) (*BuildMenuJob, error) {
	var job BuildMenuJob
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &job, nil
}

func (job *BuildMenuJob) PrintContents() {
	for k, v := range job.Contents {
		fmt.Printf("Collection name: %s:\n", k)
		prettyJSON, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling content:", err)
			continue
		}
		fmt.Println(string(prettyJSON))
		fmt.Println()
	}
}

func (job *BuildMenuJob) MarshalContents() (*string, error) {
	jsonBytes, err := json.Marshal(job.Contents)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling to JSON: %w", err)
	}

	jsonString := string(jsonBytes)
	return &jsonString, nil
}
