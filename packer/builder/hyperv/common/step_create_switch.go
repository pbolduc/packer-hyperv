// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

const (
	SwitchTypeInternal = "Internal"
	SwitchTypePrivate = "Private"
	DefaultSwitchType = SwitchTypeInternal
)

// This step creates switch for VM.
//
// Produces:
//   SwitchName string - The name of the Switch
type StepCreateSwitch struct {
	// Specifies the name of the switch to be created.
	SwitchName     string
	// Specifies the type of the switch to be created. Allowed values are Internal and Private. To create an External
	// virtual switch, specify either the NetAdapterInterfaceDescription or the NetAdapterName parameter, which
	// implicitly set the type of the virtual switch to External.
	SwitchType     string
	// Specifies the name of the network adapter to be bound to the switch to be created.
	NetAdapterName string
	// Specifies the interface description of the network adapter to be bound to the switch to be created.
	NetAdapterInterfaceDescription string
}

func (s *StepCreateSwitch) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if len(s.SwitchType) == 0 {
		s.SwitchType = DefaultSwitchType
	}

	ui.Say(fmt.Sprintf("Creating %v switch...", s.SwitchType))

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {$TestSwitch = Get-VMSwitch -Name '")
	blockBuffer.WriteString(s.SwitchName)
	blockBuffer.WriteString("' -ErrorAction SilentlyContinue; if ($TestSwitch.Count -eq 0){New-VMSwitch -Name '")
	blockBuffer.WriteString(s.SwitchName)
	blockBuffer.WriteString("' -SwitchType ")
	blockBuffer.WriteString(s.SwitchType)
	blockBuffer.WriteString(" }}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		err := fmt.Errorf("Error creating switch: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		s.SwitchName = "";
		return multistep.ActionHalt
	}

	// Set the final name in the state bag so others can use it
	state.Put("SwitchName", s.SwitchName)

	return multistep.ActionContinue
}

func (s *StepCreateSwitch) Cleanup(state multistep.StateBag) {
	if len(s.SwitchName) == 0 {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unregistering and deleting switch...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Remove-VMSwitch '")
	blockBuffer.WriteString(s.SwitchName)
	blockBuffer.WriteString("' -Force}")

	err := driver.HypervManage( blockBuffer.String() )

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting switch: %s", err))
	}
}
