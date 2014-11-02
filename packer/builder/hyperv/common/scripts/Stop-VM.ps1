param([string]$vmName)
if ((Get-VM -Name $vmName).State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
    Stop-VM -Name $vmName
}
