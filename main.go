package main

import (
	"cavdarfurkan/qr-menu-build-worker/payload"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// Env variables
var (
	RedisAddr           string
	ListenerKey         string
	AwsAccessKeyId      string
	AwsSecretAccessKey  string
	AwsRegion           string
	CloudflareApiToken  string
	CloudFlareAccountId string
	BuilderImage        string
	UnpublishImage      string
)

var (
	ctx    = context.Background()
	client = redis.NewClient(&redis.Options{Addr: RedisAddr})
)

type JobType struct {
	Type string `json:"type"`
}

func worker(wg *sync.WaitGroup, id int) {
	defer wg.Done()
	for {
		res, err := client.BRPop(ctx, 0, ListenerKey).Result()
		if err != nil {
			slog.Error("redis error", "err", err)
			continue
		}

		// Determine the job type
		var jobType JobType
		if err := json.Unmarshal([]byte(res[1]), &jobType); err != nil {
			slog.Error("failed to determine job type", "err", err)
			continue
		}

		switch jobType.Type {
		case "unpublish":
			handleUnpublishJob(res[1], id)
		case "build":
			handleBuildAndPublishJob(res[1], id)
		default:
			slog.Error("unknown job type", "job_type", jobType.Type)
		}
	}
}

func handleBuildAndPublishJob(payloadStr string, workerId int) {
	job, err := payload.NewBuildMenuJob(payloadStr)
	if err != nil {
		slog.Error("new build menu job error", "err", err)
		return
	}

	wranglerConfig := payload.NewWranglerConfig(job.SiteName)
	wranglerConfigJson, err := wranglerConfig.MarshalConfig()
	if err != nil {
		slog.Error("wrangler config marshal", "err", err)
		return
	}

	userContentsJson, err := job.MarshalContents()
	if err != nil {
		slog.Error("err", "err", err)
		return
	}

	slog.Info("worker", "id", workerId)
	slog.Info("build menu job", "job", job.Timestamp)
	fmt.Println()

	sendStatusUpdateRequest(job.StatusURL, job.Timestamp, payload.MenuJobStatusProcessing)

	cmd := exec.Command(
		"podman", "run", "--rm",
		"-e", fmt.Sprintf("THEME_LOCATION_URL=%s", job.ThemeLocationURL),
		"-e", fmt.Sprintf("WRANGLER_CONFIG=%s", wranglerConfigJson),
		"-e", fmt.Sprintf("USER_CONTENT=%s", *userContentsJson),
		"-e", fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", AwsAccessKeyId),
		"-e", fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", AwsSecretAccessKey),
		"-e", fmt.Sprintf("AWS_DEFAULT_REGION=%s", AwsRegion),
		"-e", fmt.Sprintf("CLOUDFLARE_API_TOKEN=%s", CloudflareApiToken),
		"-e", fmt.Sprintf("CLOUDFLARE_ACCOUNT_ID=%s", CloudFlareAccountId),
		BuilderImage,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("build failed:\n", string(out), err)
		sendStatusUpdateRequest(job.StatusURL, job.Timestamp, payload.MenuJobStatusFailed)
	} else {
		fmt.Println("build succeeded:\n", string(out))
		sendStatusUpdateRequest(job.StatusURL, job.Timestamp, payload.MenuJobStatusDone)
	}

	// client.LPush(ctx, "queue:build:completed:main", res[1])
	fmt.Println()
	slog.Info("DONE", "worker id", workerId)
}

func handleUnpublishJob(payloadStr string, workerId int) {
	job, err := payload.NewUnpublishMenuJob(payloadStr)
	if err != nil {
		slog.Error("new unpublish menu job error", "err", err)
		return
	}

	slog.Info("worker", "id", workerId)
	slog.Info("unpublish menu job", "site", job.SiteName, "timestamp", job.Timestamp)
	fmt.Println()

	sendStatusUpdateRequest(job.StatusURL, job.Timestamp, payload.MenuJobStatusProcessing)

	cmd := exec.Command(
		"podman", "run", "--rm",
		"-e", fmt.Sprintf("SITE_NAME=%s", job.SiteName),
		"-e", fmt.Sprintf("CLOUDFLARE_API_TOKEN=%s", CloudflareApiToken),
		"-e", fmt.Sprintf("CLOUDFLARE_ACCOUNT_ID=%s", CloudFlareAccountId),
		UnpublishImage,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("unpublish failed:\n", string(out), err)
		sendStatusUpdateRequest(job.StatusURL, job.Timestamp, payload.MenuJobStatusFailed)
	} else {
		fmt.Println("unpublish succeeded:\n", string(out))
		sendStatusUpdateRequest(job.StatusURL, job.Timestamp, payload.MenuJobStatusDone)
	}

	fmt.Println()
	slog.Info("DONE", "worker id", workerId)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	RedisAddr = os.Getenv("REDIS_ADDR")
	ListenerKey = os.Getenv("QUEUE_KEY")
	AwsAccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	AwsRegion = os.Getenv("AWS_DEFAULT_REGION")
	CloudflareApiToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	CloudFlareAccountId = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	BuilderImage = os.Getenv("BUILDER_IMAGE")
	UnpublishImage = os.Getenv("UNPUBLISH_IMAGE")
}

func main() {
	const numWorkers = 20
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	slog.Info("qr-menu-build-worker started", "workers", numWorkers)

	for i := range numWorkers {
		go worker(&wg, i+1)
	}
	wg.Wait()
}

func sendStatusUpdateRequest(status_url string, timestamp int64, status payload.MenuJobStatus) {
	resp, err := http.Post(
		status_url,
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"status": "%s"}`, status)),
	)
	if err != nil {
		fmt.Println()
		slog.Error("Job status update request", "job", timestamp, "status", status, "err", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println()
		slog.Error("Response body read error", "job", timestamp, "status", status, "err", err)
	}
	fmt.Println()
	slog.Info("Response body", "body", body)
}
