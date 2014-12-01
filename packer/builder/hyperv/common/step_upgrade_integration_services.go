// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepUpdateIntegrationServices struct {
	Username string
	Password string

	newDvdDriveProperties dvdDriveProperties
}

type dvdDriveProperties struct {
	ControllerNumber string
	ControllerLocation string
}

func (s *StepUpdateIntegrationServices) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
    comm := state.Get("communicator").(packer.Communicator)
	vmName := state.Get("vmName").(string)
	ip := state.Get("ip").(string)

	hostVersion, err := s.getHostIntegrationServicesVersion()
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	clientVersion, err := s.getClientIntegrationServicesVersion(ip)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if hostVersion == clientVersion {
		log.Println(fmt.Sprintf("Integration Services up to date with version = %v", clientVersion))
		return multistep.ActionContinue
	}

	log.Println(fmt.Sprintf("'Host Integration Services Version' = %v", hostVersion))
	log.Println(fmt.Sprintf("'VM Integration Services Version' = %v", clientVersion))

	ui.Say("Mounting Integration Services Setup Disk...")

	err = s.mountIntegrationServicesSetupDisk(vmName);
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	dvdDriveLetter, err := s.getDvdDriveLetter(vmName)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	osArchitecture, err := s.getClientOSArchitecture(comm)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	setup := dvdDriveLetter + ":\\support\\"+osArchitecture+"\\setup.exe /quiet /norestart"

	ui.Say("Run: " + setup)

	return multistep.ActionContinue
}

func (s *StepUpdateIntegrationServices) Cleanup(state multistep.StateBag) {
	vmName := state.Get("vmName").(string)

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Set-VMDvdDrive -VMName $vmName -Path $null")

	powershell := new(powershell.PowerShellCmd)
	_ = powershell.Run(script.String(), vmName)
}

func (s *StepUpdateIntegrationServices) mountIntegrationServicesSetupDisk(vmName string) error {
	isoPath := os.Getenv("WINDIR") + "\\system32\\vmguest.iso"

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName,[string]$path)")
	script.WriteLine("Set-VMDvdDrive -VMName $vmName -Path $path")

	powershell := new(powershell.PowerShellCmd)
	err := powershell.Run(script.String(), vmName, isoPath)

	return err
}

func (s *StepUpdateIntegrationServices) getHostIntegrationServicesVersion() (string, error) {
	var script powershell.ScriptBuilder
	script.WriteLine("Get-ItemProperty \"HKLM:\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion\\Virtualization\\GuestInstaller\\Version\" | Select-Object -ExpandProperty Microsoft-Hyper-V-Guest-Installer")

	powershell := new(powershell.PowerShellCmd)
	version, err := powershell.Output(script.String())

	return version, err
}

func (s *StepUpdateIntegrationServices) getClientIntegrationServicesVersion(ip string) (string, error) {
	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$username,[string]$password,[string]$computerName)")
	script.WriteLine("$securePassword = ConvertTo-SecureString $password -AsPlainText -Force")
	script.WriteLine("$credential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $username, $securePassword")
	script.WriteLine("Invoke-Command -ScriptBlock { Get-ItemProperty \"HKLM:\\software\\microsoft\\virtual machine\\auto\" | Select-Object -ExpandProperty integrationservicesversion } -ComputerName $computerName -Credential $credential")
	
	powershell := new(powershell.PowerShellCmd)
	version, err := powershell.Output(script.String(), s.Username, s.Password, ip)

	return version, err
}

func (s *StepUpdateIntegrationServices) getDvdDriveLetter(vmName string) (string, error) {
	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Get-VMDvdDrive -VMName $vmName | Select-Object -ExpandProperty Id | Split-Path -Leaf")
	
	powershell := new(powershell.PowerShellCmd)
	version, err := powershell.Output(script.String(), vmName)

	return version, err
}

func (s *StepUpdateIntegrationServices) addDvdDrive(vmName string) (dvdDriveProperties, error) {
	var dvdProperties dvdDriveProperties

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("Add-VMDvdDrive -VMName $vmName")

	powershell := new(powershell.PowerShellCmd)
	err := powershell.Run(script.String(), vmName)
	if err != nil {
		return dvdProperties, err
	}

	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName | Where-Object {$_.Path -eq $null}).ControllerLocation")
	dvdProperties.ControllerLocation, err = powershell.Output(script.String(), vmName)
	if err != nil {
		return dvdProperties, err
	}

	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName | Where-Object {$_.Path -eq $null}).ControllerNumber")
	dvdProperties.ControllerNumber, err = powershell.Output(script.String(), vmName)
	if err != nil {
		return dvdProperties, err
	}

	return dvdProperties, nil
}

func (s *StepUpdateIntegrationServices) getClientOSArchitecture(comm packer.Communicator) (string, error) {
	var cmd packer.RemoteCmd
	var stdout, stderr bytes.Buffer

	cmd.Command = "-ScriptBlock { (Get-WmiObject Win32_OperatingSystem).OSArchitecture }"
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := comm.Start(&cmd)
	if err != nil {
		return "", err
	}

	osArchitecture := stdout.String()
	if osArchitecture == "64-bit" {
		return "amd64", nil
	}
	
	if osArchitecture == "32-bit" {
		return "x86", nil
	}

	err = fmt.Errorf("Unexpected OS Architecture: %v", osArchitecture)
	return "", nil
}

func (s *StepUpdateIntegrationServices) runIntegrationServicesSetup(vmName string) (error) {
	return nil
}

