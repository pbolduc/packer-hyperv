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
	ip string
}

func (s *StepSetRemoting) Run(state multistep.StateBag) multistep.StepAction {
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error StepRemoteSession: %s"
	vmName := state.Get("vmName").(string)
	ip := state.Get("ip").(string)


	ui.Say("Adding to TrustedHosts (requires elevated mode)")

	var script ps.ScriptBuilder
	script.WriteLine("param([string]$ip)")
	script.WriteLine("Set-Item -path WSMan:\\localhost\\Client\\TrustedHosts $ip -Force -Concatenate")

	ps := new(ps.PowerShellCmd)
	err := ps.Run(script.String(), ip)

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
			RemoteHostIP: ip,
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
	s.ip = ip
	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetRemoting) Cleanup(state multistep.StateBag) {

	if s.ip == "" {
		return
	}

	var script ps.ScriptBuilder
	script.WriteLine("param([string]$ip)")
	script.WriteLine("[System.Collections.ArrayList] $hosts = (Get-Item -Path WSMan:\\localhost\\Client\\TrustedHosts).Value.Split(',')")
	script.WriteLine("$hosts.Remove($ip)")
	script.WriteLine("$newTrustedHosts = $hosts.ToArray() -Join ','")
	script.WriteLine("Set-Item -Path WSMan:\\localhost\\Client\\TrustedHosts -Value $newTrustedHosts -Force")

	ps := new(ps.PowerShellCmd)
	_ = ps.Run(script.String(), s.ip)
}
