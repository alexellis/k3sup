// Copyright Alex Ellis, OpenFaaS Ltd 2025
// Inspired by update command in openfaas/faas-cli
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/alexellis/arkade/pkg/archive"
	"github.com/alexellis/arkade/pkg/env"
	goexecute "github.com/alexellis/go-execute/v2"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/spf13/cobra"
)

// MakeGetPro creates the 'get pro' command
func MakeGetPro() *cobra.Command {
	c := &cobra.Command{
		Use:   "pro",
		Short: "Download the latest k3sup pro binary",
		Long: `The latest release version of k3sup-pro will be downloaded from a remote 
container registry.

This command will download and install k3sup-pro to /usr/local/bin by default.`,
		Example: `  # Download to the default location
  k3sup get pro

  # Download a specific version of k3sup pro
  k3sup get pro --version v0.10.0

  # Download to a custom location
  k3sup get pro --path /tmp/`,
		RunE:    runGetProE,
		PreRunE: preRunGetProE,
	}

	c.Flags().Bool("verbose", false, "Enable verbose output")
	c.Flags().String("path", "/usr/local/bin/", "Custom installation path")
	c.Flags().String("version", "latest", "Specific version to download")

	return c
}

func preRunGetProE(cmd *cobra.Command, args []string) error {
	version, _ := cmd.Flags().GetString("version")

	if len(version) == 0 {
		return fmt.Errorf(`version must be specified, or use "latest"`)
	}

	return nil
}

func runGetProE(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	customPath, _ := cmd.Flags().GetString("path")
	version, _ := cmd.Flags().GetString("version")

	// Use the provided path or default to /usr/local/bin/
	var binaryPath string
	if customPath != "/usr/local/bin/" {
		binaryPath = customPath
		if verbose {
			fmt.Printf("Using custom binary path: %s\n", binaryPath)
		}
	} else {
		binaryPath = "/usr/local/bin/"
		if verbose {
			fmt.Printf("Using default binary path: %s\n", binaryPath)
		}
	}

	arch, operatingSystem := getClientArch()
	downloadArch, downloadOS := getDownloadArch(arch, operatingSystem)

	imageRef := fmt.Sprintf("ghcr.io/openfaasltd/k3sup-pro:%s", version)

	fmt.Printf("Downloading: %s (%s/%s)\n", imageRef, downloadOS, downloadArch)

	tmpTarDir, err := os.MkdirTemp(os.TempDir(), "k3sup-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpTarDir)

	tmpTar := path.Join(tmpTarDir, "k3sup-pro.tar")

	f, err := os.Create(tmpTar)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tmpTar, err)
	}
	defer f.Close()

	img, err := crane.Pull(imageRef, crane.WithPlatform(&v1.Platform{Architecture: downloadArch, OS: downloadOS}))
	if err != nil {
		return fmt.Errorf("pulling %s: %w", imageRef, err)
	}

	if err := crane.Export(img, f); err != nil {
		return fmt.Errorf("exporting %s: %w", imageRef, err)
	}

	if verbose {
		fmt.Printf("Wrote OCI filesystem to: %s\n", tmpTar)
	}

	tarFile, err := os.Open(tmpTar)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", tmpTar, err)
	}
	defer tarFile.Close()

	// Extract to temporary directory first
	tmpExtractDir, err := os.MkdirTemp(os.TempDir(), "k3sup-extract-*")
	if err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}
	defer os.RemoveAll(tmpExtractDir)

	gzipped := false
	if err := archive.Untar(tarFile, tmpExtractDir, gzipped, true); err != nil {
		return fmt.Errorf("failed to untar %s: %w", tmpTar, err)
	}

	binaryName := "k3sup-pro"
	if runtime.GOOS == "windows" {
		binaryName = "k3sup-pro.exe"
	}

	newBinary := path.Join(tmpExtractDir, binaryName)
	if err := os.Chmod(newBinary, 0755); err != nil {
		return fmt.Errorf("failed to chmod %s: %w", newBinary, err)
	}

	// Verify the extracted binary works
	if verbose {
		fmt.Println("Verifying extracted binary..")
	}
	task := goexecute.ExecTask{
		Command: newBinary,
		Args:    []string{"version"},
	}

	res, err := task.Execute(context.Background())
	if err != nil {
		return fmt.Errorf("failed to execute extracted binary: %w", err)
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("extracted binary test failed: %s", res.Stderr)
	}

	if verbose {
		fmt.Printf("New binary version check:\n%s", res.Stdout)
	}

	// Install to target path
	targetBinary := path.Join(binaryPath, binaryName)

	// Ensure target directory exists
	if err := os.MkdirAll(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", binaryPath, err)
	}

	if err := copyFile(newBinary, targetBinary); err != nil {
		return fmt.Errorf("failed to copy binary to %s: %w", targetBinary, err)
	}
	if err := os.Chmod(targetBinary, 0755); err != nil {
		return fmt.Errorf("failed to chmod %s: %w", targetBinary, err)
	}
	fmt.Printf("Installed: %s.. OK.\n", targetBinary)

	// Final version check
	finalTask := goexecute.ExecTask{
		Command: targetBinary,
		Args:    []string{"version"},
	}

	finalRes, err := finalTask.Execute(context.Background())
	if err != nil {
		return fmt.Errorf("failed to execute updated binary: %w", err)
	}

	if finalRes.ExitCode == 0 {
		fmt.Println("Installation completed successfully!")
		if !verbose {
			fmt.Print(finalRes.Stdout)
		}
	}

	return nil
}

func getClientArch() (arch string, os string) {
	if runtime.GOOS == "windows" {
		return runtime.GOARCH, runtime.GOOS
	}

	return env.GetClientArch()
}

func getDownloadArch(clientArch, clientOS string) (arch string, os string) {
	downloadArch := strings.ToLower(clientArch)
	downloadOS := strings.ToLower(clientOS)

	if downloadArch == "x86_64" {
		downloadArch = "amd64"
	} else if downloadArch == "aarch64" {
		downloadArch = "arm64"
	}

	return downloadArch, downloadOS
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer df.Close()

	_, err = io.Copy(df, sf)
	return err
}
