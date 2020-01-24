package download

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

func findGithubRelease(url string) (string, error) {

	client := http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != 302 {
		return "", fmt.Errorf("incorrect status code: %d", res.StatusCode)
	}

	loc := res.Header.Get("Location")
	if len(loc) == 0 {
		return "", fmt.Errorf("unable to determine release version")
	}
	version := loc[strings.LastIndex(loc, "/")+1:]
	return version, nil
}

func downloadBinary(client *http.Client, url, name, outputDirectory string, untar bool) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("could not get download from %s, status code %d", url, res.StatusCode)
	}

	if res.Body == nil {
		return fmt.Errorf("no data returned from url: %s", url)
	}

	outputPath := path.Join(outputDirectory, name)
	defer res.Body.Close()
	if untar {
		r := ioutil.NopCloser(res.Body)
		if err := Untar(r, outputDirectory); err != nil {
			return err
		}
	} else {
		res, _ := ioutil.ReadAll(res.Body)

		if err := ioutil.WriteFile(outputPath, res, 0777); err != nil {
			return err
		}
	}
	return nil
}

func buildFilename(arch, osVal string) (string, string) {
	extension := ""
	arch = strings.ToLower(arch)
	osVal = strings.ToLower(osVal)

	if osVal == "windows" {
		extension = ".exe"
	}

	if arch == "arm" {
		arch = "armhf"
	}

	if osVal == "darwin" {
		arch = "-" + osVal
	} else if arch == "amd64" || arch == "x86_64" {
		arch = ""
	} else {
		arch = "-" + arch
	}

	return arch, extension
}


func moveFile(source, destination string) (error, bool) {
	src, err := os.Open(source)
	if err != nil {
		return err, false
	}
	defer src.Close()
	fi, err := src.Stat()
	if err != nil {
		return err, false
	}
	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	perm := fi.Mode() & os.ModePerm

	dst, err := os.OpenFile(destination, flag, perm)
	if err != nil {
		return err, true
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		dst.Close()
		os.Remove(destination)
		return err, false
	}
	err = dst.Close()
	if err != nil {
		return err, false
	}
	err = src.Close()
	if err != nil {
		return err, false
	}
	err = os.Remove(source)
	if err != nil {
		return err, false
	}
	return nil, false
}

func ExpandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}