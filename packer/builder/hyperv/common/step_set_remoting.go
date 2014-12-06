// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/communicator/powershell"
	ps "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepSetRemoting struct {
	Username string
	Password string

	comm packer.Communicator
	trustedHost string
}

func (s *StepSetRemoting) Run(state multistep.StateBag) multistep.StepAction {
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error StepRemoteSession: %s"
	vmName := state.Get("vmName").(string)

	s.trustedHost = state.Get("hostname").(string)
	if s.trustedHost == "" {
		s.trustedHost = state.Get("ip").(string)
	}

	ui.Say("Adding '"+s.trustedHost+"' to TrustedHosts (requires elevated mode)")

	var script ps.ScriptBuilder
	script.WriteLine("param([string]$trustedHost)")
	script.WriteLine("Set-Item -path WSMan:\\localhost\\Client\\TrustedHosts $trustedHost -Force -Concatenate")

	ps := new(ps.PowerShellCmd)
	err := ps.Run(script.String(), s.trustedHost)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	comm, err := powershell.New(
		&powershell.Config{
			Username: s.Username,
			Password: s.Password,
			RemoteHost: s.trustedHost,
			VmName: vmName,
			Ui: ui,
		})

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	packerCommunicator := packer.Communicator(comm)

	s.comm = packerCommunicator
	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetRemoting) Cleanup(state multistep.StateBag) {

	if s.trustedHost == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Removing '"+s.trustedHost+"' from TrustedHosts")

	var script ps.ScriptBuilder
	script.WriteLine("param([string]$trustedHost)")
	script.WriteLine("[System.Collections.ArrayList] $hosts = (Get-Item -Path WSMan:\\localhost\\Client\\TrustedHosts).Value.Split(',')")
	script.WriteLine("$hosts.Remove($trustedHost)")
	script.WriteLine("$newTrustedHosts = $hosts.ToArray() -Join ','")
	script.WriteLine("Set-Item -Path WSMan:\\localhost\\Client\\TrustedHosts -Value $newTrustedHosts -Force")

	ps := new(ps.PowerShellCmd)
	_ = ps.Run(script.String(), s.trustedHost)
}
