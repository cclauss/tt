package init

import (
	// Go embed blank import.
	_ "embed"
	"fmt"
	"io"
	"os"

	"github.com/apex/log"
	"github.com/mitchellh/mapstructure"
	"github.com/tarantool/tt/cli/config"
	"github.com/tarantool/tt/cli/configure"
	"github.com/tarantool/tt/cli/util"
	"gopkg.in/yaml.v2"
)

const (
	defaultDirPermissions = os.FileMode(0750)
)

// InitCtx contains information for tt config creation.
type InitCtx struct {
	// SkipConfig - if set, disables cartridge & tarantoolctl config analysis,
	// so init does not try to get directories information from exitsting config files.
	SkipConfig bool
	// ForceMode, if set, tt config is re-written without a question.
	ForceMode bool
	// reader to use for reading user input.
	reader io.Reader
	// Tarantool executable path.
	TarantoolExecutable string
}

type cartridgeOpts struct {
	LogDir  string `mapstructure:"log-dir"`
	RunDir  string `mapstructure:"run-dir"`
	DataDir string `mapstructure:"data-dir"`
}

// configData is a configuration data loaded from an existing config.
type configData struct {
	instancesEnabled   string
	logDir             string
	runDir             string
	walDir             string
	vinylDir           string
	memtxDir           string
	tarantoolctlLayout bool
}

// configLoader binds config name with load functor.
type configLoader struct {
	configName string
	load       func(*InitCtx, string) (configData, error)
}

//go:embed print_tarantoolctl_cfg.lua
var printTarantoolctlCfgLuaBytes []byte

// loadCartridgeConfig parses configPath as .cartridge.yml and fill directories info structure.
func loadCartridgeConfig(initCtx *InitCtx, configPath string) (configData, error) {
	var cartridgeConf cartridgeOpts
	rawConfigOpts, err := util.ParseYAML(configPath)
	if err != nil {
		return configData{}, fmt.Errorf("failed to parse cartridge app configuration: %s", err)
	}

	if err := mapstructure.Decode(rawConfigOpts, &cartridgeConf); err != nil {
		return configData{}, fmt.Errorf("failed to parse cartridge app configuration: %s", err)
	}

	return configData{
		runDir:             cartridgeConf.RunDir,
		logDir:             cartridgeConf.LogDir,
		walDir:             cartridgeConf.DataDir,
		vinylDir:           cartridgeConf.DataDir,
		memtxDir:           cartridgeConf.DataDir,
		tarantoolctlLayout: false,
	}, nil
}

// createDirectories creates directories specified in dirList.
func createDirectories(dirList []string) error {
	for _, dirName := range dirList {
		if dirName == "" {
			continue
		}
		if err := util.CreateDirectory(dirName, defaultDirPermissions); err != nil {
			return err
		}
		log.Debugf("'%s' directory is created.", dirName)
	}
	return nil
}

// loadTarantoolctlConfig loads data from configPath which is processed as a lua file, and fills
// directories info structure.
func loadTarantoolctlConfig(initCtx *InitCtx, configPath string) (configData, error) {
	var existingCfg configData
	if initCtx.TarantoolExecutable == "" {
		return existingCfg, fmt.Errorf("tarantool executable is not set")
	}

	out, err := util.ExecuteCommandGetOutput(initCtx.TarantoolExecutable, "",
		printTarantoolctlCfgLuaBytes, "-", configPath)
	if err != nil {
		return existingCfg, fmt.Errorf("tarantoolctl config loading error: %s", string(out))
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(out, &raw); err != nil {
		return existingCfg, fmt.Errorf("failed to parse YAML: %s", err)
	}

	for _, dir := range []struct {
		path    *string
		varName string
	}{
		{&existingCfg.walDir, "wal_dir"},
		{&existingCfg.vinylDir, "vinyl_dir"},
		{&existingCfg.memtxDir, "memtx_dir"},
		{&existingCfg.logDir, "log_dir"},
		{&existingCfg.runDir, "pid_file"},
		{&existingCfg.instancesEnabled, "instance_dir"},
	} {
		if val, ok := raw[dir.varName]; ok && val != nil {
			*dir.path = val.(string)
		}
	}
	existingCfg.tarantoolctlLayout = true
	return existingCfg, nil
}

// generateTtEnv generates environment config in configPath using configuration data provided.
func generateTtEnv(configPath string, sourceCfg configData) error {
	cfg := config.Config{
		CliConfig: configure.GetDefaultCliOpts(),
	}
	if sourceCfg.runDir != "" {
		cfg.CliConfig.App.RunDir = sourceCfg.runDir
	}
	if sourceCfg.walDir != "" {
		cfg.CliConfig.App.WalDir = sourceCfg.walDir
	}
	if sourceCfg.vinylDir != "" {
		cfg.CliConfig.App.VinylDir = sourceCfg.vinylDir
	}
	if sourceCfg.memtxDir != "" {
		cfg.CliConfig.App.MemtxDir = sourceCfg.memtxDir
	}
	if sourceCfg.logDir != "" {
		cfg.CliConfig.App.LogDir = sourceCfg.logDir
	}
	if sourceCfg.instancesEnabled != "" {
		cfg.CliConfig.App.InstancesEnabled = sourceCfg.instancesEnabled
	}
	cfg.CliConfig.App.TarantoolctlLayout = sourceCfg.tarantoolctlLayout

	if err := util.WriteYaml(configPath, cfg); err != nil {
		return err
	}

	directoriesToCreate := []string{
		cfg.CliConfig.App.InstancesEnabled,
		cfg.CliConfig.Modules.Directory,
		cfg.CliConfig.App.IncludeDir,
		cfg.CliConfig.App.BinDir,
		cfg.CliConfig.Repo.Install,
	}
	for _, templatesPathOpts := range cfg.CliConfig.Templates {
		directoriesToCreate = append(directoriesToCreate, templatesPathOpts.Path)
	}

	return createDirectories(directoriesToCreate)
}

// FillCtx initializes init context.
func FillCtx(initCtx *InitCtx) {
	initCtx.reader = os.Stdin
}

// checkExistingConfig checks tt config for existence and asks for confirmation to overwrite.
// Returns file name if init process can continue, and false otherwise. In case of error, non-nil
// error returned as second returned value.
func checkExistingConfig(initCtx *InitCtx) (string, error) {
	configName, err := util.GetYamlFileName(configure.ConfigName, false)
	if configName == "" {
		return "", err
	}

	if _, err := os.Stat(configName); err == nil {
		if initCtx.ForceMode {
			if err = os.Remove(configName); err != nil {
				return "", err
			}
		} else {
			confirmed, err := util.AskConfirm(initCtx.reader,
				fmt.Sprintf("%s already exists. Overwrite?", configName))
			if err != nil {
				return "", err
			}
			if confirmed {
				if err = os.Remove(configName); err != nil {
					return "", err
				}
			} else {
				log.Info("Init is cancelled by user.")
				return "", nil
			}
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return configName, nil
}

// Run creates tt environment config for the application in current dir.
func Run(initCtx *InitCtx) error {
	if initCtx.reader == nil {
		initCtx.reader = os.Stdin
	}

	configLoaders := []configLoader{
		{".cartridge.yml", loadCartridgeConfig},
		{".tarantoolctl", loadTarantoolctlConfig},
	}

	configName, err := checkExistingConfig(initCtx)
	if configName == "" {
		return err
	}

	var sourceCfg configData
	if !initCtx.SkipConfig {
		for _, confLoader := range configLoaders {
			if _, err = os.Stat(confLoader.configName); err != nil {
				if os.IsNotExist(err) {
					continue
				} else {
					log.Warnf("Failed to get info of '%s': %s", confLoader.configName, err)
				}
			}

			log.Infof("Found existing config '%s'", confLoader.configName)
			sourceCfg, err = confLoader.load(initCtx, confLoader.configName)
			if err != nil {
				return err
			} else {
				break
			}
		}
	}
	if !util.IsApp(".") && sourceCfg.instancesEnabled == "" {
		// Current directory is not app dir, instances enabled dir will be used.
		sourceCfg.instancesEnabled = configure.InstancesEnabledDirName
	}

	if err := generateTtEnv(configName, sourceCfg); err != nil {
		return err
	}

	log.Infof("Environment config is written to '%s'", configure.ConfigName)

	return nil
}
