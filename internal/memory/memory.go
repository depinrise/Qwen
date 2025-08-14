package memory

import (
	"Qwen/internal/ai"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// MemoryService mengelola memory permanen user dengan LLM
type MemoryService struct {
	db       *sql.DB
	aiClient *ai.Client
}

// LLMResponse represents the response from LLM for memory management
type LLMResponse struct {
	MemoryUpdate json.RawMessage `json:"memory_update"`
	Reply        string          `json:"reply"`
}

// NewMemoryService membuat instance baru MemoryService
func NewMemoryService(db *sql.DB, aiClient *ai.Client) *MemoryService {
	return &MemoryService{
		db:       db,
		aiClient: aiClient,
	}
}

// buildPrompt membuat prompt untuk LLM dengan instruksi memory management
func (m *MemoryService) buildPrompt(currentMemory string, userMessage string) string {
	systemPrompt := `You are an AI assistant connected to a persistent memory database.

For every user message, you will:
1. Read the current stored memory JSON (if any)
2. Analyze the latest user message
3. Decide if there is new or updated information to store
4. Merge new information with existing memory
5. Output the updated memory and a natural reply to the user

Memory Rules:
- Memory can contain: User profile (name, age, gender, location, language), Interests and preferences, Current goals or tasks, Past conversation summaries, Promises/commitments/unfinished discussions, Any unique facts the user shared
- Replace old values if contradicted or updated
- Avoid storing trivial or irrelevant details
- Keep JSON concise (max 2KB)
- Never invent facts — only store explicitly shared or strongly implied info

Output Format (must always follow exactly):
{
  "memory_update": { ...merged updated memory JSON... },
  "reply": "Natural, contextual reply to the user"
}

IMPORTANT: Always respond with valid JSON in the exact format above. Never include markdown formatting or explanations outside the JSON.`

	userPrompt := fmt.Sprintf(`Current Memory:
%s

User Message:
%s

Please analyze and respond with the JSON format specified.`, currentMemory, userMessage)

	return fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
}

// SaveMemory menyimpan memory JSON ke database
func (m *MemoryService) SaveMemory(userID int64, memoryJSON string) error {
	query := `
		INSERT INTO user_memories (user_id, memory_key, memory_value)
		VALUES (?, 'user_memory', ?)
		ON DUPLICATE KEY UPDATE 
		memory_value = VALUES(memory_value),
		updated_at = CURRENT_TIMESTAMP
	`

	_, err := m.db.Exec(query, userID, memoryJSON)
	if err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}

	log.Printf("💾 Memory saved for user %d: %s", userID, memoryJSON)
	return nil
}

// GetMemory mengambil memory JSON user dari database
func (m *MemoryService) GetMemory(userID int64) (string, error) {
	query := `
		SELECT memory_value 
		FROM user_memories 
		WHERE user_id = ? AND memory_key = 'user_memory'
		ORDER BY updated_at DESC
		LIMIT 1
	`

	var memoryJSON string
	err := m.db.QueryRow(query, userID).Scan(&memoryJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return empty JSON if no memory found
			return "{}", nil
		}
		return "", fmt.Errorf("failed to get memory: %w", err)
	}

	return memoryJSON, nil
}

// ResetMemory menghapus semua memory user dari database
func (m *MemoryService) ResetMemory(userID int64) error {
	query := `DELETE FROM user_memories WHERE user_id = ?`

	result, err := m.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to reset memory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	log.Printf("🗑️ Memory reset for user %d: %d records deleted", userID, rowsAffected)
	return nil
}

// ProcessMessage memproses pesan user dengan LLM untuk memory management
// Returns: reply string, memorySaved bool, error
func (m *MemoryService) ProcessMessage(userID int64, message string) (string, bool, error) {
	// Ambil memory JSON user dari database
	currentMemory, err := m.GetMemory(userID)
	if err != nil {
		log.Printf("❌ Error getting memory: %v", err)
		currentMemory = "{}" // Fallback ke empty JSON
	}

	// Build prompt untuk LLM
	prompt := m.buildPrompt(currentMemory, message)

	// Kirim ke LLM untuk analisis dan update memory
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "Selalu balas dalam format JSON valid seperti yang diminta.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	ctx := context.Background()
	response, err := m.aiClient.Chat(ctx, messages)
	if err != nil {
		log.Printf("❌ Error getting LLM response: %v", err)
		return message, false, fmt.Errorf("failed to process with LLM: %w", err)
	}

	// Parse response JSON
	llmResponse, reply, err := m.parseResponse(response)
	if err != nil {
		log.Printf("❌ Error parsing LLM response: %v", err)
		return message, false, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Simpan updated memory ke database
	memorySaved := false
	if llmResponse.MemoryUpdate != nil {
		memoryJSON := string(llmResponse.MemoryUpdate)
		if err := m.SaveMemory(userID, memoryJSON); err != nil {
			log.Printf("❌ Error saving memory: %v", err)
		} else {
			memorySaved = true
		}
	}

	return reply, memorySaved, nil
}

// parseResponse parses LLM response and extracts memory update and reply
func (m *MemoryService) parseResponse(response string) (*LLMResponse, string, error) {
	// Clean response - remove any markdown formatting
	cleanResponse := strings.TrimSpace(response)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	var llmResponse LLMResponse
	err := json.Unmarshal([]byte(cleanResponse), &llmResponse)
	if err != nil {
		return nil, "", fmt.Errorf("failed to unmarshal JSON: %w, response: %s", err, cleanResponse)
	}

	return &llmResponse, llmResponse.Reply, nil
}
