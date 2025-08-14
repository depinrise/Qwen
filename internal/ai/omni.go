package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OmniMedia struct {
	Mime       string
	DataBase64 string
}

type OmniResponse struct {
	Text     string
	AudioMP3 []byte
}

// ChatOmni sends a multimodal request (text + optional image/audio/video) and can request audio output
func (c *Client) ChatOmni(ctx context.Context, systemPrompt string, userText string, images []OmniMedia, inputAudio *OmniMedia, videoURL string, wantAudio bool) (OmniResponse, error) {
	// Build OpenAI-compatible JSON payload
	// We construct raw JSON to support multimodal parts regardless of SDK version

	// Build user content parts
	userContent := make([]map[string]any, 0, 1+len(images)+1)
	if userText != "" {
		userContent = append(userContent, map[string]any{
			"type": "input_text",
			"text": userText,
		})
	}

	for _, img := range images {
		dataURL := fmt.Sprintf("data:%s;base64,%s", img.Mime, img.DataBase64)
		userContent = append(userContent, map[string]any{
			"type":      "input_image",
			"image_url": map[string]any{"url": dataURL},
		})
	}

	if inputAudio != nil && inputAudio.DataBase64 != "" {
		// Pass input audio as base64 with declared format from MIME (e.g., audio/ogg)
		userContent = append(userContent, map[string]any{
			"type": "input_audio",
			"input_audio": map[string]any{
				"data":   inputAudio.DataBase64,
				"format": mimeToSimpleFormat(inputAudio.Mime),
			},
		})
	}

	if videoURL != "" {
		userContent = append(userContent, map[string]any{
			"type":      "input_video",
			"video_url": map[string]any{"url": videoURL},
		})
	}

	messages := []map[string]any{}
	if systemPrompt != "" {
		messages = append(messages, map[string]any{
			"role": "system",
			"content": []map[string]any{{
				"type": "text",
				"text": systemPrompt,
			}},
		})
	}
	messages = append(messages, map[string]any{
		"role":    "user",
		"content": userContent,
	})

	body := map[string]any{
		"model":    c.model,
		"messages": messages,
		"stream":   false,
	}

	if wantAudio {
		body["modalities"] = []string{"text", "audio"}
		body["audio"] = map[string]any{
			"voice":  "alloy",
			"format": "mp3",
		}
	}

	// Sampling params (align with client params)
	body["temperature"] = c.params.Temperature
	body["top_p"] = c.params.TopP
	body["top_k"] = c.params.TopK

	payload, err := json.Marshal(body)
	if err != nil {
		return OmniResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return OmniResponse{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return OmniResponse{}, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return OmniResponse{}, fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return OmniResponse{}, fmt.Errorf("bad status %d: %s", resp.StatusCode, string(respBytes))
	}

	// Parse flexible response structure
	var raw map[string]any
	if err := json.Unmarshal(respBytes, &raw); err != nil {
		return OmniResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	var out OmniResponse
	// Extract text from choices[0]
	choices, _ := raw["choices"].([]any)
	if len(choices) > 0 {
		choice, _ := choices[0].(map[string]any)
		msg, _ := choice["message"].(map[string]any)
		// message.content may be string or array of parts
		switch content := msg["content"].(type) {
		case string:
			out.Text = content
		case []any:
			for _, part := range content {
				pm, _ := part.(map[string]any)
				ptype, _ := pm["type"].(string)
				if ptype == "output_text" || ptype == "text" {
					if txt, ok := pm["text"].(string); ok {
						out.Text += txt
					}
				}
				if ptype == "output_audio" || ptype == "audio" {
					if aud, ok := pm["audio"].(map[string]any); ok {
						if data, ok := aud["data"].(string); ok {
							b, _ := base64.StdEncoding.DecodeString(data)
							out.AudioMP3 = b
						}
					}
				}
			}
		}
	}

	return out, nil
}

func mimeToSimpleFormat(mime string) string {
	// Map MIME to simple format names used by APIs
	switch mime {
	case "audio/ogg", "audio/opus":
		return "ogg"
	case "audio/wav", "audio/x-wav":
		return "wav"
	case "audio/mpeg", "audio/mp3":
		return "mp3"
	default:
		return "wav"
	}
}

func ToBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
