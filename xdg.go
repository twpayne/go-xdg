// Package xdg provides support for the XDG Base Directory Specification. See
// https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html.
package xdg

import (
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// An XDG represents an XDG configuration.
type XDG struct {
	ConfigHome string
	ConfigDirs []string
	DataHome   string
	DataDirs   []string
	CacheHome  string
	RuntimeDir string
}

// NewXDG returns a new XDG, configured from the user's environment variables
// according to the XDG specification.
func NewXDG() (*XDG, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	return newXDG(homeDir, os.Getenv), nil
}

func newXDG(homeDir string, getenv func(string) string) *XDG {
	defaultConfigHome := filepath.Join(homeDir, ".config")
	configHome := firstNonEmpty(getenv("XDG_CONFIG_HOME"), defaultConfigHome)

	defaultConfigDirs := filepath.Join("/", "etc", "xdg")
	configDirs := append([]string{configHome}, filepath.SplitList(firstNonEmpty(getenv("XDG_CONFIG_DIRS"), defaultConfigDirs))...)

	defaultDataHome := filepath.Join(homeDir, ".local", "share")
	dataHome := firstNonEmpty(getenv("XDG_DATA_HOME"), defaultDataHome)

	defaultDataDirs := strings.Join([]string{
		filepath.Join("/", "usr", "local", "share"),
		filepath.Join("/", "usr", "share"),
	}, string(filepath.ListSeparator))
	dataDirs := append([]string{dataHome}, filepath.SplitList(firstNonEmpty(getenv("XDG_DATA_DIRS"), defaultDataDirs))...)

	defaultCacheHome := filepath.Join(homeDir, ".cache")
	cacheHome := firstNonEmpty(getenv("XDG_CACHE_HOME"), defaultCacheHome)

	runtimeDir := getenv("XDG_RUNTIME_DIR")

	return &XDG{
		ConfigHome: configHome,
		ConfigDirs: configDirs,
		DataHome:   dataHome,
		DataDirs:   dataDirs,
		CacheHome:  cacheHome,
		RuntimeDir: runtimeDir,
	}
}

// OpenConfigFile opens the first configuration file with the given name found,
// its full path, and any error. If no file can be found, the error will be
// os.ErrNotExist.
//
// The file is opened with the open argument. If open is nil, os.Open is used.
func (x *XDG) OpenConfigFile(open func(string) (*os.File, error), nameComponents ...string) (*os.File, string, error) {
	return openFile(open, nameComponents, x.ConfigDirs)
}

// OpenDataFile opens the first data file with the given name found, its full
// path, and any error. If no file can be found, the error will be
// os.ErrNotExist.
//
// The file is opened with the open argument. If open is nil, os.Open is used.
func (x *XDG) OpenDataFile(open func(string) (*os.File, error), nameComponents ...string) (*os.File, string, error) {
	return openFile(open, nameComponents, x.DataDirs)
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func openFile(open func(string) (*os.File, error), nameComponents, dirs []string) (*os.File, string, error) {
	if open == nil {
		open = os.Open
	}
	for _, dir := range dirs {
		path := filepath.Join(append([]string{dir}, nameComponents...)...)
		f, err := open(path)
		switch {
		case err == nil:
			return f, path, nil
		case os.IsNotExist(err):
			continue
		default:
			return nil, path, err
		}
	}
	return nil, "", os.ErrNotExist
}
