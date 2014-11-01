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

type StepEnableIntegrationService struct {
	name string
}

func (s *StepEnableIntegrationService) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	vmName := state.Get("vmName").(string)
	s.name = "Guest Service Interface"

	ui.Say("Enabling Integration Service...")

	powershell, err := powershell.Command()
	ps1, err := Asset("scripts/Enable-VMIntegrationService.ps1")
	if err != nil {
		err := fmt.Errorf("Could not load script scripts/Enable-VMIntegrationService.ps1: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	err = powershell.RunFile(ps1, vmName, s.name)

	if err != nil {
		err := fmt.Errorf("Error enabling Integration Service: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepEnableIntegrationService) Cleanup(state multistep.StateBag) {
	// do nothing
}
