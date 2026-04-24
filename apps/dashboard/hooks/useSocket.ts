import { useEffect, useRef, useState } from 'react'
import { io, Socket } from 'socket.io-client'

interface Metrics {
  liveVisitors: number
  queueDepth: number
  rps: number
  protectedRevenue: number
  systemHealth: 'healthy' | 'degraded' | 'down'
}

export function useSocket(siteId: string) {
  const socketRef = useRef<Socket | null>(null)
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [connected, setConnected] = useState(false)

  useEffect(() => {
    socketRef.current = io(process.env.NEXT_PUBLIC_API_URL || 'http://localhost:4000', {
      auth: { token: localStorage.getItem('accessToken') }
    })

    socketRef.current.on('connect', () => setConnected(true))
    socketRef.current.on('disconnect', () => setConnected(false))

    socketRef.current.emit('subscribe', { siteId })
    socketRef.current.on('metrics', (data: Metrics) => setMetrics(data))

    return () => {
      socketRef.current?.disconnect()
    }
  }, [siteId])

  return { metrics, connected }
}
