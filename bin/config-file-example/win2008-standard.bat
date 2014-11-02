@ECHO OFF
SETLOCAL

SET ISO_URL=iso\en_windows_server_2008_datacenter_enterprise_standard_sp2_x86_dvd_342333.iso
SET ISO_CHECKSUM=

IF "%PACKER_HOME%" == "" (
	@ECHO.
	@ECHO Please set PACKER_HOME
	GOTO :eof
)

IF NOT EXIST "%PACKER_HOME%" (
	@ECHO.
	@ECHO PACKER_HOME does not exist
    @ECHO.
	@ECHO    PACKER_HOME=%PACKER_HOME%
	GOTO :eof
)

IF NOT EXIST "%PACKER_HOME%\packer-builder-hyperv-iso.exe" (
	@ECHO.
	@ECHO Copying packer-builder-hyperv-iso.exe to PACKER_HOME
	COPY "%~dp0..\packer-builder-hyperv-iso.exe" "%PACKER_HOME%\"
 )

PATH=%PACKER_HOME%;%PATH%

IF NOT EXIST "%ISO_URL%" (
    @ECHO.
	@ECHO Windows Server 2008 ISO file not found:
    @ECHO.
	@ECHO    %ISO_URL%
	GOTO :eof
)

packer.exe build -only=hyperv-iso -var "iso_url=%ISO_URL%" -var "iso_checksum=%ISO_CHECKSUM%" win2008-standard.json
