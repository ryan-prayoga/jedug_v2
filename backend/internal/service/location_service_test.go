package service

import "testing"

func TestBuildHumanLabel(t *testing.T) {
	tests := []struct {
		name        string
		primary     string
		parent      *string
		grandparent *string
		want        *string
	}{
		{
			name:    "joins primary and ancestors",
			primary: "Kecamatan Tebet",
			parent:  strPtr("Jakarta Selatan"),
			grandparent: strPtr("DKI Jakarta"),
			want:    strPtr("Kecamatan Tebet, Jakarta Selatan, DKI Jakarta"),
		},
		{
			name:    "deduplicates repeated names",
			primary: "Kecamatan Tebet",
			parent:  strPtr("Kecamatan Tebet"),
			grandparent: strPtr("DKI Jakarta"),
			want:    strPtr("Kecamatan Tebet, DKI Jakarta"),
		},
		{
			name:    "returns nil when all parts empty",
			primary: "   ",
			parent:  strPtr(""),
			grandparent: nil,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildHumanLabel(tt.primary, tt.parent, tt.grandparent)
			if !equalStringPtr(got, tt.want) {
				t.Fatalf("buildHumanLabel() = %v, want %v", valueOf(got), valueOf(tt.want))
			}
		})
	}
}

func equalStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func valueOf(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func strPtr(v string) *string {
	return &v
}
