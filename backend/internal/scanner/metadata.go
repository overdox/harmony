package scanner

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

// TrackMetadata contains extracted metadata from an audio file
type TrackMetadata struct {
	Title       string
	Artist      string
	Album       string
	AlbumArtist string
	Year        int
	TrackNumber int
	DiscNumber  int
	Genre       string
	Duration    int // in seconds
	Bitrate     int
	SampleRate  int
	Channels    int
	Format      string
	HasArtwork  bool
}

// MetadataExtractor handles metadata extraction from audio files
type MetadataExtractor struct{}

// NewMetadataExtractor creates a new MetadataExtractor
func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

// Extract extracts metadata from an audio file
func (e *MetadataExtractor) Extract(path string) (*TrackMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		// If tag reading fails, try to extract from filename
		slog.Debug("tag reading failed, using filename fallback", "path", path, "error", err)
		return e.extractFromFilename(path), nil
	}

	trackMeta := &TrackMetadata{
		Title:       metadata.Title(),
		Artist:      metadata.Artist(),
		Album:       metadata.Album(),
		AlbumArtist: metadata.AlbumArtist(),
		Year:        metadata.Year(),
		Genre:       metadata.Genre(),
		Format:      GetFormatFromPath(path),
	}

	// Extract track and disc numbers
	trackNum, totalTracks := metadata.Track()
	trackMeta.TrackNumber = trackNum
	_ = totalTracks // Not used currently

	discNum, totalDiscs := metadata.Disc()
	trackMeta.DiscNumber = discNum
	if trackMeta.DiscNumber == 0 {
		trackMeta.DiscNumber = 1
	}
	_ = totalDiscs // Not used currently

	// Check for embedded artwork
	if metadata.Picture() != nil {
		trackMeta.HasArtwork = true
	}

	// Apply fallbacks for missing metadata
	e.applyFallbacks(trackMeta, path)

	return trackMeta, nil
}

// extractFromFilename creates metadata from the filename when tags are unavailable
func (e *MetadataExtractor) extractFromFilename(path string) *TrackMetadata {
	meta := &TrackMetadata{
		Format:     GetFormatFromPath(path),
		DiscNumber: 1,
	}

	filename := filepath.Base(path)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Try to parse common filename patterns

	// Pattern: "01 - Song Title" or "01. Song Title"
	trackPattern := regexp.MustCompile(`^(\d{1,3})[\s\.\-_]+(.+)$`)
	if matches := trackPattern.FindStringSubmatch(filename); matches != nil {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			meta.TrackNumber = num
		}
		meta.Title = strings.TrimSpace(matches[2])
	} else {
		// Pattern: "Artist - Song Title"
		artistPattern := regexp.MustCompile(`^(.+?)\s*-\s*(.+)$`)
		if matches := artistPattern.FindStringSubmatch(filename); matches != nil {
			meta.Artist = strings.TrimSpace(matches[1])
			meta.Title = strings.TrimSpace(matches[2])
		} else {
			// Just use filename as title
			meta.Title = filename
		}
	}

	// Use directory structure for album/artist fallbacks
	e.applyFallbacks(meta, path)

	return meta
}

// applyFallbacks fills in missing metadata from directory structure
func (e *MetadataExtractor) applyFallbacks(meta *TrackMetadata, path string) {
	dir := filepath.Dir(path)
	dirName := filepath.Base(dir)
	parentDir := filepath.Dir(dir)
	parentDirName := filepath.Base(parentDir)

	// Fallback title to filename
	if meta.Title == "" {
		filename := filepath.Base(path)
		meta.Title = strings.TrimSuffix(filename, filepath.Ext(filename))
		// Clean up common patterns
		meta.Title = cleanTitle(meta.Title)
	}

	// Fallback album to directory name
	if meta.Album == "" {
		meta.Album = cleanAlbumName(dirName)
	}

	// Fallback artist to parent directory name
	if meta.Artist == "" {
		meta.Artist = parentDirName
		// If parent is something generic, use "Unknown Artist"
		if isGenericName(meta.Artist) {
			meta.Artist = "Unknown Artist"
		}
	}

	// Set album artist if empty
	if meta.AlbumArtist == "" {
		meta.AlbumArtist = meta.Artist
	}

	// Try to extract year from album name if missing
	if meta.Year == 0 {
		meta.Year = extractYearFromString(meta.Album)
	}

	// Try to extract year from directory name
	if meta.Year == 0 {
		meta.Year = extractYearFromString(dirName)
	}
}

// cleanTitle removes track numbers and other prefixes from a title
func cleanTitle(title string) string {
	// Remove leading track numbers like "01 - ", "01. ", "01 "
	pattern := regexp.MustCompile(`^(\d{1,3})[\s\.\-_]+`)
	title = pattern.ReplaceAllString(title, "")
	return strings.TrimSpace(title)
}

// cleanAlbumName cleans up an album name extracted from directory
func cleanAlbumName(name string) string {
	// Remove year patterns like "(2020)" or "[2020]"
	yearPattern := regexp.MustCompile(`[\[\(]\d{4}[\]\)]`)
	name = yearPattern.ReplaceAllString(name, "")

	// Remove common suffixes
	suffixes := []string{" - Album", " (Album)", " [Album]", " - EP", " (EP)", " [EP]"}
	for _, suffix := range suffixes {
		name = strings.TrimSuffix(name, suffix)
	}

	return strings.TrimSpace(name)
}

// extractYearFromString tries to extract a year from a string
func extractYearFromString(s string) int {
	// Look for 4-digit year between 1900-2099
	yearPattern := regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	if matches := yearPattern.FindStringSubmatch(s); matches != nil {
		if year, err := strconv.Atoi(matches[1]); err == nil {
			return year
		}
	}
	return 0
}

// isGenericName checks if a name is too generic to be an artist name
func isGenericName(name string) bool {
	generic := map[string]bool{
		"music":     true,
		"media":     true,
		"audio":     true,
		"downloads": true,
		"library":   true,
		"songs":     true,
		"tracks":    true,
		"albums":    true,
		"":          true,
		".":         true,
		"..":        true,
	}
	return generic[strings.ToLower(name)]
}

// ExtractEmbeddedArtwork extracts embedded artwork from an audio file
func (e *MetadataExtractor) ExtractEmbeddedArtwork(path string) ([]byte, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, "", fmt.Errorf("reading tags: %w", err)
	}

	picture := metadata.Picture()
	if picture == nil {
		return nil, "", nil
	}

	mimeType := picture.MIMEType
	if mimeType == "" {
		// Try to detect from data
		mimeType = detectImageMIME(picture.Data)
	}

	return picture.Data, mimeType, nil
}

// detectImageMIME detects the MIME type from image data
func detectImageMIME(data []byte) string {
	if len(data) < 4 {
		return "image/jpeg" // Default
	}

	// Check magic bytes
	if data[0] == 0xFF && data[1] == 0xD8 {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return "image/webp"
	}

	return "image/jpeg" // Default fallback
}
