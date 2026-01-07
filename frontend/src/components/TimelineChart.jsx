import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts'
import { TrendingUp } from 'lucide-react'

function TimelineChart({ timelineData }) {
  // Agréger les données par timestamp et sévérité
  const aggregateData = () => {
    if (!timelineData || timelineData.length === 0) {
      return []
    }

    const grouped = {}
    
    timelineData.forEach(item => {
      const time = new Date(item.timestamp).toLocaleTimeString('fr-FR', { 
        hour: '2-digit', 
        minute: '2-digit' 
      })
      
      if (!grouped[time]) {
        grouped[time] = { time, low: 0, medium: 0, high: 0, critical: 0, total: 0 }
      }
      
      grouped[time][item.severity] = (grouped[time][item.severity] || 0) + item.count
      grouped[time].total += item.count
    })

    return Object.values(grouped).sort((a, b) => {
      const timeA = a.time.split(':').map(Number)
      const timeB = b.time.split(':').map(Number)
      return timeA[0] * 60 + timeA[1] - (timeB[0] * 60 + timeB[1])
    })
  }

  const data = aggregateData()

  if (data.length === 0) {
    return (
      <div className="chart-card">
        <h3>
          <TrendingUp size={20} />
          Events Timeline
        </h3>
        <div className="no-data-chart">
          <p>No timeline data available</p>
        </div>
      </div>
    )
  }

  return (
    <div className="chart-card">
      <h3>
        <TrendingUp size={20} />
        Events Timeline (Last 24 Hours)
      </h3>
      <ResponsiveContainer width="100%" height={300}>
        <AreaChart data={data}>
          <defs>
            <linearGradient id="colorCritical" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#ef4444" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorHigh" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#f59e0b" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#f59e0b" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorMedium" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
            </linearGradient>
            <linearGradient id="colorLow" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor="#10b981" stopOpacity={0.3}/>
              <stop offset="95%" stopColor="#10b981" stopOpacity={0}/>
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis 
            dataKey="time" 
            stroke="#9ca3af"
            style={{ fontSize: '12px' }}
          />
          <YAxis 
            stroke="#9ca3af"
            style={{ fontSize: '12px' }}
          />
          <Tooltip 
            contentStyle={{ 
              backgroundColor: '#1f2937', 
              border: '1px solid #374151',
              borderRadius: '8px',
              color: '#e5e7eb'
            }}
          />
          <Legend 
            wrapperStyle={{ fontSize: '12px' }}
            iconType="circle"
          />
          <Area 
            type="monotone" 
            dataKey="critical" 
            stackId="1"
            stroke="#ef4444" 
            fill="url(#colorCritical)" 
            name="Critical"
          />
          <Area 
            type="monotone" 
            dataKey="high" 
            stackId="1"
            stroke="#f59e0b" 
            fill="url(#colorHigh)" 
            name="High"
          />
          <Area 
            type="monotone" 
            dataKey="medium" 
            stackId="1"
            stroke="#3b82f6" 
            fill="url(#colorMedium)" 
            name="Medium"
          />
          <Area 
            type="monotone" 
            dataKey="low" 
            stackId="1"
            stroke="#10b981" 
            fill="url(#colorLow)" 
            name="Low"
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}

export default TimelineChart
