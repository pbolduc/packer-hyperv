packer-hyperv - Fork Notes
==========================

Goals
-----
* Support any Hyper-V supported operating system
  - priority will go to Windows Server OSes 2008 and above
* Remove embedded PowerShell to external files with parameters
  - Use https://github.com/jteeuwen/go-bindata to embedded PowerShell scripts in Go executables
  - Could use: powershell.exe -File <file> <arg1> <arg2> ...
  - Could use: powershell.exe -EncodedCommand <base-64-encoded-block>
* Improve the strategy for detecting the end of the install
  - could add registry key HKLM\SOFTWARE\packer.io\completed to signal the end of autounattend install
* Leverage work done in https://github.com/box-cutter/windows-vm 
* Add more customization features to VM creation (more CPUs, dynamic memory, etc)
