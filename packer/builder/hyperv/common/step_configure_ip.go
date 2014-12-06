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
	"time"
	"log"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)


type StepConfigureIp struct {
}

func (s *StepConfigureIp) Run(state multistep.StateBag) multistep.StepAction {
//	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error configuring ip address: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Configuring ip address...")

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("try {")
	script.WriteLine("  $adapter = Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue")
	script.WriteLine("  $ip = $adapter.IPAddresses[0]")
	script.WriteLine("  if($ip -eq $null) {")
	script.WriteLine("    return $false")
	script.WriteLine("  }")
	script.WriteLine("} catch {")
	script.WriteLine("  return $false")
	script.WriteLine("}")
	script.WriteLine("$ip")

	count := 60
	var duration time.Duration = 1
	sleepTime := time.Minute * duration
	var ip string

	for count != 0 {
		powershell := new(powershell.PowerShellCmd)
		cmdOut, err := powershell.Output(script.String(), vmName);
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ip = strings.TrimSpace(string(cmdOut))

		if ip != "False" {
			break;
		}

		log.Println(fmt.Sprintf("Waiting for another %v minutes...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if(count == 0){
		err := fmt.Errorf(errorMsg, "IP address assigned to the adapter is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("ip address is " + ip)

	hostName, err := s.getHostName(ip);
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("hostname is " + hostName)

	state.Put("ip", ip)
	state.Put("hostname", hostName)

	return multistep.ActionContinue
}

func (s *StepConfigureIp) Cleanup(state multistep.StateBag) {
	// do nothing
}


func (s *StepConfigureIp) getHostName(ip string) (string, error) {

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$ip)")
	script.WriteLine("try {")
	script.WriteLine("  $HostName = [System.Net.Dns]::GetHostEntry($ip).HostName")
	script.WriteLine("  if ($HostName -ne $null) {")
	script.WriteLine("    $HostName = $HostName.Split('.')[0]")
	script.WriteLine("  }")
	script.WriteLine("  $HostName")
	script.WriteLine("} catch { }")

	//
	powershell := new(powershell.PowerShellCmd)

	cmdOut, err := powershell.Output(script.String(), ip);
	if err != nil {
		return "", err
	}

	return cmdOut, nil
}