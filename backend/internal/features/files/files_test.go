package files

import (
	"archive/zip"
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// --- образцы файлов ---

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func zipWith(t *testing.T, name string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create(name)
	require.NoError(t, err)
	_, _ = f.Write([]byte("<xml/>"))
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

var (
	jpegHead = append([]byte{0xFF, 0xD8, 0xFF, 0xE0}, make([]byte, 16)...)
	pdfHead  = []byte("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n")
	oleHead  = append([]byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}, make([]byte, 16)...)
	exeHead  = append([]byte("MZ\x90\x00"), make([]byte, 16)...)
)

// --- распознавание типа ---

func TestDetect_Allowed(t *testing.T) {
	cases := []struct {
		name, file string
		data       []byte
		mime       string
	}{
		{"jpeg", "photo.jpg", jpegHead, MimeJPEG},
		{"jpeg без расширения", "IMG_0001", jpegHead, MimeJPEG},
		{"png", "scan.png", pngBytes(t, 4, 3), MimePNG},
		{"pdf", "work.pdf", pdfHead, MimePDF},
		{"docx", "essay.docx", zipWith(t, "word/document.xml"), MimeDOCX},
		{"doc", "old.doc", oleHead, MimeDOC},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d, ok := detect(c.file, c.data)
			require.True(t, ok)
			assert.Equal(t, c.mime, d.mime)
		})
	}
}

func TestDetect_Rejected(t *testing.T) {
	cases := []struct {
		name, file string
		data       []byte
	}{
		{"exe, переименованный в jpg", "photo.jpg", exeHead},
		{"обычный zip под видом docx", "essay.docx", zipWith(t, "readme.txt")},
		{"старый Excel (.xls) — тоже OLE2", "table.xls", oleHead},
		{"текст", "notes.txt", []byte("hello")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, ok := detect(c.file, c.data)
			assert.False(t, ok)
		})
	}
}

func TestDetect_ImageSize(t *testing.T) {
	d, ok := detect("a.png", pngBytes(t, 40, 30))
	require.True(t, ok)
	require.NotNil(t, d.width)
	assert.Equal(t, 40, *d.width)
	assert.Equal(t, 30, *d.height)
}

// --- сервис ---

type memRepo struct{ files map[uuid.UUID]*domain.File }

func (r *memRepo) Create(_ context.Context, f *domain.File) error { cp := *f; r.files[f.ID] = &cp; return nil }
func (r *memRepo) Delete(_ context.Context, id uuid.UUID) error  { delete(r.files, id); return nil }
func (r *memRepo) ListStalePending(_ context.Context, before time.Time, limit int) ([]*domain.File, error) {
	var out []*domain.File
	for _, f := range r.files {
		if f.Status == domain.FileStatusPending && f.CreatedAt.Before(before) && len(out) < limit {
			cp := *f
			out = append(out, &cp)
		}
	}
	return out, nil
}

type memStore struct{ objects map[string][]byte }

func (s *memStore) PutObject(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	b, _ := io.ReadAll(r)
	s.objects[key] = b
	return nil
}
func (s *memStore) RemoveObject(_ context.Context, key string) error { delete(s.objects, key); return nil }

type noopSigner struct{}

func (noopSigner) PresignedGetURL(_ context.Context, key string) (string, error) { return key, nil }

func newSvc() (*Service, *memRepo, *memStore) {
	repo := &memRepo{files: map[uuid.UUID]*domain.File{}}
	store := &memStore{objects: map[string][]byte{}}
	return NewService(repo, store, noopSigner{}), repo, store
}

func TestUpload_StoresPendingWithMetadata(t *testing.T) {
	svc, repo, store := newSvc()
	owner := uuid.New()
	data := pngBytes(t, 8, 6)

	f, err := svc.Upload(context.Background(), owner, domain.FilePurposeSubmission, `C:\Users\kid\Desktop\tetrad.png`, bytes.NewReader(data))
	require.NoError(t, err)

	assert.Equal(t, domain.FileStatusPending, f.Status)
	assert.Equal(t, MimePNG, f.MimeType)
	assert.Equal(t, "tetrad.png", f.OriginalName, "путь клиента отрезан")
	assert.Equal(t, int64(len(data)), f.SizeBytes)
	require.NotNil(t, f.SHA256)
	assert.Len(t, *f.SHA256, 64)
	assert.True(t, strings.HasPrefix(f.ObjectKey, "files/"+owner.String()+"/"))
	assert.True(t, strings.HasSuffix(f.ObjectKey, ".png"), "расширение — по содержимому")
	assert.Contains(t, repo.files, f.ID)
	assert.Equal(t, data, store.objects[f.ObjectKey])
}

func TestUpload_TooLarge(t *testing.T) {
	svc, _, store := newSvc()
	big := append(append([]byte{}, jpegHead...), make([]byte, MaxFileSize)...)
	_, err := svc.Upload(context.Background(), uuid.New(), domain.FilePurposeSubmission, "big.jpg", bytes.NewReader(big))
	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Empty(t, store.objects, "в хранилище ничего не попало")
}

func TestUpload_TypeNotAllowed(t *testing.T) {
	svc, _, store := newSvc()
	_, err := svc.Upload(context.Background(), uuid.New(), domain.FilePurposeSubmission, "photo.jpg", bytes.NewReader(exeHead))
	assert.ErrorIs(t, err, domain.ErrValidation)
	assert.Empty(t, store.objects)
}

func TestCleanupPending_RemovesOnlyStaleUnattached(t *testing.T) {
	svc, repo, store := newSvc()
	now := time.Now()
	mk := func(status domain.FileStatus, age time.Duration) uuid.UUID {
		id := uuid.New()
		key := "files/x/" + id.String()
		repo.files[id] = &domain.File{ID: id, ObjectKey: key, Status: status, CreatedAt: now.Add(-age)}
		store.objects[key] = []byte("x")
		return id
	}
	stale := mk(domain.FileStatusPending, 25*time.Hour)
	fresh := mk(domain.FileStatusPending, time.Hour)
	attached := mk(domain.FileStatusAttached, 48*time.Hour)

	n, err := svc.CleanupPending(context.Background(), now)
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.NotContains(t, repo.files, stale)
	assert.Contains(t, repo.files, fresh)
	assert.Contains(t, repo.files, attached)
	assert.Len(t, store.objects, 2)
}
