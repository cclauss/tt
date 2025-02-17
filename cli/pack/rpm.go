package pack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/otiai10/copy"
	"github.com/tarantool/tt/cli/cmdcontext"
	"github.com/tarantool/tt/cli/config"
	"github.com/tarantool/tt/cli/util"
)

// rpmPacker is a structure that implements Packer interface
// with specific rpm packing behavior.
type rpmPacker struct {
}

// Run packs a bundle into rpm package.
func (packer *rpmPacker) Run(cmdCtx *cmdcontext.CmdCtx, packCtx *PackCtx,
	opts *config.CliOpts) error {
	var err error

	if err := util.CheckRequiredBinaries("cpio"); err != nil {
		return err
	}

	// Create a package directory, where it will be built.
	packageDir, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer func() {
		err := os.RemoveAll(packageDir)
		if err != nil {
			log.Warnf("Failed to remove a temporary directory %s: %s",
				packageDir, err.Error())
		}
	}()

	log.Debugf("A root for package is located in: %s", packageDir)

	// Prepare a bundle.
	bundlePath, err := prepareBundle(cmdCtx, *packCtx, opts, true)
	if err != nil {
		return err
	}
	defer func() {
		err := os.RemoveAll(bundlePath)
		if err != nil {
			log.Warnf("Failed to remove a temporary directory %s: %s",
				bundlePath, err.Error())
		}
	}()

	bundleName, err := getPackageName(packCtx, opts, "", false)
	if err != nil {
		return err
	}

	if err := copy.Copy(bundlePath, filepath.Join(packageDir, "usr", "share", "tarantool",
		bundleName)); err != nil {
		return err
	}

	rpmSuffix, err := getRPMSuffix()
	if err != nil {
		return err
	}
	resPackagePath, err := getPackageName(packCtx, opts, rpmSuffix, true)
	if err != nil {
		return err
	}

	envSystemPath := filepath.Join("/", defaultEnvPrefix, bundleName)
	err = initSystemdDir(packCtx, opts, packageDir, envSystemPath)
	if err != nil {
		return err
	}

	err = packRpm(cmdCtx, packCtx, opts, packageDir, resPackagePath)

	if err != nil {
		return fmt.Errorf("failed to create RPM package: %s", err)
	}

	log.Infof("Created result RPM package: %s", resPackagePath)

	return nil
}

// getRPMSuffix returns suffix for an RPM package.
func getRPMSuffix() (string, error) {
	arch, err := util.GetArch()
	if err != nil {
		return "", err
	}
	debSuffix := "-1" + "." + arch + ".rpm"
	return debSuffix, nil
}
