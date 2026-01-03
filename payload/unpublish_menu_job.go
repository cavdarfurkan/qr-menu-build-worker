package payload

import (
	"encoding/json"
	"fmt"
)

type UnpublishMenuJob struct {
	SiteName  string `json:"site_name"`
	StatusURL string `json:"status_url"`
	Timestamp int64  `json:"timestamp"`
}

func NewUnpublishMenuJob(payload string) (*UnpublishMenuJob, error) {
	var job UnpublishMenuJob
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &job, nil
}
