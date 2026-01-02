package scanner

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"  // GIF support
	_ "golang.org/x/image/webp" // WebP support (if available)
)

// ArtworkSize represents a predefined artwork size
type ArtworkSize struct {
	Name   string
	Width  int
	Height int
}

// Predefined artwork sizes
var (
	ArtworkSizeThumbnail = ArtworkSize{Name: "thumbnail", Width: 64, Height: 64}
	ArtworkSizeSmall     = ArtworkSize{Name: "small", Width: 150, Height: 150}
	ArtworkSizeMedium    = ArtworkSize{Name: "medium", Width: 300, Height: 300}
	ArtworkSizeLarge     = ArtworkSize{Name: "large", Width: 600, Height: 600}

	AllArtworkSizes = []ArtworkSize{
		ArtworkSizeThumbnail,
		ArtworkSizeSmall,
		ArtworkSizeMedium,
		ArtworkSizeLarge,
	}
)

// External artwork filenames to look for (in order of preference)
var ExternalArtworkFiles = []string{
	"cover.jpg",
	"cover.jpeg",
	"cover.png",
	"folder.jpg",
	"folder.jpeg",
	"folder.png",
	"album.jpg",
	"album.jpeg",
	"album.png",
	"front.jpg",
	"front.jpeg",
	"front.png",
	"artwork.jpg",
	"artwork.jpeg",
	"artwork.png",
}

// ArtworkInfo contains information about found artwork
type ArtworkInfo struct {
	Data     []byte
	MIMEType string
	Source   string // "embedded" or "external"
	Path     string // For external artwork, the file path
}

// ArtworkProcessor handles artwork extraction and processing
type ArtworkProcessor struct {
	cacheDir string
}

// NewArtworkProcessor creates a new ArtworkProcessor
func NewArtworkProcessor(cacheDir string) *ArtworkProcessor {
	return &ArtworkProcessor{
		cacheDir: cacheDir,
	}
}

// FindArtwork looks for artwork for an audio file
func (p *ArtworkProcessor) FindArtwork(audioPath string) (*ArtworkInfo, error) {
	// First, try to find external artwork in the same directory
	dir := filepath.Dir(audioPath)
	artwork, err := p.findExternalArtwork(dir)
	if err == nil && artwork != nil {
		return artwork, nil
	}

	// Then try to extract embedded artwork
	extractor := NewMetadataExtractor()
	data, mimeType, err := extractor.ExtractEmbeddedArtwork(audioPath)
	if err != nil {
		slog.Debug("no embedded artwork", "path", audioPath, "error", err)
		return nil, nil
	}
	if data == nil {
		return nil, nil
	}

	return &ArtworkInfo{
		Data:     data,
		MIMEType: mimeType,
		Source:   "embedded",
	}, nil
}

// findExternalArtwork looks for artwork files in a directory
func (p *ArtworkProcessor) findExternalArtwork(dir string) (*ArtworkInfo, error) {
	for _, filename := range ExternalArtworkFiles {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			mimeType := getMIMETypeFromFilename(filename)
			return &ArtworkInfo{
				Data:     data,
				MIMEType: mimeType,
				Source:   "external",
				Path:     path,
			}, nil
		}

		// Also try case-insensitive match
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if strings.EqualFold(entry.Name(), filename) {
				path := filepath.Join(dir, entry.Name())
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}

				mimeType := getMIMETypeFromFilename(entry.Name())
				return &ArtworkInfo{
					Data:     data,
					MIMEType: mimeType,
					Source:   "external",
					Path:     path,
				}, nil
			}
		}
	}

	return nil, nil
}

// ProcessAndCache processes artwork and caches it in multiple sizes
func (p *ArtworkProcessor) ProcessAndCache(artwork *ArtworkInfo, albumID string) (map[string]string, error) {
	if artwork == nil || len(artwork.Data) == 0 {
		return nil, nil
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(artwork.Data))
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	// Create cache directory for this album
	albumCacheDir := filepath.Join(p.cacheDir, "artwork", albumID)
	if err := os.MkdirAll(albumCacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	paths := make(map[string]string)

	// Save original
	originalPath := filepath.Join(albumCacheDir, "original.jpg")
	if err := p.saveImage(img, originalPath); err != nil {
		return nil, fmt.Errorf("saving original: %w", err)
	}
	paths["original"] = originalPath

	// Create resized versions
	for _, size := range AllArtworkSizes {
		resized := p.resize(img, size.Width, size.Height)
		path := filepath.Join(albumCacheDir, fmt.Sprintf("%s.jpg", size.Name))
		if err := p.saveImage(resized, path); err != nil {
			slog.Warn("failed to save resized image", "size", size.Name, "error", err)
			continue
		}
		paths[size.Name] = path
	}

	return paths, nil
}

// resize resizes an image to fit within the given dimensions while maintaining aspect ratio
func (p *ArtworkProcessor) resize(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Calculate new dimensions maintaining aspect ratio
	ratio := float64(srcWidth) / float64(srcHeight)
	var newWidth, newHeight int

	if ratio > 1 {
		// Wider than tall
		newWidth = maxWidth
		newHeight = int(float64(maxWidth) / ratio)
	} else {
		// Taller than wide or square
		newHeight = maxHeight
		newWidth = int(float64(maxHeight) * ratio)
	}

	// Ensure dimensions don't exceed max
	if newWidth > maxWidth {
		newWidth = maxWidth
		newHeight = int(float64(newWidth) / ratio)
	}
	if newHeight > maxHeight {
		newHeight = maxHeight
		newWidth = int(float64(newHeight) * ratio)
	}

	// Create new image with calculated dimensions
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Simple nearest-neighbor scaling (for better quality, use a dedicated imaging library)
	scaleX := float64(srcWidth) / float64(newWidth)
	scaleY := float64(srcHeight) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}
			dst.Set(x, y, img.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
		}
	}

	return dst
}

// saveImage saves an image as JPEG
func (p *ArtworkProcessor) saveImage(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	opts := &jpeg.Options{Quality: 85}
	if err := jpeg.Encode(file, img, opts); err != nil {
		return fmt.Errorf("encoding jpeg: %w", err)
	}

	return nil
}

// GetArtworkPath returns the cached artwork path for an album and size
func (p *ArtworkProcessor) GetArtworkPath(albumID string, size string) string {
	return filepath.Join(p.cacheDir, "artwork", albumID, fmt.Sprintf("%s.jpg", size))
}

// ArtworkExists checks if artwork exists for an album
func (p *ArtworkProcessor) ArtworkExists(albumID string) bool {
	path := filepath.Join(p.cacheDir, "artwork", albumID, "original.jpg")
	_, err := os.Stat(path)
	return err == nil
}

// DeleteArtwork removes cached artwork for an album
func (p *ArtworkProcessor) DeleteArtwork(albumID string) error {
	path := filepath.Join(p.cacheDir, "artwork", albumID)
	return os.RemoveAll(path)
}

// getMIMETypeFromFilename returns MIME type based on file extension
func getMIMETypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

// LoadArtwork loads artwork from cache
func (p *ArtworkProcessor) LoadArtwork(albumID string, size string) ([]byte, string, error) {
	path := p.GetArtworkPath(albumID, size)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return data, "image/jpeg", nil
}

// CopyArtwork copies artwork data to a writer
func (p *ArtworkProcessor) CopyArtwork(albumID string, size string, w io.Writer) error {
	path := p.GetArtworkPath(albumID, size)
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	return err
}

// SaveArtworkFromReader saves artwork from a reader
func (p *ArtworkProcessor) SaveArtworkFromReader(albumID string, r io.Reader, mimeType string) error {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading artwork: %w", err)
	}

	artwork := &ArtworkInfo{
		Data:     data,
		MIMEType: mimeType,
		Source:   "upload",
	}

	_, err = p.ProcessAndCache(artwork, albumID)
	return err
}

// SaveRawArtwork saves raw artwork data without processing
func (p *ArtworkProcessor) SaveRawArtwork(albumID string, data []byte, filename string) error {
	albumCacheDir := filepath.Join(p.cacheDir, "artwork", albumID)
	if err := os.MkdirAll(albumCacheDir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	path := filepath.Join(albumCacheDir, filename)
	return os.WriteFile(path, data, 0644)
}

// DecodeImage decodes image data into an image.Image
func DecodeImage(data []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", fmt.Errorf("decoding image: %w", err)
	}
	return img, format, nil
}

// EncodeJPEG encodes an image to JPEG format
func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("encoding jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

// EncodePNG encodes an image to PNG format
func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encoding png: %w", err)
	}
	return buf.Bytes(), nil
}
