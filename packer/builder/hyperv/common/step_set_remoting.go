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
	powershell "github.com/MSOpenTech/packer-hyperv/packer/communicator/powershell"
	ps "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepSetRemoting struct {
	comm packer.Communicator
}

func (s *StepSetRemoting) Run(state multistep.StateBag) multistep.StepAction {
	//driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error StepRemoteSession: %s"
	vmName := state.Get("vmName").(string)
	ip := state.Get("ip").(string)

	ps, _ := ps.Command()

	ui.Say("Adding to TrustedHosts (require elevated mode)")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("param([string]$ip)")
	blockBuffer.WriteString("Invoke-Command -scriptblock { Set-Item -path WSMan:\\localhost\\Client\\TrustedHosts $ip -Force }")

	var err error
	err = ps.RunFile(blockBuffer.Bytes(), ip)

	if err != nil {
		err := fmt.Errorf(errorMsg, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	comm, err := powershell.New(
		&powershell.Config{
			Username: "vagrant",
			Password: "vagrant",
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
	state.Put("communicator", packerCommunicator)

	return multistep.ActionContinue
}

func (s *StepSetRemoting) Cleanup(state multistep.StateBag) {
	// do nothing
}
