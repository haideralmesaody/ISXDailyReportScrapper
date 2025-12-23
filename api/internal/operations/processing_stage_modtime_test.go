package operations

import (
	"testing"
	"time"
)

func TestParseFileInfoModTime(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		wantOK bool
		want   time.Time
	}{
		{
			name:   "empty",
			input:  "",
			wantOK: false,
		},
		{
			name:   "invalid",
			input:  "not-a-date",
			wantOK: false,
		},
		{
			name:   "rfc3339",
			input:  time.Date(2025, 12, 12, 14, 15, 16, 0, time.FixedZone("TEST", 3*60*60)).Format(time.RFC3339),
			wantOK: true,
			want:   time.Date(2025, 12, 12, 14, 15, 16, 0, time.FixedZone("TEST", 3*60*60)),
		},
		{
			name:   "legacy_local",
			input:  time.Date(2025, 12, 12, 14, 15, 16, 0, time.Local).Format("2006-01-02 15:04:05"),
			wantOK: true,
			want:   time.Date(2025, 12, 12, 14, 15, 16, 0, time.Local),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseFileInfoModTime(tc.input)
			if ok != tc.wantOK {
				t.Fatalf("ok=%v want %v (got=%v)", ok, tc.wantOK, got)
			}
			if !ok {
				return
			}
			if !got.Equal(tc.want) {
				t.Fatalf("got=%s want=%s", got.Format(time.RFC3339Nano), tc.want.Format(time.RFC3339Nano))
			}
		})
	}
}
