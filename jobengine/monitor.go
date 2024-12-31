package jobengine

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type containerStats struct {
	Timestamp       string
	CPUPercentage   float64
	MemoryMB        float64
	NetworkSent     uint64
	NetworkReceived uint64
}

func monitorContainerStats(ctx context.Context, cli *client.Client, containerID string, statsCh chan<- containerStats) {
	defer close(statsCh)

	stats, err := cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		statsCh <- containerStats{} // Send empty stats on error
		return
	}
	defer stats.Body.Close()

	decoder := json.NewDecoder(stats.Body)
	var usage containerStats

	for {
		var v container.StatsResponse
		if err := decoder.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			statsCh <- containerStats{}
			return
		}

		// format timestamp with milliseconds
		usage.Timestamp = time.Now().Format("2006-01-02T15:04:05.000")
		cpuDelta := float64(v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(v.CPUStats.SystemUsage - v.PreCPUStats.SystemUsage)
		if systemDelta > 0 {
			usage.CPUPercentage = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		}

		usage.MemoryMB = float64(v.MemoryStats.Usage) / (1024 * 1024)

		usage.NetworkSent += v.Networks["eth0"].TxBytes
		usage.NetworkReceived += v.Networks["eth0"].RxBytes

		statsCh <- usage

		if v.PidsStats.Current == 0 {
			break
		}

		time.Sleep(200 * time.Millisecond)
	}
}
