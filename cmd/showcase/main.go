package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/samtaborsky/containerlib/docker"
	"github.com/samtaborsky/containerlib/types"
)

type Task struct {
	Name  string
	Image string
	Cmd   []string
}

func main() {
	ctx := context.Background()

	rt, err := docker.New("")
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer rt.Close()

	pipeline := []Task{
		{
			Name:  "Codebase linting",
			Image: "alpine:latest",
			Cmd:   []string{"echo", "Checking syntax... Passed!"},
		},
		{
			Name:  "Build",
			Image: "golang:1.20-alpine",
			Cmd:   []string{"go", "version"},
		},
		{
			Name:  "Tests (intentional failure)",
			Image: "alpine:latest",
			Cmd:   []string{"sh", "-c", "echo 'Running test suite...' && exit 1"},
		},
	}

	fmt.Println("=== PIPELINE STARTT ===")

	for i, task := range pipeline {
		fmt.Printf("\n>>> Step %d/3: %s <<<\n", i+1, task.Name)

		if err := runTask(ctx, rt, task); err != nil {
			log.Fatalf("\n[PIPELINE FAILED] Step '%s' ended with an error: %v\n", task.Name, err)
		}

		fmt.Printf("[SUCCESS] Step '%s' was successfully completed.\n", task.Name)
	}

	fmt.Println("\n=== PIPELINE SUCCESS ===")
}

func runTask(ctx context.Context, rt types.Runtime, task Task) error {
	startTime := time.Now()

	progressChan := make(chan types.ImagePullProgress, 100)

	go func() {
		for p := range progressChan {
			if p.Status == "Pull complete" || strings.HasPrefix(p.Status, "Downloaded newer image") {
				fmt.Printf("  -> [Pulling] %s: %s\n", task.Image, p.Status)
			}
		}
	}()

	err := rt.ImagePull(ctx, task.Image, &types.ImagePullOptions{Progress: progressChan})
	if err != nil {
		return fmt.Errorf("could not pull image: %w", err)
	}

	createRes, err := rt.ContainerCreate(ctx, &types.ContainerCreateConfig{
		Image: task.Image,
		Cmd:   task.Cmd,
	})
	if err != nil {
		return fmt.Errorf("could not create container: %w", err)
	}
	containerID := createRes.ID

	defer func() {
		_ = rt.ContainerRemove(context.Background(), containerID, &types.ContainerRemoveOptions{Force: true})
	}()

	if err := rt.ContainerStart(ctx, containerID, nil); err != nil {
		return fmt.Errorf("could not start container: %w", err)
	}

	statsCtx, cancelStats := context.WithCancel(ctx)
	defer cancelStats()

	statsRes, err := rt.ContainerStats(statsCtx, containerID, &types.ContainerStatsOptions{
		Stream: true,
	})
	if err != nil {
		return fmt.Errorf("could not start stats stream: %w", err)
	}

	avgRamChan := make(chan uint64)
	go func() {
		var totalRAM uint64
		var count uint64

		for {
			select {
			case stats, ok := <-statsRes.Stats:
				if !ok {
					var avg uint64
					if count > 0 {
						avg = totalRAM / count
					}
					avgRamChan <- avg
					return
				}
				totalRAM += stats.Memory.UsedMb
				count++

			case <-statsRes.Errors:
				var avg uint64
				if count > 0 {
					avg = totalRAM / count
				}
				avgRamChan <- avg
				return

			case <-statsCtx.Done():
				var avg uint64
				if count > 0 {
					avg = totalRAM / count
				}
				avgRamChan <- avg
				return
			}
		}
	}()

	_, err = rt.ContainerLogs(ctx, containerID, &types.ContainerLogsOptions{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Follow: true,
	})
	if err != nil {
		return fmt.Errorf("error while reading logs: %w", err)
	}

	statusRes, err := rt.ContainerWait(ctx, containerID, &types.ContainerWaitOptions{})
	if err != nil {
		return fmt.Errorf("error %w", err)
	}

	if statusRes.ExitCode != 0 {
		return fmt.Errorf("process ended with exit code %d", statusRes.ExitCode)
	}

	cancelStats()
	avgRam := <-avgRamChan
	duration := time.Since(startTime)

	fmt.Printf("\n[TELEMETRY] Task completed in %v | Avg RAM: %d MB\n",
		duration.Round(time.Millisecond),
		avgRam,
	)

	return nil
}
