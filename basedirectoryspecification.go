package xdg

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// A BaseDirectorySpecification represents an XDG Base Directory Specification
// configuration. See
// https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html.
type BaseDirectorySpecification struct {
	ConfigHome string
	ConfigDirs []string
	DataHome   string
	DataDirs   []string
	CacheHome  string
	RuntimeDir string
}

// NewBaseDirectorySpecification returns a new BaseDirectorySpecification,
// configured from the user's environment variables.
func NewBaseDirectorySpecification() (*BaseDirectorySpecification, error) {
	homeDir, err := userHomeDir()
	if err != nil {
		return nil, err
	}
	return newBaseDirectorySpecification(homeDir, os.Getenv), nil
}

func newBaseDirectorySpecification(homeDir string, getenv func(string) string) *BaseDirectorySpecification {
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

	return &BaseDirectorySpecification{
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
func (b *BaseDirectorySpecification) OpenConfigFile(open func(string) (*os.File, error), nameComponents ...string) (*os.File, string, error) {
	return openFile(open, nameComponents, b.ConfigDirs)
}

// OpenDataFile opens the first data file with the given name found, its full
// path, and any error. If no file can be found, the error will be
// os.ErrNotExist.
//
// The file is opened with the open argument. If open is nil, os.Open is used.
func (b *BaseDirectorySpecification) OpenDataFile(open func(string) (*os.File, error), nameComponents ...string) (*os.File, string, error) {
	return openFile(open, nameComponents, b.DataDirs)
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

// userHomeDir returns the current user's home directory.
//
// On Unix, including macOS, it returns the $HOME environment variable.
// On Windows, it returns %USERPROFILE%.
// On Plan 9, it returns the $home environment variable.
//
// FIXME this is copied from https://tip.golang.org/src/os/file.go?s=11606:11640#L379.
// Replace it with os.UserHomeDir when Go 1.12 is released.
func userHomeDir() (string, error) {
	env, enverr := "HOME", "$HOME"
	switch runtime.GOOS {
	case "windows":
		env, enverr = "USERPROFILE", "%userprofile%"
	case "plan9":
		env, enverr = "home", "$home"
	case "nacl", "android":
		return "/", nil
	case "darwin":
		if runtime.GOARCH == "arm" || runtime.GOARCH == "arm64" {
			return "/", nil
		}
	}
	if v := os.Getenv(env); v != "" {
		return v, nil
	}
	return "", errors.New(enverr + " is not defined")
}
