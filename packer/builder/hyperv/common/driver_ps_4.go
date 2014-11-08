// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"log"
	"strings"
	"runtime"
	"strconv"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)

type HypervPS4Driver struct {
}

func NewHypervPS4Driver() (Driver, error) {
	appliesTo := "Applies to Windows 8.1, Windows PowerShell 4.0, Windows Server 2012 R2 only"

	// Check this is Windows
	if runtime.GOOS != "windows" {
		err := fmt.Errorf("%s", appliesTo)
		return nil, err
	}

	ps4Driver := &HypervPS4Driver { }

	if err := ps4Driver.Verify(); err != nil {
		return nil, err
	}

	return ps4Driver, nil
}

func (d *HypervPS4Driver) Verify() error {

	if err := d.verifyPSVersion(); err != nil {
		return err
	}

	if err := d.verifyPSHypervModule(); err != nil {
		return err
	}

	if err := d.verifyElevatedMode(); err != nil {
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyPSVersion() error {

	log.Printf("Enter method: %s", "verifyPSVersion")
	// check PS is available and is of proper version
	versionCmd := "$host.version.Major"
	powershell, _ := powershell.Command()

	cmdOut, err := powershell.Output(versionCmd)
	if err != nil {
		return err
	}

	versionOutput := strings.TrimSpace(string(cmdOut))
	log.Printf("%s output: %s", versionCmd, versionOutput)

	ver, err := strconv.ParseInt(versionOutput, 10, 32)

	if  err != nil {
		return err
	}

	if ver < 4 {
		err := fmt.Errorf("%s", "Windows PowerShell version 4.0 or higher is expected")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyPSHypervModule() error {

	log.Printf("Enter method: %s", "verifyPSHypervModule")

	versionCmd := "function foo(){try{ $commands = Get-Command -Module Hyper-V;if($commands.Length -eq 0){return $false} }catch{return $false}; return $true} foo"

	powershell, _ := powershell.Command()
	cmdOut, err := powershell.OutputScriptBlock(versionCmd)
	if err != nil {
		return err
	}

	res := strings.TrimSpace(string(cmdOut))

	if(res== "False"){
		err := fmt.Errorf("%s", "PS Hyper-V module is not loaded. Make sure Hyper-V feature is on.")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyElevatedMode() error {

	log.Printf("Enter method: %s", "verifyElevatedMode")

	powershell, err := powershell.Command()

	ps1, err := Asset("scripts/is_current_user_administrator.ps1")
	if err != nil {
		err := fmt.Errorf("Could not load script scripts/IsAdministrator.ps1: %s", err)
		return err
	}

	cmdOut, err := powershell.OutputFile(ps1);
	if err != nil {
		return err
	}

	res := strings.TrimSpace(string(cmdOut))
	log.Printf("cmdOut: " + string(res))

	if(res == "False"){
		err := fmt.Errorf("%s", "Please restart your shell in elevated mode")
		return err
	}

	return nil
}
