@echo off
cd /d "%~dp0"

echo.
echo  Pack-a-mal Demo Dashboard
echo ========================================

python --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Python chua duoc cai dat!
    pause & exit /b 1
)

if not exist "venv\Scripts\python.exe" (
    echo [Setup] Tao virtual environment...
    python -m venv venv
    echo [Setup] Cai dependencies...
    venv\Scripts\pip install -r requirements.txt -q
    echo [Setup] Cau hinh UTF-8 encoding...
    (
        echo import sys
        echo try:
        echo     sys.stdout.reconfigure(encoding="utf-8", errors="replace"^)
        echo     sys.stderr.reconfigure(encoding="utf-8", errors="replace"^)
        echo except Exception:
        echo     pass
    ) > "venv\Lib\site-packages\sitecustomize.py"
)

echo [Check] Kiem tra port 5500...
for /f "tokens=5" %%p in ('netstat -ano ^| findstr ":5500 " ^| findstr "LISTENING"') do (
    echo [Kill] Dang tat tien trinh cu tren port 5500 (PID: %%p^)...
    taskkill /PID %%p /F >nul 2>&1
)
timeout /t 1 /nobreak >nul

echo [OK] Dang khoi dong server...
echo.
echo  =^=^> http://127.0.0.1:5500
echo.
echo  Nhan Ctrl+C de dung
echo ========================================
echo.

start "" http://127.0.0.1:5500
"%~dp0venv\Scripts\python.exe" "%~dp0app.py"
pause
