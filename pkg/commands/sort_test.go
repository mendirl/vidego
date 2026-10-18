package commands

import (
	"testing"
	"vidego/pkg/datatype"
)

func TestFindBase(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		base     string
		expected string
	}{
		{
			name:     "Base explicitly specified",
			path:     "/mnt/c/Downloads/video.mp4",
			base:     "/custom/base/dir",
			expected: "/custom/base/dir",
		},
		{
			name:     "Base explicitly specified with trailing slash",
			path:     "/mnt/c/Downloads/video.mp4",
			base:     "/custom/base/dir/",
			expected: "/custom/base/dir",
		},
		{
			name:     "Base empty, detected from path drive c",
			path:     "/mnt/c/Downloads/video.mp4",
			base:     "",
			expected: "/mnt/c",
		},
		{
			name:     "Base empty, detected from path drive d",
			path:     "/mnt/d/Videos/video.mp4",
			base:     "",
			expected: "/mnt/d",
		},
		{
			name:     "Base empty, default fallback",
			path:     "/some/other/path/video.mp4",
			base:     "",
			expected: "/mnt/n/T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findBase(tt.path, tt.base)
			if got != tt.expected {
				t.Errorf("findBase(%q, %q) = %q, want %q", tt.path, tt.base, got, tt.expected)
			}
		})
	}
}

func TestComputeNamedNaseFolder(t *testing.T) {
	res := computeNamedNaseFolder("/mnt/c/test/vid.mp4", "Artist", false, "/custom/base")
	if res != "/custom/base/N/Artist" {
		t.Errorf("expected /custom/base/N/Artist, got %s", res)
	}

	resUnique := computeNamedNaseFolder("/mnt/c/test/vid.mp4", "Artist", true, "/custom/base")
	if resUnique != "/custom/base/N/_ALL" {
		t.Errorf("expected /custom/base/N/_ALL, got %s", resUnique)
	}
}

func TestComputeOtherNameFolder(t *testing.T) {
	v := datatype.Video{
		Duration: 250,
	}
	res := computeOtherNameFolder("/mnt/c/test/vid.mp4", v, "/custom/base")
	if res != "/custom/base/O/O1_under05" {
		t.Errorf("expected /custom/base/O/O1_under05, got %s", res)
	}
}
