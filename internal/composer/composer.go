package composer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Composer struct {
	keys  []string
	index int
	model string
	mu    sync.Mutex
}

type GeminiRequest struct {
	Contents          []Content         `json:"contents"`
	SystemInstruction *SystemInstruction `json:"systemInstruction,omitempty"`
	GenerationConfig  GenerationConfig  `json:"generationConfig"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type GenerationConfig struct {
	ResponseMimeType string  `json:"responseMimeType"`
	Temperature      float64 `json:"temperature"`
}

type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

// NewComposer khởi tạo đối tượng Composer với danh sách keys xoay vòng
func NewComposer(keysCsv, model string) (*Composer, error) {
	var keys []string
	for _, k := range strings.Split(keysCsv, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, errors.New("không tìm thấy GEMINI_KEYS trong cấu hình")
	}

	if model == "" {
		model = "gemini-2.0-flash" // model mặc định tốc độ cao, hỗ trợ tiếng Việt cực tốt
	}

	return &Composer{
		keys:  keys,
		model: model,
	}, nil
}

// GetNextKey lấy key tiếp theo và di chuyển con trỏ xoay vòng
func (c *Composer) GetNextKey() (string, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx := c.index
	key := c.keys[idx]
	c.index = (c.index + 1) % len(c.keys)
	return key, idx
}

// Compose gọi Gemini API để sáng tác nhạc với cơ chế xoay vòng và tự động thử lại nếu lỗi limit
func (c *Composer) Compose(systemPrompt, userPrompt string) (string, error) {
	reqPayload := GeminiRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: userPrompt},
				},
			},
		},
		SystemInstruction: &SystemInstruction{
			Parts: []Part{
				{Text: systemPrompt},
			},
		},
		GenerationConfig: GenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.75,
		},
	}

	reqBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("lỗi đóng gói request payload: %w", err)
	}

	// Thử tối đa qua tất cả các keys hiện có nếu gặp lỗi Rate Limit (429) hoặc lỗi mạng
	maxRetries := len(c.keys)
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		apiKey, keyIdx := c.GetNextKey()
		log.Printf("[Composer] Đang sử dụng API Key thứ %d (vòng lặp thử lại: %d/%d)...", keyIdx+1, i+1, maxRetries)

		url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, apiKey)
		
		client := &http.Client{Timeout: 60 * time.Second}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBytes))
		if err != nil {
			lastErr = fmt.Errorf("lỗi khởi tạo request với key %d: %w", keyIdx+1, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[Composer] Lỗi kết nối với key %d: %v. Đang thử key tiếp theo...", keyIdx+1, err)
			lastErr = fmt.Errorf("lỗi kết nối: %w", err)
			continue
		}
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[Composer] Lỗi đọc phản hồi với key %d: %v", keyIdx+1, err)
			lastErr = fmt.Errorf("lỗi đọc phản hồi: %w", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("[Composer] Key thứ %d trả về status code lỗi: %d. Chi tiết: %s", keyIdx+1, resp.StatusCode, string(respBytes))
			lastErr = fmt.Errorf("API trả về lỗi %d: %s", resp.StatusCode, string(respBytes))
			// Nếu bị rate limit (429) hoặc lỗi server (5xx), xoay vòng thử tiếp
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				continue
			}
			// Nếu lỗi sai Key hoặc Bad Request khác không sửa đổi được bằng cách xoay key, vẫn xoay tiếp để phòng trường hợp key sau đúng
			continue
		}

		// Parse kết quả thành công
		var geminiResp GeminiResponse
		if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
			log.Printf("[Composer] Lỗi parse JSON phản hồi với key %d: %v", keyIdx+1, err)
			lastErr = fmt.Errorf("lỗi parse JSON phản hồi: %w", err)
			continue
		}

		if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
			log.Printf("[Composer] Key %d trả về dữ liệu trống", keyIdx+1)
			lastErr = errors.New("dữ liệu phản hồi trống từ Gemini API")
			continue
		}

		// Lấy text trả về
		responseText := geminiResp.Candidates[0].Content.Parts[0].Text
		return responseText, nil
	}

	return "", fmt.Errorf("tất cả API keys đều thất bại. Lỗi cuối cùng: %w", lastErr)
}
