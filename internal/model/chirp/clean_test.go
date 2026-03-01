package chirp

import "testing"

func TestCleanMessage(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name            string
		args            args
		wantCleaned_msg string
	}{
		{"leave empty string alone", args{""}, ""},
		{"Don't touch single clean word", args{"clean"}, "clean"},
		{"Clean single profane word", args{"kerfuffle"}, "****"},
		{
			"Clean mixed words, ignoring case",
			args{"This is a FORNAX Sharbert !!!"},
			"This is a **** **** !!!",
		},
		{
			"Can deal with multiple clean words",
			args{"Can deal with multiple clean words"},
			"Can deal with multiple clean words",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCleaned_msg := CleanMessage(tt.args.msg); gotCleaned_msg != tt.wantCleaned_msg {
				t.Errorf("CleanMessage() = %v, want %v", gotCleaned_msg, tt.wantCleaned_msg)
			}
		})
	}
}
