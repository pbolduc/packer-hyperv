@ECHO OFF
SETLOCAL
PATH=%PATH%;E:\work\go\bin

CD %~dp0
go-bindata.exe -pkg="common" scripts/...

ENDLOCAL
