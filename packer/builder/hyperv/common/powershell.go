// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"strconv"
	"bytes"
)

type PowerShell interface {
	Output(string) (string, error)
	OutputScriptBlock(string) (string, error)
	RunScriptBlock(string) (error)

	Version() (int64, error)
}

type PowerShellv4 struct {
	PowerShellPath string
}

func NewPowerShellv4() (PowerShell, error) {

	path, err := exec.LookPath("powershell")
	if err != nil {
		log.Fatal("Cannot find PowerShell in the path", err)
	}

	powershell := &PowerShellv4{ PowerShellPath: path}

	log.Printf("PowerShell path: %s", powershell.PowerShellPath)

	return powershell, nil
}

// Output runs the PowerShell command and returns its standard output. 
func (ps *PowerShellv4) Output(args string) (string, error) {
	log.Printf("Executing PowerShell: %#v", args)

	var stdout, stderr bytes.Buffer

	script := exec.Command(ps.PowerShellPath, args)
	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return stdoutString, err
}

// OutputScriptBlock runs the PowerShell script block and returns its standard output.
// The script block will be wrappend in  Invoke-Command -ScriptBlock { scriptBlock }
func (ps *PowerShellv4) OutputScriptBlock(scriptBlock string) (string, error) {
	block := fmt.Sprintf("Invoke-Command -ScriptBlock { %v }", scriptBlock)
	output, err := ps.Output(block);
	return output, err
}

// RunScriptBlock runs the PowerShell script block 
func (ps *PowerShellv4) RunScriptBlock(scriptBlock string) (error) {
	_, err := ps.OutputScriptBlock(scriptBlock);
	return err;
}

// Version gets the version of PowerShell
func (ps *PowerShellv4) Version() (int64, error) {
	versionOutput, err := ps.Output("$host.Version.Major")
	if err != nil {
		return 0, err
	}
	ver, err := strconv.ParseInt(versionOutput, 10, 16)

	return ver, nil;	
}

// TODO: move outside of the powershell package
func (ps *PowerShellv4) GetFreePhysicalMemory(block string) (freePhysicalMemory float64, err error) {

	output, err := ps.OutputScriptBlock("(Get-WmiObject Win32_OperatingSystem).FreePhysicalMemory / 1024");
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(output, 32)
    return value, err
}
