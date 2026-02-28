package chirp

import "testing"

func TestValidateMessage(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name          string
		args          args
		wantValid     bool
		wantError_msg string
	}{
		{"mensagem pequena", args{"w"}, true, ""},
		{"mensagem grande", args{"Isso é uma mensagem grande por incrível que pareça"}, true, ""},
		{"mensagem vazia", args{""}, false, "Chirp not informed"},
		{"mensagem grande demais", args{"01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789_"}, false, "Chirp is too long"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotError_msg := ValidateMessage(tt.args.msg)
			if gotValid != tt.wantValid {
				t.Errorf("ValidateMessage() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
			if gotError_msg != tt.wantError_msg {
				t.Errorf("ValidateMessage() gotError_msg = %v, want %v", gotError_msg, tt.wantError_msg)
			}
		})
	}
}
