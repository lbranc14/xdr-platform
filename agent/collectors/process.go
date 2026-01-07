package collectors

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/luigi/xdr-platform/agent/models"
	"github.com/luigi/xdr-platform/agent/utils"
)

// ProcessCollector collecte les informations sur les processus
type ProcessCollector struct {
	logger   *utils.Logger
	agentID  string
	hostname string
}

// NewProcessCollector crée un nouveau collecteur de processus
func NewProcessCollector(logger *utils.Logger, agentID, hostname string) *ProcessCollector {
	return &ProcessCollector{
		logger:   logger,
		agentID:  agentID,
		hostname: hostname,
	}
}

// Collect collecte les informations des processus actifs
func (pc *ProcessCollector) Collect() ([]*models.Event, error) {
	pc.logger.Debug("Starting process collection...")

	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var events []*models.Event

	// Collecter les informations de tous les processus
	for _, p := range processes {
		procEvent, err := pc.collectProcessInfo(p)
		if err != nil {
			// Log l'erreur mais continue avec les autres processus
			pc.logger.Debug("Failed to collect process %d: %v", p.Pid, err)
			continue
		}

		if procEvent != nil {
			events = append(events, procEvent)
		}
	}

	pc.logger.Info("Collected %d process events", len(events))
	return events, nil
}

// collectProcessInfo collecte les informations d'un processus spécifique
func (pc *ProcessCollector) collectProcessInfo(p *process.Process) (*models.Event, error) {
	name, err := p.Name()
	if err != nil {
		return nil, err
	}

	cmdline, _ := p.Cmdline()
	exe, _ := p.Exe()
	username, _ := p.Username()
	ppid, _ := p.Ppid()
	cpuPercent, _ := p.CPUPercent()
	memPercent, _ := p.MemoryPercent()
	memInfo, _ := p.MemoryInfo()
	createTime, _ := p.CreateTime()
	numThreads, _ := p.NumThreads()
	status, _ := p.Status()
	connections, _ := p.Connections()

	var memBytes uint64
	if memInfo != nil {
		memBytes = memInfo.RSS
	}

	processEvent := models.ProcessEvent{
		PID:            int(p.Pid),
		Name:           name,
		CommandLine:    cmdline,
		ExecutablePath: exe,
		ParentPID:      int(ppid),
		Username:       username,
		CPUPercent:     cpuPercent,
		MemoryPercent:  float64(memPercent),
		MemoryBytes:    memBytes,
		CreateTime:     createTime,
		NumThreads:     numThreads,
		Status:         status[0],
		Connections:    len(connections),
	}

	// Créer l'événement
	event := &models.Event{
		Timestamp:   time.Now(),
		AgentID:     pc.agentID,
		Hostname:    pc.hostname,
		EventType:   models.EventTypeProcess,
		Severity:    pc.determineSeverity(processEvent),
		ProcessName: name,
		ProcessPID:  int(p.Pid),
		Username:    username,
		RawData: map[string]interface{}{
			"process": processEvent,
		},
		Tags: pc.generateTags(processEvent),
	}

	return event, nil
}

// determineSeverity détermine la sévérité basée sur les métriques du processus
func (pc *ProcessCollector) determineSeverity(pe models.ProcessEvent) models.Severity {
	// Processus suspect : haute utilisation CPU ou mémoire
	if pe.CPUPercent > 80.0 || pe.MemoryPercent > 80.0 {
		return models.SeverityHigh
	}

	// Processus avec beaucoup de connexions réseau
	if pe.Connections > 50 {
		return models.SeverityMedium
	}

	return models.SeverityLow
}

// generateTags génère des tags basés sur le processus
func (pc *ProcessCollector) generateTags(pe models.ProcessEvent) []string {
	tags := []string{"process_monitoring"}

	// Ajouter des tags basés sur les caractéristiques
	if pe.CPUPercent > 50.0 {
		tags = append(tags, "high_cpu")
	}

	if pe.MemoryPercent > 50.0 {
		tags = append(tags, "high_memory")
	}

	if pe.Connections > 10 {
		tags = append(tags, "network_active")
	}

	return tags
}
