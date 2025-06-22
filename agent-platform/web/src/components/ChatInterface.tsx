import React, { useState, useRef, useEffect } from 'react'

interface Message {
  id: string
  type: 'user' | 'agent' | 'system'
  content: string
  timestamp: Date
  status?: 'thinking' | 'executing' | 'completed' | 'error'
  steps?: ExecutionStep[]
}

interface ExecutionStep {
  id: string
  title: string
  status: 'pending' | 'running' | 'completed' | 'error'
  details?: string
  timestamp: Date
}

interface ChatResponse {
  session_id: string
  message: string
  status: string
  steps?: ExecutionStep[]
  timestamp: string
}

interface StreamMessage {
  type: 'message' | 'step_update' | 'status_update' | 'error'
  session_id: string
  content?: string
  steps?: ExecutionStep[]
  status?: string
  error?: string
  timestamp: string
}

const ChatInterface: React.FC = () => {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      type: 'system',
      content: '👋 欢迎使用 AI Agent Platform！我是您的智能助手，可以帮助您执行各种任务。请告诉我您想要完成什么？',
      timestamp: new Date(),
      status: 'completed'
    }
  ])
  const [inputValue, setInputValue] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [currentSessionId, setCurrentSessionId] = useState<string>('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const eventSourceRef = useRef<EventSource | null>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // 清理 EventSource
  useEffect(() => {
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close()
      }
    }
  }, [])

  const sendMessageToAPI = async (message: string, sessionId?: string) => {
    try {
      const response = await fetch('/api/v1/chat', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          message: message,
          session_id: sessionId,
          stream: true,
        }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data: ChatResponse = await response.json()
      return data
    } catch (error) {
      console.error('Error sending message:', error)
      throw error
    }
  }

  const setupEventSource = (sessionId: string) => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
    }

    const eventSource = new EventSource(`/api/v1/chat/stream/${sessionId}`)
    eventSourceRef.current = eventSource

    eventSource.onmessage = (event) => {
      try {
        const data: StreamMessage = JSON.parse(event.data)
        handleStreamMessage(data)
      } catch (error) {
        console.error('Error parsing stream message:', error)
      }
    }

    eventSource.onerror = (error) => {
      console.error('EventSource error:', error)
      eventSource.close()
    }
  }

  const handleStreamMessage = (streamMsg: StreamMessage) => {
    setMessages(prev => prev.map(msg => {
      if (msg.id === `agent-${streamMsg.session_id}`) {
        const updatedMsg = { ...msg }
        
        // 更新状态
        if (streamMsg.status) {
          updatedMsg.status = streamMsg.status as any
        }
        
        // 更新步骤信息 - 对于 status_update 和 step_update 都要更新
        if (streamMsg.steps && streamMsg.steps.length > 0) {
          updatedMsg.steps = streamMsg.steps.map(step => ({
            ...step,
            timestamp: new Date(step.timestamp)
          }))
        }
        
        // 更新内容
        if (streamMsg.content && streamMsg.type !== 'step_update') {
          updatedMsg.content = streamMsg.content
        }
        
        return updatedMsg
      }
      return msg
    }))

    // 如果任务完成，停止加载状态
    if (streamMsg.status === 'completed' || streamMsg.status === 'error') {
      setIsLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputValue.trim() || isLoading) return

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: inputValue.trim(),
      timestamp: new Date()
    }

    setMessages(prev => [...prev, userMessage])
    const messageContent = inputValue.trim()
    setInputValue('')
    setIsLoading(true)

    try {
      // 发送消息到后端
      const response = await sendMessageToAPI(messageContent, currentSessionId)
      
      // 更新会话ID
      if (response.session_id !== currentSessionId) {
        setCurrentSessionId(response.session_id)
      }

      // 添加 agent 消息
      const agentMessage: Message = {
        id: `agent-${response.session_id}`,
        type: 'agent',
        content: response.message || '正在处理您的请求...',
        timestamp: new Date(response.timestamp),
        status: response.status as any,
        steps: response.steps?.map(step => ({
          ...step,
          timestamp: new Date(step.timestamp)
        })) || [
          { id: '1', title: '🧠 理解任务需求', status: 'pending', timestamp: new Date() },
          { id: '2', title: '📋 制定执行计划', status: 'pending', timestamp: new Date() },
          { id: '3', title: '⚡ 执行任务步骤', status: 'pending', timestamp: new Date() },
          { id: '4', title: '✅ 验证结果', status: 'pending', timestamp: new Date() },
        ]
      }

      setMessages(prev => [...prev, agentMessage])

      // 设置事件源监听
      setupEventSource(response.session_id)

    } catch (error) {
      console.error('Error:', error)
      setIsLoading(false)
      
      // 添加错误消息
      const errorMessage: Message = {
        id: Date.now().toString(),
        type: 'agent',
        content: '❌ 抱歉，处理您的请求时出现了错误。请稍后再试。',
        timestamp: new Date(),
        status: 'error'
      }
      setMessages(prev => [...prev, errorMessage])
    }
  }

  const getStatusIcon = (status?: string) => {
    switch (status) {
      case 'thinking':
        return <div className="flex items-center space-x-1">
          <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce"></div>
          <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
          <div className="w-2 h-2 bg-blue-500 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
        </div>
      case 'executing':
        return <div className="w-4 h-4 border-2 border-yellow-500 border-t-transparent rounded-full animate-spin"></div>
      case 'completed':
        return <div className="w-4 h-4 bg-green-500 rounded-full flex items-center justify-center">
          <div className="w-2 h-2 bg-white rounded-full"></div>
        </div>
      case 'error':
        return <div className="w-4 h-4 bg-red-500 rounded-full flex items-center justify-center">
          <div className="w-1 h-1 bg-white rounded-full"></div>
        </div>
      default:
        return null
    }
  }

  const getStepStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <div className="w-3 h-3 border-2 border-blue-400 border-t-transparent rounded-full animate-spin"></div>
      case 'completed':
        return <div className="w-3 h-3 bg-green-400 rounded-full flex items-center justify-center">
          <svg className="w-2 h-2 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
        </div>
      case 'error':
        return <div className="w-3 h-3 bg-red-400 rounded-full flex items-center justify-center">
          <svg className="w-2 h-2 text-white" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
          </svg>
        </div>
      default:
        return <div className="w-3 h-3 border-2 border-gray-300 rounded-full"></div>
    }
  }

  return (
    <div className="max-w-4xl mx-auto bg-white dark:bg-gray-900 rounded-2xl shadow-2xl overflow-hidden border border-gray-200 dark:border-gray-700">
      {/* Header */}
      <div className="bg-gradient-to-r from-blue-600 to-purple-600 p-6">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-10 h-10 bg-white bg-opacity-20 rounded-xl flex items-center justify-center">
              <span className="text-2xl">🤖</span>
            </div>
            <div>
              <h2 className="text-white text-xl font-bold">AI Agent Assistant</h2>
              <p className="text-blue-100 text-sm">智能代理助手</p>
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
            <div className="text-white text-sm">在线</div>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="h-96 overflow-y-auto p-6 space-y-6 bg-gradient-to-b from-gray-50 to-white dark:from-gray-800 dark:to-gray-900">
        {messages.map((message) => (
          <div
            key={message.id}
            className={`flex items-start space-x-3 animate-fade-in ${
              message.type === 'user' ? 'justify-end' : 'justify-start'
            }`}
          >
            {message.type !== 'user' && (
              <div className="flex-shrink-0">
                {message.type === 'agent' ? (
                  <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-xl flex items-center justify-center shadow-lg">
                    <span className="text-white text-lg">🤖</span>
                  </div>
                ) : (
                  <div className="w-10 h-10 bg-gradient-to-br from-gray-400 to-gray-600 rounded-xl flex items-center justify-center shadow-lg">
                    <span className="text-white text-xs font-bold">SYS</span>
                  </div>
                )}
              </div>
            )}
            
            <div className={`max-w-lg ${
              message.type === 'user' 
                ? 'bg-gradient-to-br from-blue-500 to-blue-600 text-white rounded-2xl rounded-br-md shadow-lg' 
                : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded-2xl rounded-bl-md shadow-lg border border-gray-200 dark:border-gray-700'
            } p-4 relative`}>
              {/* 消息状态指示器 */}
              {message.status && (
                <div className="absolute -top-2 -right-2 bg-white dark:bg-gray-800 rounded-full p-1 shadow-md">
                  {getStatusIcon(message.status)}
                </div>
              )}
              
              <div className="flex items-center justify-between mb-2">
                <span className={`text-xs font-medium ${
                  message.type === 'user' ? 'text-blue-100' : 'text-gray-500 dark:text-gray-400'
                }`}>
                  {message.timestamp.toLocaleTimeString()}
                </span>
              </div>
              
              <p className="text-sm leading-relaxed mb-2">{message.content}</p>
              
              {/* Execution Steps */}
              {message.steps && message.steps.length > 0 && (
                <div className="mt-4 p-3 bg-gray-50 dark:bg-gray-700 rounded-xl">
                  <div className="text-xs font-bold text-gray-600 dark:text-gray-300 mb-3 flex items-center">
                    <span className="mr-1">⚡</span> 执行进度
                  </div>
                  <div className="space-y-3">
                    {message.steps.map((step, index) => (
                      <div key={step.id} className="flex items-center space-x-3 group">
                        <div className="flex-shrink-0">
                          {getStepStatusIcon(step.status)}
                        </div>
                        <div className="flex-1">
                          <div className={`text-xs font-medium transition-all duration-300 ${
                            step.status === 'completed' 
                              ? 'line-through text-gray-400 dark:text-gray-500' 
                              : step.status === 'running'
                              ? 'text-blue-600 dark:text-blue-400 font-semibold'
                              : 'text-gray-600 dark:text-gray-300'
                          }`}>
                            {step.title}
                          </div>
                          {step.details && (
                            <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                              {step.details}
                            </div>
                          )}
                        </div>
                        {step.status === 'completed' && (
                          <div className="text-xs text-green-500 font-medium opacity-0 group-hover:opacity-100 transition-opacity">
                            完成
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>

            {message.type === 'user' && (
              <div className="flex-shrink-0">
                <div className="w-10 h-10 bg-gradient-to-br from-gray-400 to-gray-600 rounded-xl flex items-center justify-center shadow-lg">
                  <span className="text-white text-lg">👤</span>
                </div>
              </div>
            )}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t border-gray-200 dark:border-gray-700 p-6 bg-white dark:bg-gray-800">
        <form onSubmit={handleSubmit} className="flex space-x-4">
          <div className="flex-1 relative">
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="输入您的任务需求..."
              className="w-full border-2 border-gray-200 dark:border-gray-600 rounded-xl px-4 py-3 pr-12 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white placeholder-gray-400 transition-all duration-200"
              disabled={isLoading}
            />
            <div className="absolute right-3 top-3 text-gray-400">
              <span className="text-lg">💬</span>
            </div>
          </div>
          <button
            type="submit"
            disabled={!inputValue.trim() || isLoading}
            className="bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 disabled:from-gray-400 disabled:to-gray-500 disabled:cursor-not-allowed text-white rounded-xl px-6 py-3 transition-all duration-200 shadow-lg hover:shadow-xl transform hover:scale-105 disabled:transform-none min-w-[80px] font-medium"
          >
            {isLoading ? (
              <div className="flex items-center justify-center">
                <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
              </div>
            ) : (
              <div className="flex items-center space-x-1">
                <span>发送</span>
                <span className="text-lg">🚀</span>
              </div>
            )}
          </button>
        </form>
        
        {/* Session Info */}
        {currentSessionId && (
          <div className="mt-3 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-green-400 rounded-full"></div>
              <span>会话活跃 • ID: {currentSessionId.substring(0, 8)}...</span>
            </div>
            <div className="text-xs text-gray-400">
              AI Agent Platform v1.0
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default ChatInterface 