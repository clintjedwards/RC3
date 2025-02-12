package conf

import (
	"os"
	"sort"
	"strings"

	"github.com/fatih/structs"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type CLI struct {
	Detail  bool   `koanf:"detail"`
	Format  string `koanf:"format"`
	Host    string `koanf:"host"`
	NoColor bool   `koanf:"no_color"`
	Token   string `koanf:"token"`
}

// DefaultCLIConfig returns a pre-populated configuration struct that is used as the base for super imposing user configuration
// settings.
func DefaultCLIConfig() *CLI {
	return &CLI{
		Host:   "localhost:8080",
		Format: "pretty",
	}
}

// Get configuration for command line.
// This involves correctly finding and ordering different possible paths for the configuration file:
//
//  1. The function is intended to be called with paths gleaned from the -config flag in the cli.
//  2. If the user does not use the -config path of the path does not exist,
//     then we default to a few hard coded config path locations.
//  3. Then try to see if the user has set an envvar for the config file, which overrides
//     all previous config file paths.
//  4. Finally, whatever configuration file path is found first is the processed.
//
// Whether or not we use the configuration file we then search the environment for all environment variables:
//   - Environment variables are loaded after the config file and therefore overwrite any conflicting keys.
//   - All configuration that goes into a configuration file can also be used as an environment variable.
func InitCLIConfig(flagPath string, loadDefaults bool) (*CLI, error) {
	var config *CLI

	// First we initiate the default values for the config.
	if loadDefaults {
		config = DefaultCLIConfig()
	}

	homeDir, _ := os.UserHomeDir()
	path := searchFilePaths(possibleConfigPaths(homeDir, flagPath)...)

	// envVars top all other entries so if its not empty we just insert it over the current path
	// regardless of if we found one.
	envPath := os.Getenv("RC3_CLI_CONFIG_PATH")
	if envPath != "" {
		path = envPath
	}

	configParser := koanf.New(".")

	if path != "" {
		err := configParser.Load(file.Provider(path), toml.Parser())
		if err != nil {
			return nil, err
		}
	}

	err := configParser.Load(env.Provider("RC3_CLI_", "__", func(s string) string {
		newStr := strings.TrimPrefix(s, "RC3_CLI_")
		newStr = strings.ToLower(newStr)
		return newStr
	}), nil)
	if err != nil {
		return nil, err
	}

	err = configParser.Unmarshal("", &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetCLIEnvVars() []string {
	vars := getEnvVarsFromStruct("RC3_CLI_", structs.Fields(CLI{}))
	sort.Strings(vars)
	return vars
}
