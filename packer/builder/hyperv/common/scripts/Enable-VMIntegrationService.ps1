param([string]$vmName, [string]$integrationServiceName)
Enable-VMIntegrationService -VMName $vmName -Name $integrationServiceName
