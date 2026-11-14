package services

import "testing"

func TestFindCodeFileTypeMatchesExactFilename(t *testing.T) {
	fileType, recognized := findCodeFileType("Dockerfile")
	if !recognized || fileType.ID != "build" {
		t.Fatalf("expected Dockerfile to match build type, got %#v", fileType)
	}
}

func TestFindCodeFileTypeMatchesCaseInsensitiveExtension(t *testing.T) {
	fileType, recognized := findCodeFileType("Component.TSX")
	if !recognized || fileType.LanguageID != "typescript" {
		t.Fatalf("expected TSX to match TypeScript, got %#v", fileType)
	}
}

func TestDecodeCodeFileRejectsBinaryContent(t *testing.T) {
	_, _, valid := decodeCodeFile([]byte{'a', 0, 'b'})
	if valid {
		t.Fatal("expected content containing a null byte to be rejected")
	}
}

func TestDecodeCodeFilePreservesUTF8BOM(t *testing.T) {
	content, encoding, valid := decodeCodeFile([]byte{0xef, 0xbb, 0xbf, 'h', 'i'})
	if !valid || encoding != "utf-8-bom" || content != "hi" {
		t.Fatalf("unexpected decoded result: %q %q %t", content, encoding, valid)
	}
}

func TestCodeFileEncodingRoundTripsUTF16(t *testing.T) {
	encoded := encodeCodeFile("hello", "utf-16-le")
	content, encoding, valid := decodeCodeFile(encoded)
	if !valid || encoding != "utf-16-le" || content != "hello" {
		t.Fatalf("unexpected decoded result: %q %q %t", content, encoding, valid)
	}
}
