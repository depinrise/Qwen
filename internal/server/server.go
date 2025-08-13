package server

import (
	"Qwen/internal/ai"
	"Qwen/internal/websocket"
	"log"
	"net/http"
)

type Server struct {
	hub  *websocket.Hub
	port string
}

func NewServer(aiClient *ai.Client, port string) *Server {
	hub := websocket.NewHub(aiClient)

	return &Server{
		hub:  hub,
		port: port,
	}
}

func (s *Server) Start() error {
	// Start the WebSocket hub
	go s.hub.Run()

	// Setup routes
	http.HandleFunc("/ws", s.hub.ServeWS)
	http.HandleFunc("/", s.serveHome)
	http.HandleFunc("/health", s.healthCheck)

	log.Printf("HTTP server starting on port %s", s.port)
	return http.ListenAndServe(":"+s.port, nil)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(homeHTML))
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

const homeHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Telegram AI Bot - WebSocket Test</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        #messages {
            height: 400px;
            border: 1px solid #ddd;
            padding: 15px;
            overflow-y: auto;
            background: #fafafa;
            border-radius: 8px;
            margin-bottom: 20px;
            font-family: monospace;
        }
        .message {
            margin-bottom: 10px;
            padding: 8px 12px;
            border-radius: 6px;
            animation: fadeIn 0.3s ease-in;
        }
        .user-message {
            background: #007bff;
            color: white;
            text-align: right;
        }
        .ai-message {
            background: #e9ecef;
            color: #333;
        }
        .streaming { background: #e2e3e5; color: #383d41; }
        .complete { background: #d1ecf1; color: #0c5460; font-weight: bold; }
        .error { background: #f8d7da; color: #721c24; }
        
        .input-container {
            display: flex;
            gap: 10px;
        }
        #messageInput {
            flex: 1;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 14px;
        }
        #sendButton {
            padding: 12px 24px;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 14px;
        }
        #sendButton:hover:not(:disabled) {
            background: #0056b3;
        }
        #sendButton:disabled {
            background: #6c757d;
            cursor: not-allowed;
        }
        .status {
            text-align: center;
            margin-bottom: 20px;
            padding: 10px;
            border-radius: 6px;
        }
        .connected { background: #d4edda; color: #155724; }
        .disconnected { background: #f8d7da; color: #721c24; }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .typing-indicator {
            display: none;
            font-style: italic;
            color: #666;
            margin-top: 10px;
        }
        .typing-indicator.show {
            display: block;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Telegram AI Bot - Real-time Testing</h1>
        
        <div id="status" class="status disconnected">
            Disconnected
        </div>
        
        <div id="messages"></div>
        <div id="typingIndicator" class="typing-indicator">AI is typing...</div>
        
        <div class="input-container">
            <input type="text" id="messageInput" placeholder="Type your message here..." 
                   onkeypress="if(event.key==='Enter') sendMessage()">
            <button id="sendButton" onclick="sendMessage()">Send</button>
        </div>
    </div>

    <script>
        let ws;
        let isConnected = false;
        const messages = document.getElementById('messages');
        const messageInput = document.getElementById('messageInput');
        const sendButton = document.getElementById('sendButton');
        const status = document.getElementById('status');
        const typingIndicator = document.getElementById('typingIndicator');
        
        let currentStreamingMessage = null;
        
        function connect() {
            const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = wsProtocol + '//' + window.location.host + '/ws?user_id=web_user';
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                isConnected = true;
                status.textContent = 'Connected';
                status.className = 'status connected';
                sendButton.disabled = false;
            };
            
            ws.onclose = function() {
                isConnected = false;
                status.textContent = 'Disconnected';
                status.className = 'status disconnected';
                sendButton.disabled = true;
                
                // Auto-reconnect after 3 seconds
                setTimeout(connect, 3000);
            };
            
            ws.onmessage = function(event) {
                const message = JSON.parse(event.data);
                handleAIMessage(message);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
            };
        }
        
        function sendMessage() {
            const text = messageInput.value.trim();
            if (!text || !isConnected) return;
            
            // Display user message
            addMessage('user', text);
            
            // Send to WebSocket
            ws.send(JSON.stringify({
                type: 'user_message',
                content: text
            }));
            
            messageInput.value = '';
            currentStreamingMessage = null;
        }
        
        function handleAIMessage(message) {
            if (message.type === 'ai_response') {
                const stage = message.stage;
                
                if (stage === 'streaming') {
                    // Stream the response
                    if (!currentStreamingMessage) {
                        currentStreamingMessage = addMessage('ai', '', 'streaming');
                        showTypingIndicator();
                    }
                    currentStreamingMessage.textContent += message.content;
                    scrollToBottom();
                } else if (stage === 'complete') {
                    // Response complete
                    hideTypingIndicator();
                    if (currentStreamingMessage) {
                        currentStreamingMessage.className = 'message ai-message complete';
                    }
                    currentStreamingMessage = null;
                } else if (stage === 'error') {
                    hideTypingIndicator();
                    addMessage('ai', message.content, 'error');
                    currentStreamingMessage = null;
                }
            }
        }
        
        function addMessage(sender, text, stage = '') {
            const messageEl = document.createElement('div');
            messageEl.className = 'message ' + 
                (sender === 'user' ? 'user-message' : 'ai-message' + (stage ? ' ' + stage : ''));
            messageEl.textContent = text;
            
            messages.appendChild(messageEl);
            scrollToBottom();
            
            return messageEl;
        }
        
        function showTypingIndicator() {
            typingIndicator.classList.add('show');
        }
        
        function hideTypingIndicator() {
            typingIndicator.classList.remove('show');
        }
        
        function scrollToBottom() {
            messages.scrollTop = messages.scrollHeight;
        }
        
        // Connect on page load
        connect();
        
        // Focus input
        messageInput.focus();
    </script>
</body>
</html>`
