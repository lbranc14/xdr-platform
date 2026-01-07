package collectors

import (
	"fmt"
	"time"

	"github.com/luigi/xdr-platform/agent/models"
	"github.com/luigi/xdr-platform/agent/utils"
	"github.com/shirou/gopsutil/v3/net"
)

// NetworkCollector collecte les informations réseau
type NetworkCollector struct {
	logger   *utils.Logger
	agentID  string
	hostname string
}

// NewNetworkCollector crée un nouveau collecteur réseau
func NewNetworkCollector(logger *utils.Logger, agentID, hostname string) *NetworkCollector {
	return &NetworkCollector{
		logger:   logger,
		agentID:  agentID,
		hostname: hostname,
	}
}

// Collect collecte les connexions réseau actives
func (nc *NetworkCollector) Collect() ([]*models.Event, error) {
	nc.logger.Debug("Starting network collection...")

	connections, err := net.Connections("all")
	if err != nil {
		return nil, fmt.Errorf("failed to get network connections: %w", err)
	}

	var events []*models.Event

	for _, conn := range connections {
		// Ignorer les connexions sans IP (listening sockets, etc.)
		if conn.Laddr.IP == "" {
			continue
		}

		// Convertir le type de protocole en string
		protocol := "unknown"
		switch conn.Type {
		case 1:
			protocol = "tcp"
		case 2:
			protocol = "udp"
		case 3:
			protocol = "tcp6"
		case 4:
			protocol = "udp6"
		}

		networkEvent := models.NetworkEvent{
			Protocol:   protocol,
			SourceIP:   conn.Laddr.IP,
			SourcePort: int(conn.Laddr.Port),
			DestIP:     conn.Raddr.IP,
			DestPort:   int(conn.Raddr.Port),
			State:      conn.Status,
		}

		event := &models.Event{
			Timestamp:     time.Now(),
			AgentID:       nc.agentID,
			Hostname:      nc.hostname,
			EventType:     models.EventTypeNetwork,
			Severity:      nc.determineSeverity(networkEvent),
			SourceIP:      conn.Laddr.IP,
			DestinationIP: conn.Raddr.IP,
			ProcessPID:    int(conn.Pid),
			RawData: map[string]interface{}{
				"network": networkEvent,
			},
			Tags: nc.generateTags(networkEvent),
		}

		events = append(events, event)
	}

	nc.logger.Info("Collected %d network events", len(events))
	return events, nil
}

// determineSeverity détermine la sévérité basée sur la connexion
func (nc *NetworkCollector) determineSeverity(ne models.NetworkEvent) models.Severity {
	// Ports sensibles
	sensitivePorts := map[int]bool{
		22: true, 23: true, 3389: true, // SSH, Telnet, RDP
		445: true, 135: true, 139: true, // SMB, RPC
		1433: true, 3306: true, 5432: true, // SQL Servers
	}

	if sensitivePorts[ne.DestPort] {
		return models.SeverityMedium
	}

	// Connexions vers Internet (non-privées)
	if !nc.isPrivateIP(ne.DestIP) {
		return models.SeverityLow
	}

	return models.SeverityLow
}

// isPrivateIP vérifie si une IP est privée
func (nc *NetworkCollector) isPrivateIP(ip string) bool {
	// Vérifier que l'IP n'est pas vide
	if ip == "" {
		return false
	}

	// Simplification : vérifier les plages privées communes
	// 10.x.x.x, 172.16-31.x.x, 192.168.x.x
	if len(ip) >= 3 && ip[:3] == "10." {
		return true
	}
	if len(ip) >= 4 && ip[:4] == "172." {
		return true
	}
	if len(ip) >= 8 && ip[:8] == "192.168." {
		return true
	}
	if len(ip) >= 4 && ip[:4] == "127." {
		return true // localhost
	}
	return false
}

// generateTags génère des tags basés sur la connexion
func (nc *NetworkCollector) generateTags(ne models.NetworkEvent) []string {
	tags := []string{"network_monitoring"}

	// Ajouter des tags basés sur le protocole
	tags = append(tags, "protocol_"+ne.Protocol)

	// Ajouter des tags basés sur l'état
	if ne.State == "ESTABLISHED" {
		tags = append(tags, "active_connection")
	}

	// Connexion externe
	if !nc.isPrivateIP(ne.DestIP) {
		tags = append(tags, "external_connection")
	}

	return tags
}
