package collectors

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/luigi/xdr-platform/agent/models"
	"github.com/luigi/xdr-platform/agent/utils"
)

// SystemCollector collecte les informations système
type SystemCollector struct {
	logger   *utils.Logger
	agentID  string
	hostname string
}

// NewSystemCollector crée un nouveau collecteur système
func NewSystemCollector(logger *utils.Logger, agentID, hostname string) *SystemCollector {
	return &SystemCollector{
		logger:   logger,
		agentID:  agentID,
		hostname: hostname,
	}
}

// Collect collecte les métriques système
func (sc *SystemCollector) Collect() ([]*models.Event, error) {
	sc.logger.Debug("Starting system collection...")

	var events []*models.Event

	// Collecter les informations CPU
	cpuEvent, err := sc.collectCPUInfo()
	if err != nil {
		sc.logger.Error("Failed to collect CPU info: %v", err)
	} else {
		events = append(events, cpuEvent)
	}

	// Collecter les informations mémoire
	memEvent, err := sc.collectMemoryInfo()
	if err != nil {
		sc.logger.Error("Failed to collect memory info: %v", err)
	} else {
		events = append(events, memEvent)
	}

	// Collecter les informations disque
	diskEvent, err := sc.collectDiskInfo()
	if err != nil {
		sc.logger.Error("Failed to collect disk info: %v", err)
	} else {
		events = append(events, diskEvent)
	}

	// Collecter les informations host
	hostEvent, err := sc.collectHostInfo()
	if err != nil {
		sc.logger.Error("Failed to collect host info: %v", err)
	} else {
		events = append(events, hostEvent)
	}

	sc.logger.Info("Collected %d system events", len(events))
	return events, nil
}

// collectCPUInfo collecte les informations CPU
func (sc *SystemCollector) collectCPUInfo() (*models.Event, error) {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percent: %w", err)
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	cpuCount := runtime.NumCPU()

	systemEvent := models.SystemEvent{
		EventName:   "cpu_metrics",
		Source:      "system_collector",
		Description: "CPU usage metrics",
		Data: map[string]interface{}{
			"cpu_percent": cpuPercent[0],
			"cpu_count":   cpuCount,
			"cpu_info":    cpuInfo,
		},
	}

	severity := models.SeverityLow
	if cpuPercent[0] > 80.0 {
		severity = models.SeverityHigh
	} else if cpuPercent[0] > 60.0 {
		severity = models.SeverityMedium
	}

	event := &models.Event{
		Timestamp: time.Now(),
		AgentID:   sc.agentID,
		Hostname:  sc.hostname,
		EventType: models.EventTypeSystem,
		Severity:  severity,
		RawData: map[string]interface{}{
			"system": systemEvent,
		},
		Tags: []string{"system_metrics", "cpu"},
	}

	return event, nil
}

// collectMemoryInfo collecte les informations mémoire
func (sc *SystemCollector) collectMemoryInfo() (*models.Event, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	systemEvent := models.SystemEvent{
		EventName:   "memory_metrics",
		Source:      "system_collector",
		Description: "Memory usage metrics",
		Data: map[string]interface{}{
			"total":        vmem.Total,
			"available":    vmem.Available,
			"used":         vmem.Used,
			"used_percent": vmem.UsedPercent,
			"free":         vmem.Free,
		},
	}

	severity := models.SeverityLow
	if vmem.UsedPercent > 90.0 {
		severity = models.SeverityHigh
	} else if vmem.UsedPercent > 75.0 {
		severity = models.SeverityMedium
	}

	event := &models.Event{
		Timestamp: time.Now(),
		AgentID:   sc.agentID,
		Hostname:  sc.hostname,
		EventType: models.EventTypeSystem,
		Severity:  severity,
		RawData: map[string]interface{}{
			"system": systemEvent,
		},
		Tags: []string{"system_metrics", "memory"},
	}

	return event, nil
}

// collectDiskInfo collecte les informations disque
func (sc *SystemCollector) collectDiskInfo() (*models.Event, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	diskData := make([]map[string]interface{}, 0)
	var maxUsedPercent float64

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskData = append(diskData, map[string]interface{}{
			"device":       partition.Device,
			"mountpoint":   partition.Mountpoint,
			"fstype":       partition.Fstype,
			"total":        usage.Total,
			"used":         usage.Used,
			"free":         usage.Free,
			"used_percent": usage.UsedPercent,
		})

		if usage.UsedPercent > maxUsedPercent {
			maxUsedPercent = usage.UsedPercent
		}
	}

	systemEvent := models.SystemEvent{
		EventName:   "disk_metrics",
		Source:      "system_collector",
		Description: "Disk usage metrics",
		Data: map[string]interface{}{
			"partitions": diskData,
		},
	}

	severity := models.SeverityLow
	if maxUsedPercent > 90.0 {
		severity = models.SeverityHigh
	} else if maxUsedPercent > 80.0 {
		severity = models.SeverityMedium
	}

	event := &models.Event{
		Timestamp: time.Now(),
		AgentID:   sc.agentID,
		Hostname:  sc.hostname,
		EventType: models.EventTypeSystem,
		Severity:  severity,
		RawData: map[string]interface{}{
			"system": systemEvent,
		},
		Tags: []string{"system_metrics", "disk"},
	}

	return event, nil
}

// collectHostInfo collecte les informations de l'hôte
func (sc *SystemCollector) collectHostInfo() (*models.Event, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	systemEvent := models.SystemEvent{
		EventName:   "host_info",
		Source:      "system_collector",
		Description: "Host information",
		Data: map[string]interface{}{
			"hostname":        hostInfo.Hostname,
			"uptime":          hostInfo.Uptime,
			"boot_time":       hostInfo.BootTime,
			"procs":           hostInfo.Procs,
			"os":              hostInfo.OS,
			"platform":        hostInfo.Platform,
			"platform_family": hostInfo.PlatformFamily,
			"platform_version": hostInfo.PlatformVersion,
			"kernel_version":  hostInfo.KernelVersion,
			"kernel_arch":     hostInfo.KernelArch,
		},
	}

	event := &models.Event{
		Timestamp: time.Now(),
		AgentID:   sc.agentID,
		Hostname:  sc.hostname,
		EventType: models.EventTypeSystem,
		Severity:  models.SeverityLow,
		RawData: map[string]interface{}{
			"system": systemEvent,
		},
		Tags: []string{"system_metrics", "host_info"},
	}

	return event, nil
}
