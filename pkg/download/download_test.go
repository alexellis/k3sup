package download

import "testing"

func Test_BuildFilename_Linux_amd64(t *testing.T) {
	arch, ext := buildFilename("amd64", "linux")
	want := ""

	if want != arch+ext {
		t.Errorf("want: %s, but got: %s", want, arch+ext)
	}
}

func Test_BuildFilename_Windows_amd64(t *testing.T) {
	arch, ext := buildFilename("amd64", "windows")
	want := ".exe"

	if want != arch+ext {
		t.Errorf("want: %s, but got: %s", want, arch+ext)
	}
}

func Test_BuildFilename_Linux_arm64(t *testing.T) {
	arch, ext := buildFilename("arm64", "linux")
	want := "-arm64"

	if want != arch+ext {
		t.Errorf("want: %s, but got: %s", want, arch+ext)
	}
}

func Test_BuildFilename_Linux_armhf(t *testing.T) {
	arch, ext := buildFilename("armhf", "linux")
	want := "-armhf"

	if want != arch+ext {
		t.Errorf("want: %s, but got: %s", want, arch+ext)
	}
}

func Test_BuildFilename_Darwin_amd64(t *testing.T) {
	arch, ext := buildFilename("amd64", "darwin")
	want := "-darwin"

	if want != arch+ext {
		t.Errorf("want: %s, but got: %s", want, arch+ext)
	}
}