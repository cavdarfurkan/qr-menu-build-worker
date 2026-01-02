package main

import (
	"cavdarfurkan/qr-menu-build-worker/payload"
	"context"
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
)

var (
	ctx    = context.Background()
	client = redis.NewClient(&redis.Options{Addr: RedisAddr})
)

func worker(wg *sync.WaitGroup, id int) {
	defer wg.Done()
	for {
		res, err := client.BRPop(ctx, 0, ListenerKey).Result()
		if err != nil {
			slog.Error("redis error", "err", err)
			continue
		}

		job, err := payload.NewBuildMenuJob(res[1])
		if err != nil {
			slog.Error("new build menu job error", "err", err)
		}

		wranglerConfig := payload.NewWranglerConfig(job.SiteName)
		wranglerConfigJson, err := wranglerConfig.MarshalConfig()
		if err != nil {
			slog.Error("wrangler config marshal", "err", err)
		}

		userContentsJson, err := job.MarshalContents()
		if err != nil {
			slog.Error("err", "err", err)
		}

		slog.Info("worker", "id", id)
		slog.Info("build menu job", "job", job.Timestamp)
		fmt.Println()

		sendStatusUpdateRequest(job, payload.MenuJobStatusProcessing)

		cmd := exec.Command(
			"podman", "run", "--rm",
			"-e", fmt.Sprintf("THEME_LOCATION_URL=%s", job.ThemeLocationURL),
			"-e", fmt.Sprintf("WRANGLER_CONFIG=%s", wranglerConfigJson),
			"-e", fmt.Sprintf("USER_CONTENT=%s", *userContentsJson),
			// "-v", "/home/fcavdar/.aws/credentials:/root/.aws/credentials:ro",
			// "-v", "${WORKSPACE}:/project",
			"-e", fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", AwsAccessKeyId),
			"-e", fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", AwsSecretAccessKey),
			"-e", fmt.Sprintf("AWS_DEFAULT_REGION=%s", AwsRegion),
			"-e", fmt.Sprintf("CLOUDFLARE_API_TOKEN=%s", CloudflareApiToken),
			"-e", fmt.Sprintf("CLOUDFLARE_ACCOUNT_ID=%s", CloudFlareAccountId),
			BuilderImage,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Println("build failed:\n", string(out), err)
			sendStatusUpdateRequest(job, payload.MenuJobStatusFailed)
		} else {
			fmt.Println("build succeeded:\n", string(out))
			sendStatusUpdateRequest(job, payload.MenuJobStatusDone)
		}

		// client.LPush(ctx, "queue:build:completed:main", res[1])
		fmt.Println()
		slog.Info("DONE", "worker id", id)
	}
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

func sendStatusUpdateRequest(job *payload.BuildMenuJob, status payload.MenuJobStatus) {
	resp, err := http.Post(
		job.StatusURL,
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"status": "%s"}`, status)),
	)
	if err != nil {
		fmt.Println()
		slog.Error("Job status update request", "job", job.Timestamp, "status", status, "err", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println()
		slog.Error("Response body read error", "job", job.Timestamp, "status", status, "err", err)
	}
	fmt.Println()
	slog.Info("Response body", "body", body)
}
