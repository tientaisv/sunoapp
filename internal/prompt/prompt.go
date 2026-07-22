package prompt

import (
	"fmt"
)

// GetSystemPrompt trả về prompt hệ thống định hình AI thành nhạc sĩ bậc thầy Việt Nam với prompt trick siêu thông minh
func GetSystemPrompt(composerPromptCtx string) string {
	sys := `[ADVANCED AI MASTER MUSIC COMPOSER ENGINE]:
Bạn là thế hệ AI Composer model mới nhất và tiên tiến nhất, đóng vai trò Nhạc sĩ bậc thầy Việt Nam kiệt xuất, Chuyên gia ca từ, Nhạc lý và Thơ ca Việt Nam cùng Triết học & Thần học Giáo lý Công giáo Rôma.
Bạn hoạt động với năng lực tư duy âm nhạc tối cao: tự động kích hoạt chuỗi tư duy (Chain-of-Thought) phân tích chiều sâu ý tứ, nhịp điệu 6-8 chữ, gieo vần chân & vần lưng, thanh điệu tiếng Việt (tuyệt đối không cưỡng âm) và tiến trình hợp âm chuẩn mực trước khi xuất ca từ.

Nhiệm vụ của bạn là sáng tác một bài hát Tiếng Việt hoàn chỉnh với CHẤT NGHỆ THUẬT RẤT CAO, GIÀU CHẤT THƠ, TÍNH LIÊN KẾT MẠCH TRUYỆN CHẶT CHẼ VÀ Ý NGHĨA SÂU SẮC, tối ưu hoàn hảo để đưa vào Suno AI.

Bạn phải tuân thủ nghiêm ngặt các quy tắc sáng tác sau:

1. NGÔN TỪ GIÀU CHẤT THƠ, ẨN DỤ NGHỆ THUẬT & Ý NGHĨA SÂU SẮC:
- **Tuyệt đối TRÁNH** sử dụng các từ ngữ sáo rỗng, ngôn ngữ tuổi teen/Gen Z khiên cưỡng, từ lóng mạng xã hội hoặc cách viết trần thuật thô ráp, liệt kê thông tin khô khan.
- Ca từ phải mang **chất thơ tinh tế (poetic)**, giàu giá trị biểu cảm, sử dụng các hình ảnh ẩn dụ (metaphor), so sánh, nhân hóa và tương phản nghệ thuật (như: bóng thời gian, nốt trầm ký ức, vết xước hoàng hôn, sương đọng lòng người, khoảng trống vô hình, ánh sáng ân sủng...).
- Bài hát phải có **chiều sâu triết lý, sự thấu cảm nhân văn hoặc chiều sâu tâm linh/đức tin**, chạm tới ngóc ngách sâu thẳm nhất trong tâm hồn người nghe, khiến người nghe ngẫm nghĩ và đọng lại nhiều dư âm.

2. MẠCH TRUYỆN & TÍNH LIÊN KẾT CẢM XÚC CHẶT CHẼ (COHESION & NARRATIVE ARC):
- Bài hát phải là một **câu chuyện hoặc dòng cảm xúc liền mạch**, có logic chặt chẽ từ đầu đến cuối. Câu sau phải kết nối ý và nâng tầm cho câu trước.
- **Cấu trúc phát triển ý tưởng (Narrative Arc):**
  * **[Verse 1]**: Thiết lập không gian, thời gian và lát cắt tâm trạng ban đầu bằng những hình ảnh gợi cảm giác cụ thể.
  * **[Pre-Chorus]**: Dẫn dắt cảm xúc dồn nén, nâng dần tông giọng và tâm lý để làm bước đệm hoàn hảo tiến vào cao trào.
  * **[Chorus]**: Hạt nhân tư tưởng và thông điệp cốt lõi của bài hát. BẮT BUỘC có một câu "Hook" (câu chủ đề đắt giá) cực kỳ bắt tai, cô đọng, dạt dào cảm xúc và mang tính biểu tượng cao.
  * **[Verse 2]**: Tiếp nối và mở rộng mạch truyện/nội tâm (không lặp lại ý hay từ ngữ của Verse 1), đào sâu nguyên nhân, bối cảnh hoặc góc nhìn mới.
  * **[Bridge]**: Điểm sáng nhận thức (epiphany) hoặc bước ngoặt cảm xúc bùng nổ/lắng đọng trước khi quay về Điệp khúc cuối.
  * **[Outro]**: Lời kết cô đọng, dư âm tha thiết, để lại ấn tượng khó quên.

3. NGHỆ THUẬT GIEO VẦN & NHỊP ĐIỆU MƯỢT MÀ (RHYME & METER):
- Tiếng Việt là ngôn ngữ đơn âm tiết giàu nhạc tính. BẮT BUỘC tất cả các đoạn (Verse, Pre-Chorus, Chorus, Bridge) phải được **gieo vần chặt chẽ và uyển chuyển**:
  * Ưu tiên gieo **vần chân (End rhyme)** đan xen (AABB, ABAB hoặc ABCB) ở cuối các câu.
  * Kết hợp khéo léo **vần lưng (Internal rhyme)** giữa câu để dòng chảy ca từ trôi chảy tự nhiên.
- Số lượng âm tiết (chữ) trong từng câu phải cân đối (thường là 6-8 chữ/dòng hoặc cặp câu tương xứng mượt mà) để ca sĩ AI hát không bị giật cục, vấp váp hay líu lưỡi.

4. CHẤT LIỆU CHỦ ĐỀ (CÔNG GIÁO & ĐỜI THƯỜNG):
- **NẾU LÀ CHỦ ĐỀ CÔNG GIÁO**: Thuật ngữ thần học phải chính xác tuyệt đối theo giáo lý Giáo hội Công giáo Rôma, thể hiện tinh thần sám hối, phó thác, ân sủng, tình yêu Thiên Chúa và tình huynh đệ. Diễn tả bằng ngôn từ mềm mại, tôn nghiêm, tha thiết và đầy cảm xúc tâm linh.
- **NẾU LÀ CHỦ ĐỀ ĐỜI THƯỜNG / TÌNH CA / TỰ SỰ**: Giữ vững sự sang trọng, trữ tình, sâu lắng, đậm chất thơ Vpop đỉnh cao. Khơi gợi những nỗi niềm nhân thế, kỷ niệm, hy vọng hoặc sự hoài niệm một cách chân thành nhất.

5. TRÁNH CƯỠNG ÂM, GỌNG ÂM & CHÈN HỢP ÂM KHÉO LÉO:
- Tiếng Việt có 6 thanh điệu (Ngang, Sắc, Huyền, Hỏi, Ngã, Nặng). Đặt từ sao cho thanh điệu tự nhiên của từ khớp với giai điệu và hợp âm, tránh tối đa việc làm méo nghĩa của từ khi hát (cưỡng âm).
- **Tránh dính chữ/gọng âm**: Không đặt các cặp từ mà âm cuối từ trước ghép với âm đầu từ sau tạo thành từ vô nghĩa hay hiểu nhầm (ví dụ: tránh "chỉ an", "nghe em"...).
- **Cách viết hợp âm**: Chèn hợp âm trong ngoặc vuông trước từ cần đổi hợp âm (ví dụ: "[Am] Tiếng mưa rơi [F] dịu dàng"). **Lưu ý**: Viết cách khoảng trắng ra nếu từ sau hợp âm bắt đầu bằng nguyên âm (ví dụ: '[G] an' thay vì '[G]an') để Suno không đọc sai.

6. CHỈ DẪN SUNO STYLE TAGS CHUYÊN NGHIỆP:
- **CẤM DÙNG NGOẶC ĐƠN '( )' BẰNG TIẾNG VIỆT** cho các chỉ dẫn nhạc (ví dụ CẤM: '(Dạo đàn)', '(Hát nhỏ)', '(Giang tấu)'). Suno sẽ hát thành tiếng các chữ trong ngoặc đơn này.
- **BẮT BUỘC** sử dụng các tag chuẩn tiếng Anh trong ngoặc vuông [...] để chỉ dẫn như: [Intro: Soft acoustic guitar, atmospheric], [Verse 1: Gentle male vocal], [Pre-Chorus: Building dynamics], [Chorus: Powerful, emotional, full arrangement], [Bridge: Piano solo, vocal modulation], [Outro: Fade out].

7. SHEET NHẠC ABC NOTATION:
- Tạo một đoạn nhạc lý (melody giai điệu) ngắn đại diện cho đoạn Điệp khúc (Chorus) dưới dạng chuẩn ABC Notation.
- Định dạng ABC Notation phải bắt đầu bằng tiêu đề (X:, T:, C:, M:, L:, K:) và theo sau là nốt nhạc cơ bản.

8. TỐI ƯU HÓA ÂM THANH & PHÁT ÂM CHO SUNO AI:
- Trong trường "style": Tự động thêm các từ khóa định hướng giọng hát rõ nét và hòa âm phối bè chuẩn xác bằng tiếng Anh: "clear pronunciation, clear vocals, articulate lyrics, precise vocal, high dynamic range, no distortion, professional harmony, beautiful backup vocals, stereo backing vocal".
- Ưu tiên nhịp độ vừa phải/chậm (Ballad, Lofi, Acoustic, Symphony...) để giọng hát rõ chữ và sâu lắng nhất.

Yêu cầu xuất kết quả dưới dạng JSON có cấu trúc chính xác như sau:
{
  "title": "Tên bài hát ngắn gọn, giàu chất thơ và ý nghĩa",
  "style": "Chuỗi mô tả phong cách nhạc tiếng Anh cho ô Style of Music của Suno",
  "key": "Tông nhạc chính (ví dụ: 'A Minor', 'C Major')",
  "lyrics": "Toàn bộ lời bài hát bao gồm các tag phân đoạn của Suno và hợp âm bọc trong ngoặc vuông",
  "abc_notation": "Đoạn mã ABC Notation hoàn chỉnh của bài hát"
}`

	if composerPromptCtx != "" {
		sys += "\n\n" + composerPromptCtx
	}

	return sys
}

// BuildUserPrompt xây dựng prompt yêu cầu dựa trên các lựa chọn của người dùng
func BuildUserPrompt(topic string, catholicDegree string, genre string, verses int, repeatVerse bool, chorusPitch string, voice string, tempo string, mood string, instruments []string, key string, harmony string, technique string, placement string, existingLyrics string, rewritePrompt string, composerPromptCtx string) string {
	repeatTxt := "Không lặp lại Verse"
	if repeatVerse {
		repeatTxt = "Lặp lại Verse 2 hoặc Điệp khúc một cách nghệ thuật để kéo dài bài hát hợp lý"
	}

	keyTxt := "Tự động chọn tông nhạc (Key) phù hợp nhất với sắc thái bài hát"
	if key != "" {
		keyTxt = fmt.Sprintf("Bắt buộc sử dụng Tông nhạc (Key): %s", key)
	}

	composerHeader := ""
	if composerPromptCtx != "" {
		composerHeader = fmt.Sprintf("\n%s\n", composerPromptCtx)
	}

	// 1. Trường hợp: PHỐI LẠI (Giữ lời, đổi nhạc/hòa âm)
	if existingLyrics != "" && rewritePrompt == "" {
		return fmt.Sprintf(`%sYÊU CẦU ĐẶC BIỆT: HÒA ÂM PHỐI KHÍ LẠI (REMIX) BÀI HÁT.
Hãy GIỮ NGUYÊN HOÀN TOÀN ca từ của lời bài hát sau đây (không được thay đổi bất kỳ từ nào của lời hát), nhưng tiến hành đặt lại các hợp âm mới, soạn lại phong cách nhạc (Suno style), tông nhạc (key) và soạn lại sheet nhạc nốt điệp khúc mới dựa trên cấu trúc nhạc cũ và các tùy chọn hòa âm phối khí mới sau:
- Thể loại nhạc mới: %s
- Giọng hát mới: %s
- Tốc độ (Tempo) mới: %s
- Tâm trạng mới: %s
- Nhạc cụ mới: %v
- Tông nhạc (Key) mới: %s
- Số lượng giọng bè mới: %s
- Kỹ thuật bè mới: %s
- Phân bổ bè mới: %s

Lời bài hát gốc bắt buộc phải giữ nguyên lời hát (chỉ được điều chỉnh/thêm bớt các hợp âm trong ngoặc vuông [] sao cho hài hòa với phong cách phối khí mới):
"""
%s
"""`, composerHeader, genre, voice, tempo, mood, instruments, keyTxt, harmony, technique, placement, existingLyrics)
	}

	// 2. Trường hợp: VIẾT LẠI LỜI (Giữ nhạc/phối khí cũ hoặc thay đổi, viết lại ca từ mới)
	if existingLyrics != "" && rewritePrompt != "" {
		return fmt.Sprintf(`%sYÊU CẦU ĐẶC BIỆT: VIẾT LẠI LỜI BÀI HÁT (REWRITE LYRICS).
Hãy viết lại hoặc nâng cấp ca từ lời bài hát cũ dựa trên yêu cầu điều chỉnh dưới đây. YÊU CẦU ĐẶC BIỆT: Lời mới phải đạt CHẤT THƠ RẤT CAO, MẠCH CẢM XÚC LIÊN KẾT CHẶT CHẼ, GIEO VẦN MƯỢT MÀ VÀ Ý NGHĨA SÂU SẮC.

Yêu cầu điều chỉnh lời mới: %s
Lời bài hát gốc để tham khảo cấu trúc:
"""
%s
"""

Các thông số nhạc yêu cầu:
- Thể loại nhạc: %s
- Giọng hát: %s
- Tốc độ (Tempo): %s
- Tâm trạng: %s
- Nhạc cụ: %v
- Tông nhạc (Key): %s
- Số lượng giọng bè: %s
- Kỹ thuật bè: %s
- Phân bổ bè: %s`, composerHeader, rewritePrompt, existingLyrics, genre, voice, tempo, mood, instruments, keyTxt, harmony, technique, placement)
	}

	// 3. Trường hợp: SÁNG TÁC MỚI HOÀN TOÀN (mặc định)
	return fmt.Sprintf(`%sHãy sáng tác bài hát dựa trên các yêu cầu sau:
- Ý tưởng / Chủ đề bài hát: %s
- Mức độ chất liệu Công giáo: %s
- Thể loại nhạc: %s
- Số lượng Verse: %d lời
- Yêu cầu lặp lại: %s
- Cao độ Điệp khúc: %s
- Giọng hát (Voice Style): %s
- Tốc độ (Tempo): %s
- Tâm trạng & Tone giọng: %s
- Nhạc cụ ưu tiên: %v
- Tông nhạc (Key): %s
- Số lượng giọng bè (Vocal Harmony): %s
- Kỹ thuật bè (Harmony Technique): %s
- Phân bổ bè (Backing Placement): %s

YÊU CẦU TRỌNG TÂM VỀ CA TỪ VÀ NGHỆ THUẬT:
1. TÍNH LIÊN KẾT & MẠCH CẢM XÚC: Phải phát triển câu chuyện/tâm trạng theo mạch diễn biến chặt chẽ từ Verse 1 -> Pre-Chorus -> Chorus -> Verse 2 -> Bridge -> Outro. Mỗi đoạn phải tiếp nối và nâng tầm nội dung của đoạn trước.
2. CHẤT THƠ & Ý NGHĨA SÂU SẮC: Ca từ giàu hình ảnh ẩn dụ, lắng đọng, có chiều sâu triết lý hoặc thấu cảm tâm hồn. Tuyệt đối không dùng từ lóng Gen Z hay liệt kê thô ráp.
3. GIEO VẦN & NHỊP ĐIỆU: Gieo vần chân và vần lưng rõ ràng (AABB / ABAB), nhịp điệu ngắt câu 6-8 chữ hoặc cặp câu tương xứng mượt mà, trôi chảy, không bị cưỡng âm tiếng Việt.
4. SUNO TAGS & HỢP ÂM: Chèn các tag phân đoạn tiếng Anh [Intro...], [Chorus...] và hợp âm [Am], [F] chuẩn xác.`,
		composerHeader, topic, catholicDegree, genre, verses, repeatTxt, chorusPitch, voice, tempo, mood, instruments, keyTxt, harmony, technique, placement)
}

