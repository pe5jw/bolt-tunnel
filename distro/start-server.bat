@echo off
if not "%CMDEXTVERSION%"=="" goto run
cmd /c "%~f0" %*
exit /b %errorlevel%

:run
chcp 65001 >nul

set TS_AUTHKEY=
set HOSTNAME=bolttunnel-server
set PORT=7780
set CLIENT=bolttunnel-client
set CLIENT_PORT=7781
set GUI=localhost:8081

:: ── Server tunnel configuratie ───────────────────────────────────────────────
:: Tunnels die de server aanbiedt op het tailnet
:: Formaat A: naam:poort              → localhost:poort
:: Formaat C: naam:tailnetpoort:host:poort  → LAN apparaat
set SERVER_TUNNELS=SDRoxide:4950,BoltSDR:6443,BoltDVK:4532,TCI40001:40001,SWITCH:9090,BoltATR1000:3000

:: ── Client tunnel configuratie ───────────────────────────────────────────────
:: Tunnels die automatisch naar alle clients worden gepusht
:: Formaat: naam:clientpoort:serverpoort
set CLIENT_TUNNELS=SDRoxide:4951:4950,BoltSDR:6444:6443,BoltDVK:4533:4532,TCI40001:40001:40001,SWITCH:9091:9090,BoltATR1000:3001:3000

:: ── Commando configuratie (beschikbaar voor alle clients) ────────────────────
set COMMANDS=Start-BoltSDR:C:\scripts\start-boltsdr.bat,Stop-BoltSDR:C:\scripts\stop-boltsdr.bat,Start-SDRoxide:C:\scripts\start-sdroxide.bat,Stop-SDRoxide:C:\scripts\stop-sdroxide.bat

:: ── Server-only commando's (alleen zichtbaar in server GUI) ──────────────────
set SERVER_COMMANDS=Restart-Server:C:\scripts\restart-server.bat

:: ─────────────────────────────────────────────────────────────────────────────

if not exist "%~dp0bolttunnel-server.exe" (
    echo [FOUT] bolttunnel-server.exe niet gevonden.
    pause & exit /b 1
)

echo Bolttunnel server starten als "%HOSTNAME%"
echo Tailnet poort    : %PORT%
echo Server tunnels   : %SERVER_TUNNELS%
echo Client tunnels   : %CLIENT_TUNNELS%
echo GUI              : http://%GUI%
echo.

"%~dp0bolttunnel-server.exe" --hostname %HOSTNAME% --addr :%PORT% --client %CLIENT% --cport %CLIENT_PORT% --gui %GUI% --tunnels "%SERVER_TUNNELS%" --client-tunnels "%CLIENT_TUNNELS%" --commands "%COMMANDS%" --server-commands "%SERVER_COMMANDS%"

pause
