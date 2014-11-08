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
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
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
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if len(s.SwitchType) == 0 {
		s.SwitchType = DefaultSwitchType
	}

	powershell, _ := powershell.Command()

	ui.Say(fmt.Sprintf("Creating %v switch...", s.SwitchType))

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("param([string]$switchName,[string]$switchType)")
	blockBuffer.WriteString("$switches = Get-VMSwitch -Name $switchName -ErrorAction SilentlyContinue")
	blockBuffer.WriteString("if ($switches.Count -eq 0) {")
	blockBuffer.WriteString("  New-VMSwitch -Name $switchName -SwitchType $switchType")
	blockBuffer.WriteString("}")

	err := powershell.RunFile(blockBuffer.Bytes(), s.SwitchName, s.SwitchType)

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

	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	powershell, _ := powershell.Command()

	ui.Say("Unregistering and deleting switch...")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("param([string]$switchName)")
	blockBuffer.WriteString("Remove-VMSwitch $switchName -Force}")

	err := powershell.RunFile(blockBuffer.Bytes(), s.SwitchName)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting switch: %s", err))
	}
}
