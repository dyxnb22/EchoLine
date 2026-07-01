package pin

import (
	"testing"
)

func TestParseConvAndMessageID(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/api/conversations/550e8400-e29b-41d4-a716-446655440000/pins/6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			wantErr: false,
		},
		{
			name:    "too short",
			path:    "/api/conversations/550e8400-e29b-41d4-a716-446655440000",
			wantErr: true,
		},
		{
			name:    "invalid uuid for conv",
			path:    "/api/conversations/not-a-uuid/pins/6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			wantErr: true,
		},
		{
			name:    "invalid uuid for message",
			path:    "/api/conversations/550e8400-e29b-41d4-a716-446655440000/pins/not-a-uuid",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseConvAndMessageID(tc.path)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseConvAndMessageID(%q) err = %v, wantErr = %v", tc.path, err, tc.wantErr)
			}
		})
	}
}

func TestParseConvID(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "/api/conversations/550e8400-e29b-41d4-a716-446655440000/pins",
			wantErr: false,
		},
		{
			name:    "too short",
			path:    "/api",
			wantErr: true,
		},
		{
			name:    "invalid uuid",
			path:    "/api/conversations/not-a-uuid/pins",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConvID(tc.path)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseConvID(%q) err = %v, wantErr = %v", tc.path, err, tc.wantErr)
			}
		})
	}
}
