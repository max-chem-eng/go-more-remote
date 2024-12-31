package jobengine

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var SupportedLanguages = map[string]languageConfig{
	"python": {DefaultImage: "python:3.10", Command: []string{"python", "/tmp/job_script"}, extension: ".py"},
	"ruby":   {DefaultImage: "ruby:3.2", Command: []string{"ruby", "/tmp/job_script"}, extension: ".rb"},
	"go":     {DefaultImage: "golang:1.23.4", Command: []string{"go", "run", "/tmp/job_script"}, extension: ".go"},
	"node":   {DefaultImage: "node:16.11", Command: []string{"node", "/tmp/job_script"}, extension: ".js"},
}

var (
	dockerClient *client.Client
	dockerOnce   sync.Once
	dockerErr    error
)

type JobConfig struct {
	Language      string
	Image         string
	ScriptContent string
	Timeout       time.Duration
}

// type JobConfigOption func(*JobConfig)

type languageConfig struct {
	DefaultImage string
	Command      []string
	extension    string
}

func ExecuteJob(config JobConfig) (string, string, error) {
	config.setDefaults()

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	languageConfig, supported := SupportedLanguages[config.Language]
	if !supported {
		return "", "", fmt.Errorf("unsupported language: %s", config.Language)
	}

	cli, err := getDockerClient()
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize Docker client: %w", err)
	}

	err = pullImage(ctx, cli, config.Image)
	if err != nil {
		return "", "", fmt.Errorf("failed to pull image %s: %w", config.Image, err)
	}

	logs, stats, err := runContainer(ctx, cli, config, languageConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute job: %w", err)
	}

	return logs, stats, nil
}

func (config *JobConfig) setDefaults() {
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Minute
	}
	if config.Image == "" {
		config.Image = SupportedLanguages[config.Language].DefaultImage
	}
}

func getDockerClient() (*client.Client, error) {
	dockerOnce.Do(func() {
		dockerClient, dockerErr = client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
	})
	return dockerClient, dockerErr
}

func pullImage(ctx context.Context, cli *client.Client, imageName string) error {
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(os.Stdout, out) // Stream pull progress to stdout
	return err
}

func runContainer(ctx context.Context, cli *client.Client, config JobConfig, languageConfig languageConfig) (string, string, error) {
	scriptPath, err := saveScriptToFile(config.ScriptContent, config.Language)
	if err != nil {
		fmt.Printf("Error saving script: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(scriptPath)

	if _, err := os.Stat(scriptPath); err != nil {
		return "", "", fmt.Errorf("script file not found: %s", scriptPath)
	}

	resp, err := createContainer(ctx, cli, config, languageConfig, scriptPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", fmt.Errorf("failed to start container: %w", err)
	}

	statsCh := make(chan containerStats)
	go monitorContainerStats(ctx, cli, resp.ID, statsCh)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "", "", fmt.Errorf("container wait error: %w", err)
		}
	case <-statusCh:
	}

	logs, err := fetchContainerLogs(ctx, cli, resp.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch logs: %w", err)
	}

	var stats string
	for stat := range statsCh {
		stats += fmt.Sprintf("\nContainer Stats: %+v", stat)
	}

	err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
	if err != nil {
		return "", "", fmt.Errorf("failed to remove container: %w", err)
	}

	return logs, stats, nil
}

func createContainer(ctx context.Context, cli *client.Client, config JobConfig, languageConfig languageConfig, scriptPath string) (container.CreateResponse, error) {
	res, err := cli.ContainerCreate(ctx, &container.Config{
		Image: config.Image,
		Cmd:   languageConfig.Command,
		Tty:   false,
	}, &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/tmp/job_script", scriptPath), // Bind script to container
		},
	}, nil, nil, "")

	return res, err
}

func fetchContainerLogs(ctx context.Context, cli *client.Client, containerID string) (string, error) {
	out, err := cli.ContainerLogs(ctx, containerID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	var logsBuilder strings.Builder
	for {
		header := make([]byte, 8)
		_, err := io.ReadFull(out, header)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read container log header: %w", err)
		}

		length := int(binary.BigEndian.Uint32(header[4:]))
		logMessage := make([]byte, length)

		_, err = io.ReadFull(out, logMessage)
		if err != nil {
			return "", fmt.Errorf("failed to read container log message: %w", err)
		}

		logsBuilder.Write(logMessage)
	}

	return logsBuilder.String(), nil
}

func saveScriptToFile(content, language string) (string, error) {
	if _, supported := SupportedLanguages[language]; !supported {
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	tempFile, err := os.CreateTemp("", fmt.Sprintf("job_script_*%s", SupportedLanguages[language].extension))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write script to file: %w", err)
	}

	return tempFile.Name(), nil
}
