// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"strconv"
	"github.com/mitchellh/multistep"
	hypervcommon "github.com/MSOpenTech/packer-hyperv/packer/builder/hyperv/common"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"regexp"
	"code.google.com/p/go-uuid/uuid"
	"strings"
)

const (
	DefaultDiskSize = 127 * 1024	// 127GB
	MinDiskSize = 10 * 1024			// 10GB
	MaxDiskSize = 65536 * 1024		// 64TB

	DefaultRamSize = 1024	// 1GB
	MinRamSize = 512		// 512MB
	MaxRamSize = 32768 		// 32GB
)

// Builder implements packer.Builder and builds the actual Hyperv
// images.
type Builder struct {
	config config
	runner multistep.Runner
}

type config struct {
	// The size, in megabytes, of the hard disk to create for the VM. 
	// By default, this is 130048 (about 127 GB).
	DiskSize            uint     			`mapstructure:"disk_size"`
	// The size, in megabytes, of the computer memory in the VM. 
	// By default, this is 1024 (about 1 GB).
	RamSizeMB           uint     			`mapstructure:"ram_size_mb"`
	// A list of files to place onto a floppy disk that is attached when the 
	// VM is booted. This is most useful for unattended Windows installs, 
	// which look for an Autounattend.xml file on removable media. By default, 
	// no floppy will be attached. All files listed in this setting get 
	// placed into the root directory of the floppy and the floppy is attached 
	// as the first floppy device. Currently, no support exists for creating 
	// sub-directories on the floppy. Wildcard characters (*, ?, and []) 
	// are allowed. Directory names are also allowed, which will add all 
	// the files found in the directory to the floppy.
	FloppyFiles         []string            `mapstructure:"floppy_files"`
	// The checksum for the OS ISO file. Because ISO files are so large, 
	// this is required and Packer will verify it prior to booting a virtual 
	// machine with the ISO attached. The type of the checksum is specified 
	// with iso_checksum_type, documented below.
	ISOChecksum         string              `mapstructure:"iso_checksum"`
	// The type of the checksum specified in iso_checksum. Valid values are 
	// "none", "md5", "sha1", "sha256", or "sha512" currently. While "none" 
	// will skip checksumming, this is not recommended since ISO files are 
	// generally large and corruption does happen from time to time.
	ISOChecksumType     string              `mapstructure:"iso_checksum_type"`
	// A URL to the ISO containing the installation image. This URL can be 
	// either an HTTP URL or a file URL (or path to a file). If this is an 
	// HTTP URL, Packer will download it and cache it between runs.
	RawSingleISOUrl 	string 				`mapstructure:"iso_url"`
	// Multiple URLs for the ISO to download. Packer will try these in order. 
	// If anything goes wrong attempting to download or while downloading a 
	// single URL, it will move on to the next. All URLs must point to the 
	// same file (same checksum). By default this is empty and iso_url is 
	// used. Only one of iso_url or iso_urls can be specified.
	ISOUrls             []string            `mapstructure:"iso_urls"`
	// This is the name of the new virtual machine. 
	// By default this is "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
	VMName              string              `mapstructure:"vm_name"`


	SleepTimeMinutes 	time.Duration		`mapstructure:"wait_time_minutes"`
	ProductKey 			string				`mapstructure:"product_key"`

	common.PackerConfig           			`mapstructure:",squash"`
	hypervcommon.OutputConfig     			`mapstructure:",squash"`

	SwitchName          string 				`mapstructure:"switch_name"`

	tpl *packer.ConfigTemplate
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {

	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("%s: %v", "PackerUserVars", b.config.PackerUserVars))

	b.config.tpl.UserVars = b.config.PackerUserVars

	// Accumulate any errors and warnings
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.OutputConfig.Prepare(b.config.tpl, &b.config.PackerConfig)...)
	warnings := make([]string, 0)

	err = b.checkDiskSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	err = b.checkRamSize()
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	warning := b.checkHostAvailableMemory()
	if warning != "" {
		warnings = appendWarnings(warnings, warning)
	}


	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("pvm_%s", uuid.New())
	}

	if b.config.SwitchName == "" {
		b.config.SwitchName = fmt.Sprintf("pis_%s", uuid.New())
	}

	if b.config.SleepTimeMinutes == 0 {
		b.config.SleepTimeMinutes = 10
	} else if b.config.SleepTimeMinutes < 0 {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("wait_time_minutes: '%v' %s", int64(b.config.SleepTimeMinutes), "the value can't be negative" ))
	} else if b.config.SleepTimeMinutes > 1440 {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("wait_time_minutes: '%v' %s", uint(b.config.SleepTimeMinutes), "the value is too big" ))
	} else if b.config.SleepTimeMinutes > 120 {
		warnings = appendWarnings( warnings, fmt.Sprintf("wait_time_minutes: '%v' %s", uint(b.config.SleepTimeMinutes), "You may want to decrease the value. Usually 20 min is enough."))
	}
	log.Println(fmt.Sprintf("%s: %v", "SleepTimeMinutes", uint(b.config.SleepTimeMinutes)))

	// Errors
	templates := map[string]*string{
		"iso_url":            &b.config.RawSingleISOUrl,
		"product_key":        &b.config.ProductKey,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = b.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	// TODO: remove product key, use Autounattend.xml on a floppy instead
	pk := strings.TrimSpace(b.config.ProductKey)
	if len(pk) != 0 {
		pattern := "^[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$"
		value := pk

		match, _ := regexp.MatchString(pattern, value)
		if !match {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("product_key: Make sure the product_key follows the pattern: XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"))
		}

		warnings = appendWarnings( warnings, fmt.Sprintf("product_key: %s", "value is not empty. Packer will try to activate Windows with the product key. To do this Packer will need an Internet connection."))
	}

	log.Println(fmt.Sprintf("%s: %v","VMName", b.config.VMName))
	log.Println(fmt.Sprintf("%s: %v","SwitchName", b.config.SwitchName))
	log.Println(fmt.Sprintf("%s: %v","ProductKey", b.config.ProductKey))


	if b.config.RawSingleISOUrl == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("iso_url: The option can't be missed and a path must be specified."))
	}else if _, err := os.Stat(b.config.RawSingleISOUrl); err != nil {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("iso_url: Check the path is correct"))
	}

	log.Println(fmt.Sprintf("%s: %v","RawSingleISOUrl", b.config.RawSingleISOUrl))

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a Hyperv appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Hyperv
	driver, err := hypervcommon.NewHypervPS4Driver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Hyper-V driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		new(hypervcommon.StepCreateTempDir),
		&hypervcommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		&hypervcommon.StepCreateSwitch{
			SwitchName: b.config.SwitchName,
		},
		&hypervcommon.StepCreateVM{
			VMName: b.config.VMName,
			SwitchName: b.config.SwitchName,
			RamSizeMB: b.config.RamSizeMB,
			DiskSize: b.config.DiskSize,
		},
		new(hypervcommon.StepEnableIntegrationService),
		&hypervcommon.StepMountDvdDrive{
			RawSingleISOUrl: b.config.RawSingleISOUrl,
		},
		new(hypervcommon.StepMountFloppydrive),
		new(hypervcommon.StepStartVm),
		&hypervcommon.StepWaitForInstallToComplete{ 
			ExpectedRebootCount: 2, 
			ActionName: "Installing",
		},

		// wait for the first post-install boot to complete
		// &hypervcommon.StepSleep{ 
		// 	Minutes: 2,
		// },
		new(hypervcommon.StepConfigureIp),

		&hypervcommon.StepSetRemoting{
			Username: "vagrant",
			Password: "vagrant",
		},

		new(hypervcommon.StepCheckRemoting),

		//new(StepUpdateIntegrationServices),

		//new(StepSysprep),

		new(hypervcommon.StepUnmountFloppyDrive),
		new(hypervcommon.StepUnmountDvdDrive),
		new(hypervcommon.StepStopVm),
		&hypervcommon.StepExportVm{
			OutputDir: b.config.OutputDir,
		},

		// the clean up actions for each step will be executed reverse order
	}

	// Run the steps.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}
	b.runner.Run(state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	return hypervcommon.NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func appendWarnings(slice []string, data ...string) []string {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]string, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func (b *Builder) checkDiskSize() error {
	if b.config.DiskSize == 0 {
		b.config.DiskSize = DefaultDiskSize
	}

	log.Println(fmt.Sprintf("%s: %v", "DiskSize", b.config.DiskSize))

	if(b.config.DiskSize < MinDiskSize ){
		return fmt.Errorf("disk_size_gb: Windows server requires disk space >= %v GB, but defined: %v", MinDiskSize, b.config.DiskSize/1024)
	} else if b.config.DiskSize > MaxDiskSize {
		return fmt.Errorf("disk_size_gb: Windows server requires disk space <= %v GB, but defined: %v", MaxDiskSize, b.config.DiskSize/1024)
	}

	return nil
}

func (b *Builder) checkRamSize() error {
	if b.config.RamSizeMB == 0 {
		b.config.RamSizeMB = DefaultRamSize
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", b.config.RamSizeMB))

	if(b.config.RamSizeMB < MinRamSize ){
		return fmt.Errorf("ram_size_mb: Windows server requires memory size >= %v MB, but defined: %v", MinRamSize, b.config.RamSizeMB)
	} else if b.config.RamSizeMB > MaxRamSize {
		return fmt.Errorf("ram_size_mb: Windows server requires memory size <= %v MB, but defined: %v", MaxRamSize, b.config.RamSizeMB)
	}

	return nil
}

func (b *Builder) checkHostAvailableMemory() string {

	var script powershell.ScriptBuilder
	script.WriteLine("(Get-WmiObject Win32_OperatingSystem).FreePhysicalMemory / 1024")

	powershell := new(powershell.PowerShellCmd)
	output, _ := powershell.Output(script.String())

	freeMB, _ := strconv.ParseFloat(output, 64)

	if (freeMB - float64(b.config.RamSizeMB)) < 512 {
		return fmt.Sprintf("Hyper-V might fail to create a VM if there is no available memory in the system.")
	}

	return ""
}
