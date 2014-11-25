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

type StepStopVm struct {
}

func (s *StepStopVm) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error stopping vm: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Stopping vm...")

	powershell := new(powershell.PowerShellCmd)

	var script ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("$vm = Get-VM -Name $vmName")
	script.WriteLine("if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {")
	script.WriteLine("    Stop-VM -VM $vm")
	script.WriteLine("}")

	err := powershell.Run(script.String(), vmName)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStopVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
