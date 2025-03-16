package parser_test

import (
	"testing"

	"github.com/jxsl13/banserver/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseClientEntered(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		want     parser.ClientEntered
		wantBool bool
	}{
		{
			name: "ddnet join",
			line: "2024-12-10 22:28:11 I server: player has entered the game. ClientId=2 addr=<{123.123.123.123:27996}> sixup=0",
			want: parser.ClientEntered{
				ClientID: 2,
				IP:       "123.123.123.123",
				Port:     27996,
			},
			wantBool: true,
		},
		{
			name: "ddnet join v6",
			line: "2024-12-10 22:28:11 I server: player has entered the game. ClientId=2 addr=<{[eadc:6745:7332:1a06:a9b2:ef9e:60e8:1c3f]:27996}> sixup=0",
			want: parser.ClientEntered{
				ClientID: 2,
				IP:       "[eadc:6745:7332:1a06:a9b2:ef9e:60e8:1c3f]",
				Port:     27996,
			},
			wantBool: true,
		},
		{
			name: "vanilla join",
			line: "[2024-12-29 13:50:42][server]: player has entered the game. ClientID=2 addr=234.234.234.234:64285",
			want: parser.ClientEntered{
				ClientID: 2,
				IP:       "234.234.234.234",
				Port:     64285,
			},
			wantBool: true,
		},
		{
			name: "vanilla join v6",
			line: "[2024-12-29 13:50:42][server]: player has entered the game. ClientID=2 addr=[0c2c:7f29:0206:717f:9c6d:8e17:8934:4e6c]:64285",
			want: parser.ClientEntered{
				ClientID: 2,
				IP:       "[0c2c:7f29:0206:717f:9c6d:8e17:8934:4e6c]",
				Port:     64285,
			},
			wantBool: true,
		},
		{
			name:     "invalid line",
			line:     "some random text",
			want:     parser.ClientEntered{},
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parser.ParseClientEntered(tt.line)
			assert.Equal(t, tt.wantBool, ok)
			if ok {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
