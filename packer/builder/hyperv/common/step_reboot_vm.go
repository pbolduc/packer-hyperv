// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepRebootVm struct {
}

func (s *StepRebootVm) Run(state multistep.StateBag) multistep.StepAction {
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error rebooting vm: %s"
	vmName := state.Get("vmName").(string)

	powershell, _ := powershell.Command()

	ui.Say("Rebooting vm...")

	var script ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Restart-VM $vmName -Force")

	err := powershell.RunFile(script.Bytes(), vmName)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting the VM to complete rebooting (2 minutes)...")

	sleepTime := time.Minute * 2
	time.Sleep(sleepTime)

	return multistep.ActionContinue
}

func (s *StepRebootVm) Cleanup(state multistep.StateBag) {
	// do nothing
}
