package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeforge/backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Store struct {
	client           *minio.Client
	bucket, localDir string
}

func New(cfg config.Config) *Store {
	_ = os.MkdirAll(cfg.UploadDir, 0755)
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{Creds: credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""), Secure: cfg.MinIOUseSSL})
	s := &Store{bucket: cfg.MinIOBucket, localDir: cfg.UploadDir}
	if err != nil {
		return s
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err == nil && !exists {
		err = client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{})
	}
	if err == nil {
		s.client = client
	}
	return s
}
func (s *Store) Save(ctx context.Context, prefix string, h *multipart.FileHeader) (string, error) {
	clean := strings.ReplaceAll(filepath.Base(h.Filename), " ", "_")
	key := fmt.Sprintf("%s/%d_%s", strings.Trim(prefix, "/"), time.Now().UnixNano(), clean)
	f, err := h.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	if s.client != nil {
		_, err = s.client.PutObject(ctx, s.bucket, key, f, h.Size, minio.PutObjectOptions{ContentType: h.Header.Get("Content-Type")})
		if err == nil {
			return key, nil
		}
		if _, seekErr := f.Seek(0, io.SeekStart); seekErr != nil {
			return "", err
		}
	}
	path := filepath.Join(s.localDir, filepath.FromSlash(key))
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, f)
	return key, err
}
func (s *Store) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if s.client != nil {
		return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	}
	return os.Open(filepath.Join(s.localDir, filepath.FromSlash(key)))
}

func (s *Store) ImageDataURL(ctx context.Context, key string, maxBytes int64) (string, error) {
	reader, err := s.Open(ctx, key)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("image exceeds %d bytes", maxBytes)
	}
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(key)))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("attachment is not an image: %s", contentType)
	}
	optimized, optimizedType := optimizeForVision(data, contentType)
	return "data:" + optimizedType + ";base64," + base64.StdEncoding.EncodeToString(optimized), nil
}

func optimizeForVision(data []byte, contentType string) ([]byte, string) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, contentType
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	const maxEdge = 1280
	scale := 1.0
	if width > maxEdge || height > maxEdge {
		if width > height {
			scale = float64(maxEdge) / float64(width)
		} else {
			scale = float64(maxEdge) / float64(height)
		}
	}
	newW, newH := maxInt(1, int(float64(width)*scale)), maxInt(1, int(float64(height)*scale))
	if scale == 1 && len(data) < 600*1024 {
		return data, contentType
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		sy := bounds.Min.Y + y*height/newH
		for x := 0; x < newW; x++ {
			sx := bounds.Min.X + x*width/newW
			c := color.NRGBAModel.Convert(img.At(sx, sy)).(color.NRGBA)
			a := uint16(c.A)
			dst.SetRGBA(x, y, color.RGBA{R: uint8((uint16(c.R)*a + 255*(255-a)) / 255), G: uint8((uint16(c.G)*a + 255*(255-a)) / 255), B: uint8((uint16(c.B)*a + 255*(255-a)) / 255), A: 255})
		}
	}
	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 82}); err != nil {
		return data, contentType
	}
	return out.Bytes(), "image/jpeg"
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
