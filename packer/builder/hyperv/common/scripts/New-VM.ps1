param([string]$vmName, [string]$path, [long]$memoryStartupBytes, [long]$newVHDSizeBytes, [string]$switchName)

$vhdx = $vmName + ".vhdx"
$vhdPath = Join-Path -Path $path -ChildPath $vhdx

New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $newVHDSizeBytes -SwitchName $switchName
