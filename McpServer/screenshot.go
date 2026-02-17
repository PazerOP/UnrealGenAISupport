package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/kbinani/screenshot"
)

// captureScreenshot captures the primary monitor and returns a base64-encoded PNG string.
func captureScreenshot() (string, error) {
	n := screenshot.NumActiveDisplays()
	if n < 1 {
		return "", fmt.Errorf("no active displays found")
	}

	// Capture primary display (index 0, equivalent to mss mon=1)
	bounds := screenshot.GetDisplayBounds(0)
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return "", fmt.Errorf("capture failed: %w", err)
	}

	// Encode to PNG in memory
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("PNG encode failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
