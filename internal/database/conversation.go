package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Conversation struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatSession struct {
	ID           int64           `json:"id"`
	UserID       string          `json:"user_id"`
	SessionData  json.RawMessage `json:"session_data"`
	LastActivity time.Time       `json:"last_activity"`
	CreatedAt    time.Time       `json:"created_at"`
}

type ConversationService struct {
	db *DB
}

func NewConversationService(db *DB) *ConversationService {
	return &ConversationService{db: db}
}

// SaveConversation saves a conversation to the database
func (cs *ConversationService) SaveConversation(userID, userName, message, response string) error {
	query := `
		INSERT INTO conversations (user_id, user_name, message, response) 
		VALUES (?, ?, ?, ?)
	`

	_, err := cs.db.conn.Exec(query, userID, userName, message, response)
	if err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}

	return nil
}

// GetRecentConversations gets recent conversations for a user
func (cs *ConversationService) GetRecentConversations(userID string, limit int) ([]Conversation, error) {
	query := `
		SELECT id, user_id, user_name, message, response, created_at 
		FROM conversations 
		WHERE user_id = ? 
		ORDER BY created_at DESC 
		LIMIT ?
	`

	rows, err := cs.db.conn.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.UserID, &conv.UserName, &conv.Message, &conv.Response, &conv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// GetConversationContext builds context from recent conversations
func (cs *ConversationService) GetConversationContext(userID string, maxMessages int) string {
	conversations, err := cs.GetRecentConversations(userID, maxMessages)
	if err != nil || len(conversations) == 0 {
		return ""
	}

	var context string
	// Reverse to get chronological order (oldest first)
	for i := len(conversations) - 1; i >= 0; i-- {
		conv := conversations[i]
		context += fmt.Sprintf("User: %s\nAI: %s\n\n", conv.Message, conv.Response)
	}

	return context
}

// UpdateSession updates or creates a chat session
func (cs *ConversationService) UpdateSession(userID string, sessionData interface{}) error {
	dataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	query := `
		INSERT INTO chat_sessions (user_id, session_data) 
		VALUES (?, ?) 
		ON DUPLICATE KEY UPDATE 
		session_data = VALUES(session_data), 
		last_activity = CURRENT_TIMESTAMP
	`

	_, err = cs.db.conn.Exec(query, userID, dataJSON)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// GetSession gets a chat session
func (cs *ConversationService) GetSession(userID string) (*ChatSession, error) {
	query := `
		SELECT id, user_id, session_data, last_activity, created_at 
		FROM chat_sessions 
		WHERE user_id = ?
	`

	var session ChatSession
	err := cs.db.conn.QueryRow(query, userID).Scan(
		&session.ID, &session.UserID, &session.SessionData,
		&session.LastActivity, &session.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No session found
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// CleanOldConversations removes conversations older than specified days
func (cs *ConversationService) CleanOldConversations(days int) error {
	query := `
		DELETE FROM conversations 
		WHERE created_at < DATE_SUB(NOW(), INTERVAL ? DAY)
	`

	result, err := cs.db.conn.Exec(query, days)
	if err != nil {
		return fmt.Errorf("failed to clean old conversations: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("🧹 Cleaned %d old conversation records\n", rowsAffected)
	}

	return nil
}
