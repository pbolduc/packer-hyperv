// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"fmt"
	"strconv"
	"github.com/mitchellh/multistep"
	hypervcommon "github.com/MSOpenTech/packer-hyperv/packer/builder/hyperv/common"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"

	"github.com/mitchellh/packer/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//   vmName string - The name of the VM
type StepCreateVM struct {
	vmName string
}

func (s *StepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating virtual machine...")

	vmName := config.VMName
	path :=	state.Get("packerTempDir").(string)

	// convert the MB to bytes
	ramBytes := int64(config.RamSizeMB * 1024 * 1024)
	diskSizeBytes := int64(config.DiskSize * 1024 * 1024)

	ram := strconv.FormatInt(ramBytes, 10)
	diskSize := strconv.FormatInt(diskSizeBytes, 10)
	switchName := config.SwitchName

	powershell, _ := powershell.Command()

	var script hypervcommon.ScriptBuilder
	script.WriteLine("param([string]$vmName, [string]$path, [long]$memoryStartupBytes, [long]$newVHDSizeBytes, [string]$switchName)")
	script.WriteLine("$vhdx = $vmName + '.vhdx'")
	script.WriteLine("$vhdPath = Join-Path -Path $path -ChildPath $vhdx")
	script.WriteLine("New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $newVHDSizeBytes -SwitchName $switchName")

	err := powershell.RunFile(script.Bytes(), vmName, path, ram, diskSize, switchName)
	if err != nil {
		err := fmt.Errorf("Error creating virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the VM name property on the first command
	if s.vmName == "" {
		s.vmName = vmName
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.vmName)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmName == "" {
		return
	}

	//driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	powershell, _ := powershell.Command()

	ui.Say("Unregistering and deleting virtual machine...")

	var err error = nil

	var script hypervcommon.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Remove-VM -Name $vmName -Force")

	err = powershell.RunFile(script.Bytes(), s.vmName)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}
