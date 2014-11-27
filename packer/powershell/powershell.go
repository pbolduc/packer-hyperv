// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package powershell

import (
	"fmt"
	"log"
	"io"
	"os"
	"os/exec"
	"strings"
	"bytes"
	"io/ioutil"
)

type PowerShellCmd struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (ps *PowerShellCmd) Run(fileContents string, params ...string) error {
	_, err := ps.Output(fileContents, params...)
	return err
}

// Output runs the PowerShell command and returns its standard output. 
func (ps *PowerShellCmd) Output(fileContents string, params ...string) (string, error) {
	path, err := ps.getPowerShellPath();
	if err != nil {
		return "", nil
	}

	filename, err := saveScript(fileContents);
	if err != nil {
		return "", err
	}

	debug := os.Getenv("PACKER_POWERSHELL_DEBUG") != ""
	verbose := debug || os.Getenv("PACKER_POWERSHELL_VERBOSE") != ""

	if !debug {
		defer os.Remove(filename)
	}
	
	args := createArgs(filename, params...)

	if verbose {
		log.Printf("Run: %s %s", path, args)
	}

	var stdout, stderr bytes.Buffer
	command := exec.Command(path, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err = command.Run()

	if ps.Stdout != nil {
		stdout.WriteTo(ps.Stdout)
	}

	if ps.Stderr != nil {
		stderr.WriteTo(ps.Stderr)
	}

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	if verbose && stdoutString != "" {
		log.Printf("stdout: %s", stdoutString)
	}

	// only write the stderr string if verbose because
	// the error string will already be in the err return value.
	if verbose && stderrString != "" {
		log.Printf("stderr: %s", stderrString)
	}

	return stdoutString, err;	
}

func (ps *PowerShellCmd) getPowerShellPath() (string, error) {
	path, err := exec.LookPath("powershell")
	if err != nil {
		log.Fatal("Cannot find PowerShell in the path", err)
		return "", err
	}

	return path, nil
}

func saveScript(fileContents string) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "ps")
	if err != nil {
		return "", err
	}
	
	_, err = file.Write([]byte(fileContents))
	if err != nil {
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	newFilename := file.Name() + ".ps1"
	err = os.Rename(file.Name(), newFilename)
	if err != nil {
		return "", err
	}

	return newFilename, nil
}

func createArgs(filename string, params ...string) []string {
	args := make([]string,len(params)+4)
	args[0] = "-ExecutionPolicy"
	args[1] = "Bypass"

	args[2] = "-File"
	args[3] = filename

	for key, value := range params {
		args[key+4] = value
	}	

	return args;
}
