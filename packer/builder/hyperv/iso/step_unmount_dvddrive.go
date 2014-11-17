// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
	common "github.com/MSOpenTech/packer-hyperv/packer/builder/hyperv/common"
)


type StepUnmountDvdDrive struct {
	path string
}

func (s *StepUnmountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	//config := state.Get("config").(*config)
	//driver := state.Get("driver").(hypervcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)
	powershell, _ := powershell.Command()
	
	ui.Say("Unmounting dvd drive...")

	var script common.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Set-VMDvdDrive -VMName $vmName -Path $null")

	err := powershell.RunFile(script.Bytes(), vmName)

	if err != nil {
		err := fmt.Errorf("Error unmounting dvd drive: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.path = ""

	return multistep.ActionContinue
}

func (s *StepUnmountDvdDrive) Cleanup(state multistep.StateBag) {
}
