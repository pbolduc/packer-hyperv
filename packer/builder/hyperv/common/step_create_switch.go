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

	powershell := new(powershell.PowerShellCmd)

	ui.Say(fmt.Sprintf("Creating %v switch...", s.SwitchType))

	var script ScriptBuilder
	script.WriteLine("param([string]$switchName,[string]$switchType)")
	script.WriteLine("$switches = Get-VMSwitch -Name $switchName -ErrorAction SilentlyContinue")
	script.WriteLine("if ($switches.Count -eq 0) {")
	script.WriteLine("  New-VMSwitch -Name $switchName -SwitchType $switchType")
	script.WriteLine("}")

	err := powershell.RunFile(script.Bytes(), s.SwitchName, s.SwitchType)

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

	powershell := new(powershell.PowerShellCmd)

	ui.Say("Unregistering and deleting switch...")

	var script ScriptBuilder
	script.WriteLine("param([string]$switchName)")
	script.WriteLine("Remove-VMSwitch $switchName -Force")

	err := powershell.RunFile(script.Bytes(), s.SwitchName)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting switch: %s", err))
	}
}

func writeLine(buffer bytes.Buffer, s string) {
	buffer.WriteString(s)
	buffer.WriteString("\n")
}