import { useState, useEffect } from 'react'
import axios from 'axios'
import { Activity, Database, Shield, AlertTriangle, CheckCircle, XCircle, Download } from 'lucide-react'
import './App.css'
import './App-enhanced.css'
import EventsTable from './components/EventsTable'
import StatsCards from './components/StatsCards'
import EventsChart from './components/EventsChart'
import TimelineChart from './components/TimelineChart'
import Filters from './components/Filters'

const API_BASE_URL = ''

function App() {
  const [events, setEvents] = useState([])
  const [filteredEvents, setFilteredEvents] = useState([])
  const [stats, setStats] = useState(null)
  const [timelineData, setTimelineData] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)
  const [autoRefresh, setAutoRefresh] = useState(true)
  const [filters, setFilters] = useState({})

  // Fonction pour récupérer les événements
  const fetchEvents = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/events?limit=200`)
      setEvents(response.data.events || [])
      setFilteredEvents(response.data.events || [])
      setError(null)
    } catch (err) {
      console.error('Error fetching events:', err)
      setError('Failed to fetch events')
    }
  }

  // Fonction pour récupérer les stats
  const fetchStats = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/events/stats`)
      setStats(response.data.stats)
    } catch (err) {
      console.error('Error fetching stats:', err)
    }
  }

  // Fonction pour récupérer la timeline
  const fetchTimeline = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/events/timeline?interval=1 hour&hours=24`)
      setTimelineData(response.data.data || [])
    } catch (err) {
      console.error('Error fetching timeline:', err)
    }
  }

  // Fonction pour récupérer toutes les données
  const fetchData = async () => {
    setLoading(true)
    await Promise.all([fetchEvents(), fetchStats(), fetchTimeline()])
    setLoading(false)
  }

  // Charger les données au démarrage
  useEffect(() => {
    fetchData()
  }, [])

  // Auto-refresh toutes les 10 secondes
  useEffect(() => {
    if (!autoRefresh) return

    const interval = setInterval(() => {
      fetchData()
    }, 10000)

    return () => clearInterval(interval)
  }, [autoRefresh])

  // Appliquer les filtres
  useEffect(() => {
    let filtered = [...events]

    // Filtre par type
    if (filters.eventType) {
      filtered = filtered.filter(e => e.event_type === filters.eventType)
    }

    // Filtre par sévérité
    if (filters.severity) {
      filtered = filtered.filter(e => e.severity === filters.severity)
    }

    // Filtre par hostname
    if (filters.hostname) {
      filtered = filtered.filter(e => 
        e.hostname.toLowerCase().includes(filters.hostname.toLowerCase())
      )
    }

    // Filtre par recherche
    if (filters.searchQuery) {
      const query = filters.searchQuery.toLowerCase()
      filtered = filtered.filter(e =>
        e.hostname.toLowerCase().includes(query) ||
        e.event_type.toLowerCase().includes(query) ||
        e.severity.toLowerCase().includes(query) ||
        e.agent_id.toLowerCase().includes(query) ||
        (e.process_name && e.process_name.toLowerCase().includes(query))
      )
    }

    // Filtre par date range
    if (filters.dateRange && filters.dateRange !== '24h') {
      const now = new Date()
      let cutoff = new Date()
      
      switch(filters.dateRange) {
        case '1h':
          cutoff.setHours(now.getHours() - 1)
          break
        case '6h':
          cutoff.setHours(now.getHours() - 6)
          break
        case '7d':
          cutoff.setDate(now.getDate() - 7)
          break
        case '30d':
          cutoff.setDate(now.getDate() - 30)
          break
        default:
          cutoff.setHours(now.getHours() - 24)
      }
      
      filtered = filtered.filter(e => new Date(e.timestamp) >= cutoff)
    }

    setFilteredEvents(filtered)
  }, [filters, events])

  // Handler pour les filtres
  const handleFilterChange = (newFilters) => {
    setFilters(newFilters)
  }

  const handleResetFilters = () => {
    setFilters({})
    setFilteredEvents(events)
  }

  // Export des données en CSV
  const exportToCSV = () => {
    const headers = ['Timestamp', 'Severity', 'Type', 'Hostname', 'Agent ID', 'Process']
    const rows = filteredEvents.map(e => [
      e.timestamp,
      e.severity,
      e.event_type,
      e.hostname,
      e.agent_id,
      e.process_name || '-'
    ])

    const csvContent = [
      headers.join(','),
      ...rows.map(row => row.join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `xdr-events-${new Date().toISOString()}.csv`
    a.click()
  }

  // Calculer les stats par sévérité
  const getEventsBySeverity = () => {
    const counts = { low: 0, medium: 0, high: 0, critical: 0 }
    filteredEvents.forEach(event => {
      if (counts[event.severity] !== undefined) {
        counts[event.severity]++
      }
    })
    return counts
  }

  // Calculer les stats par type
  const getEventsByType = () => {
    const counts = {}
    filteredEvents.forEach(event => {
      counts[event.event_type] = (counts[event.event_type] || 0) + 1
    })
    return counts
  }

  const severityCounts = getEventsBySeverity()
  const typeCounts = getEventsByType()

  return (
    <div className="app">
      {/* Header */}
      <header className="header">
        <div className="header-content">
          <div className="header-left">
            <Shield className="logo-icon" size={32} />
            <div>
              <h1>XDR Platform</h1>
              <p className="subtitle">Security Operations Center</p>
            </div>
          </div>
          <div className="header-right">
            <button 
              className={`refresh-btn ${autoRefresh ? 'active' : ''}`}
              onClick={() => setAutoRefresh(!autoRefresh)}
            >
              <Activity size={16} />
              Auto-refresh {autoRefresh ? 'ON' : 'OFF'}
            </button>
            <button className="refresh-btn" onClick={fetchData}>
              Refresh Now
            </button>
            <button className="export-btn" onClick={exportToCSV}>
              <Download size={16} />
              Export CSV
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="main-content">
        {loading && !events.length ? (
          <div className="loading">
            <Activity className="spin" size={48} />
            <p>Loading dashboard...</p>
          </div>
        ) : error ? (
          <div className="error">
            <XCircle size={48} />
            <p>{error}</p>
            <button onClick={fetchData}>Retry</button>
          </div>
        ) : (
          <>
            {/* Stats Cards */}
            <StatsCards 
              totalEvents={stats?.total_events || 0}
              severityCounts={severityCounts}
              typeCounts={typeCounts}
            />

            {/* Filters */}
            <Filters 
              onFilterChange={handleFilterChange}
              onReset={handleResetFilters}
            />

            {/* Timeline Chart */}
            <TimelineChart timelineData={timelineData} />

            {/* Charts Section */}
            <div className="charts-section">
              <EventsChart events={filteredEvents} />
            </div>

            {/* Events Table */}
            <div className="table-section">
              <div className="section-header">
                <h2>Recent Events</h2>
                <span className="badge">
                  {filteredEvents.length} / {events.length} events
                </span>
              </div>
              <EventsTable events={filteredEvents} />
            </div>
          </>
        )}
      </main>

      {/* Footer */}
      <footer className="footer">
        <p>XDR Platform v1.0 - Made with ❤️ for Security Operations</p>
        {stats && (
          <p className="footer-stats">
            Last updated: {new Date(stats.last_updated).toLocaleString()}
          </p>
        )}
      </footer>
    </div>
  )
}

export default App
