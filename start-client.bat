@echo off
if not "%CMDEXTVERSION%"=="" goto run
cmd /c "%~f0" %*
exit /b %errorlevel%

:run
chcp 65001 >nul

if exist "%~dp0go\bin\go.exe" (
    set "GOROOT=%~dp0go"
    set "GOPATH=%~dp0.gopath"
    set "GOCACHE=%~dp0.gocache"
    set "PATH=%~dp0go\bin;%PATH%"
)

set TS_AUTHKEY=
set HOSTNAME=bolttunnel-client
set SERVER=bolttunnel-server
set PORT=7780
set LISTEN=7781
set GUI=127.0.0.1:8080

:: ── Log configuratie (optioneel) ───────────────────────────────────────────
:: Laat leeg om alleen op het scherm te loggen
:: Vul een pad in om ook naar een bestand te loggen
set LOGFILE=

:: ── Extra tunnel configuratie (optioneel) ────────────────────────────────────────
:: Tunnels die alleen voor DEZE client actief zijn.
:: De server pusht automatisch de standaard tunnels.
:: Formaat: naam:clientpoort:serverpoort
set EXTRA_TUNNELS=RDP:13389:3389

:: ─────────────────────────────────────────────────────────────────────────────

set ARGS=--hostname %HOSTNAME% --server %SERVER% --port %PORT% --listen :%LISTEN% --gui %GUI% --tunnels "%EXTRA_TUNNELS%" %LOGFILE_ARG%

:: Binary beschikbaar? Start zonder terminal venster
if exist "%~dp0bin\bolttunnel-client.exe" (
    if not "%LOGFILE%"=="" (set "LOGFILE_ARG=--logfile %LOGFILE%") else (set "LOGFILE_ARG=")
start "" "%~dp0bin\bolttunnel-client.exe" %ARGS%
    exit
)

:: Nog niet gebouwd - go run (met terminal, voor ontwikkeling)
call "%~dp0setup.bat"
if %errorlevel% neq 0 (
    echo Setup mislukt - client niet gestart.
    pause & exit /b 1
)

if exist "%~dp0go\bin\go.exe" (
    set "GOROOT=%~dp0go"
    set "GOPATH=%~dp0.gopath"
    set "GOCACHE=%~dp0.gocache"
    set "PATH=%~dp0go\bin;%PATH%"
)

echo Binary niet gevonden - gebruik go run (bouw eerst met build.bat)
echo.
cd /d "%~dp0cmd\client"
go run . %ARGS%
pause
