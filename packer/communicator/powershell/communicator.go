// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package powershell

import (
//	"bufio"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
//	"io/ioutil"
	"log"
	"path/filepath"
	"os"
	"container/list"
	powershell "github.com/MSOpenTech/packer-hyperv/packer/powershell"
)


type comm struct {
	config *Config
}

type Config struct {
	Username string
	Password string
	RemoteHostIP string
	VmName string
	Ui packer.Ui
}

// Creates a new packer.Communicator implementation over SSH. This takes
// an already existing TCP connection and SSH configuration.
func New(config *Config) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	username := c.config.Username
	password := c.config.Password
	remoteHost := c.config.RemoteHostIP

	log.Printf(fmt.Sprintf("Executing remote script..."))

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$username,[string]$password,[string]$computerName)")
	script.WriteLine("$securePassword = ConvertTo-SecureString $password -AsPlainText -Force")
	script.WriteLine("$credential = New-Object -TypeName System.Management.Automation.PSCredential -ArgumentList $username, $securePassword")
	script.WriteString("Invoke-Command -ComputerName $computerName ")
	script.WriteString(cmd.Command)
	script.WriteString(" -Credential $credential")

	powershell := new(powershell.PowerShellCmd)

	if cmd.Stdout  != nil {
		powershell.Stdout = cmd.Stdout
	}

	if cmd.Stderr != nil {
		powershell.Stderr = cmd.Stderr
	}

	err = powershell.Run(script.String(), username, password, remoteHost)

	return err
}

func (c *comm) Upload(string, io.Reader, *os.FileInfo) error {
	panic("not implemented for powershell")
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	ui := c.config.Ui

	if info.IsDir() {
		ui.Say(fmt.Sprintf("Uploading folder to the VM '%s' => '%s'...",  src, dst))
		err := c.uploadFolder(dst, src)
		if err != nil {
			return err
		}
	} else {
		target_file := filepath.Join(dst,filepath.Base(src))
		ui.Say(fmt.Sprintf("Uploading file to the VM '%s' => '%s'...", src, target_file))
		err := c.uploadFile(target_file, src)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *comm) Download(string, io.Writer) error {
	panic("not implemented yet")
}

func (c *comm) uploadFile(dscPath string, srcPath string) error {

	dscPath = filepath.FromSlash(dscPath)
	srcPath = filepath.FromSlash(srcPath)

	vmName := c.config.VmName

	var script powershell.ScriptBuilder
	script.WriteLine("param([string]$vmName,[string]$sourcePath,[string]$destinationPath)")
	script.WriteLine("Copy-VMFile -Name $vmName -SourcePath $sourcePath -DestinationPath $destinationPath -CreateFullPath -FileSource Host -Force")

	powershell := new(powershell.PowerShellCmd)
	err := powershell.Run(script.String(), vmName, srcPath, dscPath)

	return err
}

func (c *comm) uploadFolder(dscPath string, srcPath string ) error {
	l := list.New()

	type dstSrc struct {
		src string
		dst string
	}

	treeWalk := func(path string, info os.FileInfo, prevErr error) error {
		// If there was a prior error, return it
		if prevErr != nil {
			return prevErr
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		l.PushBack(
			dstSrc{
				src: path,
				dst: filepath.Join(dscPath,rel),
			})

		return nil
	}

	filepath.Walk(srcPath, treeWalk)

	var err error
	for e := l.Front(); e != nil; e = e.Next() {
		// do something with e.Value
		pair := e.Value.(dstSrc)
		log.Printf("'%s' ==> '%s'\n", pair.src, pair.dst)
		err = c.uploadFile(pair.dst, pair.src)
		if err != nil {
			return err
		}
	}

	return err
}

