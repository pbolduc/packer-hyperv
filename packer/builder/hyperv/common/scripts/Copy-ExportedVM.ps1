param([string]$srcPath, [string]$dstPath, [string]$vhdDirName, [string]$vmDir)
Copy-Item -Path $srcPath\$vhdDirName -Destination $dstPath -recurse;
Copy-Item -Path $srcPath\$vmDir -Destination $dstPath;
Copy-Item -Path $srcPath\$vmDir\*.xml -Destination $dstPath\$vmDir;
