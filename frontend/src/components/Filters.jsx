import { useState } from 'react'
import { Filter, X, Calendar, Shield, Server } from 'lucide-react'

function Filters({ onFilterChange, onReset }) {
  const [isOpen, setIsOpen] = useState(false)
  const [filters, setFilters] = useState({
    eventType: '',
    severity: '',
    hostname: '',
    searchQuery: '',
    dateRange: '24h'
  })

  const handleFilterChange = (key, value) => {
    const newFilters = { ...filters, [key]: value }
    setFilters(newFilters)
    onFilterChange(newFilters)
  }

  const handleReset = () => {
    const resetFilters = {
      eventType: '',
      severity: '',
      hostname: '',
      searchQuery: '',
      dateRange: '24h'
    }
    setFilters(resetFilters)
    onReset()
  }

  const activeFilterCount = Object.values(filters).filter(v => v && v !== '24h').length

  return (
    <div className="filters-container">
      {/* Search Bar */}
      <div className="search-bar">
        <input
          type="text"
          placeholder="Search events..."
          value={filters.searchQuery}
          onChange={(e) => handleFilterChange('searchQuery', e.target.value)}
          className="search-input"
        />
      </div>

      {/* Filter Toggle Button */}
      <button 
        className={`filter-toggle-btn ${isOpen ? 'active' : ''}`}
        onClick={() => setIsOpen(!isOpen)}
      >
        <Filter size={16} />
        Filters
        {activeFilterCount > 0 && (
          <span className="filter-badge">{activeFilterCount}</span>
        )}
      </button>

      {/* Filter Panel */}
      {isOpen && (
        <div className="filter-panel">
          <div className="filter-header">
            <h3>Filters</h3>
            <button className="close-btn" onClick={() => setIsOpen(false)}>
              <X size={18} />
            </button>
          </div>

          <div className="filter-grid">
            {/* Date Range */}
            <div className="filter-group">
              <label>
                <Calendar size={16} />
                Time Range
              </label>
              <select
                value={filters.dateRange}
                onChange={(e) => handleFilterChange('dateRange', e.target.value)}
              >
                <option value="1h">Last Hour</option>
                <option value="6h">Last 6 Hours</option>
                <option value="24h">Last 24 Hours</option>
                <option value="7d">Last 7 Days</option>
                <option value="30d">Last 30 Days</option>
              </select>
            </div>

            {/* Event Type */}
            <div className="filter-group">
              <label>
                <Shield size={16} />
                Event Type
              </label>
              <select
                value={filters.eventType}
                onChange={(e) => handleFilterChange('eventType', e.target.value)}
              >
                <option value="">All Types</option>
                <option value="system">System</option>
                <option value="network">Network</option>
                <option value="process">Process</option>
                <option value="file">File</option>
              </select>
            </div>

            {/* Severity */}
            <div className="filter-group">
              <label>
                <Shield size={16} />
                Severity
              </label>
              <select
                value={filters.severity}
                onChange={(e) => handleFilterChange('severity', e.target.value)}
              >
                <option value="">All Severities</option>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>

            {/* Hostname */}
            <div className="filter-group">
              <label>
                <Server size={16} />
                Hostname
              </label>
              <input
                type="text"
                placeholder="Filter by hostname..."
                value={filters.hostname}
                onChange={(e) => handleFilterChange('hostname', e.target.value)}
              />
            </div>
          </div>

          {/* Reset Button */}
          <div className="filter-actions">
            <button className="reset-btn" onClick={handleReset}>
              <X size={16} />
              Reset Filters
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default Filters
