package parser_test

import (
	"testing"
	"time"

	"github.com/jxsl13/banserver/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseClientBanned(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		want     parser.ClientBanned
		wantBool bool
	}{
		{
			name: "#1 v4",
			line: "[2025-02-16 10:39:05][net_ban]: banned '123.123.123.123' for 1 minute (Stressing network)",
			want: parser.ClientBanned{
				IP:       "123.123.123.123",
				Duration: 1 * time.Minute,
				Reason:   "Stressing network",
			},
			wantBool: true,
		},
		{
			name: "#2 v6",
			line: "[2025-02-16 10:39:05][net_ban]: banned '[36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06]' for 1 minute (Stressing network)",
			want: parser.ClientBanned{
				IP:       "36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06",
				Duration: 1 * time.Minute,
				Reason:   "Stressing network",
			},
			wantBool: true,
		},
		{
			name: "#3 v4 empty reason",
			line: "[2025-02-16 10:39:05][net_ban]: banned '123.123.123.123' for 1 minute ()",
			want: parser.ClientBanned{
				IP:       "123.123.123.123",
				Duration: 1 * time.Minute,
				Reason:   "",
			},
			wantBool: true,
		},
		{
			name: "#4 v6 empty reason",
			line: "[2025-02-16 10:39:05][net_ban]: banned '[36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06]' for 1 minute ()",
			want: parser.ClientBanned{
				IP:       "36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06",
				Duration: 1 * time.Minute,
				Reason:   "",
			},
			wantBool: true,
		},
		{
			name: "#5 v4 day",
			line: "[2025-02-16 10:39:05][net_ban]: banned '123.123.123.123' for 1440 minute ()",
			want: parser.ClientBanned{
				IP:       "123.123.123.123",
				Duration: 1440 * time.Minute,
				Reason:   "",
			},
			wantBool: true,
		},
		{
			name: "#6 v6 day",
			line: "[2025-02-16 10:39:05][net_ban]: banned '[36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06]' for 1440 minute ()",
			want: parser.ClientBanned{
				IP:       "36bc:94f6:4608:14b4:f72a:8aa9:c75f:4e06",
				Duration: 1440 * time.Minute,
				Reason:   "",
			},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parser.ParseClientBanned(tt.line)
			assert.Equal(t, tt.wantBool, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
