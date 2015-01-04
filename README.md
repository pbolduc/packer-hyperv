packer-hyperv
=============

Packer is an open source tool for creating identical machine images for multiple platforms from a single source configuration. For an introduction to Packer, check out documentation at http://www.packer.io/intro/index.html.

This is a Hyperv plugin for Packer.io to enable windows users to build custom images given an ISO. 

ISO's can be downloaded off technet or MSDN (if you have a subscription for the latter).
http://www.microsoft.com/en-us/evalcenter/evaluate-windows-server-2012-r2

The hyper-v plugin enables you to build a Windows Server Vagrant box for the hyper-v provider only. 

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
            "switch_name": "Virtual WLAN",
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

Additionally, as indicated above, if you obtain a windows license, you can specify the product key within your .json configuration and the plugin will register your copy of windows. 

Note: The plugin has to be run on a Windows workstation 8.1 or higher and must have hyper-v enabled. 


Additional Examples can be found on my fork of [Box Cutter Windows VM](https://github.com/pbolduc/windows-vm) repository.

