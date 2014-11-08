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

	powershell, err := powershell.Command()
	ps1, err := Asset("scripts/stop_vm.ps1")
	if err != nil {
		err := fmt.Errorf("Could not load script scripts/Stop-VM.ps1: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	err = powershell.RunFile(ps1, vmName)
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
