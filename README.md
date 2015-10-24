2015-10-24 Update
=================

I would recommend anyone interested in using Hyper-V with Packer take a look at the great
work [Taliesin Sisson](https://github.com/taliesins) did to take my changes and bring it
up standards so that it can be merged into the main Packer solution. 

[Hyper-V builder for Packer pull request](https://github.com/mitchellh/packer/pull/2576)


packer-hyperv
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is a Hyperv plugin for Packer.io to enable windows users to build custom images given an ISO.

ISO's can be downloaded off technet or MSDN (if you have a subscription for the latter).
[Windows Server Evaluations](http://www.microsoft.com/en-us/evalcenter/evaluate-windows-server-2012-r2)

The hyper-v plugin enables you to build a Windows Server [Vagrant](https://www.vagrantup.com/) box for the hyper-v provider only. 

The bin folder has an example JSON to help specify the new hyperv configuration.

    "builders": [
        {
            "vm_name": "win2012r2-standard",
            "type": "hyperv-iso",
            "iso_url": "{{ user `iso_url` }}",
            "iso_checksum": "{{ user `iso_checksum` }}",
            "iso_checksum_type": "sha1",
            "ssh_username": "vagrant",
            "ssh_password": "vagrant",
            "ssh_wait_timeout": "10000s",
            "switch_name": "",
            "product_key": "xxxxx-xxxxx-xxxxx-xxxxx-xxxxx",
            "floppy_files": [
                "floppy/win2012r2-standard/Autounattend.xml",
                "floppy/00-run-all-scripts.cmd",
                "floppy/install-winrm.cmd",
                "floppy/powerconfig.bat",
                "floppy/01-install-wget.cmd",
                "floppy/_download.cmd",
                "floppy/_packer_config.cmd",
                "floppy/passwordchange.bat",
                "floppy/openssh.bat",
                "floppy/z-install-integration-services.bat",
                "floppy/zz-start-sshd.cmd",
                "floppy/oracle-cert.cer",
                "floppy/zzzz-shutdown.bat"
            ]
        }
    ]

Current version of the hyper-v builder **requires that the VM be shutdown after the install/build phase is complete**.  The *zzzz-shutdown.bat* script in the example above does this.  When the virtual machine is started, the communicator must be available for provisioning stage. 

~~Additionally, as indicated above, if you obtain a windows license, you can specify the product key within your .json configuration and the plugin will register your copy of windows.~~

Note: The plugin has to be run on a Windows workstation 8.1 or higher and must have hyper-v enabled. 

Examples can be found on my fork of [Box Cutter Windows VM](https://github.com/pbolduc/windows-vm) repository.


# Configuration Reference

## Required:

* **vm_name** (string) - The name of the virtual machine
* **type** (string) - Must be *hyperv-iso*
* **iso_url** (string) - A URL to the ISO containing the installation image. This URL can be either an HTTP URL or a file URL (or path to a file). If this is an HTTP URL, Packer will download it and cache it between runs

## Optional:

* **switch_name** (string) - The Hyper-V virtual switch name to bind to the virtual machine.  If not specified, the external virtual switch connected fastest (based on link speed) network adapter is used. If no virtual switch can be detected, a temporary internal switch will be created.
* **floppy_files** (array of strings) - A list of files to place onto a floppy disk that is attached when the VM is booted. This is most useful for unattended Windows installs, which look for an **Autounattend.xml** file on removable media. By default, no floppy will be attached. All files listed in this setting get placed into the root directory of the floppy and the floppy is attached as the first floppy device. Currently, no support exists for creating sub-directories on the floppy. Wildcard characters (*, ?, and []) are allowed. Directory names are also allowed, which will add all the files found in the directory to the floppy.
* **ssh_username** (string) - The username to use to SSH into the machine once the OS is installed.
* **ssh_password** (string) - The password to use to SSH into the machine once the OS is installed.
* **ssh_wait_timeout** (string) - How long to wait for SSH to be available.
* **communicator** (string) - Can be either **ssh** or **winrm**.  Default is ssh.  winrm **not** currently implemented.
* **product_key** (string) - Windows product key to set.  Your floppy_files must contain a Autounattend.xml entry.