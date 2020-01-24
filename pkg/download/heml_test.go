package download

import (
	"testing"
)

func Test_getHelmURL(t *testing.T) {
	got := getHelmURL("amd64", "darwin", "v2.14.3")
	want := "https://get.helm.sh/helm-v2.14.3-darwin-amd64.tar.gz"

	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}