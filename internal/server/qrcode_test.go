package server

import (
	"strings"
	"testing"
)

func TestGenerateQRCode(t *testing.T) {
	url := "https://test.trycloudflare.com?token=ayo_testtoken"

	qr, err := GenerateQRCode(url)
	if err != nil {
		t.Fatalf("failed to generate QR code: %v", err)
	}

	if qr == "" {
		t.Error("expected non-empty QR code")
	}

	// QR code should contain block characters
	if !strings.ContainsAny(qr, "█▀▄ ") {
		t.Error("QR code should contain block characters")
	}
}

func TestGenerateQRCodeWithURL(t *testing.T) {
	baseURL := "https://abc123.trycloudflare.com"
	token := "ayo_testtoken123"

	qr, payload, err := GenerateQRCodeWithURL(baseURL, token)
	if err != nil {
		t.Fatalf("failed to generate QR code: %v", err)
	}

	if qr == "" {
		t.Error("expected non-empty QR code")
	}

	expectedURL := baseURL + "?token=" + token
	if payload.URL != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, payload.URL)
	}

	if payload.Token != token {
		t.Errorf("expected token %q, got %q", token, payload.Token)
	}
}

func TestFormatQRDisplay(t *testing.T) {
	qr := "██████\n██  ██\n██████"
	url := "https://test.com"
	token := "ayo_token"

	display := FormatQRDisplay(qr, url, token)

	if !strings.Contains(display, "Scan this QR code") {
		t.Error("expected scan instructions")
	}

	if !strings.Contains(display, url) {
		t.Error("expected URL in display")
	}

	if !strings.Contains(display, token) {
		t.Error("expected token in display")
	}
}
