package files

import (
	"archive/zip"
	"bytes"
	"image"
	_ "image/jpeg" // регистрирует декодер заголовка JPEG для image.DecodeConfig
	_ "image/png"  // то же для PNG
	"path/filepath"
	"strings"
)

// Разрешённые типы (DECISIONS.md, 2026-10-07): фото JPG/PNG, PDF, Word.
// HEIC с iPhone приложение конвертирует в JPG до загрузки.
const (
	MimeJPEG = "image/jpeg"
	MimePNG  = "image/png"
	MimePDF  = "application/pdf"
	MimeDOCX = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MimeDOC  = "application/msword"
)

var (
	sigJPEG = []byte{0xFF, 0xD8, 0xFF}
	sigPNG  = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	sigPDF  = []byte("%PDF-")
	sigZIP  = []byte("PK\x03\x04")
	sigOLE2 = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
)

// detected — результат распознавания содержимого файла.
type detected struct {
	mime   string
	ext    string
	width  *int
	height *int
}

// detect определяет тип файла по содержимому (сигнатуре), а не по расширению:
// переименованный .exe → .jpg не пройдёт. Возвращает ok=false для неразрешённых типов.
func detect(name string, data []byte) (detected, bool) {
	switch {
	case bytes.HasPrefix(data, sigJPEG):
		return withImageSize(detected{mime: MimeJPEG, ext: ".jpg"}, data), true
	case bytes.HasPrefix(data, sigPNG):
		return withImageSize(detected{mime: MimePNG, ext: ".png"}, data), true
	case bytes.HasPrefix(data, sigPDF):
		return detected{mime: MimePDF, ext: ".pdf"}, true
	case bytes.HasPrefix(data, sigZIP) && isDOCX(data):
		return detected{mime: MimeDOCX, ext: ".docx"}, true
	case bytes.HasPrefix(data, sigOLE2) && strings.EqualFold(filepath.Ext(name), ".doc"):
		// OLE2 — общий контейнер старого Office (.doc/.xls/.ppt). Отличить .doc
		// по содержимому без разбора контейнера нельзя, поэтому требуем и расширение.
		return detected{mime: MimeDOC, ext: ".doc"}, true
	}
	return detected{}, false
}

// isDOCX: .docx — это ZIP, внутри которого есть word/document.xml.
// Обычный ZIP-архив под видом .docx не пройдёт.
func isDOCX(data []byte) bool {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return false
	}
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			return true
		}
	}
	return false
}

// withImageSize дописывает размеры изображения из его заголовка. Битый
// заголовок не повод отклонять файл — размеры просто останутся пустыми.
func withImageSize(d detected, data []byte) detected {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err == nil {
		d.width, d.height = &cfg.Width, &cfg.Height
	}
	return d
}
