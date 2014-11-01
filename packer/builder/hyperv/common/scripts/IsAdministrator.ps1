$identity=[System.Security.Principal.WindowsIdentity]::GetCurrent();
$principal=new-object System.Security.Principal.WindowsPrincipal($identity);
$administratorRole=[System.Security.Principal.WindowsBuiltInRole]::Administrator;
return $principal.IsInRole($administratorRole);