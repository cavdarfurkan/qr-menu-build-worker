package payload

type MenuJobStatus string

const (
	MenuJobStatusPending    MenuJobStatus = "PENDING"
	MenuJobStatusProcessing MenuJobStatus = "PROCESSING"
	MenuJobStatusDone       MenuJobStatus = "DONE"
	MenuJobStatusFailed     MenuJobStatus = "FAILED"
)
