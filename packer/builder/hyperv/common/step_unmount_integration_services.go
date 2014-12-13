// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepUnmountIntegrationServices struct {
}

func (s *StepUnmountIntegrationServices) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// todo: should this message say removing the dvd?
	ui.Say("Unmounting Integration Services Setup Disk...")

	controllerNumber := state.Get("integration.services.dvd.controller.number").(string)
	controllerLocation := state.Get("integration.services.dvd.controller.location").(string)

	var script powershell.ScriptBuilder
	powershell := new(powershell.PowerShellCmd)

	script.WriteLine("param([string]$vmName,[int]$controllerNumber,[int]$controllerLocation)")
	script.WriteLine("Remove-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation")
	err := powershell.Run(script.String(), vmName, controllerNumber, controllerLocation)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUnmountIntegrationServices) Cleanup(state multistep.StateBag) {
}
