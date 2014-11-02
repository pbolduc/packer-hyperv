param([string]$vmName, [string]$path)
Export-VM -Name $vmName -Path $path
