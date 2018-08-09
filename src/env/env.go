package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/imdario/mergo"
)

// Env is the structure of a configuration for an environment.
type Env struct {
	Name         string        `yaml:"-" json:"-" env:"-"`
	Password     string        `yaml:"password,omitempty" json:"password,omitempty" env:"THEMEKIT_PASSWORD"`
	ThemeID      string        `yaml:"theme_id,omitempty" json:"theme_id,omitempty" env:"THEMEKIT_THEME_ID"`
	Domain       string        `yaml:"store" json:"store" env:"THEMEKIT_STORE"`
	Directory    string        `yaml:"directory,omitempty" json:"directory,omitempty" env:"THEMEKIT_DIRECTORY"`
	IgnoredFiles []string      `yaml:"ignore_files,omitempty" json:"ignore_files,omitempty" env:"THEMEKIT_IGNORE_FILES" envSeparator:":"`
	Proxy        string        `yaml:"proxy,omitempty" json:"proxy,omitempty" env:"THEMEKIT_PROXY"`
	Ignores      []string      `yaml:"ignores,omitempty" json:"ignores,omitempty" env:"THEMEKIT_IGNORES" envSeparator:":"`
	Timeout      time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty" env:"THEMEKIT_TIMEOUT"`
	ReadOnly     bool          `yaml:"readonly,omitempty" json:"readonly,omitempty" env:"-"`
	Notify       string        `yaml:"notify,omitempty" json:"notify,omitempty" env:"THEMEKIT_NOTIFY"`
}

//Default is the default values for a environment
var Default = Env{
	Name:    "development",
	Timeout: 30 * time.Second,
}

func init() {
	Default.Directory, _ = os.Getwd()
}

func newEnv(name string, initial Env, overrides ...Env) (*Env, error) {
	newConfig := &Env{Name: name}
	for _, override := range overrides {
		mergo.Merge(newConfig, &override)
	}
	mergo.Merge(newConfig, &initial)
	mergo.Merge(newConfig, &Default)
	return newConfig, newConfig.validate()
}

func (env *Env) validate() error {
	errors := []string{}

	env.ThemeID = strings.ToLower(strings.TrimSpace(env.ThemeID))
	if env.ThemeID != "" {
		if env.ThemeID == "live" {
			env.ThemeID = ""
		} else if _, err := strconv.ParseInt(env.ThemeID, 10, 64); err != nil {
			errors = append(errors, "invalid theme_id")
		}
	}

	if len(env.Domain) == 0 {
		errors = append(errors, "missing store domain")
	} else if !strings.HasSuffix(env.Domain, "myshopify.com") {
		errors = append(errors, "invalid store domain must end in '.myshopify.com'")
	}

	if len(env.Password) == 0 {
		errors = append(errors, "missing password")
	}

	if fi, err := os.Stat(filepath.Clean(env.Directory)); err != nil {
		errors = append(errors, fmt.Sprintf("invalid project directory %v", err))
	} else if fi.Mode()&os.ModeSymlink != 0 {
		var symlinkErr error
		if env.Directory, symlinkErr = filepath.EvalSymlinks(filepath.Clean(env.Directory)); symlinkErr != nil {
			errors = append(errors, fmt.Sprintf("invalid project directory: %s", symlinkErr.Error()))
		}
	} else if !fi.Mode().IsDir() {
		errors = append(errors, fmt.Sprintf("Directory config %v is not a directory", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid environment [%s]: (%v)", env.Name, strings.Join(errors, ","))
	}

	return nil
}
