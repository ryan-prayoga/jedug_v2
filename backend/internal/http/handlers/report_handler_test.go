package handlers

import "testing"

func TestValidateReportBodyInvalidLocation(t *testing.T) {
	tests := []struct {
		name string
		body submitReportBody
		want string
	}{
		{
			name: "invalid latitude",
			body: submitReportBody{
				AnonToken: "anon",
				Latitude:  -91,
				Longitude: 106.8,
				Severity:  3,
				Media: []reportMediaInput{
					{
						ObjectKey: "issues/test.jpg",
						MimeType:  "image/jpeg",
						SizeBytes: 1024,
					},
				},
			},
			want: "latitude must be between -90 and 90",
		},
		{
			name: "invalid longitude",
			body: submitReportBody{
				AnonToken: "anon",
				Latitude:  -6.2,
				Longitude: 181,
				Severity:  3,
				Media: []reportMediaInput{
					{
						ObjectKey: "issues/test.jpg",
						MimeType:  "image/jpeg",
						SizeBytes: 1024,
					},
				},
			},
			want: "longitude must be between -180 and 180",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReportBody(&tc.body)
			if err == nil {
				t.Fatalf("expected error but got nil")
			}
			if err.Error() != tc.want {
				t.Fatalf("error mismatch: got %q want %q", err.Error(), tc.want)
			}
		})
	}
}
