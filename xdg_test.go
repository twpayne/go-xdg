package xdg

import (
	"os"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/twpayne/go-vfs/vfst"
)

func TestNewXDG(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  map[string]string
		want *XDG
	}{
		{
			name: "default",
			env:  nil,
			want: &XDG{
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
			env: map[string]string{
				"XDG_CONFIG_HOME": "/my/user/config",
			},
			want: &XDG{
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
			env: map[string]string{
				"XDG_CONFIG_DIRS": "/config/dir/1:/config/dir/2",
			},
			want: &XDG{
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
			env: map[string]string{
				"XDG_DATA_HOME": "/my/user/data",
			},
			want: &XDG{
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
			env: map[string]string{
				"XDG_DATA_DIRS": "/data/dir/1:/data/dir/2",
			},
			want: &XDG{
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
			env: map[string]string{
				"XDG_CACHE_HOME": "/my/user/cache",
			},
			want: &XDG{
				ConfigHome: "/home/user/.config",
				ConfigDirs: []string{"/home/user/.config", "/etc/xdg"},
				DataHome:   "/home/user/.local/share",
				DataDirs:   []string{"/home/user/.local/share", "/usr/local/share", "/usr/share"},
				CacheHome:  "/my/user/cache",
				RuntimeDir: "",
			},
		},
		{
			name: "all",
			env: map[string]string{
				"XDG_CONFIG_HOME": "/my/user/config",
				"XDG_CONFIG_DIRS": "/config/dir/1:/config/dir/2",
				"XDG_DATA_HOME":   "/my/user/data",
				"XDG_DATA_DIRS":   "/data/dir/1:/data/dir/2",
				"XDG_CACHE_HOME":  "/my/user/cache",
				"XDG_RUNTIME_DIR": "/my/user/runtime",
			},
			want: &XDG{
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
			got := newXDG("/home/user", func(key string) string {
				return tc.env[key]
			})
			if diff, equal := messagediff.PrettyDiff(tc.want, got); !equal {
				t.Errorf("newXDG(...) == %+v, want %+v\n%s", got, tc.want, diff)
			}
		})
	}
}

func TestOpenConfigFile(t *testing.T) {
	for _, tc := range []struct {
		name     string
		env      map[string]string
		root     interface{}
		wantName string
		wantErr  error
	}{
		{
			name: "first_dir",
			root: map[string]interface{}{
				"/home/user/.config/go-xdg.conf": "# contents of first go-xdg.conf\n",
				"/etc/xdg/go-xdg.conf":           "# contents of second go-xdg.conf\n",
			},
			wantName: "/home/user/.config/go-xdg.conf",
		},
		{
			name: "second_dir",
			root: map[string]interface{}{
				"/etc/xdg/go-xdg.conf": "# contents of second go-xdg.conf\n",
			},
			wantName: "/etc/xdg/go-xdg.conf",
		},
		{
			name:    "not_found",
			root:    map[string]interface{}{},
			wantErr: os.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xdg := newXDG("/home/user", func(key string) string {
				return tc.env[key]
			})
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfst.NewTestFS(...) == %v, want <nil>", err)
			}
			gotFile, gotName, gotErr := xdg.OpenConfigFile(fs.Open, "go-xdg.conf")
			defer gotFile.Close()
			if (gotFile == nil && tc.wantErr == nil) || gotName != tc.wantName || gotErr != tc.wantErr {
				t.Errorf("xdg.OpenConfigFile(...) == %v, %q, %v, want !<nil>, %q, %v", gotFile, gotName, gotErr, tc.wantName, tc.wantErr)
			}
		})
	}
}

func TestOpenDataFile(t *testing.T) {
	for _, tc := range []struct {
		name     string
		env      map[string]string
		root     interface{}
		wantName string
		wantErr  error
	}{
		{
			name: "first_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/go-xdg.dat": "# contents of first go-xdg.dat\n",
				"/usr/local/share/go-xdg.dat":        "# contents of second go-xdg.dat\n",
				"/usr/share/go-xdg.dat":              "# contents of third go-xdg.dat\n",
			},
			wantName: "/home/user/.local/share/go-xdg.dat",
		},
		{
			name: "second_dir",
			root: map[string]interface{}{
				"/usr/local/share/go-xdg.dat": "# contents of second go-xdg.dat\n",
				"/usr/share/go-xdg.dat":       "# contents of third go-xdg.dat\n",
			},
			wantName: "/usr/local/share/go-xdg.dat",
		},
		{
			name: "third_dir",
			root: map[string]interface{}{
				"/usr/share/go-xdg.dat": "# contents of third go-xdg.dat\n",
			},
			wantName: "/usr/share/go-xdg.dat",
		},
		{
			name:    "not_found",
			root:    map[string]interface{}{},
			wantErr: os.ErrNotExist,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			xdg := newXDG("/home/user", func(key string) string {
				return tc.env[key]
			})
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfst.NewTestFS(...) == %v, want <nil>", err)
			}
			gotFile, gotName, gotErr := xdg.OpenDataFile(fs.Open, "go-xdg.dat")
			defer gotFile.Close()
			if (gotFile == nil && tc.wantErr == nil) || gotName != tc.wantName || gotErr != tc.wantErr {
				t.Errorf("xdg.OpenConfigFile(...) == %v, %q, %v, want !<nil>, %q, %v", gotFile, gotName, gotErr, tc.wantName, tc.wantErr)
			}
		})
	}
}
