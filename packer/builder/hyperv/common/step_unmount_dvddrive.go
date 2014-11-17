// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)


type StepUnmountDvdDrive struct {
	path string
}

func (s *StepUnmountDvdDrive) Run(state multistep.StateBag) multistep.StepAction {
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)
	powershell := new(powershell.PowerShellCmd)
	
	ui.Say("Unmounting dvd drive...")

	var script ScriptBuilder
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
