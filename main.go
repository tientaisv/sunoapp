package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"taonhac/internal/composer"
	"taonhac/internal/prompt"
	"taonhac/internal/storage"
)

//go:embed public/*
var publicFS embed.FS

var (
	appPassword  = ""
	sessionToken = ""
)

func initAuth() {
	appPassword = os.Getenv("APP_PASSWORD")
	if appPassword != "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err == nil {
			sessionToken = fmt.Sprintf("%x", b)
		} else {
			sessionToken = "default-fallback-session-token"
		}
		log.Println("[Auth] Cơ chế bảo mật bằng mật khẩu đã được kích hoạt.")
	} else {
		log.Println("[Auth] Ứng dụng chạy ở chế độ công cộng (không yêu cầu mật khẩu).")
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if appPassword == "" {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path
		// Các tài nguyên public được phép truy cập tự do
		if path == "/login.html" || path == "/style.css" || path == "/api/login" || path == "/favicon.ico" || path == "/api/auth-status" {
			next.ServeHTTP(w, r)
			return
		}

		// Kiểm tra cookie session_token
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie.Value == sessionToken {
			next.ServeHTTP(w, r)
			return
		}

		// Nếu chưa đăng nhập:
		// 1. Đối với API: Trả về 401 Unauthorized
		if strings.HasPrefix(path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Chưa đăng nhập hoặc phiên đã hết hạn"})
			return
		}

		// 2. Đối với giao diện: Chuyển hướng về login.html
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})
}

// loadEnv đọc file .env thủ công nếu có để tránh phụ thuộc thư viện ngoài
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("[Env] Không tìm thấy file .env, sử dụng biến môi trường hệ thống.")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Loại bỏ dấu nháy đơn/kép xung quanh giá trị nếu có
			val = strings.Trim(val, `"'`)
			os.Setenv(key, val)
		}
	}
	log.Println("[Env] Đã tải cấu hình từ file .env thành công.")
}

type ComposeRequest struct {
	Topic          string   `json:"topic"`
	CatholicDegree string   `json:"catholicDegree"`
	Genre          string   `json:"genre"`
	Verses         int      `json:"verses"`
	RepeatVerse    bool     `json:"repeatVerse"`
	ChorusPitch    string   `json:"chorusPitch"`
	Voice          string   `json:"voice"`
	Tempo          string   `json:"tempo"`
	Mood           string   `json:"mood"`
	Instruments    []string `json:"instruments"`
	Key            string   `json:"key"`
	VocalHarmony   string   `json:"vocalHarmony"`
	VocalTechnique string   `json:"vocalTechnique"`
	VocalPlacement string   `json:"vocalPlacement"`
	ExistingLyrics string   `json:"existingLyrics"`
	RewritePrompt  string   `json:"rewritePrompt"`
}

type SunoGenerateRequest struct {
	AuthToken        string `json:"auth_token"`
	BrowserToken     string `json:"browser_token"`
	DeviceID         string `json:"device_id"`
	SunoToken        string `json:"suno_token"`
	SongID           string `json:"song_id"`
	Prompt           string `json:"prompt"`
	Tags             string `json:"tags"`
	Title            string `json:"title"`
	ModelVersion     string `json:"model_version"`
	MakeInstrumental bool   `json:"make_instrumental"`
	AccountEmail     string `json:"account_email"`
}

type SunoFeedRequest struct {
	AuthToken    string   `json:"auth_token"`
	BrowserToken string   `json:"browser_token"`
	DeviceID     string   `json:"device_id"`
	ClipIDs      []string `json:"clip_ids"`
	SongID       string   `json:"song_id"`
	AccountEmail string   `json:"account_email"`
}

type SunoClipAPIResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	AudioURL  string `json:"audio_url"`
	VideoURL  string `json:"video_url"`
	ImageURL  string `json:"image_url"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Metadata  struct {
		Prompt string `json:"prompt"`
	} `json:"metadata"`
}

type SunoGenerateAPIResponse struct {
	Clips []SunoClipAPIResponse `json:"clips"`
}

func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "1b482b56-52b2-4fff-8192-56dffb406b0d"
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func apiClipToStorageClip(apiClip SunoClipAPIResponse, email string) storage.SunoClip {
	createdAt := time.Now()
	if apiClip.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, apiClip.CreatedAt); err == nil {
			createdAt = t
		} else if t, err := time.Parse("2006-01-02T15:04:05.000Z", apiClip.CreatedAt); err == nil {
			createdAt = t
		}
	}
	return storage.SunoClip{
		ID:           apiClip.ID,
		Title:        apiClip.Title,
		AudioURL:     apiClip.AudioURL,
		VideoURL:     apiClip.VideoURL,
		ImageURL:     apiClip.ImageURL,
		Status:       apiClip.Status,
		Prompt:       apiClip.Metadata.Prompt,
		CreatedAt:    createdAt,
		AccountEmail: email,
	}
}

func callSunoAPI(method, endpoint string, body []byte, authToken, browserToken, deviceID string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en")
	req.Header.Set("content-type", "application/json")
	
	auth := authToken
	if auth != "" {
		if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			auth = "Bearer " + auth
		}
		req.Header.Set("authorization", auth)
	}
	
	if browserToken != "" {
		req.Header.Set("browser-token", browserToken)
	}
	if deviceID != "" {
		req.Header.Set("device-id", deviceID)
	}
	
	req.Header.Set("origin", "https://suno.com")
	req.Header.Set("referer", "https://suno.com/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/149.0.0.0 Safari/537.36")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return respBytes, fmt.Errorf("Suno API error status %d", resp.StatusCode)
	}
	
	return respBytes, nil
}

func main() {
	loadEnv()
	initAuth()

	// Đọc cấu hình cổng và keys
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Cổng chạy mặc định bên trong container
	}

	keysCsv := os.Getenv("GEMINI_KEYS")
	if keysCsv == "" {
		log.Println("[CẢNH BÁO] Chưa cấu hình GEMINI_KEYS trong biến môi trường!")
	}

	model := os.Getenv("GEMINI_MODEL")

	// Khởi tạo Composer client
	comp, err := composer.NewComposer(keysCsv, model)
	if err != nil {
		log.Printf("[LỖI KHỞI TẠO COMPOSER] %v", err)
	}

	// Khởi tạo Storage Manager để lưu trữ các bài hát
	mgr, err := storage.NewManager("data")
	if err != nil {
		log.Fatalf("[LỖI KHỞI TẠO STORAGE] %v", err)
	}

	// Đăng ký routes
	mux := http.NewServeMux()

	// Route API tạo nhạc trên Suno AI (Tự động)
	mux.HandleFunc("/api/suno/generate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		var req SunoGenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu yêu cầu không hợp lệ: " + err.Error()})
			return
		}

		if req.AuthToken == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Thiếu Authorization token của Suno"})
			return
		}

		modelVersion := req.ModelVersion
		if modelVersion == "" {
			modelVersion = "chirp-fenix" // default to v5.5
		}

		// Build payload for Suno
		transactionUUID := generateUUID()
		lyricsProjectID := generateUUID()

		sunoPayload := map[string]interface{}{
			"token":             req.SunoToken,
			"generation_type":   "TEXT",
			"title":             req.Title,
			"tags":              req.Tags,
			"negative_tags":     "",
			"mv":                modelVersion,
			"prompt":            req.Prompt,
			"make_instrumental": req.MakeInstrumental,
			"metadata": map[string]interface{}{
				"web_client_pathname":          "/create",
				"is_max_mode":                  false,
				"create_mode":                  "custom",
				"user_tier":                    "4497580c-f4eb-4f86-9f0e-960eb7c48d7d",
				"create_session_token":         "3d8d709b-97f1-4867-acfb-a014c499b58d",
				"disable_volume_normalization": false,
			},
			"override_fields":   []interface{}{},
			"transaction_uuid":  transactionUUID,
			"token_provider":    1,
			"lyrics_project_id": lyricsProjectID,
		}

		payloadBytes, err := json.Marshal(sunoPayload)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi mã hóa payload: " + err.Error()})
			return
		}

		respBytes, err := callSunoAPI("POST", "https://studio-api-prod.suno.com/api/generate/v2-web/", payloadBytes, req.AuthToken, req.BrowserToken, req.DeviceID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := err.Error()
			if len(respBytes) > 0 {
				errMsg = string(respBytes)
			}
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi gọi Suno API: " + errMsg})
			return
		}

		var sunoResp SunoGenerateAPIResponse
		if err := json.Unmarshal(respBytes, &sunoResp); err != nil {
			// Thử unmarshal dạng mảng trực tiếp đề phòng
			var clips []SunoClipAPIResponse
			if err2 := json.Unmarshal(respBytes, &clips); err2 == nil {
				sunoResp.Clips = clips
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi giải mã phản hồi của Suno: " + err.Error()})
				return
			}
		}

		if len(sunoResp.Clips) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Suno không trả về bất kỳ clip nhạc nào. Vui lòng kiểm tra lại credits tài khoản."})
			return
		}

		// Cập nhật bài hát gốc nếu có song_id
		if req.SongID != "" {
			song, err := mgr.Get(req.SongID)
			if err == nil {
				var storageClips []storage.SunoClip
				for _, c := range sunoResp.Clips {
					storageClips = append(storageClips, apiClipToStorageClip(c, req.AccountEmail))
				}
				song.SunoClips = storageClips
				mgr.Save(song)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	})

	// Route API polling trạng thái bài hát từ Suno
	mux.HandleFunc("/api/suno/feed", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		var req SunoFeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu yêu cầu không hợp lệ: " + err.Error()})
			return
		}

		if req.AuthToken == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Thiếu Authorization token của Suno"})
			return
		}

		if len(req.ClipIDs) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Thiếu Clip IDs cần kiểm tra"})
			return
		}

		// Build payload
		sunoPayload := map[string]interface{}{
			"filters": map[string]interface{}{
				"ids": map[string]interface{}{
					"presence": "True",
					"clipIds":  req.ClipIDs,
				},
			},
			"limit": len(req.ClipIDs),
		}

		payloadBytes, err := json.Marshal(sunoPayload)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi mã hóa payload: " + err.Error()})
			return
		}

		respBytes, err := callSunoAPI("POST", "https://studio-api-prod.suno.com/api/feed/v3", payloadBytes, req.AuthToken, req.BrowserToken, req.DeviceID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			errMsg := err.Error()
			if len(respBytes) > 0 {
				errMsg = string(respBytes)
			}
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi gọi Suno API: " + errMsg})
			return
		}

		var apiClips []SunoClipAPIResponse
		if err := json.Unmarshal(respBytes, &apiClips); err != nil {
			var wrap struct {
				Clips []SunoClipAPIResponse `json:"clips"`
			}
			if err2 := json.Unmarshal(respBytes, &wrap); err2 == nil {
				apiClips = wrap.Clips
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi giải mã phản hồi của Suno: " + err.Error()})
				return
			}
		}

		// Cập nhật bài hát gốc nếu có song_id và trạng thái cập nhật
		if req.SongID != "" && len(apiClips) > 0 {
			song, err := mgr.Get(req.SongID)
			if err == nil {
				var storageClips []storage.SunoClip
				for _, c := range apiClips {
					storageClips = append(storageClips, apiClipToStorageClip(c, req.AccountEmail))
				}
				song.SunoClips = storageClips
				mgr.Save(song)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	})

	// Route API sáng tác nhạc
	mux.HandleFunc("/api/compose", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		if comp == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Ứng dụng chưa được cấu hình API Keys. Vui lòng kiểm tra file .env.",
			})
			return
		}

		var req ComposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Dữ liệu yêu cầu không hợp lệ: " + err.Error(),
			})
			return
		}

		// Giới hạn số lượng verse hợp lý (từ 1 đến 5)
		if req.Verses < 1 {
			req.Verses = 2
		} else if req.Verses > 5 {
			req.Verses = 5
		}

		log.Printf("[API] Nhận yêu cầu sáng tác bài hát với chủ đề: '%s', thể loại: '%s'", req.Topic, req.Genre)

		// Tạo prompt
		systemPrompt := prompt.GetSystemPrompt()
		userPrompt := prompt.BuildUserPrompt(
			req.Topic,
			req.CatholicDegree,
			req.Genre,
			req.Verses,
			req.RepeatVerse,
			req.ChorusPitch,
			req.Voice,
			req.Tempo,
			req.Mood,
			req.Instruments,
			req.Key,
			req.VocalHarmony,
			req.VocalTechnique,
			req.VocalPlacement,
			req.ExistingLyrics,
			req.RewritePrompt,
		)

		// Gọi AI sáng tác nhạc
		respJSON, err := comp.Compose(systemPrompt, userPrompt)
		if err != nil {
			log.Printf("[API LỖI] %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Sáng tác thất bại: " + err.Error(),
			})
			return
		}

		// Parse kết quả từ AI để đóng gói lưu trữ
		var aiResult struct {
			Title       string `json:"title"`
			Style       string `json:"style"`
			Key         string `json:"key"`
			Lyrics      string `json:"lyrics"`
			AbcNotation string `json:"abc_notation"`
		}
		if err := json.Unmarshal([]byte(respJSON), &aiResult); err != nil {
			log.Printf("[API CẢNH BÁO] Không thể giải mã kết quả AI: %v. Trả về text gốc.", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(respJSON))
			return
		}

		// Lưu tự động bài hát vào lịch sử
		songId := fmt.Sprintf("%d", time.Now().UnixNano()/1000000) // Epoch millisecond
		
		// Lấy tiêu đề từ AI hoặc tự tạo từ chủ đề làm phương án dự phòng
		title := strings.TrimSpace(aiResult.Title)
		if title == "" {
			words := strings.Fields(req.Topic)
			if len(words) > 5 {
				title = strings.Join(words[:5], " ") + "..."
			} else if len(words) > 0 {
				title = req.Topic
			} else {
				title = "Bài hát không tên"
			}
		}

		savedSong := storage.SavedSong{
			ID:             songId,
			Title:          title,
			CreatedAt:      time.Now(),
			Style:          aiResult.Style,
			Key:            aiResult.Key,
			Lyrics:         aiResult.Lyrics,
			Topic:          req.Topic,
			CatholicDegree: req.CatholicDegree,
			Genre:          req.Genre,
			Verses:         req.Verses,
			RepeatVerse:    req.RepeatVerse,
			ChorusPitch:    req.ChorusPitch,
			Voice:          req.Voice,
			Tempo:          req.Tempo,
			Mood:           req.Mood,
			Instruments:    req.Instruments,
			AbcNotation:    aiResult.AbcNotation,
			VocalHarmony:   req.VocalHarmony,
			VocalTechnique: req.VocalTechnique,
			VocalPlacement: req.VocalPlacement,
		}

		if err := mgr.Save(savedSong); err != nil {
			log.Printf("[API CẢNH BÁO] Lỗi lưu bài hát mới: %v", err)
		}

		// Trả kết quả kèm ID bài hát về cho Frontend
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(savedSong)
	})

	// Route API quản lý danh sách bài hát (GET, POST, DELETE)
	mux.HandleFunc("/api/songs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 1. Lấy danh sách hoặc chi tiết bài hát
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			if id != "" {
				song, err := mgr.Get(id)
				if err != nil {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"error": "Không tìm thấy bài hát"})
					return
				}
				json.NewEncoder(w).Encode(song)
				return
			}

			songs, err := mgr.List()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi đọc dữ liệu: " + err.Error()})
				return
			}
			json.NewEncoder(w).Encode(songs)
			return
		}

		// 2. Cập nhật bài hát (Chỉnh sửa trực tiếp lời)
		if r.Method == http.MethodPost {
			var inputSong storage.SavedSong
			if err := json.NewDecoder(r.Body).Decode(&inputSong); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu không hợp lệ"})
				return
			}

			if inputSong.ID == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "ID bài hát không được để trống"})
				return
			}

			// Lấy bản ghi gốc để cập nhật
			song, err := mgr.Get(inputSong.ID)
			if err != nil {
				// Nếu không thấy thì tạo mới hoàn toàn
				song = inputSong
				if song.CreatedAt.IsZero() {
					song.CreatedAt = time.Now()
				}
			} else {
				// Cập nhật các trường chỉnh sửa
				song.Title = inputSong.Title
				song.Lyrics = inputSong.Lyrics
				song.Style = inputSong.Style
				song.Key = inputSong.Key
				song.AbcNotation = inputSong.AbcNotation
				song.SunoClips = inputSong.SunoClips
			}

			if err := mgr.Save(song); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi lưu dữ liệu: " + err.Error()})
				return
			}

			json.NewEncoder(w).Encode(song)
			return
		}

		// 3. Xóa bài hát
		if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "ID bài hát không được để trống"})
				return
			}

			if err := mgr.Delete(id); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Không thể xóa bài hát: " + err.Error()})
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"success": "Đã xóa bài hát khỏi lịch sử"})
			return
		}

		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
	})

	// Route API đăng nhập hệ thống
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu yêu cầu không hợp lệ"})
			return
		}

		if req.Password == appPassword {
			// Lưu session bằng cookie bảo mật
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    sessionToken,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   60 * 60 * 24 * 7, // Hạn dùng 7 ngày
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"success": "Đăng nhập thành công"})
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Mật khẩu không chính xác"})
		}
	})

	// Route API đăng xuất
	mux.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1, // Xóa cookie lập tức
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"success": "Đăng xuất thành công"})
	})

	// Route API kiểm tra trạng thái bảo mật của server
	mux.HandleFunc("/api/auth-status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"auth_enabled": appPassword != "",
		})
	})

	// Route API tải nhạc Suno trực tiếp về máy và tự động đổi tên file
	mux.HandleFunc("/api/suno/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		audioURL := r.URL.Query().Get("url")
		fileName := r.URL.Query().Get("name")

		if audioURL == "" {
			http.Error(w, "Thiếu tham số url", http.StatusBadRequest)
			return
		}

		if fileName == "" {
			fileName = "suno-song"
		}

		// Làm sạch tên file (xóa ký tự nguy hiểm)
		fileName = strings.ReplaceAll(fileName, "\n", "")
		fileName = strings.ReplaceAll(fileName, "\r", "")
		fileName = strings.ReplaceAll(fileName, "\"", "")
		fileName = strings.ReplaceAll(fileName, "'", "")
		fileName = strings.ReplaceAll(fileName, "/", "-")
		fileName = strings.ReplaceAll(fileName, "\\", "-")
		fileName = strings.TrimSpace(fileName)

		if !strings.HasSuffix(strings.ToLower(fileName), ".mp3") {
			fileName += ".mp3"
		}

		// Gửi yêu cầu lấy file audio từ Suno CDN
		resp, err := http.Get(audioURL)
		if err != nil {
			http.Error(w, "Không thể tải file nhạc: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Suno CDN trả về mã trạng thái %d", resp.StatusCode), resp.StatusCode)
			return
		}

		// Thiết lập header trả về dạng file đính kèm với tên bài hát
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
		w.Header().Set("Content-Type", "audio/mpeg")
		if resp.ContentLength > 0 {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
		}

		// Stream toàn bộ file audio về client trực tiếp
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Println("[Download] Lỗi khi đang truyền luồng tải file:", err)
		}
	})

	// Phục vụ frontend static files bằng Go embed
	publicSubFS, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("Lỗi cấu hình Static File Server: %v", err)
	}
	fileServer := http.FileServer(http.FS(publicSubFS))
	mux.Handle("/", fileServer)

	// Khởi chạy server
	addr := ":" + port
	log.Printf("=== SUNO MUSIC COMPOSER ===")
	log.Printf("Server đang chạy tại http://localhost:%s (hoặc cổng host 31)", port)
	log.Printf("Đang lắng nghe kết nối...")

	// Wrap ServeMux trong authMiddleware để bảo mật toàn bộ ứng dụng
	if err := http.ListenAndServe(addr, authMiddleware(mux)); err != nil {
		log.Fatalf("Không thể khởi chạy server: %v", err)
	}
}
