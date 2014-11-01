// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package powershell

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"bytes"
	"io/ioutil"
)

type PowerShellCmd struct {
	PowerShellPath string
}

func Command() (*PowerShellCmd, error) {

	_, err := exec.LookPath("powershell")
	if err != nil {
		log.Fatal("Cannot find PowerShell in the path", err)
	}

	ps := &PowerShellCmd{ 
		PowerShellPath: "powershell",
	}

	log.Printf("PowerShell path: %s", ps.PowerShellPath)

	return ps, nil
}

// Output runs the PowerShell command and returns its standard output. 
func (ps *PowerShellCmd) Output(args string) (string, error) {

	command := exec.Command(ps.PowerShellPath, args)

	var stdout, stderr bytes.Buffer

	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	return stdoutString, err
}

// OutputScriptBlock runs the PowerShell script block and returns its standard output.
// The script block will be wrappend in  Invoke-Command -ScriptBlock { scriptBlock }
func (ps *PowerShellCmd) OutputScriptBlock(scriptBlock string) (string, error) {
	block := fmt.Sprintf("Invoke-Command -ScriptBlock { %v }", scriptBlock)
	output, err := ps.Output(block);
	return output, err
}

// RunScriptBlock runs the PowerShell script block 
func (ps *PowerShellCmd) RunScriptBlock(scriptBlock string) (error) {
	_, err := ps.OutputScriptBlock(scriptBlock);
	return err;
}

func (ps *PowerShellCmd) OutputFile(fileContents []byte, params ...string) (string, error) {
	filename := saveScript(fileContents);
	defer os.Remove(filename)
	
	args := createFileArgs(filename, params...)

	log.Printf("Run: %s %s", ps.PowerShellPath, args)

	var stdout, stderr bytes.Buffer
	command := exec.Command(ps.PowerShellPath, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

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

	return stdoutString, err;	
}

func (ps *PowerShellCmd) RunFile(fileContents []byte, params ...string) (error) {
	_, err := ps.OutputFile(fileContents, params...)
	return err;
}

// Version gets the major version of PowerShell
func (ps *PowerShellCmd) Version() (int64, error) {
	versionOutput, err := ps.Output("$host.Version.Major")
	if err != nil {
		return 0, err
	}
	ver, err := strconv.ParseInt(versionOutput, 10, 16)

	return ver, nil;	
}

// TODO: move outside of the powershell package
func (ps *PowerShellCmd) GetFreePhysicalMemory(block string) (freePhysicalMemory float64, err error) {

	output, err := ps.OutputScriptBlock("(Get-WmiObject Win32_OperatingSystem).FreePhysicalMemory / 1024");
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseFloat(output, 32)
    return value, err
}


func saveScript(fileContents []byte) string {
	// TODO: check error state (disk could be full)
	file, _ := ioutil.TempFile(os.TempDir(), "ps")
	file.Write(fileContents)
	_ = file.Close()

	newFilename := file.Name() + ".ps1"
	os.Rename(file.Name(), newFilename)

	return newFilename
}

func createFileArgs(filename string, params ...string) []string {
	args := make([]string,len(params)+2)
	args[0] = "-File"
	args[1] = filename

	for key, value := range params {
		args[key+2] = value
	}	

	return args;
}