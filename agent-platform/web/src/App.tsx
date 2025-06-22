import React from 'react'
import ChatInterface from './components/ChatInterface'

function App() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
      {/* Background Pattern */}
      <div className="absolute inset-0 bg-grid-pattern opacity-5"></div>
      
      <div className="relative z-10 container mx-auto px-4 py-8">
        <header className="text-center mb-8 animate-fade-in">
          <div className="flex items-center justify-center mb-4">
            <div className="w-16 h-16 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-2xl">
              <span className="text-3xl">🤖</span>
            </div>
          </div>
          <h1 className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent mb-3">
            AI Agent Platform
          </h1>
          <p className="text-gray-600 dark:text-gray-300 text-lg max-w-2xl mx-auto">
            🚀 智能代理助手 - 让AI帮您完成任务，体验未来的工作方式
          </p>
          <div className="flex items-center justify-center space-x-6 mt-4 text-sm text-gray-500 dark:text-gray-400">
            <div className="flex items-center space-x-1">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              <span>实时响应</span>
            </div>
            <div className="flex items-center space-x-1">
              <div className="w-2 h-2 bg-blue-400 rounded-full animate-pulse"></div>
              <span>智能理解</span>
            </div>
            <div className="flex items-center space-x-1">
              <div className="w-2 h-2 bg-purple-400 rounded-full animate-pulse"></div>
              <span>任务执行</span>
            </div>
          </div>
        </header>
        
        <main className="animate-fade-in" style={{animationDelay: '0.2s'}}>
          <ChatInterface />
        </main>
        
        <footer className="text-center mt-12 text-gray-400 dark:text-gray-500 animate-fade-in" style={{animationDelay: '0.4s'}}>
          <div className="flex items-center justify-center space-x-4 mb-2">
            <span className="text-sm">Powered by</span>
            <div className="flex items-center space-x-2">
              <span className="text-blue-500 font-semibold">React</span>
              <span className="text-gray-300">•</span>
              <span className="text-green-500 font-semibold">TypeScript</span>
              <span className="text-gray-300">•</span>
              <span className="text-purple-500 font-semibold">Go</span>
            </div>
          </div>
          <p className="text-sm">&copy; 2025 AI Agent Platform. Made with ❤️ for the future</p>
        </footer>
      </div>
    </div>
  )
}

export default App 