param([string]$vmName, [string]$Path, [long]$MemoryStartupBytes, [long]$NewVHDSizeBytes, [string]$SwitchName)

$vhdx = $vmName + ".vhdx"
$vhdPath = Join-Path -Path $Path -ChildPath $vhdx

New-VM -Name $vmName -Path $Path -MemoryStartupBytes $MemoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $NewVHDSizeBytes -SwitchName $SwitchName
