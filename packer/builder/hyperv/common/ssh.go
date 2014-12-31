package common

import (
	"fmt"
	"time"
	"log"
	"strings"

	gossh "code.google.com/p/go.crypto/ssh"
	commonssh "github.com/mitchellh/packer/common/ssh"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"github.com/MSOpenTech/packer-hyperv/packer/powershell/hyperv"
)

func SSHAddress(state multistep.StateBag) (string, error) {
	//sshHostPort := state.Get("sshHostPort").(uint)
	sshHostPort := 22
	ip, _ := getVMAddress(state)
	return fmt.Sprintf("%v:%d", ip, sshHostPort), nil
}

func SSHConfigFunc(config SSHConfig) func(multistep.StateBag) (*gossh.ClientConfig, error) {
	return func(state multistep.StateBag) (*gossh.ClientConfig, error) {
		auth := []gossh.AuthMethod{
			gossh.Password(config.SSHPassword),
			gossh.KeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHPassword)),
		}

		if config.SSHKeyPath != "" {
			signer, err := commonssh.FileSigner(config.SSHKeyPath)
			if err != nil {
				return nil, err
			}

			auth = append(auth, gossh.PublicKeys(signer))
		}

		return &gossh.ClientConfig{
			User: config.SSHUser,
			Auth: auth,
		}, nil
	}
}

func getVMAddress(state multistep.StateBag) (string, error) {

	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	errorMsg := "Could not get ip address for VM."

	count := 60
	var duration time.Duration = 1
	sleepTime := time.Minute * duration
	var ip string

	for count != 0 {

		address, err := hyperv.GetVirtualMachineNetworkAdapterAddress(vmName)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return "", err
		}

		ip = strings.TrimSpace(string(address))

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
		return "", err
	}

	ui.Say("ip address is " + ip)
	state.Put("ip", ip)

	return ip, nil

}
