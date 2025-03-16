package parser_test

import (
	"reflect"
	"testing"

	"github.com/jxsl13/banserver/parser"
)

func TestParseClientDropped(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		want     parser.ClientDropped
		wantBool bool
	}{
		{
			name: "vanilla drop",
			line: "[2024-12-29 13:50:46][server]: client dropped. cid=5 addr=123.123.123.123:64285 reason=''",
			want: parser.ClientDropped{
				ClientID: 5,
				IP:       "123.123.123.123",
				Port:     64285,
				Reason:   "",
			},
			wantBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parser.ParseClientDropped(tt.line)
			if ok != tt.wantBool {
				t.Errorf("ParseClientDropped() ok = %v, want %v", ok, tt.wantBool)
			}
			if ok && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseClientDropped() = %v, want %v", got, tt.want)
			}
		})
	}

}
