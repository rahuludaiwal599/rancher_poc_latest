package paths

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rancher-sandbox/rancher-desktop/src/go/rdctl/pkg/directories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaths(t *testing.T) {
	t.Run("should return correct paths without environment variables set", func(t *testing.T) {
		// Ensure that these variables are not set in the testing environment
		environment := map[string]string{
			"RD_LOGS_DIR":  "",
			"LOCALAPPDATA": "",
			"APPDATA":      "",
		}
		for key, value := range environment {
			t.Setenv(key, value)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("Unexpected error getting user home directory: %s", err)
		}
		expectedPaths := Paths{
			AppHome:         filepath.Join(homeDir, "AppData", "Local", appName),
			AltAppHome:      filepath.Join(homeDir, "AppData", "Local", appName),
			Config:          filepath.Join(homeDir, "AppData", "Local", appName),
			Logs:            filepath.Join(homeDir, "AppData", "Local", appName, "logs"),
			Cache:           filepath.Join(homeDir, "AppData", "Local", appName, "cache"),
			WslDistro:       filepath.Join(homeDir, "AppData", "Local", appName, "distro"),
			WslDistroData:   filepath.Join(homeDir, "AppData", "Local", appName, "distro-data"),
			Resources:       fakeResourcesPath,
			ExtensionRoot:   filepath.Join(homeDir, "AppData", "Local", appName, "extensions"),
			Snapshots:       filepath.Join(homeDir, "AppData", "Local", appName, "snapshots"),
			ContainerdShims: filepath.Join(homeDir, "AppData", "Local", appName, "containerd-shims"),
			OldUserData:     filepath.Join(homeDir, "AppData", "Local", appName, "cache", "Rancher Desktop"),
		}
		actualPaths, err := GetPaths(mockGetResourcesPath)
		if err != nil {
			t.Errorf("Unexpected error getting actual paths: %s", err)
		}
		if actualPaths != expectedPaths {
			t.Errorf("Actual paths does not match expected paths\nActual paths: %#v\nExpected paths: %#v", actualPaths, expectedPaths)
		}
	})

	t.Run("should return correct paths with environment variables set", func(t *testing.T) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("Unexpected error getting user home directory: %s", err)
		}
		environment := map[string]string{
			"RD_LOGS_DIR":  filepath.Join(homeDir, "mockRdLogsDir"),
			"LOCALAPPDATA": filepath.Join(homeDir, "mockLocalAppData"),
			"APPDATA":      filepath.Join(homeDir, "mockAppData"),
		}
		for key, value := range environment {
			t.Setenv(key, value)
		}

		expectedPaths := Paths{
			AppHome:         filepath.Join(environment["LOCALAPPDATA"], appName),
			AltAppHome:      filepath.Join(environment["LOCALAPPDATA"], appName),
			Config:          filepath.Join(environment["LOCALAPPDATA"], appName),
			Logs:            environment["RD_LOGS_DIR"],
			Cache:           filepath.Join(environment["LOCALAPPDATA"], appName, "cache"),
			WslDistro:       filepath.Join(environment["LOCALAPPDATA"], appName, "distro"),
			WslDistroData:   filepath.Join(environment["LOCALAPPDATA"], appName, "distro-data"),
			Resources:       fakeResourcesPath,
			ExtensionRoot:   filepath.Join(environment["LOCALAPPDATA"], appName, "extensions"),
			Snapshots:       filepath.Join(environment["LOCALAPPDATA"], appName, "snapshots"),
			ContainerdShims: filepath.Join(environment["LOCALAPPDATA"], appName, "containerd-shims"),
			OldUserData:     filepath.Join(environment["LOCALAPPDATA"], appName, "cache", "Rancher Desktop"),
		}
		actualPaths, err := GetPaths(mockGetResourcesPath)
		if err != nil {
			t.Errorf("Unexpected error getting actual paths: %s", err)
		}
		if actualPaths != expectedPaths {
			t.Errorf("Actual paths does not match expected paths\nActual paths: %#v\nExpected paths: %#v", actualPaths, expectedPaths)
		}
	})
}

// Given an application directory, create the rdctl executable at the expected
// path and return its path.
func makeRdctl(t *testing.T, appDir string) string {
	rdctlPath := filepath.Join(appDir, "resources/resources/win32/bin/rdctl.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(rdctlPath), 0o755))
	rdctl, err := os.OpenFile(rdctlPath, os.O_CREATE|os.O_WRONLY, 0o755)
	require.NoError(t, err)
	assert.NoError(t, rdctl.Close())
	return rdctlPath
}

// Given an application directory, create the main executable at the expected
// path and return its path.
func makeExecutable(t *testing.T, appDir string) string {
	executablePath := filepath.Join(appDir, "Rancher Desktop.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
	executable, err := os.OpenFile(executablePath, os.O_CREATE|os.O_WRONLY, 0o755)
	require.NoError(t, err)
	assert.NoError(t, executable.Close())
	return executablePath
}

func TestGetRDLaunchPath(t *testing.T) {
	t.Run("from bundled application", func(t *testing.T) {
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		rdctlPath := makeRdctl(t, dir)
		executablePath := makeExecutable(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		actual, err := GetRDLaunchPath(ctx)
		require.NoError(t, err)
		assert.Equal(t, executablePath, actual)
	})
	t.Run("from application directory", func(t *testing.T) {
		appDataDir, err := directories.GetLocalAppDataDirectory()
		require.NoError(t, err)
		executablePath := filepath.Join(appDataDir, "Programs/Rancher Desktop/Rancher Desktop.exe")
		if _, err := os.Stat(executablePath); errors.Is(err, os.ErrNotExist) {
			t.Skip("Application does not exist")
		}
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		rdctlPath := makeRdctl(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		actual, err := GetRDLaunchPath(ctx)
		require.NoError(t, err)
		assert.Equal(t, executablePath, actual)
	})
	t.Run("fail to find suitable application", func(t *testing.T) {
		appDataDir, err := directories.GetLocalAppDataDirectory()
		require.NoError(t, err)
		executablePath := filepath.Join(appDataDir, "Programs/Rancher Desktop/Rancher Desktop.exe")
		if _, err := os.Stat(executablePath); err == nil {
			t.Skip("Application exists")
		}
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		rdctlPath := makeRdctl(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		_, err = GetRDLaunchPath(ctx)
		assert.Error(t, err)
	})
}

func TestGetMainExecutable(t *testing.T) {
	t.Run("packaged application", func(t *testing.T) {
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		rdctlPath := makeRdctl(t, dir)
		executablePath := makeExecutable(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		actual, err := GetMainExecutable(ctx)
		require.NoError(t, err)
		assert.Equal(t, executablePath, actual)
	})
	t.Run("development build", func(t *testing.T) {
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		rdctlPath := makeRdctl(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		executablePath := filepath.Join(dir, "node_modules/electron/dist/electron.exe")
		require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
		f, err := os.OpenFile(executablePath, os.O_CREATE|os.O_WRONLY, 0o755)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		actual, err := GetMainExecutable(ctx)
		require.NoError(t, err)
		assert.Equal(t, executablePath, actual)
	})
}
