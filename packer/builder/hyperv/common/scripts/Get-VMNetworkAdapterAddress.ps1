param([string]$vmName)
try { 
	$adapter = Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue;
	$ip = $adapter.IPAddresses[0];
	if($ip -eq $null) {
		return $false
	}
}
catch {
	return $false
} 
return $ip
