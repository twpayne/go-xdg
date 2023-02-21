package xdg_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v4/vfst"

	xdg "github.com/twpayne/go-xdg/v6"
)

func TestNewBaseDirectorySpecification(t *testing.T) {
	for _, tc := range []struct {
		name     string
		getenv   xdg.GetenvFunc
		expected *xdg.BaseDirectorySpecification
	}{
		{
			name: "default",
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "config_home",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CONFIG_HOME": "/my/user/config",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/my/user/config",
				ConfigDirs: []string{"/my/user/config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "config_dirs",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CONFIG_DIRS": "/config/dir/1:/config/dir/2",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/config/dir/1", "/config/dir/2"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "data_home",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_DATA_HOME": "/my/user/data",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/my/user/data",
				DataDirs:   []string{"/my/user/data", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "data_dirs",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_DATA_DIRS": "/data/dir/1:/data/dir/2",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/data/dir/1", "/data/dir/2"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "cache_home",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CACHE_HOME": "/my/user/cache",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/my/user/cache",
				RuntimeDir: "",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "runtime_dir",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_RUNTIME_DIR": "/my/user/runtime",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "/my/user/runtime",
				StateHome:  "/home/user/.local/state",
			},
		},
		{
			name: "state_dir",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_STATE_HOME": "/my/user/state",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
				StateHome:  "/my/user/state",
			},
		},
		{
			name: "all",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CONFIG_HOME": "/my/user/config",
				"XDG_CONFIG_DIRS": "/config/dir/1:/config/dir/2",
				"XDG_DATA_HOME":   "/my/user/data",
				"XDG_DATA_DIRS":   "/data/dir/1:/data/dir/2",
				"XDG_CACHE_HOME":  "/my/user/cache",
				"XDG_RUNTIME_DIR": "/my/user/runtime",
				"XDG_STATE_HOME":  "/my/user/state",
			}),
			expected: &xdg.BaseDirectorySpecification{
				ConfigHome: "/my/user/config",
				ConfigDirs: []string{"/my/user/config", "/config/dir/1", "/config/dir/2"},
				DataHome:   "/my/user/data",
				DataDirs:   []string{"/my/user/data", "/data/dir/1", "/data/dir/2"},
				CacheHome:  "/my/user/cache",
				RuntimeDir: "/my/user/runtime",
				StateHome:  "/my/user/state",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := xdg.NewTestBaseDirectorySpecification("/home/user", tc.getenv)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestOpenConfigFile(t *testing.T) {
	for _, tc := range []struct {
		name         string
		root         interface{}
		expectedName string
		expectedErr  error
	}{
		{
			name: "first_dir",
			root: map[string]interface{}{
				"/home/user/.config/go-xdg.conf": "# contents of first go-xdg.conf\n",
				"/etc/xdg/go-xdg.conf":           "# contents of second go-xdg.conf\n",
			},
			expectedName: "/home/user/.config/go-xdg.conf",
		},
		{
			name: "second_dir",
			root: map[string]interface{}{
				"/etc/xdg/go-xdg.conf": "# contents of second go-xdg.conf\n",
			},
			expectedName: "/etc/xdg/go-xdg.conf",
		},
		{
			name:        "not_found",
			root:        map[string]interface{}{},
			expectedErr: fs.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xdg := xdg.NewTestBaseDirectorySpecification("/home/user", nil)
			fileSystem, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			require.NoError(t, err)
			actualFile, actualName, err := xdg.OpenConfigFile(fileSystem, "go-xdg.conf")
			if err == nil {
				defer func() {
					assert.NoError(t, actualFile.Close())
				}()
			}
			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedName, actualName)
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
		})
	}
}

func TestOpenDataFile(t *testing.T) {
	for _, tc := range []struct {
		name         string
		root         interface{}
		expectedName string
		expectedErr  error
	}{
		{
			name: "first_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/go-xdg.dat": "# contents of first go-xdg.dat\n",
				"/usr/local/share/go-xdg.dat":        "# contents of second go-xdg.dat\n",
				"/usr/share/go-xdg.dat":              "# contents of third go-xdg.dat\n",
			},
			expectedName: "/home/user/.local/share/go-xdg.dat",
		},
		{
			name: "second_dir",
			root: map[string]interface{}{
				"/usr/local/share/go-xdg.dat": "# contents of second go-xdg.dat\n",
				"/usr/share/go-xdg.dat":       "# contents of third go-xdg.dat\n",
			},
			expectedName: "/usr/local/share/go-xdg.dat",
		},
		{
			name: "third_dir",
			root: map[string]interface{}{
				"/usr/share/go-xdg.dat": "# contents of third go-xdg.dat\n",
			},
			expectedName: "/usr/share/go-xdg.dat",
		},
		{
			name:        "not_found",
			root:        map[string]interface{}{},
			expectedErr: fs.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xdg := xdg.NewTestBaseDirectorySpecification("/home/user", nil)
			fileSystem, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			require.NoError(t, err)
			actualFile, actualName, err := xdg.OpenDataFile(fileSystem, "go-xdg.dat")
			if err == nil {
				defer func() {
					assert.NoError(t, actualFile.Close())
				}()
			}
			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedName, actualName)
			} else {
				assert.Equal(t, tc.expectedErr, err)
			}
		})
	}
}

func makeGetenvFunc(env map[string]string) xdg.GetenvFunc {
	return func(key string) string {
		return env[key]
	}
}
