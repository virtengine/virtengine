package keeper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"

	// Import standard image decoders
	_ "image/jpeg"
	_ "image/png"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Media Types
// ============================================================================

// MediaType represents the type of media contained in a scope payload
type MediaType string

const (
	// MediaTypeImage represents image data (JPEG, PNG, WebP)
	MediaTypeImage MediaType = "image"

	// MediaTypeVideo represents video data (MP4, WebM, AVI)
	MediaTypeVideo MediaType = "video"

	// MediaTypeJSON represents structured JSON data
	MediaTypeJSON MediaType = "json"

	// MediaTypeUnknown represents unknown/unsupported media
	MediaTypeUnknown MediaType = "unknown"
)

// ImageFormat represents supported image formats
type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatWebP ImageFormat = "webp"
)

// VideoFormat represents supported video formats
type VideoFormat string

const (
	VideoFormatMP4  VideoFormat = "mp4"
	VideoFormatWebM VideoFormat = "webm"
	VideoFormatAVI  VideoFormat = "avi"
)

// ============================================================================
// Parsed Media Payload
// ============================================================================

// ParsedMediaPayload contains parsed content from a decrypted scope
type ParsedMediaPayload struct {
	// MediaType is the detected media type
	MediaType MediaType

	// ScopeType is the type of scope this payload came from
	ScopeType types.ScopeType

	// ImageData contains decoded image if MediaType is image
	ImageData *ParsedImageData

	// VideoData contains video metadata if MediaType is video
	VideoData *ParsedVideoData

	// JSONData contains parsed JSON if MediaType is json
	JSONData map[string]interface{}

	// RawData is the original raw bytes
	RawData []byte

	// ParseErrors contains any non-fatal errors during parsing
	ParseErrors []string
}

// ParsedImageData contains decoded image information
type ParsedImageData struct {
	// Format is the detected image format
	Format ImageFormat

	// Width is the image width in pixels
	Width int

	// Height is the image height in pixels
	Height int

	// Image is the decoded image (may be nil if decode failed)
	Image image.Image

	// RawBytes is the raw image bytes
	RawBytes []byte
}

// ParsedVideoData contains video metadata
type ParsedVideoData struct {
	// Format is the detected video format
	Format VideoFormat

	// RawBytes is the raw video bytes
	RawBytes []byte

	// EstimatedFrameCount is an estimate based on file size
	EstimatedFrameCount int
}

// ============================================================================
// Media Parser
// ============================================================================

// MediaParser parses decrypted scope payloads into typed media content
type MediaParser struct {
	// MaxImageSize is the maximum allowed image size in bytes (default 10MB)
	MaxImageSize int

	// MaxVideoSize is the maximum allowed video size in bytes (default 50MB)
	MaxVideoSize int

	// MaxJSONSize is the maximum allowed JSON size in bytes (default 1MB)
	MaxJSONSize int
}

// NewMediaParser creates a new media parser with default settings
func NewMediaParser() *MediaParser {
	return &MediaParser{
		MaxImageSize: 10 * 1024 * 1024, // 10MB
		MaxVideoSize: 50 * 1024 * 1024, // 50MB
		MaxJSONSize:  1 * 1024 * 1024,  // 1MB
	}
}

// Parse parses a decrypted scope payload into structured media content
func (mp *MediaParser) Parse(decrypted DecryptedScope) (*ParsedMediaPayload, error) {
	if len(decrypted.Plaintext) == 0 {
		return nil, errors.New("empty payload")
	}

	result := &ParsedMediaPayload{
		ScopeType:   decrypted.ScopeType,
		RawData:     decrypted.Plaintext,
		ParseErrors: make([]string, 0),
	}

	// Detect and parse based on scope type expectations
	switch decrypted.ScopeType {
	case types.ScopeTypeIDDocument, types.ScopeTypeSelfie:
		return mp.parseImagePayload(result)

	case types.ScopeTypeFaceVideo:
		return mp.parseVideoPayload(result)

	case types.ScopeTypeSSOMetadata, types.ScopeTypeEmailProof, types.ScopeTypeDomainVerify:
		return mp.parseJSONPayload(result)

	case types.ScopeTypeBiometric:
		// Biometric can be image, video, or structured data
		return mp.parseAutoDetect(result)

	case types.ScopeTypeSMSProof:
		return mp.parseJSONPayload(result)

	default:
		return mp.parseAutoDetect(result)
	}
}

// parseImagePayload parses an image payload
func (mp *MediaParser) parseImagePayload(result *ParsedMediaPayload) (*ParsedMediaPayload, error) {
	data := result.RawData

	if len(data) > mp.MaxImageSize {
		return nil, fmt.Errorf("image size %d exceeds maximum %d bytes", len(data), mp.MaxImageSize)
	}

	// Detect image format
	format := mp.detectImageFormat(data)
	if format == "" {
		return nil, errors.New("unrecognized image format")
	}

	result.MediaType = MediaTypeImage
	result.ImageData = &ParsedImageData{
		Format:   format,
		RawBytes: data,
	}

	// Try to decode the image to get dimensions
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		result.ParseErrors = append(result.ParseErrors, fmt.Sprintf("image decode warning: %v", err))
	} else {
		result.ImageData.Image = img
		bounds := img.Bounds()
		result.ImageData.Width = bounds.Dx()
		result.ImageData.Height = bounds.Dy()
	}

	return result, nil
}

// parseVideoPayload parses a video payload
func (mp *MediaParser) parseVideoPayload(result *ParsedMediaPayload) (*ParsedMediaPayload, error) {
	data := result.RawData

	if len(data) > mp.MaxVideoSize {
		return nil, fmt.Errorf("video size %d exceeds maximum %d bytes", len(data), mp.MaxVideoSize)
	}

	// Detect video format
	format := mp.detectVideoFormat(data)
	if format == "" {
		return nil, errors.New("unrecognized video format")
	}

	result.MediaType = MediaTypeVideo
	result.VideoData = &ParsedVideoData{
		Format:   format,
		RawBytes: data,
		// Estimate frame count based on typical video size
		// Assume ~30fps, ~500KB per second for compressed video
		EstimatedFrameCount: len(data) / (500 * 1024 / 30),
	}

	if result.VideoData.EstimatedFrameCount < 1 {
		result.VideoData.EstimatedFrameCount = 1
	}

	return result, nil
}

// parseJSONPayload parses a JSON payload
func (mp *MediaParser) parseJSONPayload(result *ParsedMediaPayload) (*ParsedMediaPayload, error) {
	data := result.RawData

	if len(data) > mp.MaxJSONSize {
		return nil, fmt.Errorf("JSON size %d exceeds maximum %d bytes", len(data), mp.MaxJSONSize)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	result.MediaType = MediaTypeJSON
	result.JSONData = parsed

	return result, nil
}

// parseAutoDetect attempts to auto-detect the media type
func (mp *MediaParser) parseAutoDetect(result *ParsedMediaPayload) (*ParsedMediaPayload, error) {
	data := result.RawData

	// Try image first
	if format := mp.detectImageFormat(data); format != "" {
		return mp.parseImagePayload(result)
	}

	// Try video
	if format := mp.detectVideoFormat(data); format != "" {
		return mp.parseVideoPayload(result)
	}

	// Try JSON
	if len(data) > 0 && (data[0] == '{' || data[0] == '[') {
		parsed, err := mp.parseJSONPayload(result)
		if err == nil {
			return parsed, nil
		}
	}

	// Unknown format
	result.MediaType = MediaTypeUnknown
	return result, nil
}

// detectImageFormat detects the image format from magic bytes
func (mp *MediaParser) detectImageFormat(data []byte) ImageFormat {
	if len(data) < 8 {
		return ""
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return ImageFormatJPEG
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if bytes.HasPrefix(data, pngMagic) {
		return ImageFormatPNG
	}

	// WebP: RIFF....WEBP
	if len(data) >= 12 &&
		data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'W' && data[9] == 'E' && data[10] == 'B' && data[11] == 'P' {
		return ImageFormatWebP
	}

	return ""
}

// detectVideoFormat detects the video format from magic bytes
func (mp *MediaParser) detectVideoFormat(data []byte) VideoFormat {
	if len(data) < 12 {
		return ""
	}

	// MP4/MOV: ftyp box
	if string(data[4:8]) == "ftyp" {
		return VideoFormatMP4
	}

	// WebM: EBML header (1A 45 DF A3)
	if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
		return VideoFormatWebM
	}

	// AVI: RIFF....AVI
	if data[0] == 'R' && data[1] == 'I' && data[2] == 'F' && data[3] == 'F' &&
		data[8] == 'A' && data[9] == 'V' && data[10] == 'I' {
		return VideoFormatAVI
	}

	return ""
}

// ============================================================================
// Validation Helpers
// ============================================================================

// ValidateImageDimensions checks if image dimensions are within acceptable range
func ValidateImageDimensions(width, height int) error {
	const (
		minDim = 64   // Minimum 64x64
		maxDim = 8192 // Maximum 8192x8192
	)

	if width < minDim || height < minDim {
		return fmt.Errorf("image dimensions %dx%d below minimum %dx%d", width, height, minDim, minDim)
	}

	if width > maxDim || height > maxDim {
		return fmt.Errorf("image dimensions %dx%d exceed maximum %dx%d", width, height, maxDim, maxDim)
	}

	return nil
}

// ValidateVideoSize checks if video has sufficient content for liveness
func ValidateVideoSize(estimatedFrames int) error {
	const minFrames = 15 // Need at least ~0.5 second at 30fps

	if estimatedFrames < minFrames {
		return fmt.Errorf("video too short: estimated %d frames, need at least %d", estimatedFrames, minFrames)
	}

	return nil
}

// IsValidScopePayload checks if a parsed payload is valid for its scope type
func IsValidScopePayload(payload *ParsedMediaPayload) (bool, string) {
	if payload == nil {
		return false, "nil payload"
	}

	switch payload.ScopeType {
	case types.ScopeTypeIDDocument:
		if payload.MediaType != MediaTypeImage {
			return false, fmt.Sprintf("ID document expected image, got %s", payload.MediaType)
		}
		if payload.ImageData == nil {
			return false, "ID document missing image data"
		}
		if err := ValidateImageDimensions(payload.ImageData.Width, payload.ImageData.Height); err != nil {
			return false, err.Error()
		}

	case types.ScopeTypeSelfie:
		if payload.MediaType != MediaTypeImage {
			return false, fmt.Sprintf("selfie expected image, got %s", payload.MediaType)
		}
		if payload.ImageData == nil {
			return false, "selfie missing image data"
		}

	case types.ScopeTypeFaceVideo:
		if payload.MediaType != MediaTypeVideo {
			return false, fmt.Sprintf("face video expected video, got %s", payload.MediaType)
		}
		if payload.VideoData == nil {
			return false, "face video missing video data"
		}
		if err := ValidateVideoSize(payload.VideoData.EstimatedFrameCount); err != nil {
			return false, err.Error()
		}

	case types.ScopeTypeSSOMetadata:
		if payload.MediaType != MediaTypeJSON {
			return false, fmt.Sprintf("SSO metadata expected JSON, got %s", payload.MediaType)
		}
		if payload.JSONData == nil {
			return false, "SSO metadata missing JSON data"
		}
	}

	return true, ""
}
