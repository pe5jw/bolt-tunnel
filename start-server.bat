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
set HOSTNAME=bolttunnel-server
set PORT=7780
set CLIENT=bolttunnel-client
set CLIENT_PORT=7781
set GUI=localhost:8081

:: ── Server tunnel configuratie ───────────────────────────────────────────────
:: Formaat A: naam:poort              → localhost:poort
:: Formaat C: naam:tailnetpoort:host:poort  → LAN apparaat
set SERVER_TUNNELS=SDRoxide:4950,BoltSDR:6443,BoltDVK:4532,TCI40001:40001,SWITCH:9090,BoltATR1000:3000

:: ── Client tunnel configuratie ───────────────────────────────────────────────
:: Wordt automatisch naar alle clients gepusht
:: Formaat: naam:clientpoort:serverpoort
set CLIENT_TUNNELS=SDRoxide:4951:4950,BoltSDR:6444:6443,BoltDVK:4533:4532,TCI40001:40001:40001,SWITCH:9091:9090,BoltATR1000:3001:3000

:: ── Commando's voor alle clients ─────────────────────────────────────────────
set COMMANDS=Start-BoltSDR:C:\scripts\start-boltsdr.bat,Stop-BoltSDR:C:\scripts\stop-boltsdr.bat,Start-SDRoxide:C:\scripts\start-sdroxide.bat,Stop-SDRoxide:C:\scripts\stop-sdroxide.bat

:: ── Server-only commando's ────────────────────────────────────────────────────
set SERVER_COMMANDS=Restart-Server:C:\scripts\restart-server.bat

:: ─────────────────────────────────────────────────────────────────────────────

set ARGS=--hostname %HOSTNAME% --addr :%PORT% --client %CLIENT% --cport %CLIENT_PORT% --gui %GUI% --tunnels "%SERVER_TUNNELS%" --client-tunnels "%CLIENT_TUNNELS%" --commands "%COMMANDS%" --server-commands "%SERVER_COMMANDS%"

if exist "%~dp0bin\bolttunnel-server.exe" (
    echo Gebruik: bin\bolttunnel-server.exe
    "%~dp0bin\bolttunnel-server.exe" %ARGS%
) else (
    call "%~dp0setup.bat"
    if %errorlevel% neq 0 (
        echo Setup mislukt - server niet gestart.
        pause & exit /b 1
    )
    if exist "%~dp0go\bin\go.exe" (
        set "GOROOT=%~dp0go"
        set "GOPATH=%~dp0.gopath"
        set "GOCACHE=%~dp0.gocache"
        set "PATH=%~dp0go\bin;%PATH%"
    )
    echo Gebruik: go run (bouw eerst met build.bat)
    cd /d "%~dp0cmd\server"
    go run . %ARGS%
)
pause
