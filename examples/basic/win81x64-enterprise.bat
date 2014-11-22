@ECHO OFF
SETLOCAL

SET ISO_URL=iso\en_windows_8_1_enterprise_x64_dvd_2971902.iso
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
	@ECHO Windows 8.1 Enterprise ISO file not found:
    @ECHO.
	@ECHO    %ISO_URL%
	GOTO :eof
)

IF NOT EXIST "box" MKDIR box

packer.exe build -only=hyperv-iso -var "iso_url=%ISO_URL%" -var "iso_checksum=%ISO_CHECKSUM%" win81x64-enterprise.json
