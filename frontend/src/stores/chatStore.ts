import { create } from 'zustand'
import type { ChatSession, ChatMessage } from '@/types'

interface ChatState {
  sessions: ChatSession[]
  currentSessionId: number | null
  messages: ChatMessage[]
  isStreaming: boolean
  setSessions: (sessions: ChatSession[]) => void
  setCurrentSession: (id: number | null) => void
  setMessages: (messages: ChatMessage[]) => void
  addMessage: (msg: ChatMessage) => void
  appendToLastAssistant: (chunk: string) => void
  setStreaming: (v: boolean) => void
}

export const useChatStore = create<ChatState>((set, get) => ({
  sessions: [],
  currentSessionId: null,
  messages: [],
  isStreaming: false,

  setSessions: (sessions) => set({ sessions }),
  setCurrentSession: (id) => set({ currentSessionId: id }),

  setMessages: (messages) => set({ messages }),

  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),

  appendToLastAssistant: (chunk) =>
    set((s) => {
      const msgs = [...s.messages]
      const last = msgs[msgs.length - 1]
      if (last && last.role === 'assistant') {
        msgs[msgs.length - 1] = { ...last, content: last.content + chunk }
      }
      return { messages: msgs }
    }),

  setStreaming: (v) => set({ isStreaming: v }),
}))
