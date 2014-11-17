// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
	"strconv"
	"time"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepWaitForInstallToComplete struct {
	ExpectedRebootCount uint
	ActionName string
}

func (s *StepWaitForInstallToComplete) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	if(len(s.ActionName)>0){
		ui.Say(s.ActionName + "! Waiting for OS install to complete...")
	}

	var rebootCount uint
	var lastUptime uint64

	var script ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VM -Name $vmName).Uptime.TotalSeconds")

	uptimeScript := script.Bytes()

	for rebootCount < s.ExpectedRebootCount {
		powershell, err := powershell.Command()
		cmdOut, err := powershell.OutputFile(uptimeScript, vmName);
		if err != nil {
			err := fmt.Errorf("Error checking uptime: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		uptime, _ := strconv.ParseUint(strings.TrimSpace(string(cmdOut)), 10, 64)
		if uint64(uptime) < lastUptime {
			rebootCount++
			ui.Say(s.ActionName + "  -> Detected reboot "+fmt.Sprintf("%v",rebootCount)+" after "+fmt.Sprintf("%v",lastUptime)+" seconds...")
		} else {
			//ui.Say(s.ActionName + "  ->  Uptime "+fmt.Sprintf("%v",uptime)+" seconds...")
		}

		lastUptime = uptime

		if (rebootCount < s.ExpectedRebootCount) {
			time.Sleep(time.Second);
		}
	}


	return multistep.ActionContinue
}

func (s *StepWaitForInstallToComplete) Cleanup(state multistep.StateBag) {

}
