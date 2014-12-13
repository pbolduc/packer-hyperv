// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	//"fmt"
	"os"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type StepMountSecondaryDvdImages struct {
	Files [] string
	dvdProperties []dvdControllerProperties
}

type dvdControllerProperties struct {
	controllerNumber string
	controllerLocation string
}

func (s *StepMountSecondaryDvdImages) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Mounting secondary DVD images...")

	// should be able to mount up to 60 additional iso images using SCSI
	// but Windows would only allow a max of 22 due to available drive letters
	// Will Windows assign DVD drives to A: and B: ?
	dvdProperties, err := s.mountFiles(vmName);
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	
	state.Put("secondary.dvd.properties", dvdProperties)

	return multistep.ActionContinue
}

func (s *StepMountSecondaryDvdImages) Cleanup(state multistep.StateBag) {

}


func (s *StepMountSecondaryDvdImages) mountFiles(vmName string) ([]dvdControllerProperties, error) {

	var dvdProperties []dvdControllerProperties

	properties, err := s.addAndMountIntegrationServicesSetupDisk(vmName)
	if err != nil {
		return dvdProperties, err
	}

	dvdProperties = append(dvdProperties, properties)

	for _, value := range s.Files {
		properties, err := s.addAndMountDvdDisk(vmName, value)
		if err != nil {
			return dvdProperties, err
		}

		dvdProperties = append(dvdProperties, properties)
	}

	return dvdProperties, nil
}


func (s *StepMountSecondaryDvdImages) addAndMountIntegrationServicesSetupDisk(vmName string) (dvdControllerProperties, error) {

	isoPath := os.Getenv("WINDIR") + "\\system32\\vmguest.iso"
	properties, err := s.addAndMountDvdDisk(vmName, isoPath)
	if err != nil {
		return properties, err
	}

	return properties, nil
}




func (s *StepMountSecondaryDvdImages) addAndMountDvdDisk(vmName string, isoPath string) (dvdControllerProperties, error) {

	var properties dvdControllerProperties
	var script powershell.ScriptBuilder
	powershell := new(powershell.PowerShellCmd)

	// get the controller number that the OS install disk is mounted on	
	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName).ControllerNumber")
	controllerNumber, err := powershell.Output(script.String(), vmName)
	if err != nil {
		return properties, err
	}

	script.Reset()
	script.WriteLine("param([string]$vmName,[int]$controllerNumber)")
	script.WriteLine("Add-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber")
	err = powershell.Run(script.String(), vmName, controllerNumber)
	if err != nil {
		return properties, err
	}

	// we could try to get the controller location and number in one call, but this way we do not
	// need to parse the output
	script.Reset()
	script.WriteLine("param([string]$vmName)")
	script.WriteLine("(Get-VMDvdDrive -VMName $vmName | Where-Object {$_.Path -eq $null}).ControllerLocation")
	controllerLocation, err := powershell.Output(script.String(), vmName)
	if err != nil {
		return properties, err
	}

	script.Reset()
	script.WriteLine("param([string]$vmName,[string]$path,[string]$controllerNumber,[string]$controllerLocation)")
	script.WriteLine("Set-VMDvdDrive -VMName $vmName -Path $path -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation")

	err = powershell.Run(script.String(), vmName, isoPath, controllerNumber, controllerLocation)
	if err != nil {
		return properties, err
	}

	properties.controllerNumber = controllerNumber
	properties.controllerLocation = controllerLocation

	return properties, nil
}
