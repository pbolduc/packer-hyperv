param([string]$ip)
Set-Item -Path WSMan:\\localhost\\Client\\TrustedHosts -Value $ip -Force -Concatenate
