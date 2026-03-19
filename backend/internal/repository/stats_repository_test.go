package repository

import (
	"testing"

	"jedug_backend/internal/domain"
)

func TestBuildStatsScope(t *testing.T) {
	provinceID := int64(31)
	regencyID := int64(3171)
	provinceLabel := "DKI Jakarta"
	regencyLabel := "Jakarta Selatan, DKI Jakarta"

	tests := []struct {
		name      string
		query     domain.PublicStatsQuery
		label     *string
		isDefault bool
		wantKind  string
		wantLabel string
	}{
		{
			name:      "global scope without label",
			query:     domain.PublicStatsQuery{},
			isDefault: true,
			wantKind:  "global",
			wantLabel: "Semua wilayah publik",
		},
		{
			name: "province scope keeps province label",
			query: domain.PublicStatsQuery{
				ProvinceID: &provinceID,
			},
			label:     &provinceLabel,
			isDefault: false,
			wantKind:  "province",
			wantLabel: provinceLabel,
		},
		{
			name: "regency scope wins over province scope",
			query: domain.PublicStatsQuery{
				ProvinceID: &provinceID,
				RegencyID:  &regencyID,
			},
			label:     &regencyLabel,
			isDefault: true,
			wantKind:  "regency",
			wantLabel: regencyLabel,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildStatsScope(tc.query, tc.label, tc.isDefault)
			if got.Kind != tc.wantKind {
				t.Fatalf("kind mismatch: got %q want %q", got.Kind, tc.wantKind)
			}
			if got.Label != tc.wantLabel {
				t.Fatalf("label mismatch: got %q want %q", got.Label, tc.wantLabel)
			}
			if got.IsDefault != tc.isDefault {
				t.Fatalf("is_default mismatch: got %t want %t", got.IsDefault, tc.isDefault)
			}
		})
	}
}
