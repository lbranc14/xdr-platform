import { Activity, AlertTriangle, Shield, Database } from 'lucide-react'

function StatsCards({ totalEvents, severityCounts, typeCounts }) {
  return (
    <div className="stats-grid">
      {/* Total Events */}
      <div className="stat-card">
        <div className="stat-icon blue">
          <Database size={24} />
        </div>
        <div className="stat-content">
          <p className="stat-label">Total Events</p>
          <p className="stat-value">{totalEvents.toLocaleString()}</p>
        </div>
      </div>

      {/* Critical Events */}
      <div className="stat-card">
        <div className="stat-icon red">
          <AlertTriangle size={24} />
        </div>
        <div className="stat-content">
          <p className="stat-label">Critical</p>
          <p className="stat-value">{severityCounts.critical}</p>
        </div>
      </div>

      {/* High Severity */}
      <div className="stat-card">
        <div className="stat-icon orange">
          <Shield size={24} />
        </div>
        <div className="stat-content">
          <p className="stat-label">High Severity</p>
          <p className="stat-value">{severityCounts.high}</p>
        </div>
      </div>

      {/* Active Monitoring */}
      <div className="stat-card">
        <div className="stat-icon green">
          <Activity size={24} />
        </div>
        <div className="stat-content">
          <p className="stat-label">Event Types</p>
          <p className="stat-value">{Object.keys(typeCounts).length}</p>
        </div>
      </div>
    </div>
  )
}

export default StatsCards
