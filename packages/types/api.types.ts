export interface Site {
  id: string
  domain: string
  originUrl: string
  rpsLimit: number
  burstSize: number
  active: boolean
  createdAt: string
}

export interface Tenant {
  id: string
  name: string
  email: string
  plan: 'STARTER' | 'GROWTH' | 'BUSINESS' | 'ENTERPRISE'
}

export interface LiveMetrics {
  liveVisitors: number
  queueDepth: number
  rps: number
  protectedRevenue: number
  systemHealth: 'healthy' | 'degraded' | 'down'
  circuitState: 'CLOSED' | 'OPEN' | 'HALF_OPEN'
}

export interface AnalyticsData {
  hour: string
  reqCount: number
  queuedCount: number
  blockedCount: number
  avgLatencyMs: number
}
