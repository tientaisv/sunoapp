// Khởi chạy Lucide Icons
document.addEventListener("DOMContentLoaded", async () => {
    lucide.createIcons();
    initUIHandlers();
    await fetchSunoAccounts();
    initSunoIntegration();
    checkAuthStatus();
});

// Lưu trữ dữ liệu bài hát đã tạo
let currentSongData = {
    key: "",
    style: "",
    lyrics: ""
};

// Khởi tạo các sự kiện giao diện
function initUIHandlers() {
    // 1. Xử lý Single Select cho các nhóm Button
    setupGroupSelection("catholicDegree");
    setupGroupSelection("verses");
    setupGroupSelection("chorusPitch");
    setupGroupSelection("tempo");
    setupGroupSelection("vocalHarmony");
    setupGroupSelection("vocalTechnique");
    setupGroupSelection("vocalPlacement");

    // 2. Xử lý Single Select cho Chips (Genre, Voice, Mood)
    setupChipSingleSelection("genre");
    setupChipSingleSelection("voiceStyle");
    setupChipSingleSelection("mood");

    // 3. Xử lý Multi Select cho Nhạc cụ
    setupChipMultiSelection("instruments");

    // 4. Xử lý chọn màu sắc không khí âm nhạc
    const colorBtns = document.querySelectorAll(".color-btn");
    const customColorPicker = document.getElementById("customColorPicker");
    const customColorName = document.getElementById("customColorName");

    function updateThemeColor(hexColor) {
        document.documentElement.style.setProperty("--color-primary", hexColor);
        const glow1 = document.querySelector(".glow-bg-1");
        if (glow1) {
            glow1.style.background = `radial-gradient(circle, ${hexColor} 0%, transparent 70%)`;
        }
    }

    colorBtns.forEach(btn => {
        btn.addEventListener("click", () => {
            colorBtns.forEach(b => b.classList.remove("active"));
            btn.classList.add("active");
            
            // Reset ô tự nhập
            customColorName.value = "";
            
            const themeColor = btn.style.getPropertyValue("--color-theme");
            updateThemeColor(themeColor);
            customColorPicker.value = themeColor;
        });
    });

    // Lắng nghe thay đổi màu sắc từ bảng chọn màu
    customColorPicker.addEventListener("input", (e) => {
        colorBtns.forEach(b => b.classList.remove("active"));
        updateThemeColor(e.target.value);
    });

    // Lắng nghe sự kiện người dùng gõ tên màu sắc cảm xúc
    customColorName.addEventListener("input", () => {
        colorBtns.forEach(b => b.classList.remove("active"));
        updateThemeColor(customColorPicker.value);
    });

    // 5. Xử lý Click Gợi ý nhanh
    const suggestChips = document.querySelectorAll(".suggest-chip");
    const topicTextarea = document.getElementById("topic");
    suggestChips.forEach(chip => {
        chip.addEventListener("click", () => {
            topicTextarea.value = chip.textContent;
            
            // Tự động chuyển mức độ Công giáo thành "Đậm chất Công giáo" khi chọn các gợi ý thánh ca
            const catholicGroup = document.getElementById("catholicDegree");
            const firstBtn = catholicGroup.querySelector(".btn-item");
            if (firstBtn) {
                catholicGroup.querySelectorAll(".btn-item").forEach(b => b.classList.remove("active"));
                firstBtn.classList.add("active");
            }
            
            // Tự động chọn thể loại "Thánh ca Đương đại" hoặc "Thánh ca Truyền thống"
            const genreGroup = document.getElementById("genre");
            const tcChip = genreGroup.querySelector('[data-val="Thánh ca Đương đại"]') || genreGroup.querySelector('[data-val="Thánh ca Truyền thống"]');
            if (tcChip) {
                genreGroup.querySelectorAll(".chip-item").forEach(c => c.classList.remove("active"));
                tcChip.classList.add("active");
            }

            topicTextarea.focus();
            
            // Hiệu ứng nhấp nháy viền khi chọn gợi ý
            topicTextarea.style.borderColor = "var(--color-primary)";
            setTimeout(() => {
                topicTextarea.style.borderColor = "var(--border-color)";
            }, 1000);
        });
    });

    // 6. Xử lý bật/tắt Hợp âm
    const chordToggle = document.getElementById("chordToggle");
    chordToggle.addEventListener("change", () => {
        if (currentSongData.lyrics) {
            renderLyrics(currentSongData.lyrics, chordToggle.checked);
        }
    });

    // 7. Xử lý nút Tạo bài hát
    const composeBtn = document.getElementById("composeBtn");
    composeBtn.addEventListener("click", handleCompose);

    // 8. Xử lý nút Copy
    const copyBtns = document.querySelectorAll(".btn-copy");
    copyBtns.forEach(btn => {
        btn.addEventListener("click", () => {
            const targetId = btn.getAttribute("data-target");
            let textToCopy = "";

            if (targetId === "styleText") {
                textToCopy = document.getElementById("styleText").textContent;
            } else if (targetId === "lyricsText") {
                textToCopy = getVisibleLyrics();
            }

            copyToClipboard(textToCopy, btn);
        });
    });

    // 9. Xử lý chuyển đổi Tab (Tác phẩm & Lịch sử)
    const tabBtns = document.querySelectorAll(".tab-btn");
    tabBtns.forEach(btn => {
        btn.addEventListener("click", () => {
            tabBtns.forEach(b => b.classList.remove("active"));
            btn.classList.add("active");

            const tabTarget = btn.getAttribute("data-tab");
            document.querySelectorAll(".tab-content").forEach(tc => {
                tc.style.display = "none";
                tc.classList.remove("active");
            });

            const activeContent = document.getElementById("tab-" + tabTarget);
            if (activeContent) {
                activeContent.style.display = "block";
                activeContent.classList.add("active");
            }

            // Nếu mở tab Lịch sử, tải lại danh sách
            if (tabTarget === "history-library") {
                loadHistory();
            }

            // Đồng bộ sang mobile bottom navigation
            const appContent = document.querySelector(".app-content");
            if (appContent) {
                if (tabTarget === "current-song") {
                    appContent.classList.remove("mobile-show-config", "mobile-show-history");
                    appContent.classList.add("mobile-show-result");
                    updateMobileTabActive("result");
                } else if (tabTarget === "history-library") {
                    appContent.classList.remove("mobile-show-config", "mobile-show-result");
                    appContent.classList.add("mobile-show-history");
                    updateMobileTabActive("history");
                }
            }
        });
    });

    // 10. Xử lý các nút Chỉnh sửa bài hát hiện tại
    const btnEditSong = document.getElementById("btnEditSong");
    const btnSaveSong = document.getElementById("btnSaveSong");
    const btnCancelEdit = document.getElementById("btnCancelEdit");
    const btnDownloadPDF = document.getElementById("btnDownloadPDF");
    const btnRemixSong = document.getElementById("btnRemixSong");
    const btnRewriteSong = document.getElementById("btnRewriteSong");

    const resTitleDisplay = document.getElementById("resTitleDisplay");
    const titleInput = document.getElementById("titleInput");
    const styleText = document.getElementById("styleText");
    const styleInput = document.getElementById("styleInput");
    const lyricsTextWrapper = document.getElementById("lyricsTextWrapper");
    const lyricsEditWrapper = document.getElementById("lyricsEditWrapper");
    const lyricsEditInput = document.getElementById("lyricsEditInput");

    const abcEditCard = document.getElementById("abcEditCard");
    const abcEditInput = document.getElementById("abcEditInput");

    btnEditSong.addEventListener("click", () => {
        // Tắt hiển thị tĩnh, bật chế độ sửa
        resTitleDisplay.style.display = "none";
        titleInput.style.display = "block";
        titleInput.value = currentSongData.title;

        styleText.style.display = "none";
        styleInput.style.display = "block";
        styleInput.value = currentSongData.style;

        lyricsTextWrapper.style.display = "none";
        lyricsEditWrapper.style.display = "block";
        lyricsEditInput.value = currentSongData.lyrics; // Lấy lời gốc kèm hợp âm để sửa

        abcEditCard.style.display = "block";
        abcEditInput.value = currentSongData.abcNotation || "";

        btnEditSong.style.display = "none";
        btnSaveSong.style.display = "inline-flex";
        btnCancelEdit.style.display = "inline-flex";
        btnDownloadPDF.style.display = "none";
        if (btnRemixSong) btnRemixSong.style.display = "none";
        if (btnRewriteSong) btnRewriteSong.style.display = "none";
    });

    btnCancelEdit.addEventListener("click", () => {
        // Quay lại hiển thị tĩnh
        resTitleDisplay.style.display = "block";
        titleInput.style.display = "none";

        styleText.style.display = "block";
        styleInput.style.display = "none";

        lyricsTextWrapper.style.display = "block";
        lyricsEditWrapper.style.display = "none";

        abcEditCard.style.display = "none";

        btnEditSong.style.display = "inline-flex";
        btnSaveSong.style.display = "none";
        btnCancelEdit.style.display = "none";
        btnDownloadPDF.style.display = "inline-flex";
        if (btnRemixSong) btnRemixSong.style.display = "inline-flex";
        if (btnRewriteSong) btnRewriteSong.style.display = "inline-flex";
    });

    // Sự kiện Remix & Rewrite
    if (btnRemixSong) {
        btnRemixSong.addEventListener("click", () => {
            handleCompose("remix");
        });
    }

    if (btnRewriteSong) {
        btnRewriteSong.addEventListener("click", () => {
            const promptText = prompt("Nhập ý tưởng/chủ đề hoặc yêu cầu điều chỉnh để viết lại lời mới cho bài hát này (ví dụ: 'viết lại lời sang chủ đề hy vọng', 'viết lại lời về ơn cha mẹ...'):");
            if (promptText && promptText.trim() !== "") {
                handleCompose("rewrite", promptText.trim());
            }
        });
    }

    btnSaveSong.addEventListener("click", async () => {
        const editedTitle = titleInput.value.trim();
        const editedStyle = styleInput.value.trim();
        const editedLyrics = lyricsEditInput.value.trim();
        const editedAbc = abcEditInput.value.trim();

        if (!editedTitle) {
            alert("Tiêu đề bài hát không được để trống!");
            return;
        }
        if (!editedLyrics) {
            alert("Lời bài hát không được để trống!");
            return;
        }

        btnSaveSong.disabled = true;
        btnSaveSong.innerHTML = `<i data-lucide="loader-2" class="animate-spin"></i> Đang lưu...`;
        lucide.createIcons();

        try {
            const response = await fetch("/api/songs", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    id: currentSongData.id,
                    title: editedTitle,
                    style: editedStyle,
                    lyrics: editedLyrics,
                    abcNotation: editedAbc
                })
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || "Không thể cập nhật bài hát.");
            }

            // Cập nhật lại state
            currentSongData.title = data.title;
            currentSongData.style = data.style;
            currentSongData.lyrics = data.lyrics;
            currentSongData.abcNotation = data.abcNotation || "";

            // Render lại giao diện
            resTitleDisplay.textContent = currentSongData.title;
            styleText.textContent = currentSongData.style;
            const showChords = document.getElementById("chordToggle").checked;
            renderLyrics(currentSongData.lyrics, showChords);
            renderSheetMusic(currentSongData.abcNotation);

            btnCancelEdit.click();
            updateHistoryCount();
        } catch (error) {
            alert("Lỗi khi lưu: " + error.message);
        } finally {
            btnSaveSong.disabled = false;
            btnSaveSong.innerHTML = `<i data-lucide="save"></i> Lưu chỉnh sửa`;
            lucide.createIcons();
        }
    });

    // Sự kiện tải PDF
    btnDownloadPDF.addEventListener("click", downloadPDF);

    // 11. Các sự kiện phụ
    document.getElementById("btnRefreshHistory").addEventListener("click", loadHistory);

    // Tải trước số lượng bài hát đã sáng tác
    updateHistoryCount();

    // Khởi tạo điều hướng trên thiết bị di động
    initMobileNavigation();
}

// Thiết lập chọn nhiều nút trong group (Cho phép chọn nhiều)
function setupGroupSelection(groupId) {
    const container = document.getElementById(groupId);
    if (!container) return;
    
    const items = container.querySelectorAll(".btn-item");
    items.forEach(item => {
        item.addEventListener("click", () => {
            item.classList.toggle("active");
        });
    });
}

// Thiết lập chọn nhiều chips trong tập hợp chips (Cho phép chọn nhiều)
function setupChipSingleSelection(containerId) {
    const container = document.getElementById(containerId);
    if (!container) return;

    const chips = container.querySelectorAll(".chip-item");
    chips.forEach(chip => {
        chip.addEventListener("click", () => {
            chip.classList.toggle("active");
        });
    });
}

// Thiết lập chọn nhiều chips (Multi Select)
function setupChipMultiSelection(containerId) {
    const container = document.getElementById(containerId);
    if (!container) return;

    const chips = container.querySelectorAll(".chip-item");
    chips.forEach(chip => {
        chip.addEventListener("click", () => {
            chip.classList.toggle("active");
        });
    });
}

// Lấy danh sách các giá trị đang hoạt động của Button Group dưới dạng chuỗi nối nhau bằng dấu phẩy
function getGroupValue(groupId) {
    const container = document.getElementById(groupId);
    if (!container) return "";
    const actives = container.querySelectorAll(".btn-item.active");
    const values = [];
    actives.forEach(a => values.push(a.getAttribute("data-val")));
    return values.join(", ");
}

// Lấy danh sách các giá trị đang hoạt động của Chips dưới dạng chuỗi nối nhau bằng dấu phẩy
function getChipValue(containerId) {
    const container = document.getElementById(containerId);
    if (!container) return "";
    const actives = container.querySelectorAll(".chip-item.active");
    const values = [];
    actives.forEach(a => values.push(a.getAttribute("data-val")));
    return values.join(", ");
}

// Lấy danh sách các giá trị đang hoạt động của Chips (Multi Select)
function getMultiChipValues(containerId) {
    const container = document.getElementById(containerId);
    if (!container) return [];
    const actives = container.querySelectorAll(".chip-item.active");
    const values = [];
    actives.forEach(a => values.push(a.getAttribute("data-val")));
    return values;
}

// Hàm gửi request sáng tác bài hát lên Backend Go
async function handleCompose(mode, extraParam) {
    const topic = document.getElementById("topic").value.trim();
    if (!topic) {
        alert("Vui lòng nhập ý tưởng hoặc chủ đề bài hát trước khi tạo!");
        document.getElementById("topic").focus();
        return;
    }

    // Thu thập dữ liệu form
    const catholicDegree = getGroupValue("catholicDegree");
    const versesVal = getGroupValue("verses");
    let verses = 2;
    if (versesVal) {
        const nums = versesVal.split(",").map(v => parseInt(v.trim(), 10)).filter(v => !isNaN(v));
        if (nums.length > 0) {
            verses = Math.max(...nums);
        }
    }
    const repeatVerse = document.getElementById("repeatVerse").checked;
    const chorusPitch = getGroupValue("chorusPitch");
    const musicKey = document.getElementById("musicKey").value;
    const tempo = getGroupValue("tempo");
    
    const genre = getChipValue("genre");
    const voice = getChipValue("voiceStyle");
    const mood = getChipValue("mood");
    const instruments = getMultiChipValues("instruments");

    const vocalHarmony = getGroupValue("vocalHarmony");
    const vocalTechnique = getGroupValue("vocalTechnique");
    const vocalPlacement = getGroupValue("vocalPlacement");

    // Lấy màu sắc cảm xúc hiện tại
    const activeColorBtn = document.querySelector(".color-btn.active");
    let emotionColor = "";
    if (activeColorBtn) {
        emotionColor = activeColorBtn.textContent.trim();
    } else {
        const customColorNameVal = document.getElementById("customColorName").value.trim();
        const hexVal = document.getElementById("customColorPicker").value;
        if (customColorNameVal) {
            emotionColor = `${customColorNameVal} (Mã màu: ${hexVal})`;
        } else {
            emotionColor = `Màu sắc tùy chọn (Mã màu: ${hexVal})`;
        }
    }
    const moodWithColor = mood ? `${mood} (${emotionColor})` : emotionColor;

    // Hiển thị trạng thái Loading
    document.getElementById("emptyState").style.display = "none";
    document.getElementById("resultContainer").style.display = "none";
    document.getElementById("loadingState").style.display = "flex";
    document.getElementById("composeBtn").disabled = true;

    // Tự động chuyển sang Tab Kết quả trên di động để thấy trạng thái loading
    const appContent = document.querySelector(".app-content");
    if (appContent) {
        appContent.classList.remove("mobile-show-config", "mobile-show-history");
        appContent.classList.add("mobile-show-result");
        updateMobileTabActive("result");
    }
    
    // Tùy biến text loading
    let loadingText = "Đang sáng tác...";
    if (mode === "remix") loadingText = "Đang phối lại nhạc...";
    if (mode === "rewrite") loadingText = "Đang viết lại lời mới...";
    document.getElementById("composeBtn").innerHTML = `<i data-lucide="loader-2" class="animate-spin"></i> ${loadingText}`;
    lucide.createIcons(); // reload icon loader-2 để có hiệu ứng xoay

    try {
        const response = await fetch("/api/compose", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                topic: topic,
                catholicDegree: catholicDegree,
                genre: genre || "Acoustic / Pop Ballad",
                verses: verses,
                repeatVerse: repeatVerse,
                chorusPitch: chorusPitch,
                voice: voice || "Nam trầm ấm",
                tempo: tempo || "Vừa (Moderate)",
                mood: moodWithColor,
                instruments: instruments,
                key: musicKey,
                vocalHarmony: vocalHarmony,
                vocalTechnique: vocalTechnique,
                vocalPlacement: vocalPlacement,
                existingLyrics: (mode === "remix" || mode === "rewrite") ? currentSongData.lyrics : "",
                rewritePrompt: mode === "rewrite" ? extraParam : ""
            })
        });

        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || "Gặp sự cố kết nối tới máy chủ.");
        }

        // Lưu trữ kết quả nhận được
        currentSongData.id = data.id || "";
        currentSongData.title = data.title || "";
        currentSongData.key = data.key || musicKey || "Tự động";
        currentSongData.style = data.style || "";
        currentSongData.lyrics = data.lyrics || "";
        currentSongData.abcNotation = data.abcNotation || "";
        currentSongData.vocalHarmony = data.vocalHarmony || "";
        currentSongData.vocalTechnique = data.vocalTechnique || "";
        currentSongData.vocalPlacement = data.vocalPlacement || "";

        // Hiển thị lên UI
        document.getElementById("resTitleDisplay").textContent = currentSongData.title;
        document.getElementById("resKey").textContent = currentSongData.key;
        document.getElementById("styleText").textContent = currentSongData.style;
        
        const idText = document.getElementById("resSongId");
        const idWrapper = document.getElementById("resSongIdWrapper");
        if (currentSongData.id) {
            idText.textContent = currentSongData.id;
            idWrapper.style.display = "inline-flex";
        } else {
            idWrapper.style.display = "none";
        }
        
        const showChords = document.getElementById("chordToggle").checked;
        renderLyrics(currentSongData.lyrics, showChords);
        renderSheetMusic(currentSongData.abcNotation);
        
        // Cập nhật lại đếm lịch sử
        updateHistoryCount();

        // Chuyển đổi trạng thái giao diện sang Kết quả
        document.getElementById("loadingState").style.display = "none";
        document.getElementById("resultContainer").style.display = "flex";
        
        // Cuộn mượt mà xuống phần kết quả
        document.getElementById("resultContainer").scrollIntoView({ behavior: 'smooth' });

    } catch (error) {
        alert("Lỗi: " + error.message);
        document.getElementById("loadingState").style.display = "none";
        document.getElementById("emptyState").style.display = "flex";
    } finally {
        document.getElementById("composeBtn").disabled = false;
        document.getElementById("composeBtn").innerHTML = `<i data-lucide="wand-2"></i> Tạo bài hát`;
        lucide.createIcons();
    }
}

// Xử lý và format lời bài hát
function renderLyrics(lyricsText, showChords) {
    const container = document.getElementById("lyricsText");
    if (!lyricsText) {
        container.innerHTML = "";
        return;
    }

    // 1. Phân biệt hợp âm và tag phân đoạn của Suno
    // Quy tắc: Hợp âm có dạng [Bm], [C], [F#m7], [D/F#] (không có khoảng trắng, ngắn hơn 8 ký tự)
    // Phân đoạn Suno có dạng [Intro...], [Verse 1...], [Chorus...] (chứa khoảng trắng hoặc dài hơn 8 ký tự)
    
    // Tạo bản sao lời nhạc để xử lý
    let formattedText = lyricsText;

    if (!showChords) {
        // Loại bỏ hoàn toàn các hợp âm trong ngoặc vuông
        // Tìm các cặp [] có độ dài ngắn và không chứa dấu cách
        formattedText = formattedText.replace(/\[([A-G][a-zA-Z0-9#\/+]*?)\]/g, "");
    }

    // 2. Chuyển đổi ký tự HTML để tránh lỗi XSS và hiển thị chuẩn xác
    let escapedText = escapeHTML(formattedText);

    // 3. Highlight các tag phân đoạn của Suno (e.g. [Intro...], [Verse 1...])
    // Thay thế các tag [Phân đoạn...] thành class .suno-tag để hiển thị màu vàng và xuống hàng đẹp mắt
    escapedText = escapedText.replace(/\[(Intro|Verse|Chorus|Pre-Chorus|Bridge|Outro|Instrumental|Solo|Drop|Fade|Key)(.*?)\]/gi, (match) => {
        return `<span class="suno-tag">${match}</span>`;
    });

    // 4. Nếu hiển thị hợp âm, highlight các tag hợp âm còn lại
    if (showChords) {
        escapedText = escapedText.replace(/\[([A-G][a-zA-Z0-9#\/+]*?)\]/g, (match) => {
            return `<span class="chord">${match}</span>`;
        });
    }

    container.innerHTML = escapedText;
}

// Lấy lời bài hát đang hiển thị dạng text thuần tuý (dùng để copy)
function getVisibleLyrics() {
    const container = document.getElementById("lyricsText");
    if (!container) return "";
    
    // Nếu hiện hợp âm, copy lời gốc có chèn ngoặc vuông hợp âm [Bm]
    // Nếu tắt hợp âm, ta copy lời sạch
    const showChords = document.getElementById("chordToggle").checked;
    
    if (showChords) {
        return currentSongData.lyrics;
    } else {
        // Strip hợp âm
        return currentSongData.lyrics.replace(/\[([A-G][a-zA-Z0-9#\/+]*?)\]/g, "");
    }
}

// Copy văn bản vào Clipboard (Có fallback cho HTTP không bảo mật)
function copyToClipboard(text, buttonElement) {
    if (!text) return;
    
    function showSuccess() {
        const originalHTML = buttonElement.innerHTML;
        buttonElement.innerHTML = `<i data-lucide="check"></i> Đã chép!`;
        buttonElement.classList.add("copied");
        lucide.createIcons();

        setTimeout(() => {
            buttonElement.innerHTML = originalHTML;
            buttonElement.classList.remove("copied");
            lucide.createIcons();
        }, 2000);
    }

    if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).then(showSuccess).catch(err => {
            console.error("Lỗi Clipboard API:", err);
            fallbackCopy(text);
        });
    } else {
        fallbackCopy(text);
    }

    function fallbackCopy(str) {
        try {
            const textarea = document.createElement("textarea");
            textarea.value = str;
            textarea.style.position = "fixed";
            textarea.style.left = "-9999px";
            document.body.appendChild(textarea);
            textarea.select();
            textarea.setSelectionRange(0, 99999); // cho mobile
            
            const successful = document.execCommand("copy");
            document.body.removeChild(textarea);
            
            if (successful) {
                showSuccess();
            } else {
                throw new Error("execCommand copy returned false");
            }
        } catch (err) {
            console.error("Fallback copy error:", err);
            alert("Trình duyệt không hỗ trợ tự động sao chép. Vui lòng bôi đen văn bản và sao chép thủ công.");
        }
    }
}

// Escape HTML ký tự đặc biệt
function escapeHTML(str) {
    return str.replace(/[&<>'"]/g, 
        tag => ({
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            "'": '&#39;',
            '"': '&quot;'
        }[tag] || tag)
    );
}

// Cập nhật số lượng bài hát trên Tab
async function updateHistoryCount() {
    try {
        const response = await fetch("/api/songs");
        if (!response.ok) return;
        const data = await response.json();
        const countSpan = document.getElementById("historyCount");
        if (countSpan) {
            countSpan.textContent = data.length || 0;
        }
    } catch (e) {
        console.error("Lỗi lấy số lượng lịch sử:", e);
    }
}

// Tải và hiển thị danh sách bài hát trong lịch sử
async function loadHistory() {
    const listContainer = document.getElementById("historyList");
    if (!listContainer) return;

    listContainer.innerHTML = `
        <div style="text-align: center; padding: 40px; color: var(--text-secondary);">
            <i data-lucide="loader-2" class="animate-spin" style="width: 24px; height: 24px; margin: 0 auto 10px;"></i>
            <p>Đang đọc lịch sử...</p>
        </div>
    `;
    lucide.createIcons();

    try {
        const response = await fetch("/api/songs");
        if (!response.ok) throw new Error("Không thể kết nối API");
        const songs = await response.json();

        if (songs.length === 0) {
            listContainer.innerHTML = `
                <div style="text-align: center; padding: 60px 20px; color: var(--text-secondary);">
                    <i data-lucide="folder-open" style="width: 48px; height: 48px; margin: 0 auto 16px; opacity: 0.5;"></i>
                    <p>Chưa có bài hát nào được sáng tác.</p>
                </div>
            `;
            lucide.createIcons();
            return;
        }

        listContainer.innerHTML = "";
        songs.forEach(song => {
            const card = document.createElement("div");
            card.className = "history-card";
            
            const dateStr = new Date(song.createdAt).toLocaleString('vi-VN', {
                year: 'numeric', month: 'numeric', day: 'numeric',
                hour: '2-digit', minute: '2-digit'
            });

            card.innerHTML = `
                <div class="history-card-header">
                    <span class="history-card-title">${escapeHTML(song.title)}</span>
                    <span class="history-card-date">${dateStr}</span>
                </div>
                <div class="history-card-topic">${escapeHTML(song.topic)}</div>
                <div class="history-card-meta">
                    <span class="history-card-badge">Tông: ${song.key || "Tự động"}</span>
                    <span class="history-card-badge">Thể loại: ${song.genre}</span>
                    <span class="history-card-badge">Giọng: ${song.voice}</span>
                    <span class="history-card-badge">Nhịp: ${song.tempo}</span>
                </div>
                <div class="history-card-actions">
                    <button class="btn-action btn-load-song" data-id="${song.id}">
                        <i data-lucide="eye"></i> Xem & Sửa
                    </button>
                    <button class="btn-action btn-remix-song" data-id="${song.id}">
                        <i data-lucide="rotate-ccw"></i> Nạp cấu hình
                    </button>
                    <button class="btn-action btn-danger btn-delete-song" data-id="${song.id}">
                        <i data-lucide="trash-2"></i> Xóa
                    </button>
                </div>
            `;
            listContainer.appendChild(card);
        });

        lucide.createIcons();

        // Đăng ký sự kiện cho các nút
        listContainer.querySelectorAll(".btn-load-song").forEach(btn => {
            btn.addEventListener("click", () => viewSavedSong(btn.getAttribute("data-id")));
        });
        
        listContainer.querySelectorAll(".btn-remix-song").forEach(btn => {
            btn.addEventListener("click", () => loadSongConfig(btn.getAttribute("data-id")));
        });

        listContainer.querySelectorAll(".btn-delete-song").forEach(btn => {
            btn.addEventListener("click", () => deleteSavedSong(btn.getAttribute("data-id")));
        });

    } catch (error) {
        listContainer.innerHTML = `
            <div style="text-align: center; padding: 40px; color: var(--color-danger);">
                Lỗi tải lịch sử: ${error.message}
            </div>
        `;
    }
}

// Xem chi tiết bài hát cũ (Load vào Tab 1 để xem và sửa)
async function viewSavedSong(id) {
    try {
        const response = await fetch(`/api/songs?id=${id}`);
        if (!response.ok) throw new Error("Lỗi tải chi tiết bài hát");
        const song = await response.json();

        // Nạp vào state
        currentSongData.id = song.id;
        currentSongData.title = song.title;
        currentSongData.key = song.key || "Tự động";
        currentSongData.style = song.style;
        currentSongData.lyrics = song.lyrics;
        currentSongData.abcNotation = song.abcNotation || "";
        currentSongData.vocalHarmony = song.vocalHarmony || "";
        currentSongData.vocalTechnique = song.vocalTechnique || "";
        currentSongData.vocalPlacement = song.vocalPlacement || "";
        currentSongData.sunoClips = song.sunoClips || [];

        // Render UI
        document.getElementById("resTitleDisplay").textContent = currentSongData.title;
        document.getElementById("resKey").textContent = currentSongData.key;
        document.getElementById("styleText").textContent = currentSongData.style;
        
        const idText = document.getElementById("resSongId");
        const idWrapper = document.getElementById("resSongIdWrapper");
        idText.textContent = currentSongData.id;
        idWrapper.style.display = "inline-flex";

        const showChords = document.getElementById("chordToggle").checked;
        renderLyrics(currentSongData.lyrics, showChords);
        renderSheetMusic(currentSongData.abcNotation);
        
        // Render Suno clips if any exist
        renderSunoClips(currentSongData.sunoClips);

        // Tự động kích hoạt polling nếu có clip chưa hoàn thành (queued hoặc streaming)
        const unfinishedClips = currentSongData.sunoClips.filter(c => c.status !== "complete" && c.status !== "failed");
        if (unfinishedClips.length > 0) {
            const clipIds = unfinishedClips.map(c => c.id);
            const accountEmail = unfinishedClips[0].accountEmail;
            const accounts = getSunoAccounts();
            const account = accounts.find(a => a.email === accountEmail) || accounts[0];
            if (account) {
                const controls = document.getElementById("sunoGenControls");
                const loading = document.getElementById("sunoGenLoading");
                const loadingStatus = document.getElementById("sunoLoadingStatus");
                const loadingSubtext = document.getElementById("sunoLoadingSubtext");
                
                if (controls) controls.style.display = "none";
                if (loading) loading.style.display = "flex";
                if (loadingStatus) loadingStatus.textContent = "Đang kiểm tra trạng thái clips...";
                if (loadingSubtext) loadingSubtext.textContent = `Tài khoản: ${account.email}`;
                
                pollSunoStatus(clipIds, currentSongData.id, account);
            }
        }

        // Chuyển sang Tab 1
        const currentSongTabBtn = document.querySelector('.tab-btn[data-tab="current-song"]');
        if (currentSongTabBtn) {
            currentSongTabBtn.click();
        }

        // Hiển thị kết quả, ẩn empty state
        document.getElementById("emptyState").style.display = "none";
        document.getElementById("loadingState").style.display = "none";
        document.getElementById("resultContainer").style.display = "flex";

    } catch (e) {
        alert("Không thể tải bài hát: " + e.message);
    }
}

// Nạp lại các thiết bị lọc/ý tưởng của bài hát cũ vào Form trái
async function loadSongConfig(id) {
    try {
        const response = await fetch(`/api/songs?id=${id}`);
        if (!response.ok) throw new Error("Lỗi tải cấu hình bài hát");
        const song = await response.json();

        // 1. Nhập ý tưởng
        document.getElementById("topic").value = song.topic;

        // 2. Mức độ Công giáo
        setGroupActiveValue("catholicDegree", song.catholicDegree);

        // 3. Số lượng Lời & Checkbox
        setGroupActiveValue("verses", song.verses);
        document.getElementById("repeatVerse").checked = song.repeatVerse;

        // 4. Cao độ Điệp khúc & Tông Key
        setGroupActiveValue("chorusPitch", song.chorusPitch);
        document.getElementById("musicKey").value = song.key;

        // 5. Tốc độ Tempo
        setGroupActiveValue("tempo", song.tempo);

        // 6. Thể loại, Giọng hát, Tâm trạng (Chips)
        setChipActiveValue("genre", song.genre);
        setChipActiveValue("voiceStyle", song.voice);
        
        // Phân tích tâm trạng và phục hồi màu sắc
        let cleanMood = song.mood;
        if (song.mood.includes("(")) {
            const parts = song.mood.split("(");
            cleanMood = parts[0].trim();
            
            // Phục hồi màu sắc tùy chỉnh
            const colorPart = parts[1].replace(")", "").trim();
            const hexMatch = colorPart.match(/#([a-fA-F0-9]{6})/);
            
            if (hexMatch) {
                const hexColor = hexMatch[0];
                const customName = colorPart.split("Mã màu:")[0].trim();
                
                // Active custom color
                document.getElementById("customColorPicker").value = hexColor;
                document.getElementById("customColorName").value = customName;
                
                // Tắt tất cả màu preset
                document.querySelectorAll(".color-btn").forEach(b => b.classList.remove("active"));
                
                // Cập nhật màu giao diện
                updateThemeColor(hexColor);
            }
        }
        setChipActiveValue("mood", cleanMood);

        // 7. Nhạc cụ (Multi select)
        setMultiChipActiveValues("instruments", song.instruments || []);

        // 8. Phục hồi cấu hình phối bè
        setGroupActiveValue("vocalHarmony", song.vocalHarmony);
        setGroupActiveValue("vocalTechnique", song.vocalTechnique);
        setGroupActiveValue("vocalPlacement", song.vocalPlacement);

        // Chuyển sang Tab 1 để người dùng sẵn sàng nhấn nút Sáng Tác
        const currentSongTabBtn = document.querySelector('.tab-btn[data-tab="current-song"]');
        if (currentSongTabBtn) {
            currentSongTabBtn.click();
        }

        // Tự động chuyển sang Tab Sáng tác trên di động vì họ đã nạp cấu hình để sửa đổi
        const appContent = document.querySelector(".app-content");
        if (appContent) {
            appContent.classList.remove("mobile-show-result", "mobile-show-history");
            appContent.classList.add("mobile-show-config");
            updateMobileTabActive("config");
        }

        // Highlight nhẹ textarea để báo hiệu đã load
        const topicArea = document.getElementById("topic");
        topicArea.style.boxShadow = "0 0 15px var(--color-primary)";
        setTimeout(() => {
            topicArea.style.boxShadow = "none";
        }, 1200);

        alert("Đã khôi phục toàn bộ cấu hình bộ lọc của bài hát này sang bảng bên trái!");
    } catch (e) {
        alert("Lỗi nạp cấu hình: " + e.message);
    }
}

// Xóa bài hát khỏi lịch sử
async function deleteSavedSong(id) {
    if (!confirm("Bạn có chắc chắn muốn xóa bài hát này khỏi lịch sử lưu trữ không?")) {
        return;
    }

    try {
        const response = await fetch(`/api/songs?id=${id}`, {
            method: "DELETE"
        });
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || "Lỗi xóa bài hát");
        }

        // Nếu bài hát đang xem chính là bài vừa xóa, ẩn kết quả đi
        if (currentSongData.id === id) {
            currentSongData = { id: "", title: "", key: "", style: "", lyrics: "" };
            document.getElementById("resultContainer").style.display = "none";
            document.getElementById("emptyState").style.display = "flex";
        }

        loadHistory();
        updateHistoryCount();
    } catch (e) {
        alert("Không thể xóa: " + e.message);
    }
}

// Helper set active cho Button Group (Hỗ trợ nhiều giá trị cách nhau bằng dấu phẩy)
function setGroupActiveValue(groupId, val) {
    const container = document.getElementById(groupId);
    if (!container) return;
    
    let vals = [];
    if (typeof val === 'string') {
        vals = val.split(",").map(v => v.trim());
    } else if (Array.isArray(val)) {
        vals = val;
    } else if (val !== undefined && val !== null) {
        vals = [val.toString().trim()];
    }

    container.querySelectorAll(".btn-item").forEach(item => {
        const itemVal = item.getAttribute("data-val");
        if (vals.includes(itemVal)) {
            item.classList.add("active");
        } else {
            item.classList.remove("active");
        }
    });
}

// Helper set active cho Chips (Hỗ trợ nhiều giá trị cách nhau bằng dấu phẩy)
function setChipActiveValue(containerId, val) {
    const container = document.getElementById(containerId);
    if (!container) return;

    let vals = [];
    if (typeof val === 'string') {
        vals = val.split(",").map(v => v.trim());
    } else if (Array.isArray(val)) {
        vals = val;
    } else if (val !== undefined && val !== null) {
        vals = [val.toString().trim()];
    }

    container.querySelectorAll(".chip-item").forEach(chip => {
        const chipVal = chip.getAttribute("data-val");
        if (vals.includes(chipVal)) {
            chip.classList.add("active");
        } else {
            chip.classList.remove("active");
        }
    });
}

// Helper set active cho Multi select Chips
function setMultiChipActiveValues(containerId, values) {
    const container = document.getElementById(containerId);
    if (!container) return;
    container.querySelectorAll(".chip-item").forEach(chip => {
        const chipVal = chip.getAttribute("data-val");
        if (values.includes(chipVal)) {
            chip.classList.add("active");
        } else {
            chip.classList.remove("active");
        }
    });
}

// Vẽ khuôn nhạc (Sheet nhạc nốt) bằng abcjs
function renderSheetMusic(abcString) {
    const card = document.getElementById("sheetMusicCard");
    const container = document.getElementById("abcSheetMusic");
    
    if (!container) return;

    if (!abcString || abcString.trim() === "") {
        if (card) card.style.display = "none";
        container.innerHTML = "";
        return;
    }

    if (card) card.style.display = "block";
    try {
        ABCJS.renderAbc("abcSheetMusic", abcString, {
            responsive: "resize",
            paddingtop: 0,
            paddingbottom: 0,
            paddingright: 10,
            paddingleft: 10,
            scale: 0.9,
            add_classes: true
        });
    } catch (e) {
        console.error("Lỗi vẽ sheet nhạc abcjs:", e);
        if (card) card.style.display = "none";
        container.innerHTML = "";
    }
}

// Helper xử lý làm sạch lời bài hát và định dạng hợp âm chữ đỏ cho PDF
function processLyricsForPDF(rawLyrics) {
    const lines = rawLyrics.split("\n");
    const outputLines = [];

    for (let line of lines) {
        let lineProcessed = line;
        
        // Kiểm tra xem dòng này có chứa lời hát thực tế (không phải nhãn hay hợp âm) hay không.
        // Hợp âm và nhãn nằm trong ngoặc vuông, nếu xóa hết thì phần chữ còn lại phải dài hơn 0.
        const lyricTextOnly = line.replace(/\[.*?\]/g, "").trim();
        const hasLyricsContent = lyricTextOnly.length > 0;

        // Nếu dòng này không chứa ca từ hát thực tế (ví dụ dòng dạo hợp âm [Bm] [G] [A] hay dòng nhãn [Intro]), ta bỏ qua cả dòng
        if (!hasLyricsContent) {
            continue;
        }

        // Tìm các thẻ trong ngoặc vuông [x]
        const bracketRegex = /\[(.*?)\]/g;
        const matches = [...line.matchAll(bracketRegex)];

        for (let match of matches) {
            const innerText = match[1].trim();
            
            // Nhận diện xem có phải nhãn phân đoạn/nhạc dạo/hook/solo/interlude hay không
            const isKnownTag = /^(Intro|Verse|Chorus|Pre-Chorus|Bridge|Outro|Instrumental|Solo|Hook|Drop|Fade|Key|End|Guitar|Piano|Drum|Soft|Interlude|Transition)/i.test(innerText);
            // Hợp âm: Thường ngắn dưới 8 ký tự, bắt đầu bằng A-G
            const isChord = !isKnownTag && innerText.length <= 8 && /^[A-G][a-zA-Z0-9#\/+]*$/.test(innerText);

            if (isChord) {
                // Đổi thành chữ đỏ đậm, bỏ ngoặc vuông
                const chordSpan = `<span style="color: #dc2626; font-weight: bold; font-family: 'Outfit', sans-serif; font-size: 12.5px; margin: 0 2px;">${innerText}</span>`;
                lineProcessed = lineProcessed.replace(match[0], chordSpan);
            } else {
                // Xóa bỏ các nhãn phân đoạn, nhạc dạo, hook, solo
                lineProcessed = lineProcessed.replace(match[0], "");
            }
        }

        outputLines.push(lineProcessed);
    }

    // Ghép lại và dọn dẹp dòng trống thừa
    return outputLines.join("\n").replace(/\n{3,}/g, "\n\n");
}

// Tạo và tải file PDF lời, hợp âm kèm sheet nhạc nốt
function downloadPDF() {
    if (!currentSongData.lyrics) {
        alert("Chưa có bài hát nào được nạp hiển thị để tải PDF!");
        return;
    }

    const element = document.createElement("div");
    element.style.padding = "35px 45px";
    element.style.background = "#ffffff";
    element.style.color = "#1e293b";
    element.style.fontFamily = "'Outfit', 'Plus Jakarta Sans', sans-serif";

    // 1. Tiêu đề
    const titleEl = document.createElement("h1");
    titleEl.textContent = currentSongData.title || "Tác phẩm không tên";
    titleEl.style.textAlign = "center";
    titleEl.style.fontSize = "26px";
    titleEl.style.fontWeight = "700";
    titleEl.style.color = "#000000";
    titleEl.style.marginBottom = "6px";
    element.appendChild(titleEl);

    // 2. Tác giả
    const authorEl = document.createElement("p");
    authorEl.textContent = "Tác giả: Người con tội lỗi";
    authorEl.style.textAlign = "center";
    authorEl.style.fontSize = "13px";
    authorEl.style.color = "#475569";
    authorEl.style.fontStyle = "italic";
    authorEl.style.marginBottom = "24px";
    element.appendChild(authorEl);

    // 3. Tông nhạc (Key)
    const metaEl = document.createElement("p");
    metaEl.innerHTML = `Điệu thức / Tông nhạc (Key): <strong>${currentSongData.key || "Tự động"}</strong>`;
    metaEl.style.fontSize = "11px";
    metaEl.style.color = "#475569";
    metaEl.style.marginBottom = "20px";
    metaEl.style.borderBottom = "1px solid #e2e8f0";
    metaEl.style.paddingBottom = "8px";
    element.appendChild(metaEl);

    // 4. Lời bài hát & Hợp âm (chữ đỏ, lọc sạch nhãn và dòng chỉ chứa hợp âm)
    const cleanLyricsHTML = processLyricsForPDF(currentSongData.lyrics);

    const lyricsPre = document.createElement("pre");
    lyricsPre.innerHTML = cleanLyricsHTML;
    lyricsPre.style.fontFamily = "'Plus Jakarta Sans', sans-serif";
    lyricsPre.style.fontSize = "13px";
    lyricsPre.style.whiteSpace = "pre-wrap";
    lyricsPre.style.lineHeight = "1.8";
    lyricsPre.style.color = "#1e293b";
    element.appendChild(lyricsPre);

    // 5. Footer
    const footerEl = document.createElement("p");
    footerEl.textContent = "Sáng tác tự động bởi Suno Music Composer AI - Tác phẩm bản quyền thuộc Người con tội lỗi";
    footerEl.style.fontSize = "9px";
    footerEl.style.color = "#94a3b8";
    footerEl.style.textAlign = "center";
    footerEl.style.marginTop = "30px";
    footerEl.style.borderTop = "1px solid #f1f5f9";
    footerEl.style.paddingTop = "10px";
    element.appendChild(footerEl);

    // Tạo hộp chứa ẩn tạm thời có kích thước bằng 0 để trình duyệt layout vẽ nhưng không hiện ra cho người dùng
    const printContainer = document.createElement("div");
    printContainer.style.position = "fixed";
    printContainer.style.left = "0";
    printContainer.style.top = "0";
    printContainer.style.width = "0";
    printContainer.style.height = "0";
    printContainer.style.overflow = "hidden";
    printContainer.style.zIndex = "-9999";
    
    element.style.width = "800px"; // Cố định chiều rộng tương đương độ phân giải in A4
    printContainer.appendChild(element);
    document.body.appendChild(printContainer);

    // Cấu hình kết xuất PDF (in liên tục trên trang A4 dọc)
    const opt = {
        margin:       10,
        filename:     `${currentSongData.title || "Bai-hat"}.pdf`,
        image:        { type: 'jpeg', quality: 0.98 },
        html2canvas:  { scale: 2, useCORS: true, logging: false },
        jsPDF:        { unit: 'mm', format: 'a4', orientation: 'portrait' }
    };

    html2pdf().set(opt).from(element).save().then(() => {
        document.body.removeChild(printContainer);
    }).catch(err => {
        console.error("Lỗi tải PDF:", err);
        document.body.removeChild(printContainer);
    });
}

// ==========================================
// SUNO INTEGRATION LOGIC
// ==========================================

const sunoPollIntervals = {};

// Khởi tạo các sự kiện cho tích hợp Suno
function initSunoIntegration() {
    const modal = document.getElementById("sunoConfigModal");
    const btnHeaderOpen = document.getElementById("headerSunoConfigBtn");
    const btnInlineOpen = document.getElementById("btnOpenSunoConfigInline");
    const btnClose = document.getElementById("btnCloseSunoConfig");
    const btnSave = document.getElementById("btnSaveSunoConfig");
    const btnImport = document.getElementById("btnImportSunoCurl");
    const btnManualAdd = document.getElementById("btnManualAddAccount");
    const btnGenMusic = document.getElementById("btnGenSunoMusic");
    
    // Cập nhật giao diện ban đầu
    updateSunoUIState();
    
    // Tải cấu hình Model Version từ localStorage
    const savedModelVersion = localStorage.getItem("suno_model_version") || "chirp-fenix";
    const modelSelect = document.getElementById("cfgModelVersion");
    if (modelSelect) modelSelect.value = savedModelVersion;
    
    // Mở modal cấu hình
    const openModal = () => {
        document.getElementById("sunoCurlInput").value = "";
        modal.classList.add("active");
        renderSunoAccounts();
    };
    
    if (btnHeaderOpen) btnHeaderOpen.addEventListener("click", openModal);
    if (btnInlineOpen) btnInlineOpen.addEventListener("click", openModal);
    
    // Đóng modal
    const closeModal = () => modal.classList.remove("active");
    if (btnClose) btnClose.addEventListener("click", closeModal);
    if (btnSave) btnSave.addEventListener("click", () => {
        const mvSelect = document.getElementById("cfgModelVersion");
        if (mvSelect) {
            localStorage.setItem("suno_model_version", mvSelect.value);
        }
        closeModal();
        updateSunoUIState();
    });
    
    // Thêm tài khoản từ cURL
    if (btnImport) btnImport.addEventListener("click", () => {
        const curlText = document.getElementById("sunoCurlInput").value;
        if (!curlText.trim()) {
            alert("Vui lòng nhập lệnh cURL.");
            return;
        }
        
        const config = parseSunoCurl(curlText);
        
        if (config.authToken) {
            addSunoAccount(config);
            document.getElementById("sunoCurlInput").value = "";
            renderSunoAccounts();
            updateSunoUIState();
            alert("Đã thêm tài khoản từ cURL thành công!");
        } else if (config.cookie) {
            // Nếu chỉ có cookie (ví dụ cURL từ clerk.suno.com), cập nhật cho tài khoản hiện tại
            const accounts = getSunoAccounts();
            if (accounts.length > 0) {
                const currentIndex = parseInt(localStorage.getItem("suno_current_account_index") || "0", 10) % accounts.length;
                accounts[currentIndex].cookie = config.cookie;
                saveSunoAccounts(accounts);
                document.getElementById("sunoCurlInput").value = "";
                renderSunoAccounts();
                alert(`Đã cập nhật Cookie thành công cho tài khoản: ${accounts[currentIndex].email}!`);
            } else {
                alert("Chưa có tài khoản nào. Vui lòng thêm cURL có chứa Authorization token trước.");
            }
        } else {
            alert("Không tìm thấy thông tin Authorization Bearer token hoặc Cookie trong lệnh cURL. Vui lòng kiểm tra lại lệnh cURL của bạn.");
        }
    });
    
    const btnImportCookieJson = document.getElementById("btnImportCookieJson");
    const sunoCookieJsonInput = document.getElementById("sunoCookieJsonInput");

    if (btnImportCookieJson && sunoCookieJsonInput) {
        btnImportCookieJson.addEventListener("click", () => {
            sunoCookieJsonInput.click();
        });

        sunoCookieJsonInput.addEventListener("change", async (e) => {
            const file = e.target.files[0];
            if (!file) return;

            const reader = new FileReader();
            reader.onload = async (event) => {
                try {
                    const cookies = JSON.parse(event.target.result);
                    if (!Array.isArray(cookies)) {
                        throw new Error("Định dạng JSON không hợp lệ. Phải là một mảng Cookie.");
                    }

                    const cookieStr = cookies
                        .filter(c => c.name === '__client' || c.name === '__client_uat')
                        .map(c => `${c.name}=${c.value}`)
                        .join('; ');

                    if (!cookieStr) {
                        alert("Không tìm thấy Cookie __client nào trong file JSON.");
                        return;
                    }

                    // Cải tiến: Dùng cookie gọi API để lấy luôn JWT mới thay vì chỉ cập nhật
                    const btnImportCookieJson = document.getElementById("btnImportCookieJson");
                    const originalHtml = btnImportCookieJson.innerHTML;
                    btnImportCookieJson.innerHTML = `<i data-lucide="loader" class="spin" style="width: 14px; height: 14px;"></i> Đang lấy Token...`;
                    btnImportCookieJson.disabled = true;
                    lucide.createIcons();

                    try {
                        const response = await fetch("/api/suno/refresh", {
                            method: "POST",
                            headers: { "Content-Type": "application/json" },
                            body: JSON.stringify({ cookie: cookieStr })
                        });
                        
                        const data = await response.json();
                        if (!response.ok) {
                            throw new Error(data.error || "Lỗi không xác định khi làm mới token");
                        }

                        // Đã lấy được JWT mới từ cookie
                        const newJwt = data.new_auth_token;
                        const newCookieStr = data.new_cookie || cookieStr;
                        
                        addSunoAccount({
                            authToken: newJwt,
                            cookie: newCookieStr
                        });
                        
                        renderSunoAccounts();
                        updateSunoUIState();
                        alert("Tuyệt vời! Đã tự động tạo/cập nhật tài khoản hoàn chỉnh chỉ từ file Cookie JSON.");
                    } catch (apiErr) {
                        alert("Lỗi khi dùng Cookie lấy JWT: " + apiErr.message);
                    } finally {
                        btnImportCookieJson.innerHTML = originalHtml;
                        btnImportCookieJson.disabled = false;
                        lucide.createIcons();
                    }
                } catch (err) {
                    alert("Lỗi đọc file JSON: " + err.message);
                }
                
                sunoCookieJsonInput.value = "";
            };
            reader.readAsText(file);
        });
    }

    // Thêm tài khoản thủ công từ Form
    if (btnManualAdd) btnManualAdd.addEventListener("click", () => {
        const auth = document.getElementById("cfgAuthToken").value.trim();
        const browserToken = document.getElementById("cfgBrowserToken").value.trim();
        const deviceId = document.getElementById("cfgDeviceId").value.trim();
        const userTier = document.getElementById("cfgUserTier").value.trim();
        const sessionToken = document.getElementById("cfgSessionToken").value.trim();
        const bodyToken = document.getElementById("cfgBodyToken").value.trim();
        const cookie = document.getElementById("cfgCookie").value.trim();
        
        if (!auth) {
            alert("Vui lòng điền Authorization token!");
            return;
        }
        
        const config = {
            authToken: auth,
            browserToken: browserToken,
            deviceId: deviceId,
            userTier: userTier,
            createSessionToken: sessionToken,
            sunoToken: bodyToken,
            cookie: cookie
        };
        
        addSunoAccount(config);
        
        // Reset manual form fields
        document.getElementById("cfgAuthToken").value = "";
        document.getElementById("cfgBrowserToken").value = "";
        document.getElementById("cfgDeviceId").value = "";
        document.getElementById("cfgUserTier").value = "";
        document.getElementById("cfgSessionToken").value = "";
        document.getElementById("cfgBodyToken").value = "";
        document.getElementById("cfgCookie").value = "";
        
        renderSunoAccounts();
        updateSunoUIState();
        alert("Đã thêm tài khoản thủ công thành công!");
    });
    
    // Sáng tác nhạc trên Suno
    if (btnGenMusic) btnGenMusic.addEventListener("click", generateSunoMusic);
}

// Phân tích cú pháp cURL từ Suno
function parseSunoCurl(curlText) {
    const config = {};
    
    // Tìm các header -H hoặc --header, hỗ trợ cả prefix $ và dấu nháy phức tạp
    const hRegex = /(?:-H|--header)\s+[\$]?(?:(['"])(.*?)\1|([^\s'"]+))/gi;
    let match;
    while ((match = hRegex.exec(curlText)) !== null) {
        const headerValue = match[2] || match[3];
        if (!headerValue) continue;
        
        const colonIndex = headerValue.indexOf(':');
        if (colonIndex > -1) {
            const key = headerValue.substring(0, colonIndex).trim().toLowerCase();
            const value = headerValue.substring(colonIndex + 1).trim();
            if (key === 'authorization') {
                config.authToken = value;
            } else if (key === 'browser-token') {
                config.browserToken = value;
            } else if (key === 'device-id') {
                config.deviceId = value;
            } else if (key === 'cookie') {
                config.cookie = value;
            }
        }
    }
    
    // Tìm phần body dữ liệu (bắt đầu tìm từ vị trí của --data để tránh lấy nhầm JSON trong header browser-token)
    const dataIndex = curlText.search(/--data(-raw|-binary)?/i);
    if (dataIndex > -1) {
        const startSearchIndex = dataIndex;
        const startBrace = curlText.indexOf('{', startSearchIndex);
        const endBrace = curlText.lastIndexOf('}');
        if (startBrace > -1 && endBrace > startBrace) {
            const rawJson = curlText.substring(startBrace, endBrace + 1);
            try {
                let cleanedJson = rawJson.replace(/\\'/g, "'").replace(/\\"/g, '"');
                // Thử parse JSON trực tiếp
                try {
                    const parsedData = JSON.parse(rawJson);
                    if (parsedData.token) config.sunoToken = parsedData.token;
                    if (parsedData.metadata) {
                        config.userTier = parsedData.metadata.user_tier;
                        config.createSessionToken = parsedData.metadata.create_session_token;
                    }
                } catch (e) {
                    // Parse JSON đã dọn dẹp dấu nháy
                    const parsedData = JSON.parse(cleanedJson);
                    if (parsedData.token) config.sunoToken = parsedData.token;
                    if (parsedData.metadata) {
                        config.userTier = parsedData.metadata.user_tier;
                        config.createSessionToken = parsedData.metadata.create_session_token;
                    }
                }
            } catch (e) {
                console.error("Lỗi phân tích cú pháp body JSON cURL:", e);
            }
        }
    }
    
    return config;
}

// Quản lý tài khoản trên Server (Sync & Local Cache)
let localSunoAccounts = [];

async function fetchSunoAccounts() {
    try {
        const res = await fetch("/api/accounts");
        if (res.ok) {
            localSunoAccounts = await res.json();
            if (!localSunoAccounts) localSunoAccounts = [];
        }
    } catch (e) {
        console.error("Lỗi đồng bộ tài khoản từ server:", e);
    }
}

function getSunoAccounts() {
    return localSunoAccounts;
}

async function saveSunoAccountToServer(acc) {
    try {
        await fetch("/api/accounts", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(acc)
        });
    } catch (e) {
        console.error("Lỗi lưu tài khoản lên server:", e);
    }
}

async function saveSunoAccounts(accounts) {
    // Lặp qua mảng lưu từng tài khoản lên server nếu cần tương thích
    for (const acc of accounts) {
        await saveSunoAccountToServer(acc);
    }
    localSunoAccounts = accounts;
}

async function updateStoredSunoToken(email, newAuthToken, newCookie) {
    if (!email) return;
    let accounts = getSunoAccounts();
    const idx = accounts.findIndex(acc => acc.email === email);
    if (idx > -1) {
        let changed = false;
        if (newAuthToken && accounts[idx].authToken !== newAuthToken) {
            accounts[idx].authToken = newAuthToken;
            const jwtDecoded = decodeJWT(newAuthToken);
            if (jwtDecoded && jwtDecoded.exp) {
                accounts[idx].expiry = jwtDecoded.exp * 1000;
            }
            changed = true;
        }
        if (newCookie && accounts[idx].cookie !== newCookie) {
            accounts[idx].cookie = newCookie;
            changed = true;
            console.log(`Đã tự động cập nhật cookie mới cho tài khoản: ${email}`);
        }
        if (changed) {
            // Cập nhật memory ngay
            localSunoAccounts = accounts;
            // Đồng bộ lên server bất đồng bộ (không block UI)
            saveSunoAccountToServer(accounts[idx]);
            renderSunoAccounts();
            updateSunoUIState();
            console.log(`Đã tự động gia hạn token cho tài khoản: ${email}`);
        }
    }
}

async function addSunoAccount(config) {
    let accounts = getSunoAccounts();
    
    // Giải mã JWT để lấy email và expiry
    const jwtDecoded = decodeJWT(config.authToken);
    let email = "Tài khoản Suno";
    let exp = null;
    
    if (jwtDecoded) {
        email = jwtDecoded["suno.com/claims/email"] || jwtDecoded["https://suno.ai/claims/email"] || jwtDecoded["suno/username"] || "Tài khoản Suno";
        if (jwtDecoded.exp) {
            exp = jwtDecoded.exp * 1000;
        }
    }
    
    // Lấy lại các thông tin cũ nếu có (để không bị mất khi dán cURL generate)
    let oldCookie = "";
    let oldBrowserToken = "";
    let oldDeviceId = "";
    let oldUserTier = "4497580c-f4eb-4f86-9f0e-960eb7c48d7d";
    let oldSessionToken = "3d8d709b-97f1-4867-acfb-a014c499b58d";
    
    const existingAcc = accounts.find(acc => acc.email === email);
    let accId = existingAcc ? existingAcc.id : "acc_" + Date.now();

    if (email !== "Tài khoản Suno" && existingAcc) {
        oldCookie = existingAcc.cookie || "";
        oldBrowserToken = existingAcc.browserToken || "";
        oldDeviceId = existingAcc.deviceId || "";
        if (existingAcc.userTier) oldUserTier = existingAcc.userTier;
        if (existingAcc.createSessionToken) oldSessionToken = existingAcc.createSessionToken;
    }
    
    const newAcc = {
        id: accId,
        email: email,
        authToken: config.authToken,
        browserToken: config.browserToken || oldBrowserToken || "",
        deviceId: config.deviceId || oldDeviceId || "",
        cookie: config.cookie || oldCookie || "",
        userTier: config.userTier || oldUserTier,
        createSessionToken: config.createSessionToken || oldSessionToken,
        bodyToken: config.sunoToken || "",
        expiry: exp,
        addedAt: existingAcc ? existingAcc.addedAt : Date.now()
    };
    
    if (existingAcc) {
        accounts = accounts.map(a => a.email === email ? newAcc : a);
    } else {
        accounts.push(newAcc);
    }
    
    // Optimistic UI Update
    localSunoAccounts = accounts;
    // Cập nhật server
    saveSunoAccountToServer(newAcc);
}

async function deleteSunoAccount(id) {
    if (!confirm("Bạn có chắc chắn muốn xóa tài khoản này khỏi danh sách?")) return;
    
    let accounts = getSunoAccounts();
    const idx = accounts.findIndex(acc => acc.id === id);
    if (idx === -1) return;
    
    // Optimistic Delete
    accounts.splice(idx, 1);
    localSunoAccounts = accounts;
    
    const currentIndex = parseInt(localStorage.getItem("suno_current_account_index") || "0", 10);
    if (accounts.length > 0) {
        localStorage.setItem("suno_current_account_index", currentIndex % accounts.length);
    } else {
        localStorage.setItem("suno_current_account_index", "0");
    }
    
    renderSunoAccounts();
    updateSunoUIState();

    // Gửi yêu cầu xóa lên server
    try {
        await fetch(`/api/accounts?id=${id}`, { method: "DELETE" });
    } catch (e) {
        console.error("Lỗi xóa tài khoản trên server:", e);
    }
}

function decodeJWT(token) {
    try {
        const cleanToken = token.replace(/^bearer\s+/i, "").trim();
        const parts = cleanToken.split('.');
        if (parts.length < 2) return null;
        
        const base64Url = parts[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
        }).join(''));
        
        return JSON.parse(jsonPayload);
    } catch (e) {
        console.error("Lỗi giải mã JWT token:", e);
        return null;
    }
}

// Render giao diện danh sách tài khoản
function renderSunoAccounts() {
    const listContainer = document.getElementById("sunoAccountsList");
    const countSpan = document.getElementById("sunoAccountsCount");
    
    if (!listContainer) return;
    
    const accounts = getSunoAccounts();
    const currentIndex = parseInt(localStorage.getItem("suno_current_account_index") || "0", 10);
    
    if (countSpan) {
        countSpan.textContent = accounts.length + " tài khoản";
    }
    
    if (accounts.length === 0) {
        listContainer.innerHTML = `
            <div style="text-align: center; padding: 20px; color: var(--text-muted); font-size: 0.8rem; border: 1px dashed var(--border-color); border-radius: 10px;">
                Chưa có tài khoản nào được thêm. Vui lòng dán cURL ở trên để bắt đầu!
            </div>
        `;
        return;
    }
    
    listContainer.innerHTML = "";
    
    accounts.forEach((acc, idx) => {
        const item = document.createElement("div");
        const isActiveRotation = (idx === currentIndex % accounts.length);
        item.className = "suno-account-item" + (isActiveRotation ? " active-rotation" : "");
        
        const isExpired = acc.expiry && (Date.now() > acc.expiry);
        const statusClass = isExpired ? "account-status-expired" : "account-status-active";
        const statusText = isExpired ? "Hết hạn" : "Đang chạy";
        
        let expiryText = "Không xác định";
        if (acc.expiry) {
            expiryText = new Date(acc.expiry).toLocaleString('vi-VN', {
                month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit'
            });
        }
        
        item.innerHTML = `
            <div class="suno-account-details">
                <div class="suno-account-email">
                    <i data-lucide="mail" style="width: 14px; height: 14px; color: var(--text-secondary);"></i>
                    <span>${escapeHTML(acc.email)}</span>
                    ${isActiveRotation ? `<span class="suno-account-badge">Active</span>` : ''}
                </div>
                <div class="suno-account-expiry">Hạn: ${expiryText} | <span class="suno-account-status ${statusClass}">${statusText}</span></div>
            </div>
            <div class="suno-account-actions" style="display: flex; gap: 6px;">
                <button type="button" class="suno-account-btn-refresh suno-clip-btn" data-id="${acc.id}" style="padding: 4px 8px; border-color: rgba(59, 130, 246, 0.2); color: var(--color-primary); background: rgba(59, 130, 246, 0.05);">
                    <i data-lucide="refresh-cw" style="width: 12px; height: 12px;"></i> Làm mới
                </button>
                <button type="button" class="suno-account-btn-delete suno-clip-btn" data-id="${acc.id}" style="padding: 4px 8px; border-color: rgba(239, 68, 68, 0.2); color: var(--color-danger); background: rgba(239, 68, 68, 0.05);">
                    <i data-lucide="trash-2" style="width: 12px; height: 12px;"></i> Xóa
                </button>
            </div>
        `;
        
        listContainer.appendChild(item);
    });
    
    listContainer.querySelectorAll(".suno-account-btn-delete").forEach(btn => {
        btn.addEventListener("click", (e) => {
            e.stopPropagation();
            const id = btn.getAttribute("data-id");
            deleteSunoAccount(id);
        });
    });

    listContainer.querySelectorAll(".suno-account-btn-refresh").forEach(btn => {
        btn.addEventListener("click", async (e) => {
            e.stopPropagation();
            const id = btn.getAttribute("data-id");
            await forceRefreshAccount(id);
        });
    });
    
    lucide.createIcons();
}

// Gọi API thủ công để làm mới session bằng cookie
async function forceRefreshAccount(id) {
    const accounts = getSunoAccounts();
    const acc = accounts.find(a => a.id === id);
    if (!acc) return;
    
    if (!acc.cookie) {
        alert("Tài khoản này không lưu Cookie để làm mới! Vui lòng thêm lại bằng cURL.");
        return;
    }
    
    const btn = document.querySelector(`.suno-account-btn-refresh[data-id="${id}"]`);
    let originalHtml = "";
    if (btn) {
        originalHtml = btn.innerHTML;
        btn.innerHTML = `<i data-lucide="loader" class="spin" style="width: 12px; height: 12px;"></i> Đang chạy...`;
        lucide.createIcons();
        btn.disabled = true;
    }
    
    try {
        const response = await fetch("/api/suno/refresh", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                cookie: acc.cookie
            })
        });
        
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || `Lỗi ${response.status}`);
        }
        
        if (data.new_auth_token || data.new_cookie) {
            updateStoredSunoToken(acc.email, data.new_auth_token, data.new_cookie);
            if (data.new_auth_token) acc.authToken = data.new_auth_token;
            if (data.new_cookie) acc.cookie = data.new_cookie;
            alert(`Làm mới session thành công cho tài khoản ${acc.email}!`);
        } else {
            throw new Error("Không nhận được token mới từ server.");
        }
    } catch (err) {
        console.error("Lỗi gia hạn session:", err);
        alert(`Lỗi gia hạn session: ${err.message}`);
    } finally {
        if (btn) {
            btn.innerHTML = originalHtml;
            lucide.createIcons();
            btn.disabled = false;
        }
    }
}

// Cập nhật dropdown chọn tài khoản tạo nhạc
function updateSunoAccountSelector() {
    const select = document.getElementById("sunoAccountSelect");
    if (!select) return;
    
    const accounts = getSunoAccounts();
    const selectedVal = select.value || "random";
    
    select.innerHTML = '<option value="random">Xoay vòng ngẫu nhiên tài khoản</option>';
    
    accounts.forEach(acc => {
        const isExpired = acc.expiry && (Date.now() > acc.expiry);
        const statusText = isExpired ? " (Hết hạn)" : "";
        const opt = document.createElement("option");
        opt.value = acc.id;
        opt.textContent = `${acc.email}${statusText}`;
        if (isExpired) {
            opt.style.color = "var(--color-danger)";
        }
        select.appendChild(opt);
    });
    
    const exists = accounts.some(acc => acc.id === selectedVal);
    if (exists || selectedVal === "random") {
        select.value = selectedVal;
    } else {
        select.value = "random";
    }
}

// Cập nhật giao diện cảnh báo cấu hình Suno
function updateSunoUIState() {
    const accounts = getSunoAccounts();
    const warning = document.getElementById("sunoConfigWarning");
    const controls = document.getElementById("sunoGenControls");
    
    updateSunoAccountSelector();
    
    if (accounts.length > 0) {
        if (warning) warning.style.display = "none";
        if (controls) controls.style.display = "block";
    } else {
        if (warning) warning.style.display = "block";
        if (controls) controls.style.display = "none";
    }
}

// Sáng tác nhạc trên Suno (Xoay vòng và Thử lại tự động)
async function generateSunoMusic() {
    const accounts = getSunoAccounts();
    if (accounts.length === 0) {
        alert("Vui lòng cấu hình ít nhất 1 tài khoản Suno trong phần Cấu hình trước!");
        return;
    }
    
    const select = document.getElementById("sunoAccountSelect");
    const selectedVal = select ? select.value : "random";
    
    let startIndex = 0;
    if (selectedVal === "random") {
        startIndex = Math.floor(Math.random() * accounts.length);
    } else {
        const foundIdx = accounts.findIndex(acc => acc.id === selectedVal);
        if (foundIdx > -1) {
            startIndex = foundIdx;
        } else {
            startIndex = Math.floor(Math.random() * accounts.length);
        }
    }
    
    const controls = document.getElementById("sunoGenControls");
    const loading = document.getElementById("sunoGenLoading");
    const loadingStatus = document.getElementById("sunoLoadingStatus");
    const loadingSubtext = document.getElementById("sunoLoadingSubtext");
    
    controls.style.display = "none";
    loading.style.display = "flex";
    
    async function tryGenerate(attemptCount, currentIdx) {
        const idx = currentIdx % accounts.length;
        const account = accounts[idx];
        
        // Cập nhật index đang chạy sang tài khoản vừa chọn
        localStorage.setItem("suno_current_account_index", idx);
        
        loadingStatus.textContent = `Đang gửi yêu cầu tạo nhạc lên Suno...`;
        loadingSubtext.textContent = `Sử dụng tài khoản: ${account.email} (${attemptCount + 1}/${accounts.length})`;
        
        try {
            const modelVersion = localStorage.getItem("suno_model_version") || "chirp-fenix";
            const chkInstrumental = document.getElementById("chkSunoInstrumental");
            const makeInstrumental = chkInstrumental ? chkInstrumental.checked : false;
            
            const response = await fetch("/api/suno/generate", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    auth_token: account.authToken,
                    browser_token: account.browserToken,
                    device_id: account.deviceId,
                    suno_token: account.bodyToken || "",
                    user_tier: account.userTier || "",
                    create_session_token: account.createSessionToken || "",
                    song_id: currentSongData.id || "",
                    prompt: currentSongData.lyrics || "",
                    tags: currentSongData.style || "",
                    title: currentSongData.title || "Bài nhạc AI",
                    model_version: modelVersion,
                    make_instrumental: makeInstrumental,
                    account_email: account.email,
                    cookie: account.cookie || ""
                })
            });
            
            const data = await response.json();
            
            if (data.new_auth_token || data.new_cookie) {
                updateStoredSunoToken(account.email, data.new_auth_token, data.new_cookie);
                if (data.new_auth_token) account.authToken = data.new_auth_token;
                if (data.new_cookie) account.cookie = data.new_cookie;
            }
            
            if (!response.ok) {
                throw new Error(data.error || `Lỗi phản hồi ${response.status}`);
            }
            
            let clips = [];
            if (Array.isArray(data)) {
                clips = data;
            } else if (data && Array.isArray(data.clips)) {
                clips = data.clips;
            }
            
            if (clips.length === 0) {
                throw new Error("Không nhận được clips nhạc nào từ Suno.");
            }
            
            const clipIds = clips.map(c => c.id).filter(id => !!id);
            
            const formatted = clips.map(c => ({
                id: c.id,
                title: c.title,
                audioUrl: c.audio_url || c.audioUrl || "",
                videoUrl: c.video_url || c.videoUrl || "",
                imageUrl: c.image_url || c.imageUrl || "",
                status: c.status || "queued",
                prompt: c.metadata ? c.metadata.prompt : (c.prompt || ""),
                createdAt: new Date(),
                accountEmail: account.email,
                driveUrl: c.driveUrl || c.drive_url || ""
            }));
            
            // Trộn vào danh sách hiện tại thay vì ghi đè hoàn toàn
            const merged = [...(currentSongData.sunoClips || [])];
            formatted.forEach(nc => {
                const idx = merged.findIndex(x => x.id === nc.id);
                if (idx > -1) {
                    if (!nc.driveUrl && merged[idx].driveUrl) {
                        nc.driveUrl = merged[idx].driveUrl;
                    }
                    merged[idx] = nc;
                } else {
                    merged.push(nc);
                }
            });
            
            currentSongData.sunoClips = merged;
            renderSunoClips(merged);
            
            loadingStatus.textContent = "Đang xếp hàng tạo nhạc trên Suno...";
            loadingSubtext.textContent = "Nhạc sĩ AI đang kết xuất giai điệu. Có hai phiên bản đang được xử lý.";
            
            pollSunoStatus(clipIds, currentSongData.id, account);
            
        } catch (err) {
            console.warn(`Lỗi tạo nhạc với tài khoản ${account.email}:`, err.message);
            
            if (attemptCount + 1 < accounts.length) {
                await tryGenerate(attemptCount + 1, idx + 1);
            } else {
                alert(`Tất cả tài khoản Suno đều không thể tạo nhạc. Lỗi cuối cùng: ${err.message}`);
                controls.style.display = "block";
                loading.style.display = "none";
            }
        }
    }
    
    await tryGenerate(0, startIndex);
}

// Vòng lặp kiểm tra tiến độ của Suno bài hát (Polling)
function pollSunoStatus(clipIds, songId, account) {
    if (!songId) return; // Không có ID bài hát thì không thể tracking hoặc lưu trữ
    
    if (sunoPollIntervals[songId]) {
        clearInterval(sunoPollIntervals[songId]);
    }
    
    const controls = document.getElementById("sunoGenControls");
    const loading = document.getElementById("sunoGenLoading");
    const loadingStatus = document.getElementById("sunoLoadingStatus");
    const loadingSubtext = document.getElementById("sunoLoadingSubtext");
    
    let dots = 0;
    
    sunoPollIntervals[songId] = setInterval(async () => {
        dots = (dots + 1) % 4;
        const dotStr = ".".repeat(dots);
        
        try {
            const currentAccount = getSunoAccounts().find(a => a.email === account.email) || account;
            const response = await fetch("/api/suno/feed", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json"
                },
                body: JSON.stringify({
                    auth_token: currentAccount.authToken,
                    browser_token: currentAccount.browserToken,
                    device_id: currentAccount.deviceId,
                    clip_ids: clipIds,
                    song_id: songId || "",
                    account_email: currentAccount.email,
                    cookie: currentAccount.cookie || ""
                })
            });
            
            if (!response.ok) return; // Tiếp tục thử lại ở lần sau
            
            const data = await response.json();
            
            if (data.new_auth_token || data.new_cookie) {
                updateStoredSunoToken(account.email, data.new_auth_token, data.new_cookie);
                if (data.new_auth_token) account.authToken = data.new_auth_token;
                if (data.new_cookie) account.cookie = data.new_cookie;
            }
            
            let clips = [];
            if (Array.isArray(data)) {
                clips = data;
            } else if (data && Array.isArray(data.clips)) {
                clips = data.clips;
            }
            
            if (clips.length === 0) return;
            
            // Format clips cho Client
            const formattedClips = clips.map(c => {
                let existing = null;
                if (songId === currentSongData.id) {
                    existing = currentSongData.sunoClips.find(x => x.id === c.id);
                }
                return {
                    id: c.id,
                    title: c.title,
                    audioUrl: c.audio_url || c.audioUrl || "",
                    videoUrl: c.video_url || c.videoUrl || "",
                    imageUrl: c.image_url || c.imageUrl || "",
                    status: c.status || "queued",
                    prompt: c.metadata ? c.metadata.prompt : (c.prompt || ""),
                    createdAt: c.created_at ? new Date(c.created_at) : (existing ? existing.createdAt : new Date()),
                    accountEmail: account.email,
                    driveUrl: (existing && existing.driveUrl) || c.driveUrl || c.drive_url || ""
                };
            });
            
            // Chỉ cập nhật currentSongData và UI nếu bài hát đang xem trùng với bài hát đang poll
            if (songId === currentSongData.id) {
                const merged = [...(currentSongData.sunoClips || [])];
                formattedClips.forEach(nc => {
                    const idx = merged.findIndex(x => x.id === nc.id);
                    if (idx > -1) {
                        // Không đè driveUrl nếu bên nc trống nhưng merged có
                        if (!nc.driveUrl && merged[idx].driveUrl) {
                            nc.driveUrl = merged[idx].driveUrl;
                        }
                        merged[idx] = nc;
                    } else {
                        merged.push(nc);
                    }
                });
                
                currentSongData.sunoClips = merged;
                renderSunoClips(merged);
            }
            
            // Kiểm tra trạng thái hoàn tất
            const allFinished = formattedClips.every(c => c.status === "complete" || c.status === "failed");
            const anyStreaming = formattedClips.some(c => c.status === "streaming");
            
            if (songId === currentSongData.id) {
                if (anyStreaming) {
                    loadingStatus.textContent = "Đang kết xuất giai điệu" + dotStr;
                    loadingSubtext.textContent = "Nhạc sĩ AI đang kết xuất nhạc. Vui lòng chờ hoàn thành để nghe thử.";
                } else {
                    loadingStatus.textContent = "Đang xử lý giai điệu" + dotStr;
                }
            }
            
            if (allFinished) {
                clearInterval(sunoPollIntervals[songId]);
                delete sunoPollIntervals[songId];
                
                if (songId === currentSongData.id) {
                    loading.style.display = "none";
                    controls.style.display = "block";
                    
                    // Cập nhật lại list lịch sử nếu đang xem
                    updateHistoryCount();
                }
            }
            
        } catch (e) {
            console.error("Lỗi polling Suno status cho songId " + songId + ":", e);
        }
    }, 5000); // Polling mỗi 5 giây
}

let lastRenderedSongId = null;

// Hiển thị danh sách các clips nhạc Suno đã tạo
function renderSunoClips(clips) {
    const wrapper = document.getElementById("sunoGeneratedClips");
    const container = document.getElementById("sunoClipsList");
    
    if (!clips || clips.length === 0) {
        wrapper.style.display = "none";
        container.innerHTML = "";
        lastRenderedSongId = currentSongData.id;
        return;
    }
    
    wrapper.style.display = "block";
    
    // Nếu chuyển sang bài hát khác, xóa sạch list cũ để vẽ lại từ đầu
    if (lastRenderedSongId !== currentSongData.id) {
        container.innerHTML = "";
        lastRenderedSongId = currentSongData.id;
    }
    
    // Xóa bớt clip không còn trong danh sách mới (nếu có)
    const clipIdsInList = new Set(clips.map(c => c.id));
    Array.from(container.children).forEach(child => {
        const idAttr = child.id || "";
        if (idAttr.startsWith("suno-clip-")) {
            const clipId = idAttr.replace("suno-clip-", "");
            if (!clipIdsInList.has(clipId)) {
                child.remove();
            }
        }
    });
    
    clips.forEach(clip => {
        let item = document.getElementById(`suno-clip-${clip.id}`);
        const isNew = !item;
        
        if (isNew) {
            item = document.createElement("div");
            item.id = `suno-clip-${clip.id}`;
            item.className = "suno-clip-item";
            item.innerHTML = `
                <div class="suno-clip-cover-wrapper"></div>
                <div class="suno-clip-info">
                    <div class="suno-clip-title"></div>
                    <div class="suno-clip-account" style="font-size: 0.72rem; color: var(--text-muted); display: none; align-items: center; gap: 4px; margin-top: 2px;"></div>
                    <div class="suno-clip-status" style="margin-top: 4px;"></div>
                    
                    <div class="suno-clip-player-wrapper" style="display: none; margin-top: 8px;">
                        <audio controls class="suno-clip-audio" style="width: 100%;"></audio>
                        <div class="suno-clip-actions" id="actions-${clip.id}" style="display: none;"></div>
                        <div class="suno-clip-progress-wrapper" id="progress-wrapper-${clip.id}" style="display: none;">
                            <div class="suno-clip-progress-container">
                                <div class="suno-clip-progress-header">
                                    <span class="suno-clip-progress-text" id="progress-text-${clip.id}">
                                        <i data-lucide="loader-2" class="spin" style="width: 12px; height: 12px;"></i>
                                        Đang tải file từ Suno...
                                    </span>
                                    <span class="suno-clip-progress-pct" id="progress-pct-${clip.id}">0%</span>
                                </div>
                                <div class="suno-clip-progress-bar-bg">
                                    <div class="suno-clip-progress-bar-fill" id="progress-fill-${clip.id}"></div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            `;
            container.appendChild(item);
        }
        
        // 1. Cập nhật ảnh bìa (Cover Image)
        const coverWrapper = item.querySelector(".suno-clip-cover-wrapper");
        if (clip.imageUrl) {
            const currentImg = coverWrapper.querySelector("img");
            if (!currentImg || currentImg.src !== clip.imageUrl) {
                coverWrapper.innerHTML = `<img src="${clip.imageUrl}" class="suno-clip-cover" alt="Art">`;
            }
        } else {
            const currentIcon = coverWrapper.querySelector("i");
            if (!currentIcon) {
                coverWrapper.innerHTML = `<i data-lucide="music" style="width: 24px; height: 24px; color: var(--text-muted);"></i>`;
                if (window.lucide) window.lucide.createIcons({ node: coverWrapper });
            }
        }
        
        // 2. Cập nhật tiêu đề (Title)
        const titleEl = item.querySelector(".suno-clip-title");
        const formattedTitle = clip.title || 'Bài nhạc Suno';
        if (titleEl.textContent !== formattedTitle) {
            titleEl.textContent = formattedTitle;
            titleEl.title = formattedTitle;
        }
        
        // 3. Cập nhật email tài khoản (Account email)
        const accountEl = item.querySelector(".suno-clip-account");
        if (clip.accountEmail) {
            accountEl.style.display = "flex";
            const expectedHTML = `<i data-lucide="user" style="width: 11px; height: 11px;"></i> ${escapeHTML(clip.accountEmail)}`;
            if (accountEl.innerHTML !== expectedHTML) {
                accountEl.innerHTML = expectedHTML;
                if (window.lucide) window.lucide.createIcons({ node: accountEl });
            }
        } else {
            accountEl.style.display = "none";
        }
        
        // 4. Cập nhật trạng thái hiển thị (Status text và class)
        let statusText = "Đang chờ";
        if (clip.status === "complete") statusText = "Hoàn thành";
        if (clip.status === "streaming") statusText = "Đang phát Stream...";
        if (clip.status === "failed") statusText = "Thất bại";
        
        const statusEl = item.querySelector(".suno-clip-status");
        if (statusEl.textContent !== statusText) {
            statusEl.textContent = statusText;
        }
        const expectedClass = `suno-clip-status status-${clip.status || 'queued'}`;
        if (statusEl.className !== expectedClass) {
            statusEl.className = expectedClass;
        }
        
        // 5. Cập nhật thẻ audio (chỉ hỗ trợ complete và streaming)
        const playerWrapper = item.querySelector(".suno-clip-player-wrapper");
        const audioEl = item.querySelector(".suno-clip-audio");
        const canPlay = (clip.status === "complete" || clip.status === "streaming") && clip.audioUrl;
        
        if (canPlay) {
            playerWrapper.style.display = "block";
            // So sánh URL tuyệt đối để tránh việc gán lại làm ngắt nhạc đang phát
            const absoluteAudioUrl = new URL(clip.audioUrl, window.location.href).href;
            if (audioEl.src !== absoluteAudioUrl) {
                audioEl.src = clip.audioUrl;
            }
        } else {
            playerWrapper.style.display = "none";
            if (audioEl.src) {
                audioEl.src = "";
            }
        }
        
        // 6. Cập nhật các nút thao tác (chỉ khi clip hoàn thành)
        const actionsEl = item.querySelector(".suno-clip-actions");
        if (clip.status === "complete" && clip.audioUrl) {
            actionsEl.style.display = "flex";
            
            const downloadUrl = `/api/suno/download?url=${encodeURIComponent(clip.audioUrl)}&name=${encodeURIComponent((currentSongData.title || clip.title || 'music').trim())}`;
            const videoHTML = clip.videoUrl ? `
                <a href="${clip.videoUrl}" target="_blank" class="suno-clip-btn">
                    <i data-lucide="external-link"></i> Xem Video
                </a>
            ` : '';
            const driveHTML = clip.driveUrl ? `
                <a href="${clip.driveUrl}" target="_blank" class="suno-clip-btn suno-clip-btn-drive">
                    <i data-lucide="hard-drive"></i> Mở Drive
                </a>
            ` : `
                <button type="button" class="suno-clip-btn" onclick="publishToDrive('${clip.id}', '${escapeHTML(clip.audioUrl)}', this)">
                    <i data-lucide="cloud-upload"></i> Đăng lên Drive
                </button>
            `;
            
            const expectedActionsHTML = `
                <a href="${downloadUrl}" target="_blank" class="suno-clip-btn suno-clip-btn-primary">
                    <i data-lucide="download"></i> Tải MP3
                </a>
                ${videoHTML}
                ${driveHTML}
            `;
            
            // Check nếu chưa được render hoặc driveUrl thay đổi thì mới render lại nút bấm
            const currentDriveBtn = actionsEl.querySelector(".suno-clip-btn-drive");
            const hasDriveUrl = !!clip.driveUrl;
            const hadDriveUrl = !!currentDriveBtn;
            
            if (actionsEl.innerHTML === "" || hasDriveUrl !== hadDriveUrl) {
                actionsEl.innerHTML = expectedActionsHTML;
                if (window.lucide) window.lucide.createIcons({ node: actionsEl });
            }
        } else {
            actionsEl.style.display = "none";
        }
    });
}

// Kiểm tra trạng thái đăng nhập hệ thống
async function checkAuthStatus() {
    try {
        const res = await fetch("/api/auth-status");
        if (!res.ok) return;
        const data = await res.json();
        
        const logoutBtn = document.getElementById("headerLogoutBtn");
        if (data.auth_enabled && logoutBtn) {
            logoutBtn.style.display = "flex";
            logoutBtn.addEventListener("click", async () => {
                if (confirm("Bạn có chắc chắn muốn đăng xuất không?")) {
                    try {
                        const logoutRes = await fetch("/api/logout", { method: "POST" });
                        if (logoutRes.ok) {
                            window.location.reload();
                        }
                    } catch (e) {
                        console.error("Lỗi đăng xuất:", e);
                    }
                }
            });
        }
    } catch (e) {
        console.error("Lỗi kiểm tra cấu hình auth:", e);
    }
}

// Đăng tải nhạc lên Google Drive qua rclone bằng SSE
function publishToDrive(clipId, audioUrl, buttonEl) {
    if (!currentSongData || !currentSongData.id) {
        alert("Lỗi: Không xác định được thông tin tác phẩm hiện tại.");
        return;
    }

    const actionsContainer = document.getElementById(`actions-${clipId}`);
    const progressWrapper = document.getElementById(`progress-wrapper-${clipId}`);
    const progressText = document.getElementById(`progress-text-${clipId}`);
    const progressPct = document.getElementById(`progress-pct-${clipId}`);
    const progressFill = document.getElementById(`progress-fill-${clipId}`);

    if (!actionsContainer || !progressWrapper) return;

    // Ẩn thanh hành động, hiện thanh tiến trình
    actionsContainer.style.display = "none";
    progressWrapper.style.display = "block";

    const updateUI = (text, percent) => {
        if (progressText) progressText.innerHTML = `<i data-lucide="loader-2" class="spin" style="width: 12px; height: 12px;"></i> ${text}`;
        if (progressPct) progressPct.textContent = `${percent}%`;
        if (progressFill) progressFill.style.width = `${percent}%`;
        lucide.createIcons();
    };

    updateUI("Bắt đầu kết nối server...", 0);

    const queryParams = new URLSearchParams({
        song_id: currentSongData.id,
        clip_id: clipId,
        audio_url: audioUrl
    });

    const eventSource = new EventSource(`/api/suno/publish-drive?${queryParams.toString()}`);

    eventSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            if (data.status === "downloading") {
                updateUI(data.message || "Đang tải file từ Suno...", data.progress);
            } else if (data.status === "uploading") {
                updateUI(data.message || "Đang chuyển file lên Google Drive...", data.progress);
            } else if (data.status === "finalizing") {
                updateUI(data.message || "Đang thiết lập liên kết...", data.progress);
            } else if (data.status === "success") {
                eventSource.close();
                const driveUrl = data.drive_url;
                
                // Cập nhật dữ liệu client
                const clip = currentSongData.sunoClips.find(c => c.id === clipId);
                if (clip) {
                    clip.driveUrl = driveUrl;
                }

                // Render lại danh sách clip
                renderSunoClips(currentSongData.sunoClips);
                alert("Đã chuyển nhạc lên Google Drive thành công!");
            } else if (data.status === "error") {
                eventSource.close();
                alert(`Lỗi upload: ${data.message}`);
                // Khôi phục nút bấm
                progressWrapper.style.display = "none";
                actionsContainer.style.display = "flex";
            }
        } catch (e) {
            console.error("Lỗi parse SSE event:", e);
        }
    };

    eventSource.onerror = function(err) {
        console.error("Lỗi kết nối SSE:", err);
        eventSource.close();
        alert("Lỗi kết nối Server-Sent Events khi đang truyền dữ liệu.");
        // Khôi phục nút bấm
        progressWrapper.style.display = "none";
        actionsContainer.style.display = "flex";
    };
}

// --- Các hàm hỗ trợ điều hướng trên thiết bị di động ---
function initMobileNavigation() {
    const mobileNavItems = document.querySelectorAll(".mobile-nav-item");
    const appContent = document.querySelector(".app-content");
    
    if (!mobileNavItems.length || !appContent) return;
    
    mobileNavItems.forEach(item => {
        item.addEventListener("click", () => {
            const target = item.getAttribute("data-mobile-tab");
            
            if (target === "config") {
                appContent.classList.remove("mobile-show-result", "mobile-show-history");
                appContent.classList.add("mobile-show-config");
                updateMobileTabActive("config");
            } else if (target === "result") {
                const currentSongTabBtn = document.querySelector('.tab-btn[data-tab="current-song"]');
                if (currentSongTabBtn) {
                    currentSongTabBtn.click();
                } else {
                    appContent.classList.remove("mobile-show-config", "mobile-show-history");
                    appContent.classList.add("mobile-show-result");
                    updateMobileTabActive("result");
                }
            } else if (target === "history") {
                const historyTabBtn = document.querySelector('.tab-btn[data-tab="history-library"]');
                if (historyTabBtn) {
                    historyTabBtn.click();
                } else {
                    appContent.classList.remove("mobile-show-config", "mobile-show-result");
                    appContent.classList.add("mobile-show-history");
                    updateMobileTabActive("history");
                }
            }
        });
    });

    // Đồng bộ số lượng lịch sử vào tab di động
    const historyCountObserver = new MutationObserver(() => {
        syncHistoryCountToMobile();
    });
    
    const historyCountEl = document.getElementById("historyCount");
    if (historyCountEl) {
        historyCountObserver.observe(historyCountEl, { childList: true, characterData: true, subtree: true });
        syncHistoryCountToMobile();
    }
}

function syncHistoryCountToMobile() {
    const historyCountEl = document.getElementById("historyCount");
    const mobileNavItems = document.querySelectorAll(".mobile-nav-item[data-mobile-tab='history'] span");
    if (historyCountEl && mobileNavItems.length) {
        mobileNavItems[0].textContent = `Lịch sử (${historyCountEl.textContent})`;
    }
}

function updateMobileTabActive(activeTab) {
    const mobileNavItems = document.querySelectorAll(".mobile-nav-item");
    mobileNavItems.forEach(item => {
        if (item.getAttribute("data-mobile-tab") === activeTab) {
            item.classList.add("active");
        } else {
            item.classList.remove("active");
        }
    });
}
