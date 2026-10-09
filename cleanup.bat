@echo off
:: Zorg dat dit altijd in cmd.exe draait, ook als gestart vanuit PowerShell
if not "%CMDEXTVERSION%"=="" goto run
cmd /c "%~f0" %*
exit /b %errorlevel%

:run
chcp 65001 >nul

echo ========================================
echo  Bolttunnel cleanup
echo ========================================
echo.
echo Dit verwijdert:
echo   - Portable Go installatie  (.\go\)
echo   - Module cache             (.\.gopath\)
echo   - Build cache              (.\.gocache\)
echo   - go.sum (wordt opnieuw aangemaakt)
echo   - tsnet state (Tailscale node registratie)
echo.
set /p CONFIRM=Doorgaan? (j/n): 
if /i not "%CONFIRM%"=="j" (
    echo Geannuleerd.
    pause
    exit /b 0
)

echo.
echo Verwijderen...

if exist "%~dp0go"          ( rd /s /q "%~dp0go"          && echo   [OK] .\go\ )
if exist "%~dp0.gopath"     ( rd /s /q "%~dp0.gopath"     && echo   [OK] .\.gopath\ )
if exist "%~dp0.gocache"    ( rd /s /q "%~dp0.gocache"    && echo   [OK] .\.gocache\ )
if exist "%~dp0go.sum"      ( del /q   "%~dp0go.sum"      && echo   [OK] go.sum )

:: tsnet state voor client en server
if exist "%APPDATA%\tsnet-client" ( rd /s /q "%APPDATA%\tsnet-client" && echo   [OK] tsnet-client state )
if exist "%APPDATA%\tsnet-server" ( rd /s /q "%APPDATA%\tsnet-server" && echo   [OK] tsnet-server state )

:: Go build temp bestanden
for /d %%d in ("%TEMP%\go-build*") do (
    rd /s /q "%%d" >nul 2>&1
)
echo   [OK] Go build temp bestanden

echo.
echo Klaar - alles opgeruimd.
echo Volgende start downloadt Go en dependencies opnieuw.
echo.
pause
