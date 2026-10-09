@echo off
chcp 65001 >nul

echo ========================================
echo  Bolttunnel - Go verwijderen
echo ========================================
echo.
echo Dit verwijdert de portable Go installatie en caches.
echo De bolttunnel broncode en bat-bestanden blijven intact.
echo.

if not exist "%~dp0go" (
    echo Go map niet gevonden - al verwijderd?
    pause
    exit /b 0
)

echo Verwijderen:
echo   %~dp0go
echo   %~dp0.gopath
echo   %~dp0.gocache
echo.
set /p CONFIRM=Doorgaan? (j/n): 
if /i not "%CONFIRM%"=="j" (
    echo Geannuleerd.
    pause
    exit /b 0
)

echo.
if exist "%~dp0go"       rd /s /q "%~dp0go"
if exist "%~dp0.gopath"  rd /s /q "%~dp0.gopath"
if exist "%~dp0.gocache" rd /s /q "%~dp0.gocache"

echo Klaar - Go verwijderd.
echo Volgende keer dat je start-client.bat of start-server.bat
echo aanroept wordt Go automatisch opnieuw gedownload.
echo.
pause
