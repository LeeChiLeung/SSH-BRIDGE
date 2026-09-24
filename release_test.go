package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseBuildProducesVersionedArtifacts(t *testing.T) {
	t.Parallel()

	distDir := t.TempDir()
	version := "0.1.0-test"
	runReleaseCommand(t, distDir, map[string]string{
		"VERSION": version,
		"TARGETS": "linux/amd64 windows/amd64",
	}, "build")

	assertFileExists(t, filepath.Join(distDir, "timber-ssh-go_"+version+"_linux_amd64"))
	assertFileExists(t, filepath.Join(distDir, "timber-ssh-go_"+version+"_windows_amd64.exe"))

	output := runCommand(t, distDir, nil, filepath.Join(distDir, "timber-ssh-go_"+version+"_linux_amd64"), "--version")
	if !strings.Contains(output, version) {
		t.Fatalf("binary --version output missing injected version %q:\n%s", version, output)
	}
}

func TestReleaseBuildDerivesVersionWhenUnset(t *testing.T) {
	t.Parallel()

	distDir := t.TempDir()
	runReleaseCommand(t, distDir, map[string]string{
		"TARGETS": "linux/amd64",
	}, "build")

	version := derivedVersionForTests(t)
	assertFileExists(t, filepath.Join(distDir, "timber-ssh-go_"+version+"_linux_amd64"))
}

func TestReleaseDebProducesPackage(t *testing.T) {
	t.Parallel()

	distDir := t.TempDir()
	version := "0.1.0-test"
	runReleaseCommand(t, distDir, map[string]string{
		"VERSION": version,
	}, "deb", "amd64")

	debPath := filepath.Join(distDir, "timber-ssh-go_"+version+"_linux_amd64.deb")
	assertFileExists(t, debPath)

	output := runCommand(t, distDir, nil, "ar", "t", debPath)
	if !strings.Contains(output, "control.tar.gz") || !strings.Contains(output, "data.tar.gz") {
		t.Fatalf("deb archive contents missing expected entries:\n%s", output)
	}
}

func TestReleaseAppImageProducesArtifact(t *testing.T) {
	t.Parallel()

	distDir := t.TempDir()
	version := "0.1.0-test"
	fakeTool := filepath.Join(t.TempDir(), "appimagetool")
	if err := os.WriteFile(fakeTool, []byte("#!/usr/bin/env bash\nset -euo pipefail\nout=\"${@: -1}\"\n: > \"$out\"\n"), 0o755); err != nil {
		t.Fatalf("write fake appimagetool: %v", err)
	}
	runtimeFile := filepath.Join(t.TempDir(), "runtime-x86_64")
	if err := os.WriteFile(runtimeFile, []byte("runtime"), 0o644); err != nil {
		t.Fatalf("write fake runtime: %v", err)
	}

	runReleaseCommand(t, distDir, map[string]string{
		"VERSION":               version,
		"APPIMAGETOOL_BIN":      fakeTool,
		"APPIMAGE_RUNTIME_FILE": runtimeFile,
		"KEEP_WORK":             "1",
	}, "appimage", "amd64")

	assertFileExists(t, filepath.Join(distDir, "timber-ssh-go_"+version+"_linux_amd64.AppImage"))
	assertFileExists(t, filepath.Join(distDir, ".work", "appimage-amd64", "AppDir", "usr", "bin", "timber-ssh-go"))
	appRun := mustReadFile(t, filepath.Join(distDir, ".work", "appimage-amd64", "AppDir", "AppRun"))
	if !strings.Contains(appRun, "--tray --web-addr 127.0.0.1:8765") {
		t.Fatalf("AppRun does not default to tray mode on port 8765:\n%s", appRun)
	}

	if _, err := os.Stat(filepath.Join(distDir, ".work", "appimage-amd64", "AppDir", "profiles.json")); !os.IsNotExist(err) {
		t.Fatalf("appimage workdir unexpectedly contains profiles.json")
	}
}

func runReleaseCommand(t *testing.T, distDir string, extraEnv map[string]string, args ...string) string {
	t.Helper()

	env := map[string]string{
		"DIST_DIR": distDir,
		"GOCACHE":  filepath.Join(t.TempDir(), "go-build"),
	}
	for key, value := range extraEnv {
		env[key] = value
	}
	return runCommand(t, "", env, "bash", append([]string{"scripts/release.sh"}, args...)...)
}

func runCommand(t *testing.T, dir string, extraEnv map[string]string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = repoRoot(t)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = os.Environ()
	for key, value := range extraEnv {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s failed: %v\n%s", name, strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return root
}

func derivedVersionForTests(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("bash", "-lc", "if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then v=$(git describe --tags --dirty --always 2>/dev/null || true); if [ -n \"$v\" ]; then printf '%s' \"${v#v}\"; exit 0; fi; fi; printf '0.0.0-dev'")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("derive version failed: %v\n%s", err, string(out))
	}
	return string(out)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("expected file %s, got directory", path)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	return string(data)
}
