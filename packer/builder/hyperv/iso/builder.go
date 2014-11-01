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

	"github.com/mitchellh/multistep"
	hypervcommon "github.com/MSOpenTech/packer-hyperv/packer/builder/hyperv/common"
//	msbldcommon "github.com/MSOpenTech/packer-hyperv/packer/builder/common"
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
	DiskSize            uint     			`mapstructure:"disk_size"`
	RamSizeMB           uint     			`mapstructure:"ram_size_mb"`
	FloppyFiles         []string            `mapstructure:"floppy_files"`	
	GuestOSType         string   			`mapstructure:"guest_os_type"`
	ISOChecksum         string              `mapstructure:"iso_checksum"`
	ISOChecksumType     string              `mapstructure:"iso_checksum_type"`
	ISOUrls             []string            `mapstructure:"iso_urls"`
	VMName              string              `mapstructure:"vm_name"`

	RawSingleISOUrl 	string 				`mapstructure:"iso_url"`

	SleepTimeMinutes 	time.Duration		`mapstructure:"wait_time_minutes"`
	ProductKey 			string				`mapstructure:"product_key"`

	common.PackerConfig           			`mapstructure:",squash"`
	hypervcommon.OutputConfig     			`mapstructure:",squash"`

	SwitchName          string

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

	if b.config.DiskSize == 0 {
		b.config.DiskSize = DefaultDiskSize
	}
	log.Println(fmt.Sprintf("%s: %v", "DiskSize", b.config.DiskSize))

	if(b.config.DiskSize < MinDiskSize ){
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("disk_size_gb: Windows server requires disk space >= %v GB, but defined: %v", MinDiskSize, b.config.DiskSize /1024))
	} else if b.config.DiskSize > MaxDiskSize {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("disk_size_gb: Windows server requires disk space <= %v GB, but defined: %v", MaxDiskSize, b.config.DiskSize/1024))
	}

	if b.config.RamSizeMB == 0 {
		b.config.RamSizeMB = DefaultRamSize
	}

	log.Println(fmt.Sprintf("%s: %v", "RamSize", b.config.RamSizeMB))

	if(b.config.RamSizeMB < MinRamSize ){
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("ram_size_mb: Windows server requires memory size >= %v MB, but defined: %v", MinRamSize, b.config.RamSizeMB))
	} else if b.config.RamSizeMB > MaxRamSize {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("ram_size_mb: Windows server requires memory size <= %v MB, but defined: %v", MaxRamSize, b.config.RamSizeMB))
	}

	// todo: get host memory using PowerShell: Invoke-Command -ScriptBlock { (Get-WmiObject Win32_OperatingSystem).FreePhysicalMemory / 1024
	warnings = appendWarnings( warnings, fmt.Sprintf("Hyper-V might fail to create a VM if there is no available memory in the system."))


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


// 	guestOSTypesIsValid := false
// 	guestOSTypes := []string{
// 		WS2012R2DC,
// //		WS2012R2St,
// 	}

// 	log.Println(fmt.Sprintf("%s: %v","GuestOSType", b.config.GuestOSType))

// 	for _, guestType := range guestOSTypes {
// 		if b.config.GuestOSType == guestType {
// 			guestOSTypesIsValid = true
// 			break
// 		}
// 	}

// 	if !guestOSTypesIsValid {
// 		errs = packer.MultiErrorAppend(errs,
// 			fmt.Errorf("guest_os_type: The value is invalid. Must be one of: %v", guestOSTypes))
// 	}

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

		&common.StepCreateFloppy{ Files: b.config.FloppyFiles },
		&hypervcommon.StepCreateSwitch{ SwitchName: b.config.SwitchName },

		new(StepCreateVM),
		new(hypervcommon.StepEnableIntegrationService),
		new(StepMountDvdDrive),
		new(StepMountFloppydrive),
//		new(hypervcommon.StepConfigureVlan),
		new(hypervcommon.StepStartVm),
		&hypervcommon.StepSleep{ Minutes: b.config.SleepTimeMinutes, ActionName: "Installing" },

		new(hypervcommon.StepConfigureIp),
		new(hypervcommon.StepSetRemoting),
		new(common.StepProvision),
//		new(StepInstallProductKey),

		new(StepExportVm),

//		new(hypervcommon.StepConfigureIp),
//		new(hypervcommon.StepSetRemoting),
//		new(hypervcommon.StepCheckRemoting),
//		new(msbldcommon.StepSysprep),
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

