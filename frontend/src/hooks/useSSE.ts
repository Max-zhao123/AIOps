import { useRef, useCallback, useState } from 'react'

interface SSEOptions {
  onMessage: (data: string) => void
  onComplete: () => void
  onError: (err: Event) => void
}

export function useSSE() {
  const readerRef = useRef<ReadableStreamDefaultReader<Uint8Array> | null>(null)
  const [isConnected, setIsConnected] = useState(false)

  const connect = useCallback(async (url: string, token: string | null, options: SSEOptions) => {
    try {
      const headers: Record<string, string> = { Accept: 'text/event-stream' }
      if (token) headers.Authorization = `Bearer ${token}`

      const response = await fetch(url, { headers })
      if (!response.ok) throw new Error(`SSE error: ${response.status}`)

      const reader = response.body?.getReader()
      if (!reader) throw new Error('No reader')

      readerRef.current = reader
      setIsConnected(true)

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) {
          options.onComplete()
          break
        }

        buffer += decoder.decode(value, { stream: true })
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6).trim()
            if (data === '[DONE]') {
              options.onComplete()
              return
            }
            options.onMessage(data)
          } else if (line.startsWith('data:')) {
            const data = line.slice(5).trim()
            if (data === '[DONE]') {
              options.onComplete()
              return
            }
            options.onMessage(data)
          }
        }
      }
    } catch (err) {
      options.onError(err as Event)
    } finally {
      setIsConnected(false)
    }
  }, [])

  const disconnect = useCallback(() => {
    readerRef.current?.cancel()
    readerRef.current = null
    setIsConnected(false)
  }, [])

  return { connect, disconnect, isConnected }
}
