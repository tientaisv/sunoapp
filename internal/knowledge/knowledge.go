package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ComposerKnowledge struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	EraGenre       string   `json:"era_genre"`
	Description    string   `json:"description"`
	LyricalStyle   string   `json:"lyrical_style"`
	MusicalStyle   string   `json:"musical_style"`
	KeyThemes      []string `json:"key_themes"`
	Metaphors      []string `json:"metaphors"`
	SampleAnalysis string   `json:"sample_analysis"`
	IsCustom       bool     `json:"is_custom,omitempty"`
}

type Manager struct {
	filePath string
	mu       sync.RWMutex
	items    []ComposerKnowledge
}

func NewManager(filePath string) (*Manager, error) {
	m := &Manager{filePath: filePath}
	if err := m.load(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Nếu file chưa tồn tại, thử tạo thư mục cha
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("không thể tạo thư mục dữ liệu: %w", err)
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			m.items = []ComposerKnowledge{}
			return nil
		}
		return err
	}

	var items []ComposerKnowledge
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("lỗi giải mã JSON kiến thức nhạc sĩ: %w", err)
	}

	m.items = items
	return nil
}

func (m *Manager) save() error {
	data, err := json.MarshalIndent(m.items, "", "  ")
	if err != nil {
		return fmt.Errorf("lỗi mã hóa JSON kiến thức nhạc sĩ: %w", err)
	}
	return os.WriteFile(m.filePath, data, 0644)
}

func (m *Manager) List() []ComposerKnowledge {
	m.mu.RLock()
	defer m.mu.RUnlock()
	copied := make([]ComposerKnowledge, len(m.items))
	copy(copied, m.items)
	return copied
}

func (m *Manager) Get(id string) (ComposerKnowledge, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, item := range m.items {
		if item.ID == id {
			return item, true
		}
	}
	return ComposerKnowledge{}, false
}

func (m *Manager) AddOrUpdate(item ComposerKnowledge) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item.ID == "" {
		// Tự tạo ID từ tên nếu chưa có
		cleanName := strings.ToLower(strings.TrimSpace(item.Name))
		cleanName = strings.ReplaceAll(cleanName, " ", "_")
		item.ID = fmt.Sprintf("custom_%s", cleanName)
	}

	found := false
	for i, existing := range m.items {
		if existing.ID == item.ID {
			m.items[i] = item
			found = true
			break
		}
	}

	if !found {
		item.IsCustom = true
		m.items = append(m.items, item)
	}

	return m.save()
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	index := -1
	for i, item := range m.items {
		if item.ID == id {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("không tìm thấy nhạc sĩ với ID: %s", id)
	}

	m.items = append(m.items, m.items[index+1:]...)
	return m.save()
}

// BuildPromptContext tạo đoạn Prompt nhúng kiến thức chuyên sâu của nhạc sĩ được chọn
func (m *Manager) BuildPromptContext(id string) string {
	if id == "" || id == "auto" {
		return `[HỌC TỔNG HỢP CÁC BẬC THẦY ÂM NHẠC VIỆT NAM]:
Học hỏi và thẩm thấu sự tinh tế từ tất cả các nhạc sĩ đại thụ Việt Nam (Trịnh Công Sơn, Phạm Duy, Văn Cao, Việt Anh, Vũ Thành An, Trần Tiến, Ngô Thụy Miên...):
- Ca từ: Đạt chất thơ đỉnh cao, ẩn dụ đẹp, gieo vần chân & vần lưng tự nhiên, nhịp điệu trôi chảy.
- Nhạc lý: Hòa âm phong phú, tiến trình hợp âm sang trọng, không cưỡng âm.`
	}

	ck, found := m.Get(id)
	if !found {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== KIẾN THỨC NGHỆ THUẬT & PHONG CÁCH SÁNG TÁC THEO NHẠC SĨ %s ===\n", strings.ToUpper(ck.Name)))
	sb.WriteString(fmt.Sprintf("- Thời kỳ & Thể loại: %s\n", ck.EraGenre))
	sb.WriteString(fmt.Sprintf("- Triết lý & Định hướng: %s\n", ck.Description))
	sb.WriteString(fmt.Sprintf("- Phong cách Ca từ & Gieo vần: %s\n", ck.LyricalStyle))
	sb.WriteString(fmt.Sprintf("- Tư duy Nhạc lý & Hợp âm: %s\n", ck.MusicalStyle))
	if len(ck.KeyThemes) > 0 {
		sb.WriteString(fmt.Sprintf("- Đề tài chủ đạo: %s\n", strings.Join(ck.KeyThemes, ", ")))
	}
	if len(ck.Metaphors) > 0 {
		sb.WriteString(fmt.Sprintf("- Hình ảnh ẩn dụ đặc trưng: %s\n", strings.Join(ck.Metaphors, ", ")))
	}
	if ck.SampleAnalysis != "" {
		sb.WriteString(fmt.Sprintf("- Phân tích mẫu ca từ / hòa âm: %s\n", ck.SampleAnalysis))
	}
	sb.WriteString(fmt.Sprintf("\nYÊU CẦU BẮT BUỘC: Hãy thẩm thấu toàn bộ kiến thức và hồn thơ ca của Nhạc sĩ %s ở trên để sáng tác ca từ, gieo vần và đặt hợp âm cho bài hát này đạt đỉnh cao nghệ thuật tương ứng.\n", ck.Name))

	return sb.String()
}
