package qr

import (
	"bytes"
	"image/png"
	"strings"
	"testing"
)

var pngHeader = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

func TestGenerate(t *testing.T) {
	data, err := Generate("https://example.com/items/1", 256)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("len(data) = 0, want non-zero")
	}
	if len(data) < len(pngHeader) {
		t.Fatalf("len(data) = %d, want >= %d", len(data), len(pngHeader))
	}
	if !bytes.Equal(data[:len(pngHeader)], pngHeader) {
		t.Errorf("png header = %v, want %v", data[:len(pngHeader)], pngHeader)
	}
}

func TestGenerate_ShortContent(t *testing.T) {
	data, err := Generate("A", 100)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(data) < len(pngHeader) {
		t.Fatalf("len(data) = %d, want >= %d", len(data), len(pngHeader))
	}
	if !bytes.Equal(data[:len(pngHeader)], pngHeader) {
		t.Errorf("png header = %v, want %v", data[:len(pngHeader)], pngHeader)
	}
}

func TestGenerate_MaxContent(t *testing.T) {
	content := strings.Repeat("a", 78)
	data, err := Generate(content, 200)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(data) == 0 {
		t.Errorf("len(data) = 0, want non-zero")
	}
}

func TestGenerate_TooLong(t *testing.T) {
	content := strings.Repeat("b", 79)
	_, err := Generate(content, 200)
	if err == nil {
		t.Fatalf("Generate error = nil, want error")
	}
	if !strings.Contains(err.Error(), "too long") {
		t.Errorf("error = %q, want contains %q", err.Error(), "too long")
	}
}

func TestGenerate_SmallSize(t *testing.T) {
	data, err := Generate("small", 1)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(data) < len(pngHeader) {
		t.Fatalf("len(data) = %d, want >= %d", len(data), len(pngHeader))
	}
	if !bytes.Equal(data[:len(pngHeader)], pngHeader) {
		t.Errorf("png header = %v, want %v", data[:len(pngHeader)], pngHeader)
	}
}

func TestGenerate_DecodePNG(t *testing.T) {
	data, err := Generate("decode", 128)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("png.Decode error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("bounds = %v, want non-zero size", bounds)
	}
}
