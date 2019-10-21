package cmd

import "testing"

func Test_getValuesSuffix_arm64(t *testing.T) {
	want := "-arm64"
	got := getValuesSuffix("arm64")
	if want != got {
		t.Errorf("suffix, want: %s, got: %s", want, got)
	}
}

func Test_getValuesSuffix_aarch64(t *testing.T) {
	want := "-arm64"
	got := getValuesSuffix("aarch64")
	if want != got {
		t.Errorf("suffix, want: %s, got: %s", want, got)
	}
}
