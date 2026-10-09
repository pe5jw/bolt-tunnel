@echo off
if not "%CMDEXTVERSION%"=="" goto run
cmd /c "%~f0" %*
exit /b %errorlevel%

:run
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
:: Voeg hier toe wat niet voor iedereen beschikbaar moet zijn.
:: Formaat: naam:clientpoort:serverpoort
set EXTRA_TUNNELS=RDP:13389:3389

:: ─────────────────────────────────────────────────────────────────────────────

if not exist "%~dp0bolttunnel-client.exe" (
    echo [FOUT] bolttunnel-client.exe niet gevonden.
    pause & exit /b 1
)

if not "%LOGFILE%"=="" (set "LOGFILE_ARG=--logfile %LOGFILE%") else (set "LOGFILE_ARG=")
start "" "%~dp0bolttunnel-client.exe" --hostname %HOSTNAME% --server %SERVER% --port %PORT% --listen :%LISTEN% --gui %GUI% --tunnels "%EXTRA_TUNNELS%" %LOGFILE_ARG%
exit
