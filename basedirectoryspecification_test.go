package xdg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v2/vfst"
)

func TestNewBaseDirectorySpecification(t *testing.T) {
	for _, tc := range []struct {
		name     string
		getenv   GetenvFunc
		expected *BaseDirectorySpecification
	}{
		{
			name: "default",
			expected: &BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
			},
		},
		{
			name: "config_home",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CONFIG_HOME": "/my/user/config",
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/my/user/config",
				ConfigDirs: []string{"/my/user/config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
			},
		},
		{
			name: "config_dirs",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CONFIG_DIRS": "/config/dir/1:/config/dir/2",
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/config/dir/1", "/config/dir/2"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
			},
		},
		{
			name: "data_home",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_DATA_HOME": "/my/user/data",
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/my/user/data",
				DataDirs:   []string{"/my/user/data", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
			},
		},
		{
			name: "data_dirs",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_DATA_DIRS": "/data/dir/1:/data/dir/2",
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/data/dir/1", "/data/dir/2"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "",
			},
		},
		{
			name: "cache_home",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_CACHE_HOME": "/my/user/cache",
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/my/user/cache",
				RuntimeDir: "",
			},
		},
		{
			name: "runtime_dir",
			getenv: makeGetenvFunc(map[string]string{
				"XDG_RUNTIME_DIR": "/my/user/runtime",
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/home/user/.cache",
				RuntimeDir: "/my/user/runtime",
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
			}),
			expected: &BaseDirectorySpecification{
				ConfigHome: "/my/user/config",
				ConfigDirs: []string{"/my/user/config", "/config/dir/1", "/config/dir/2"},
				DataHome:   "/my/user/data",
				DataDirs:   []string{"/my/user/data", "/data/dir/1", "/data/dir/2"},
				CacheHome:  "/my/user/cache",
				RuntimeDir: "/my/user/runtime",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual := NewTestBaseDirectorySpecification("/home/user", tc.getenv)
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
			expectedErr: os.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xdg := NewTestBaseDirectorySpecification("/home/user", nil)
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			require.NoError(t, err)
			actualFile, actualName, err := xdg.OpenConfigFile(fs.Open, "go-xdg.conf")
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
			expectedErr: os.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xdg := NewTestBaseDirectorySpecification("/home/user", nil)
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			require.NoError(t, err)
			actualFile, actualName, err := xdg.OpenDataFile(fs.Open, "go-xdg.dat")
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

func makeGetenvFunc(env map[string]string) GetenvFunc {
	return func(key string) string {
		return env[key]
	}
}
