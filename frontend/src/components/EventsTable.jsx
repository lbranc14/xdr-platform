import { AlertTriangle, Info, AlertCircle, XOctagon } from 'lucide-react'

function EventsTable({ events }) {
  const getSeverityIcon = (severity) => {
    switch (severity) {
      case 'critical':
        return <XOctagon size={16} className="severity-icon critical" />
      case 'high':
        return <AlertTriangle size={16} className="severity-icon high" />
      case 'medium':
        return <AlertCircle size={16} className="severity-icon medium" />
      case 'low':
        return <Info size={16} className="severity-icon low" />
      default:
        return <Info size={16} className="severity-icon" />
    }
  }

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleString('fr-FR', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  }

  return (
    <div className="table-container">
      <table className="events-table">
        <thead>
          <tr>
            <th>Timestamp</th>
            <th>Severity</th>
            <th>Type</th>
            <th>Hostname</th>
            <th>Agent ID</th>
            <th>Process</th>
            <th>Details</th>
          </tr>
        </thead>
        <tbody>
          {events.length === 0 ? (
            <tr>
              <td colSpan="7" className="no-data">
                No events found
              </td>
            </tr>
          ) : (
            events.map((event, index) => (
              <tr key={index} className="event-row">
                <td className="timestamp">{formatTimestamp(event.timestamp)}</td>
                <td>
                  <span className={`severity-badge ${event.severity}`}>
                    {getSeverityIcon(event.severity)}
                    {event.severity}
                  </span>
                </td>
                <td>
                  <span className="type-badge">{event.event_type}</span>
                </td>
                <td className="hostname">{event.hostname}</td>
                <td className="agent-id">{event.agent_id}</td>
                <td className="process-name">
                  {event.process_name || '-'}
                  {event.process_pid && ` (${event.process_pid})`}
                </td>
                <td className="tags">
                  {event.tags && event.tags.length > 0 ? (
                    event.tags.slice(0, 2).map((tag, i) => (
                      <span key={i} className="tag">{tag}</span>
                    ))
                  ) : (
                    '-'
                  )}
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}

export default EventsTable
