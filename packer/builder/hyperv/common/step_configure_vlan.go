// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/MSOpenTech/packer-hyperv/packer/powershell/hyperv"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

//const (
//	vlanId = "130"
//)

//* added block
type StepConfigureVlan struct {
	VlanID string
}

func (s *StepConfigureVlan) Run(state multistep.StateBag) multistep.StepAction {
	//config := state.Get("config").(*config)
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error configuring vlan: %s"
	vmName := state.Get("vmName").(string)
	//switchName := state.Get("SwitchName").(string)

	ui.Say("Configuring vlan...")

	/*err := hyperv.SetNetworkAdapterVlanId(switchName, vlanId)
	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}*/
	//* added block
	if s.VlanID == "" {
		ui.Say("Coundn't config vlan ... ")
	}
	// change vlad param
	//err := hyperv.SetVirtualMachineVlanId(vmName, vlanId)
	err := hyperv.SetVirtualMachineVlanId(vmName, s.VlanID)
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
