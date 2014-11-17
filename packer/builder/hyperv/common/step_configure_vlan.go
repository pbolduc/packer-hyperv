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


const(
	vlanId = "1724"
)

type StepConfigureVlan struct {
}

func (s *StepConfigureVlan) Run(state multistep.StateBag) multistep.StepAction {
	//config := state.Get("config").(*config)
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error configuring vlan: %s"
	vmName := state.Get("vmName").(string)
	switchName := state.Get("SwitchName").(string)

	powershell := new(powershell.PowerShellCmd)

	ui.Say("Configuring vlan...")

	var script ScriptBuilder
	script.WriteLine("param([string]$networkAdapterName,[string]$vlanId)")
	script.WriteLine("Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName $networkAdapterName -Access -VlanId $vlanId")

	err := powershell.RunFile(script.Bytes(), switchName, vlanId)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	script.Reset()
	script.WriteLine("param([string]$vmName,[string]$vlanId)")
	script.WriteLine("Set-VMNetworkAdapterVlan -VMName $vmName -Access -VlanId $vlanId")

	err = powershell.RunFile(script.Bytes(), vmName, vlanId)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigureVlan) Cleanup(state multistep.StateBag) {
	//do nothing
}
