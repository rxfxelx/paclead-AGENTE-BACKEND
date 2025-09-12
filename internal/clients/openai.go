package clients

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAI struct {
	Key         string
	AssistantID string
	http        *http.Client
}

func NewOpenAI(key, asst string) *OpenAI {
	return &OpenAI{
		Key:         key,
		AssistantID: asst,
		http:        &http.Client{},
	}
}

func (c *OpenAI) newReq(ctx context.Context, method, url string, body any) (*http.Request, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("OpenAI-Beta", "assistants=v2")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *OpenAI) CreateThread(ctx context.Context) (string, error) {
	req, _ := c.newReq(ctx, "POST", "https://api.openai.com/v1/threads", map[string]any{})
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct{ ID string `json:"id"` }
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *OpenAI) CreateMessage(ctx context.Context, threadID string, role string, content any) error {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages", threadID)
	req, _ := c.newReq(ctx, "POST", url, map[string]any{
		"role":    role,
		"content": content,
	})
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return nil
}

func (c *OpenAI) CreateRun(ctx context.Context, threadID string) (string, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", threadID)
	req, _ := c.newReq(ctx, "POST", url, map[string]any{
		"assistant_id": c.AssistantID,
	})
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct{ ID string `json:"id"` }
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// Novo: permite override das 'instructions' (prompt dinâmico por cliente)
func (c *OpenAI) CreateRunWithInstructions(ctx context.Context, threadID, instructions string) (string, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/runs", threadID)
	body := map[string]any{
		"assistant_id": c.AssistantID,
	}
	if strings.TrimSpace(instructions) != "" {
		body["instructions"] = instructions
	}
	req, _ := c.newReq(ctx, "POST", url, body)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct{ ID string `json:"id"` }
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *OpenAI) LastMessageText(ctx context.Context, threadID string) (string, error) {
	url := fmt.Sprintf("https://api.openai.com/v1/threads/%s/messages?order=desc&limit=1", threadID)
	req, _ := c.newReq(ctx, "GET", url, nil)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var out struct {
		Data []struct {
			Content []struct {
				Type string `json:"type"`
				Text struct {
					Value string `json:"value"`
				} `json:"text"`
			} `json:"content"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Data) > 0 && len(out.Data[0].Content) > 0 {
		return out.Data[0].Content[0].Text.Value, nil
	}
	return "", nil
}

func (c *OpenAI) TextToSpeech(ctx context.Context, text string) (string, error) {
	// Returns base64 string (mp3)
	body := map[string]any{
		"model":        "gpt-4o-mini-tts",
		"input":        text,
		"voice":        "ballad",
		"format":       "mp3",
		"instructions": "always speak in an animated and inspiring way, ALWAYS in Brazilian Portuguese",
	}
	req, _ := c.newReq(ctx, "POST", "https://api.openai.com/v1/audio/speech", body)
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	// Some gateways return raw mp3; for consistency we base64-encode if not already JSON.
	if len(data) > 0 && data[0] == '{' {
		var tmp struct{ Data string `json:"data"` }
		if json.Unmarshal(data, &tmp) == nil && tmp.Data != "" {
			return tmp.Data, nil
		}
	}
	return base64.StdEncoding.EncodeToString(data), nil
}
