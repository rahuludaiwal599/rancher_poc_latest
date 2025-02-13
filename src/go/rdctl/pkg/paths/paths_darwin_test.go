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
		t.Setenv("RD_LOGS_DIR", "")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("Unexpected error getting user home directory: %s", err)
		}
		expectedPaths := Paths{
			AppHome:                 filepath.Join(homeDir, "Library", "Application Support", appName),
			AltAppHome:              filepath.Join(homeDir, ".rd"),
			Config:                  filepath.Join(homeDir, "Library", "Preferences", appName),
			Logs:                    filepath.Join(homeDir, "Library", "Logs", appName),
			Cache:                   filepath.Join(homeDir, "Library", "Caches", appName),
			Lima:                    filepath.Join(homeDir, "Library", "Application Support", appName, "lima"),
			Integration:             filepath.Join(homeDir, ".rd", "bin"),
			Resources:               fakeResourcesPath,
			DeploymentProfileSystem: filepath.Join("/Library", "Preferences"),
			DeploymentProfileUser:   filepath.Join(homeDir, "Library", "Preferences"),
			ExtensionRoot:           filepath.Join(homeDir, "Library", "Application Support", appName, "extensions"),
			Snapshots:               filepath.Join(homeDir, "Library", "Application Support", appName, "snapshots"),
			ContainerdShims:         filepath.Join(homeDir, "Library", "Application Support", appName, "containerd-shims"),
			OldUserData:             filepath.Join(homeDir, "Library", "Application Support", "Rancher Desktop"),
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
		rdLogsDir := filepath.Join(homeDir, "anotherLogsDir")
		t.Setenv("RD_LOGS_DIR", rdLogsDir)
		expectedPaths := Paths{
			AppHome:                 filepath.Join(homeDir, "Library", "Application Support", appName),
			AltAppHome:              filepath.Join(homeDir, ".rd"),
			Config:                  filepath.Join(homeDir, "Library", "Preferences", appName),
			Logs:                    rdLogsDir,
			Cache:                   filepath.Join(homeDir, "Library", "Caches", appName),
			Lima:                    filepath.Join(homeDir, "Library", "Application Support", appName, "lima"),
			Integration:             filepath.Join(homeDir, ".rd", "bin"),
			Resources:               fakeResourcesPath,
			DeploymentProfileSystem: filepath.Join("/Library", "Preferences"),
			DeploymentProfileUser:   filepath.Join(homeDir, "Library", "Preferences"),
			ExtensionRoot:           filepath.Join(homeDir, "Library", "Application Support", appName, "extensions"),
			Snapshots:               filepath.Join(homeDir, "Library", "Application Support", appName, "snapshots"),
			ContainerdShims:         filepath.Join(homeDir, "Library", "Application Support", appName, "containerd-shims"),
			OldUserData:             filepath.Join(homeDir, "Library", "Application Support", "Rancher Desktop"),
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
	rdctlPath := filepath.Join(appDir, "Contents/Resources/resources/darwin/bin/rdctl")
	require.NoError(t, os.MkdirAll(filepath.Dir(rdctlPath), 0o755))
	rdctl, err := os.OpenFile(rdctlPath, os.O_CREATE|os.O_WRONLY, 0o755)
	require.NoError(t, err)
	assert.NoError(t, rdctl.Close())
	return rdctlPath
}

// Given an application directory, create the main executable at the expected
// path and return its path.
func makeExecutable(t *testing.T, appDir string) string {
	executablePath := filepath.Join(appDir, "Contents/MacOS/Rancher Desktop")
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
		_ = makeExecutable(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		actual, err := GetRDLaunchPath(ctx)
		require.NoError(t, err)
		assert.Equal(t, dir, actual)
	})
	t.Run("from application directory", func(t *testing.T) {
		appDir := "/Applications/Rancher Desktop.app"
		executablePath := filepath.Join(appDir, "Contents/MacOS/Rancher Desktop")
		if _, err := os.Stat(executablePath); errors.Is(err, os.ErrNotExist) {
			t.Skip("Application does not exist")
		}
		dir, err := filepath.EvalSymlinks(t.TempDir())
		require.NoError(t, err)
		rdctlPath := makeRdctl(t, dir)
		ctx := directories.OverrideRdctlPath(context.Background(), rdctlPath)
		actual, err := GetRDLaunchPath(ctx)
		require.NoError(t, err)
		assert.Equal(t, appDir, actual)
	})
	t.Run("fail to find suitable application", func(t *testing.T) {
		appDir := "/Applications/Rancher Desktop.app"
		executablePath := filepath.Join(appDir, "Contents/MacOS/Rancher Desktop")
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
		executablePath := filepath.Join(dir, "node_modules/electron/dist/Electron.app/Contents/MacOS/Electron")
		require.NoError(t, os.MkdirAll(filepath.Dir(executablePath), 0o755))
		f, err := os.OpenFile(executablePath, os.O_CREATE|os.O_WRONLY, 0o755)
		require.NoError(t, err)
		require.NoError(t, f.Close())
		actual, err := GetMainExecutable(ctx)
		require.NoError(t, err)
		assert.Equal(t, executablePath, actual)
	})
}
