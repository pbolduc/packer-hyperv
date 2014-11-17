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

type StepDisableVlan struct {
}

func (s *StepDisableVlan) Run(state multistep.StateBag) multistep.StepAction {
	//config := state.Get("config").(*config)
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error disabling vlan: %s"
	vmName := state.Get("vmName").(string)
	switchName := state.Get("SwitchName").(string)
	var err error

	powershell := new(powershell.PowerShellCmd)

	ui.Say("Disabling vlan...")

	var script ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Set-VMNetworkAdapterVlan -VMName $vmName -Untagged")

	err = powershell.RunFile(script.Bytes(), vmName)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	script.Reset()
	script.WriteLine("param([string]$switchName)")
	script.WriteLine("Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName $switchName -Untagged")

	err = powershell.RunFile(script.Bytes(), switchName)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepDisableVlan) Cleanup(state multistep.StateBag) {
	//do nothing
}
