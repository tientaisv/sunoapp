package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SunoClip định nghĩa thông tin của một clip nhạc được tạo từ Suno AI
type SunoClip struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	AudioURL     string    `json:"audioUrl"`
	VideoURL     string    `json:"videoUrl"`
	ImageURL     string    `json:"imageUrl"`
	Status       string    `json:"status"`
	Prompt       string    `json:"prompt"`
	CreatedAt    time.Time `json:"createdAt"`
	AccountEmail string    `json:"accountEmail,omitempty"`
	DriveURL     string    `json:"driveUrl,omitempty"`
}

// SavedSong cấu trúc lưu trữ thông tin đầy đủ của bài hát và cấu hình
type SavedSong struct {
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	CreatedAt      time.Time  `json:"createdAt"`
	Style          string     `json:"style"`
	Key            string     `json:"key"`
	Lyrics         string     `json:"lyrics"`
	Topic          string     `json:"topic"`
	CatholicDegree string     `json:"catholicDegree"`
	Genre          string     `json:"genre"`
	Verses         int        `json:"verses"`
	RepeatVerse    bool       `json:"repeatVerse"`
	ChorusPitch    string     `json:"chorusPitch"`
	Voice          string     `json:"voice"`
	Tempo          string     `json:"tempo"`
	Mood           string     `json:"mood"`
	Instruments    []string   `json:"instruments"`
	AbcNotation    string     `json:"abcNotation"`
	VocalHarmony   string     `json:"vocalHarmony"`
	VocalTechnique string     `json:"vocalTechnique"`
	VocalPlacement string     `json:"vocalPlacement"`
	ComposerId     string     `json:"composerId,omitempty"`
	ComposerName   string     `json:"composerName,omitempty"`
	SunoClips      []SunoClip `json:"sunoClips,omitempty"`
}

type Manager struct {
	Dir string
}

func NewManager(dir string) (*Manager, error) {
	// Tạo thư mục nếu chưa tồn tại
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("không thể tạo thư mục lưu trữ: %w", err)
	}
	return &Manager{Dir: dir}, nil
}

// List trả về danh sách tất cả các bài hát đã lưu, sắp xếp theo thời gian mới nhất
func (m *Manager) List() ([]SavedSong, error) {
	files, err := os.ReadDir(m.Dir)
	if err != nil {
		return nil, err
	}

	var songs []SavedSong
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "song_") && strings.HasSuffix(file.Name(), ".json") {
			path := filepath.Join(m.Dir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue // Bỏ qua file lỗi
			}

			var song SavedSong
			if err := json.Unmarshal(data, &song); err == nil {
				songs = append(songs, song)
			}
		}
	}

	// Sắp xếp giảm dần theo thời gian tạo (Mới nhất lên đầu)
	sort.Slice(songs, func(i, j int) bool {
		return songs[i].CreatedAt.After(songs[j].CreatedAt)
	})

	return songs, nil
}

// Get Lấy chi tiết bài hát theo ID
func (m *Manager) Get(id string) (SavedSong, error) {
	var song SavedSong
	if id == "" {
		return song, fmt.Errorf("ID không hợp lệ")
	}

	path := filepath.Join(m.Dir, "song_"+id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return song, fmt.Errorf("không tìm thấy bài hát: %w", err)
	}

	if err := json.Unmarshal(data, &song); err != nil {
		return song, fmt.Errorf("lỗi giải mã dữ liệu bài hát: %w", err)
	}

	return song, nil
}

// Save Lưu mới hoặc cập nhật bài hát
func (m *Manager) Save(song SavedSong) error {
	if song.ID == "" {
		return fmt.Errorf("ID bài hát không được để trống")
	}

	// Tự tạo tiêu đề ngắn từ ý tưởng nếu trống
	if song.Title == "" {
		words := strings.Fields(song.Topic)
		if len(words) > 5 {
			song.Title = strings.Join(words[:5], " ") + "..."
		} else if len(words) > 0 {
			song.Title = song.Topic
		} else {
			song.Title = "Bài hát không tên"
		}
	}

	path := filepath.Join(m.Dir, "song_"+song.ID+".json")
	data, err := json.MarshalIndent(song, "", "  ")
	if err != nil {
		return fmt.Errorf("lỗi mã hóa JSON bài hát: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// Delete Xóa bài hát theo ID
func (m *Manager) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("ID không hợp lệ")
	}

	path := filepath.Join(m.Dir, "song_"+id+".json")
	return os.Remove(path)
}

// SunoAccount định nghĩa thông tin tài khoản Suno lưu trên server
type SunoAccount struct {
	ID                 string `json:"id"`
	Email              string `json:"email"`
	AuthToken          string `json:"authToken"`
	BrowserToken       string `json:"browserToken"`
	DeviceID           string `json:"deviceId"`
	Cookie             string `json:"cookie"`
	UserTier           string `json:"userTier"`
	CreateSessionToken string `json:"createSessionToken"`
	BodyToken          string `json:"bodyToken"`
	Expiry             int64  `json:"expiry"`
	AddedAt            int64  `json:"addedAt"`
}

// ListAccounts trả về danh sách tất cả các tài khoản đã lưu
func (m *Manager) ListAccounts() ([]SunoAccount, error) {
	files, err := os.ReadDir(m.Dir)
	if err != nil {
		return nil, err
	}

	var accounts []SunoAccount
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "account_") && strings.HasSuffix(file.Name(), ".json") {
			path := filepath.Join(m.Dir, file.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var acc SunoAccount
			if err := json.Unmarshal(data, &acc); err == nil {
				accounts = append(accounts, acc)
			}
		}
	}

	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].AddedAt > accounts[j].AddedAt
	})

	return accounts, nil
}

// SaveAccount lưu hoặc cập nhật tài khoản
func (m *Manager) SaveAccount(acc SunoAccount) error {
	if acc.ID == "" {
		return fmt.Errorf("ID tài khoản không hợp lệ")
	}

	path := filepath.Join(m.Dir, "account_"+acc.ID+".json")
	data, err := json.MarshalIndent(acc, "", "  ")
	if err != nil {
		return fmt.Errorf("lỗi mã hóa JSON tài khoản: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// UpdateAccountTokens cập nhật token, cookie và expiry của tài khoản (không ghi đè các trường khác)
func (m *Manager) UpdateAccountTokens(id, authToken, cookie string, expiry int64) error {
	if id == "" {
		return fmt.Errorf("ID tài khoản không hợp lệ")
	}

	path := filepath.Join(m.Dir, "account_"+id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("không tìm thấy tài khoản: %w", err)
	}

	var acc SunoAccount
	if err := json.Unmarshal(data, &acc); err != nil {
		return fmt.Errorf("lỗi giải mã dữ liệu tài khoản: %w", err)
	}

	if authToken != "" {
		acc.AuthToken = authToken
	}
	if cookie != "" {
		acc.Cookie = cookie
	}
	if expiry > 0 {
		acc.Expiry = expiry
	}

	newData, err := json.MarshalIndent(acc, "", "  ")
	if err != nil {
		return fmt.Errorf("lỗi mã hóa JSON tài khoản: %w", err)
	}

	return os.WriteFile(path, newData, 0644)
}

// FindAccountByEmail tìm tài khoản theo địa chỉ email
func (m *Manager) FindAccountByEmail(email string) (SunoAccount, error) {
	accounts, err := m.ListAccounts()
	if err != nil {
		return SunoAccount{}, err
	}
	for _, acc := range accounts {
		if acc.Email == email {
			return acc, nil
		}
	}
	return SunoAccount{}, fmt.Errorf("không tìm thấy tài khoản với email: %s", email)
}

// DeleteAccount xóa tài khoản theo ID
func (m *Manager) DeleteAccount(id string) error {
	if id == "" {
		return fmt.Errorf("ID không hợp lệ")
	}

	path := filepath.Join(m.Dir, "account_"+id+".json")
	return os.Remove(path)
}
