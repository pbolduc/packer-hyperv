packer-hyperv - Fork Notes
==========================

Goals
-----
* Support any Hyper-V supported operating system
  - priority will go to Windows Server OSes 2008 and above
* Remove embedded PowerShell to external files with parameters
* Improve the strategy for detecting the end of the install
  - could add registry key HKLM\SOFTWARE\packer.io\completed to signal the end of autounattend install
