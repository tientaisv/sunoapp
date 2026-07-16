package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	AuthToken          string `json:"auth_token"`
	BrowserToken       string `json:"browser_token"`
	DeviceID           string `json:"device_id"`
	SunoToken          string `json:"suno_token"`
	UserTier           string `json:"user_tier"`
	CreateSessionToken string `json:"create_session_token"`
	SongID             string `json:"song_id"`
	Prompt             string `json:"prompt"`
	Tags               string `json:"tags"`
	Title              string `json:"title"`
	ModelVersion       string `json:"model_version"`
	MakeInstrumental   bool   `json:"make_instrumental"`
	AccountEmail       string `json:"account_email"`
	Cookie             string `json:"cookie"`
}

type SunoFeedRequest struct {
	AuthToken    string   `json:"auth_token"`
	BrowserToken string   `json:"browser_token"`
	DeviceID     string   `json:"device_id"`
	ClipIDs      []string `json:"clip_ids"`
	SongID       string   `json:"song_id"`
	AccountEmail string   `json:"account_email"`
	Cookie       string   `json:"cookie"`
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

func getJWTExpiry(token string) (time.Time, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid token format")
	}
	payload := parts[1]

	data, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		if l := len(payload) % 4; l > 0 {
			payload += strings.Repeat("=", 4-l)
		}
		data, err = base64.URLEncoding.DecodeString(payload)
		if err != nil {
			data, err = base64.StdEncoding.DecodeString(payload)
			if err != nil {
				return time.Time{}, err
			}
		}
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(data, &claims); err != nil {
		return time.Time{}, err
	}

	return time.Unix(claims.Exp, 0), nil
}

func mergeCookies(oldCookieHeader string, newCookies []*http.Cookie) string {
	if len(newCookies) == 0 {
		return oldCookieHeader
	}

	// Parse old cookies into a map
	dummyReq, _ := http.NewRequest("GET", "https://suno.com", nil)
	dummyReq.Header.Set("Cookie", oldCookieHeader)
	parsedOldCookies := dummyReq.Cookies()

	cookieMap := make(map[string]*http.Cookie)
	for _, c := range parsedOldCookies {
		cookieMap[c.Name] = c
	}

	// Merge new cookies
	for _, nc := range newCookies {
		// If MaxAge < 0 or Expires is in the past, delete the cookie
		if nc.MaxAge < 0 || (!nc.Expires.IsZero() && nc.Expires.Before(time.Now())) {
			delete(cookieMap, nc.Name)
		} else {
			cookieMap[nc.Name] = nc
		}
	}

	// Serialize back to Cookie header format
	var parts []string
	for _, c := range cookieMap {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(parts, "; ")
}

func refreshSunoToken(cookie string) (string, string, error) {
	if cookie == "" {
		return "", "", fmt.Errorf("cookie is empty")
	}

	clerkClientURL := "https://auth.suno.com/v1/client?__clerk_api_version=2025-11-10&_clerk_js_version=5.117.0"
	req1, err := http.NewRequest("GET", clerkClientURL, nil)
	if err != nil {
		return "", "", err
	}
	req1.Header.Set("Cookie", cookie)
	req1.Header.Set("Origin", "https://suno.com")
	req1.Header.Set("Referer", "https://suno.com/")
	req1.Header.Set("Accept", "*/*")
	req1.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp1, err := client.Do(req1)
	if err != nil {
		return "", "", fmt.Errorf("error calling Clerk client API: %w", err)
	}
	defer resp1.Body.Close()

	bodyBytes, err := io.ReadAll(resp1.Body)
	if err != nil {
		return "", "", err
	}

	if resp1.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("clerk client API returned status %d: %s", resp1.StatusCode, string(bodyBytes))
	}

	// Merge cookies from resp1
	currentCookie := mergeCookies(cookie, resp1.Cookies())
	if currentCookie != cookie {
		log.Println("[Cookie Rotation] Đã cập nhật Cookie từ Clerk Client API")
	}

	var clientData struct {
		Response struct {
			LastActiveSessionID string `json:"last_active_session_id"`
		} `json:"response"`
		Client struct {
			LastActiveSessionID string `json:"last_active_session_id"`
		} `json:"client"`
	}

	if err := json.Unmarshal(bodyBytes, &clientData); err != nil {
		return "", "", fmt.Errorf("error parsing Clerk client data: %w", err)
	}

	sessionID := clientData.Response.LastActiveSessionID
	if sessionID == "" {
		sessionID = clientData.Client.LastActiveSessionID
	}

	if sessionID == "" {
		return "", "", fmt.Errorf("no active session found in Clerk response: %s", string(bodyBytes))
	}

	tokenURL := fmt.Sprintf("https://auth.suno.com/v1/client/sessions/%s/tokens?__clerk_api_version=2025-11-10&_clerk_js_version=5.117.0", sessionID)
	req2, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", "", err
	}
	req2.Header.Set("Cookie", currentCookie)
	req2.Header.Set("Origin", "https://suno.com")
	req2.Header.Set("Referer", "https://suno.com/")
	req2.Header.Set("Accept", "*/*")
	req2.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp2, err := client.Do(req2)
	if err != nil {
		return "", "", fmt.Errorf("error calling Clerk token API: %w", err)
	}
	defer resp2.Body.Close()

	bodyBytes2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", "", err
	}

	if resp2.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("clerk token API returned status %d: %s", resp2.StatusCode, string(bodyBytes2))
	}

	// Merge cookies from resp2
	finalCookie := mergeCookies(currentCookie, resp2.Cookies())
	if finalCookie != currentCookie {
		log.Println("[Cookie Rotation] Đã cập nhật Cookie từ Clerk Token API")
	}

	var tokenData struct {
		JWT string `json:"jwt"`
	}
	if err := json.Unmarshal(bodyBytes2, &tokenData); err != nil {
		return "", "", fmt.Errorf("error parsing Clerk token data: %w", err)
	}

	if tokenData.JWT == "" {
		return "", "", fmt.Errorf("empty JWT returned from Clerk: %s", string(bodyBytes2))
	}

	return "Bearer " + tokenData.JWT, finalCookie, nil
}

func getOrRefreshSunoToken(authToken string, cookie string) (string, string, bool, error) {
	if cookie == "" {
		return authToken, cookie, false, nil
	}

	expiry, err := getJWTExpiry(authToken)
	if err != nil {
		newToken, newCookie, refreshErr := refreshSunoToken(cookie)
		if refreshErr != nil {
			return authToken, cookie, false, fmt.Errorf("token invalid and refresh failed: %v", refreshErr)
		}
		return newToken, newCookie, true, nil
	}

	if time.Until(expiry) < 45*time.Second {
		newToken, newCookie, refreshErr := refreshSunoToken(cookie)
		if refreshErr != nil {
			return authToken, cookie, false, fmt.Errorf("token expiring soon and refresh failed: %v", refreshErr)
		}
		return newToken, newCookie, true, nil
	}

	return authToken, cookie, false, nil
}

// refreshAndSaveAccount làm mới token cho một account cụ thể và lưu lại kết quả vào storage
func refreshAndSaveAccount(mgr *storage.Manager, acc storage.SunoAccount) error {
	if acc.Cookie == "" {
		return fmt.Errorf("account %s không có cookie", acc.Email)
	}

	newToken, newCookie, err := refreshSunoToken(acc.Cookie)
	if err != nil {
		return fmt.Errorf("lỗi refresh token cho %s: %w", acc.Email, err)
	}

	// Tính thời hạn mới của token
	var newExpiry int64
	if exp, err := getJWTExpiry(newToken); err == nil {
		newExpiry = exp.Unix()
	}

	if err := mgr.UpdateAccountTokens(acc.ID, newToken, newCookie, newExpiry); err != nil {
		return fmt.Errorf("lỗi lưu token mới cho %s: %w", acc.Email, err)
	}

	log.Printf("[AutoRefresh] Đã refresh token thành công cho tài khoản: %s", acc.Email)
	return nil
}

// startTokenAutoRefresh khởi chạy goroutine làm mới token tự động mỗi 30 phút
func startTokenAutoRefresh(mgr *storage.Manager) {
	refreshAll := func() {
		accounts, err := mgr.ListAccounts()
		if err != nil {
			log.Printf("[AutoRefresh] Lỗi đọc danh sách tài khoản: %v", err)
			return
		}

		if len(accounts) == 0 {
			return
		}

		successCount := 0
		for _, acc := range accounts {
			if acc.Cookie == "" {
				continue
			}
			if err := refreshAndSaveAccount(mgr, acc); err != nil {
				log.Printf("[AutoRefresh] %v", err)
			} else {
				successCount++
			}
		}
		if successCount > 0 {
			log.Printf("[AutoRefresh] Đã làm mới token cho %d/%d tài khoản", successCount, len(accounts))
		}
	}

	// Chạy ngay lần đầu khi server khởi động
	go func() {
		time.Sleep(5 * time.Second) // Chờ server ổn định
		refreshAll()
		// Sau đó chạy định kỳ mỗi 30 phút
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			refreshAll()
		}
	}()

	log.Println("[AutoRefresh] Đã khởi động hệ thống làm mới token tự động (mỗi 30 phút)")
}

func mergeClips(existing []storage.SunoClip, newClips []storage.SunoClip) []storage.SunoClip {
	clipMap := make(map[string]int)
	for i, c := range existing {
		clipMap[c.ID] = i
	}

	result := make([]storage.SunoClip, len(existing))
	copy(result, existing)

	for _, nc := range newClips {
		if idx, found := clipMap[nc.ID]; found {
			if nc.DriveURL == "" && result[idx].DriveURL != "" {
				nc.DriveURL = result[idx].DriveURL
			}
			result[idx] = nc
		} else {
			result = append(result, nc)
		}
	}
	return result
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

	// Khởi động hệ thống làm mới token tự động ngầm
	startTokenAutoRefresh(mgr)

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
				"user_tier":                    func() string { if req.UserTier != "" { return req.UserTier }; return "4497580c-f4eb-4f86-9f0e-960eb7c48d7d" }(),
				"create_session_token":         func() string { if req.CreateSessionToken != "" { return req.CreateSessionToken }; return "3d8d709b-97f1-4867-acfb-a014c499b58d" }(),
				"disable_volume_normalization": false,
			},
			"override_fields":   []interface{}{},
			"transaction_uuid":  transactionUUID,
			"token_provider":    1,
			"lyrics_project_id": lyricsProjectID,
		}
		if req.SunoToken != "" {
			sunoPayload["token"] = req.SunoToken
		}

		payloadBytes, err := json.Marshal(sunoPayload)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi mã hóa payload: " + err.Error()})
			return
		}

		tokenRefreshed := false
		
		// 1. Initial check & refresh if expiring soon
		currentAuthToken, newCookie, refreshed, err := getOrRefreshSunoToken(req.AuthToken, req.Cookie)
		if err == nil && refreshed {
			req.AuthToken = currentAuthToken
			req.Cookie = newCookie
			tokenRefreshed = true
		}

		// 2. Call Suno API
		respBytes, err := callSunoAPI("POST", "https://studio-api-prod.suno.com/api/generate/v2-web/", payloadBytes, req.AuthToken, req.BrowserToken, req.DeviceID)
		
		// 3. Retry if 401 and cookie exists
		if err != nil && req.Cookie != "" && strings.Contains(err.Error(), "status 401") {
			log.Println("Suno API trả về 401. Thực hiện ép buộc làm mới token và thử lại...")
			newToken, forcedNewCookie, refreshErr := refreshSunoToken(req.Cookie)
			if refreshErr == nil {
				req.AuthToken = newToken
				req.Cookie = forcedNewCookie
				tokenRefreshed = true
				respBytes, err = callSunoAPI("POST", "https://studio-api-prod.suno.com/api/generate/v2-web/", payloadBytes, req.AuthToken, req.BrowserToken, req.DeviceID)
			} else {
				log.Printf("Ép buộc làm mới token thất bại: %v", refreshErr)
			}
		}

		// 4. Lưu token + cookie mới vào storage nếu có refresh
		if tokenRefreshed && req.AccountEmail != "" {
			go func(email, newAuthToken, newCookieVal string) {
				if acc, err := mgr.FindAccountByEmail(email); err == nil {
					var expiry int64
					if exp, err := getJWTExpiry(newAuthToken); err == nil {
						expiry = exp.Unix()
					}
					if err := mgr.UpdateAccountTokens(acc.ID, newAuthToken, newCookieVal, expiry); err != nil {
						log.Printf("[Generate] Lỗi lưu token mới cho %s: %v", email, err)
					} else {
						log.Printf("[Generate] Đã lưu token mới vào storage cho tài khoản: %s", email)
					}
				}
			}(req.AccountEmail, req.AuthToken, req.Cookie)
		}

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

		if req.SongID != "" {
			song, err := mgr.Get(req.SongID)
			if err == nil {
				var storageClips []storage.SunoClip
				for _, c := range sunoResp.Clips {
					storageClips = append(storageClips, apiClipToStorageClip(c, req.AccountEmail))
				}
				song.SunoClips = mergeClips(song.SunoClips, storageClips)
				mgr.Save(song)
			}
		}

		var clientResp struct {
			Clips        []SunoClipAPIResponse `json:"clips"`
			NewAuthToken string                `json:"new_auth_token,omitempty"`
			NewCookie    string                `json:"new_cookie,omitempty"`
		}
		clientResp.Clips = sunoResp.Clips
		if tokenRefreshed {
			clientResp.NewAuthToken = req.AuthToken
			clientResp.NewCookie = req.Cookie
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(clientResp)
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

		tokenRefreshed := false

		// 1. Initial check & refresh if expiring soon
		currentAuthToken, newCookie, refreshed, err := getOrRefreshSunoToken(req.AuthToken, req.Cookie)
		if err == nil && refreshed {
			req.AuthToken = currentAuthToken
			req.Cookie = newCookie
			tokenRefreshed = true
		}

		// 2. Call Suno API
		respBytes, err := callSunoAPI("POST", "https://studio-api-prod.suno.com/api/feed/v3", payloadBytes, req.AuthToken, req.BrowserToken, req.DeviceID)

		// 3. Retry if 401 and cookie exists
		if err != nil && req.Cookie != "" && strings.Contains(err.Error(), "status 401") {
			log.Println("Suno API feed trả về 401. Thực hiện ép buộc làm mới token và thử lại...")
			newToken, forcedNewCookie, refreshErr := refreshSunoToken(req.Cookie)
			if refreshErr == nil {
				req.AuthToken = newToken
				req.Cookie = forcedNewCookie
				tokenRefreshed = true
				respBytes, err = callSunoAPI("POST", "https://studio-api-prod.suno.com/api/feed/v3", payloadBytes, req.AuthToken, req.BrowserToken, req.DeviceID)
			} else {
				log.Printf("Ép buộc làm mới token thất bại: %v", refreshErr)
			}
		}

		// 4. Lưu token + cookie mới vào storage nếu có refresh
		if tokenRefreshed && req.AccountEmail != "" {
			go func(email, newAuthToken, newCookieVal string) {
				if acc, err := mgr.FindAccountByEmail(email); err == nil {
					var expiry int64
					if exp, err := getJWTExpiry(newAuthToken); err == nil {
						expiry = exp.Unix()
					}
					if err := mgr.UpdateAccountTokens(acc.ID, newAuthToken, newCookieVal, expiry); err != nil {
						log.Printf("[Feed] Lỗi lưu token mới cho %s: %v", email, err)
					} else {
						log.Printf("[Feed] Đã lưu token mới vào storage cho tài khoản: %s", email)
					}
				}
			}(req.AccountEmail, req.AuthToken, req.Cookie)
		}

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
				song.SunoClips = mergeClips(song.SunoClips, storageClips)
				mgr.Save(song)
			}
		}

		var clientResp struct {
			Clips        []SunoClipAPIResponse `json:"clips"`
			NewAuthToken string                `json:"new_auth_token,omitempty"`
			NewCookie    string                `json:"new_cookie,omitempty"`
		}
		clientResp.Clips = apiClips
		if tokenRefreshed {
			clientResp.NewAuthToken = req.AuthToken
			clientResp.NewCookie = req.Cookie
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(clientResp)
	})

	// Route API làm mới token từ Cookie
	mux.HandleFunc("/api/suno/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Cookie string `json:"cookie"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu yêu cầu không hợp lệ: " + err.Error()})
			return
		}

		if req.Cookie == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Thiếu Cookie tài khoản để làm mới"})
			return
		}

		newToken, newCookie, err := refreshSunoToken(req.Cookie)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi làm mới token: " + err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"new_auth_token": newToken,
			"new_cookie":     newCookie,
		})
	})

	// Route API làm mới token thủ công cho tất cả tài khoản
	mux.HandleFunc("/api/suno/accounts/refresh-all", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		accounts, err := mgr.ListAccounts()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi đọc danh sách tài khoản: " + err.Error()})
			return
		}

		type refreshResult struct {
			Email   string `json:"email"`
			Success bool   `json:"success"`
			Error   string `json:"error,omitempty"`
		}

		var results []refreshResult
		for _, acc := range accounts {
			if acc.Cookie == "" {
				results = append(results, refreshResult{Email: acc.Email, Success: false, Error: "Không có cookie"})
				continue
			}
			if err := refreshAndSaveAccount(mgr, acc); err != nil {
				results = append(results, refreshResult{Email: acc.Email, Success: false, Error: err.Error()})
			} else {
				results = append(results, refreshResult{Email: acc.Email, Success: true})
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": results,
			"total":   len(accounts),
		})
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

	// Route API quản lý danh sách tài khoản (GET, POST, DELETE)
	mux.HandleFunc("/api/accounts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			accounts, err := mgr.ListAccounts()
			if err != nil {
				accounts = []storage.SunoAccount{}
			}
			// Khởi tạo mảng rỗng thay vì nil để trả về JSON []
			if accounts == nil {
				accounts = []storage.SunoAccount{}
			}
			json.NewEncoder(w).Encode(accounts)
			return
		}

		if r.Method == http.MethodPost {
			var acc storage.SunoAccount
			if err := json.NewDecoder(r.Body).Decode(&acc); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Dữ liệu không hợp lệ"})
				return
			}
			if acc.ID == "" {
				if acc.Email != "" {
					acc.ID = strings.ReplaceAll(acc.Email, "@", "_at_")
				} else {
					acc.ID = fmt.Sprintf("%d", time.Now().UnixNano())
				}
			}
			if acc.AddedAt == 0 {
				acc.AddedAt = time.Now().UnixMilli()
			}
			if err := mgr.SaveAccount(acc); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi lưu tài khoản"})
				return
			}
			json.NewEncoder(w).Encode(acc)
			return
		}

		if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			if id == "" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Thiếu ID"})
				return
			}
			if err := mgr.DeleteAccount(id); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Lỗi xóa tài khoản"})
				return
			}
			json.NewEncoder(w).Encode(map[string]string{"success": "Xóa thành công"})
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
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

	// Route API publish file lên Google Drive qua rclone (SSE stream progress)
	mux.HandleFunc("/api/suno/publish-drive", func(w http.ResponseWriter, r *http.Request) {
		// Thiết lập headers cho kết nối SSE
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Không hỗ trợ Streaming!", http.StatusInternalServerError)
			return
		}

		sendEvent := func(status string, progress int, msg string, extra map[string]interface{}) {
			payload := map[string]interface{}{
				"status":   status,
				"progress": progress,
				"message":  msg,
			}
			for k, v := range extra {
				payload[k] = v
			}
			data, _ := json.Marshal(payload)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}

		songID := r.URL.Query().Get("song_id")
		clipID := r.URL.Query().Get("clip_id")
		audioURL := r.URL.Query().Get("audio_url")

		if songID == "" || clipID == "" || audioURL == "" {
			sendEvent("error", 0, "Thiếu tham số bắt buộc: song_id, clip_id, hoặc audio_url", nil)
			return
		}

		// Lấy thông tin bài hát để đặt tên file
		song, err := mgr.Get(songID)
		if err != nil {
			sendEvent("error", 0, "Không tìm thấy bài hát trong hệ thống: "+err.Error(), nil)
			return
		}

		clipTitle := "suno-song"
		clipIdx := -1
		for i, c := range song.SunoClips {
			if c.ID == clipID {
				clipIdx = i
				if c.Title != "" {
					clipTitle = c.Title
				}
				break
			}
		}
		if clipTitle == "suno-song" && song.Title != "" {
			clipTitle = song.Title
		}

		// Làm sạch tên file
		fileName := clipTitle
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

		sendEvent("downloading", 0, "Đang tải file từ Suno CDN...", nil)

		// 1. Tải file từ CDN về file tạm local
		tempFile, err := os.CreateTemp("", "suno-audio-*.mp3")
		if err != nil {
			sendEvent("error", 0, "Không thể tạo file tạm local: "+err.Error(), nil)
			return
		}
		tempFilePath := tempFile.Name()
		defer os.Remove(tempFilePath) // Dọn dẹp phòng hờ nếu move lỗi
		defer tempFile.Close()

		resp, err := http.Get(audioURL)
		if err != nil {
			sendEvent("error", 0, "Không thể kết nối tải nhạc từ Suno CDN: "+err.Error(), nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			sendEvent("error", 0, fmt.Sprintf("Suno CDN trả về mã trạng thái lỗi %d", resp.StatusCode), nil)
			return
		}

		contentLength := resp.ContentLength
		buffer := make([]byte, 32*1024) // 32KB chunk
		var totalDownloaded int64

		for {
			n, readErr := resp.Body.Read(buffer)
			if n > 0 {
				_, writeErr := tempFile.Write(buffer[:n])
				if writeErr != nil {
					sendEvent("error", 0, "Không thể ghi dữ liệu file tạm: "+writeErr.Error(), nil)
					return
				}
				totalDownloaded += int64(n)
				if contentLength > 0 {
					pct := int((totalDownloaded * 30) / contentLength) // Chiếm 0% - 30% tổng thanh tiến trình
					sendEvent("downloading", pct, fmt.Sprintf("Đang tải file nhạc: %d%%", int((totalDownloaded*100)/contentLength)), nil)
				} else {
					sendEvent("downloading", 15, "Đang tải file nhạc...", nil)
				}
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				sendEvent("error", 0, "Lỗi khi đang đọc dữ liệu tải nhạc: "+readErr.Error(), nil)
				return
			}
		}
		tempFile.Close()

		sendEvent("uploading", 30, "Đang chuẩn bị chuyển dữ liệu lên Drive...", nil)

		// 2. Chạy rclone moveto để chuyển file lên Drive
		cmd := exec.Command("rclone", "moveto", tempFilePath, "vtw:"+fileName,
			"--drive-root-folder-id", "1tp9JwMMe1_BDJDlebs0OUHBR-MCgrWyv",
			"--config", "data/rclone.conf",
			"--use-json-log",
			"--stats", "200ms",
		)

		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			sendEvent("error", 30, "Không thể mở luồng ghi nhận tiến trình rclone: "+err.Error(), nil)
			return
		}

		if err := cmd.Start(); err != nil {
			sendEvent("error", 30, "Không thể khởi chạy tiến trình rclone: "+err.Error(), nil)
			return
		}

		// Đọc log stderr của rclone để parse phần trăm tiến trình
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			var logData struct {
				Stats struct {
					Bytes        int64 `json:"bytes"`
					TotalBytes   int64 `json:"totalBytes"`
					Transferring []struct {
						Percentage int `json:"percentage"`
					} `json:"transferring"`
				} `json:"stats"`
			}
			
			if err := json.Unmarshal([]byte(line), &logData); err == nil {
				if len(logData.Stats.Transferring) > 0 {
					pctUp := logData.Stats.Transferring[0].Percentage
					pctOverall := 30 + int((pctUp * 65) / 100) // Chiếm 30% - 95% tổng thanh tiến trình
					sendEvent("uploading", pctOverall, fmt.Sprintf("Đang tải lên Google Drive: %d%%", pctUp), nil)
				} else if logData.Stats.TotalBytes > 0 {
					// Fallback tính toán
					pctUp := int((logData.Stats.Bytes * 100) / logData.Stats.TotalBytes)
					pctOverall := 30 + int((pctUp * 65) / 100)
					sendEvent("uploading", pctOverall, fmt.Sprintf("Đang tải lên Google Drive: %d%%", pctUp), nil)
				}
			}
		}

		if err := cmd.Wait(); err != nil {
			sendEvent("error", 0, "Lỗi upload lên Google Drive: "+err.Error(), nil)
			return
		}

		sendEvent("finalizing", 95, "Đang thiết lập liên kết chia sẻ công khai...", nil)

		// 3. Chạy rclone link để sinh link public
		linkCmd := exec.Command("rclone", "link", "vtw:"+fileName,
			"--drive-root-folder-id", "1tp9JwMMe1_BDJDlebs0OUHBR-MCgrWyv",
			"--config", "data/rclone.conf",
		)
		var linkStdout, linkStderr bytes.Buffer
		linkCmd.Stdout = &linkStdout
		linkCmd.Stderr = &linkStderr

		if err := linkCmd.Run(); err != nil {
			sendEvent("error", 95, "Lỗi sinh liên kết chia sẻ từ Google Drive: "+linkStderr.String(), nil)
			return
		}

		driveURL := strings.TrimSpace(linkStdout.String())
		if driveURL == "" {
			sendEvent("error", 95, "Không nhận được liên kết phản hồi từ rclone", nil)
			return
		}

		// 4. Lưu liên kết vào song metadata
		if clipIdx >= 0 {
			song.SunoClips[clipIdx].DriveURL = driveURL
			if err := mgr.Save(song); err != nil {
				log.Printf("[Drive] Lỗi cập nhật bài hát sau khi upload: %v", err)
			}
		}

		sendEvent("success", 100, "Đã tải lên và chia sẻ thành công!", map[string]interface{}{
			"drive_url": driveURL,
		})
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
