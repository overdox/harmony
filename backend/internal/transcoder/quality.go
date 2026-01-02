package transcoder

import (
	"strings"
)

// QualityInfo provides quality information for API responses
type QualityInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Bitrate     int    `json:"bitrate,omitempty"`
	Format      string `json:"format,omitempty"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
}

// StreamQualityOptions provides available streaming options for a track
type StreamQualityOptions struct {
	Original   QualityInfo   `json:"original"`
	Qualities  []QualityInfo `json:"qualities"`
	Recommended string       `json:"recommended"`
}

// GetQualityInfo returns quality information for a profile
func GetQualityInfo(profile Profile, available bool) QualityInfo {
	info := QualityInfo{
		Name:      profile.Name,
		Bitrate:   profile.Bitrate,
		Format:    profile.Format,
		Available: available,
	}

	switch profile.Name {
	case "original":
		info.DisplayName = "Original"
		info.Description = "Original quality, no transcoding"
	case "high":
		info.DisplayName = "High"
		info.Description = "320 kbps MP3"
	case "medium":
		info.DisplayName = "Medium"
		info.Description = "192 kbps MP3"
	case "low":
		info.DisplayName = "Low"
		info.Description = "128 kbps MP3, optimized for slow connections"
	case "high-ogg":
		info.DisplayName = "High (OGG)"
		info.Description = "320 kbps OGG Vorbis"
	case "medium-ogg":
		info.DisplayName = "Medium (OGG)"
		info.Description = "192 kbps OGG Vorbis"
	case "low-ogg":
		info.DisplayName = "Low (OGG)"
		info.Description = "128 kbps OGG Vorbis"
	}

	return info
}

// GetStreamQualityOptions returns all available quality options for streaming
func GetStreamQualityOptions(transcoderAvailable bool, originalFormat string, originalBitrate int) StreamQualityOptions {
	options := StreamQualityOptions{
		Original: QualityInfo{
			Name:        "original",
			DisplayName: "Original",
			Format:      originalFormat,
			Bitrate:     originalBitrate,
			Description: "Original quality, no transcoding",
			Available:   true,
		},
		Recommended: "original",
	}

	if transcoderAvailable {
		options.Qualities = []QualityInfo{
			GetQualityInfo(ProfileHigh, true),
			GetQualityInfo(ProfileMedium, true),
			GetQualityInfo(ProfileLow, true),
		}

		// Recommend based on original bitrate
		if originalBitrate > 0 {
			if originalBitrate <= 128 {
				options.Recommended = "original"
			} else if originalBitrate <= 192 {
				options.Recommended = "medium"
			} else if originalBitrate <= 320 {
				options.Recommended = "high"
			}
		}
	}

	return options
}

// QualitySelector helps select appropriate quality based on various factors
type QualitySelector struct {
	transcoderAvailable bool
	defaultQuality      string
}

// NewQualitySelector creates a new QualitySelector
func NewQualitySelector(transcoderAvailable bool) *QualitySelector {
	return &QualitySelector{
		transcoderAvailable: transcoderAvailable,
		defaultQuality:      "original",
	}
}

// SelectQuality selects the best quality based on various factors
func (s *QualitySelector) SelectQuality(
	requestedQuality string,
	saveData bool,
	effectiveConnectionType string,
	downlink float64,
) string {
	// If quality explicitly requested, use it
	if requestedQuality != "" {
		return s.validateQuality(requestedQuality)
	}

	// Save-Data mode
	if saveData {
		return "low"
	}

	// Based on effective connection type
	switch strings.ToLower(effectiveConnectionType) {
	case "slow-2g", "2g":
		return "low"
	case "3g":
		return "medium"
	case "4g":
		if s.transcoderAvailable {
			return "high"
		}
		return "original"
	}

	// Based on downlink speed (Mbps)
	if downlink > 0 {
		if downlink < 0.5 {
			return "low"
		} else if downlink < 1.5 {
			return "medium"
		} else if downlink < 5 {
			return "high"
		}
	}

	return s.defaultQuality
}

// validateQuality ensures the requested quality is valid
func (s *QualitySelector) validateQuality(quality string) string {
	quality = strings.ToLower(quality)

	// Check if it's a valid profile
	if _, err := GetProfile(quality); err == nil {
		// If transcoder not available, fall back to original
		if !s.transcoderAvailable && quality != "original" {
			return "original"
		}
		return quality
	}

	return s.defaultQuality
}

// ClientHints represents client hint headers for quality selection
type ClientHints struct {
	SaveData                bool
	EffectiveConnectionType string
	Downlink                float64
	RTT                     int
	DeviceMemory            float64
}

// ParseClientHints parses client hint headers
func ParseClientHints(headers map[string]string) ClientHints {
	hints := ClientHints{}

	if headers["Save-Data"] == "on" {
		hints.SaveData = true
	}

	hints.EffectiveConnectionType = headers["ECT"]

	// Parse Downlink (Mbps)
	if dl, ok := headers["Downlink"]; ok {
		var downlink float64
		if _, err := parseFloat(dl, &downlink); err == nil {
			hints.Downlink = downlink
		}
	}

	// Parse RTT (ms)
	if rtt, ok := headers["RTT"]; ok {
		var rttVal int
		if _, err := parseInt(rtt, &rttVal); err == nil {
			hints.RTT = rttVal
		}
	}

	return hints
}

// Helper functions for parsing
func parseFloat(s string, v *float64) (bool, error) {
	// Simple implementation - in production use strconv
	return false, nil
}

func parseInt(s string, v *int) (bool, error) {
	// Simple implementation - in production use strconv
	return false, nil
}

// BitrateRecommendation provides bitrate recommendations
type BitrateRecommendation struct {
	Format        string
	MinBitrate    int
	MaxBitrate    int
	Recommended   int
	LosslessOK    bool
}

// GetBitrateRecommendation returns bitrate recommendation for a format
func GetBitrateRecommendation(format string) BitrateRecommendation {
	format = strings.ToLower(format)

	switch format {
	case "mp3":
		return BitrateRecommendation{
			Format:      "mp3",
			MinBitrate:  96,
			MaxBitrate:  320,
			Recommended: 256,
			LosslessOK:  false,
		}
	case "ogg":
		return BitrateRecommendation{
			Format:      "ogg",
			MinBitrate:  96,
			MaxBitrate:  500,
			Recommended: 256,
			LosslessOK:  false,
		}
	case "flac":
		return BitrateRecommendation{
			Format:      "flac",
			MinBitrate:  800,
			MaxBitrate:  1411,
			Recommended: 1000,
			LosslessOK:  true,
		}
	case "wav":
		return BitrateRecommendation{
			Format:      "wav",
			MinBitrate:  1411,
			MaxBitrate:  1411,
			Recommended: 1411,
			LosslessOK:  true,
		}
	default:
		return BitrateRecommendation{
			Format:      format,
			MinBitrate:  128,
			MaxBitrate:  320,
			Recommended: 256,
			LosslessOK:  false,
		}
	}
}
