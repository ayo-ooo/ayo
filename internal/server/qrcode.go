package server

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
	"rsc.io/qr"
)

// QRPayload contains connection info for display purposes.
// The QR code itself encodes just the URL with token as query param.
type QRPayload struct {
	URL   string // Full URL including token query param
	Token string // Token for display
}

// GenerateQRCode creates a QR code containing the URL.
// Returns the QR code as an ASCII string suitable for terminal display.
func GenerateQRCode(url string) (string, error) {
	var buf bytes.Buffer
	config := qrterminal.Config{
		Level:          qr.M,
		Writer:         &buf,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		QuietZone:      2,
	}
	qrterminal.GenerateWithConfig(url, config)

	return buf.String(), nil
}

// GenerateQRCodeToStdout writes the QR code directly to stdout with optimal rendering.
func GenerateQRCodeToStdout(url string) {
	qrterminal.GenerateHalfBlock(url, qr.M, os.Stdout)
}

// GenerateQRCodeWithURL generates a QR code and returns both the ASCII art
// and the payload for fallback display.
func GenerateQRCodeWithURL(baseURL, token string) (ascii string, payload QRPayload, err error) {
	// Build URL with token as query param
	fullURL := baseURL + "?token=" + token

	payload = QRPayload{
		URL:   fullURL,
		Token: token,
	}

	ascii, err = GenerateQRCode(fullURL)
	if err != nil {
		return "", payload, err
	}

	return ascii, payload, nil
}

// FormatQRDisplay creates a nicely formatted QR code display for the terminal.
func FormatQRDisplay(qrASCII, url, token string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("  Scan this QR code with the ayo app to connect:\n")
	sb.WriteString("\n")

	// Indent the QR code for nicer display
	for _, line := range strings.Split(qrASCII, "\n") {
		if line != "" {
			sb.WriteString("  ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString("  Or connect manually:\n")
	sb.WriteString(fmt.Sprintf("    URL:   %s\n", url))
	sb.WriteString(fmt.Sprintf("    Token: %s\n", token))
	sb.WriteString("\n")

	return sb.String()
}
