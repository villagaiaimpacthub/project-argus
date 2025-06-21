@echo off
setlocal EnableDelayedExpansion

REM Project Argus Launcher Script for Windows
REM Makes it easy to monitor any project directory

echo.
echo   ╔═══════════════════════════════════════════════════════════╗
echo   ║                    🚀 PROJECT ARGUS                      ║
echo   ║              Real-time Project Intelligence               ║
echo   ╚═══════════════════════════════════════════════════════════╝
echo.

REM Default values
set "WORKSPACE="
set "PORT=3002"
set "HELP="

REM Parse command line arguments
:parse_args
if "%~1"=="" goto :args_done
if "%~1"=="-h" set "HELP=true" & shift & goto :parse_args
if "%~1"=="--help" set "HELP=true" & shift & goto :parse_args
if "%~1"=="help" set "HELP=true" & shift & goto :parse_args
if "%~1"=="-p" set "PORT=%~2" & shift & shift & goto :parse_args
if "%~1"=="--port" set "PORT=%~2" & shift & shift & goto :parse_args
if "%~1"=="-w" set "WORKSPACE=%~2" & shift & shift & goto :parse_args
if "%~1"=="--workspace" set "WORKSPACE=%~2" & shift & shift & goto :parse_args
if "!WORKSPACE!"=="" set "WORKSPACE=%~1"
shift
goto :parse_args

:args_done

REM Show help if requested
if "%HELP%"=="true" (
    echo USAGE:
    echo   argus.bat [OPTIONS] [WORKSPACE_PATH]
    echo.
    echo OPTIONS:
    echo   -h, --help              Show this help message
    echo   -p, --port PORT         Port to run on ^(default: 3002^)
    echo   -w, --workspace PATH    Project directory to monitor
    echo.
    echo EXAMPLES:
    echo   argus.bat                                    # Monitor current directory
    echo   argus.bat C:\Projects\my-app                # Monitor specific directory
    echo   argus.bat --port 3003 ..\my-app             # Monitor on custom port
    echo   argus.bat -w C:\Users\You\Projects\task-dash # Using --workspace flag
    echo.
    echo QUICK START:
    echo   1. Run this script with your project path
    echo   2. Open http://localhost:%PORT% in your browser
    echo   3. Or open websocket_test.html for the test dashboard
    echo.
    exit /b 0
)

REM Use current directory if no workspace specified
if "!WORKSPACE!"=="" set "WORKSPACE=."

REM Convert to absolute path
for %%i in ("!WORKSPACE!") do set "WORKSPACE=%%~fi"

REM Validate workspace exists
if not exist "!WORKSPACE!" (
    echo ❌ Error: Directory '!WORKSPACE!' does not exist
    exit /b 1
)

REM Check if Go is installed
where go >nul 2>nul
if errorlevel 1 (
    echo ❌ Error: Go is not installed or not in PATH
    echo 💡 Install Go from: https://golang.org/dl/
    exit /b 1
)

REM Check if main.go exists
if not exist "%~dp0main.go" (
    echo ❌ Error: main.go not found in script directory
    echo 💡 Make sure you're running this from the Project Argus directory
    exit /b 1
)

echo 🎯 Target Directory: !WORKSPACE!
echo 🌐 Server Port: !PORT!
echo 📂 Project Type: %PROJECT_TYPE%
echo.

REM Detect project type
if exist "!WORKSPACE!\package.json" (
    set "PROJECT_TYPE=Node.js/JavaScript"
) else if exist "!WORKSPACE!\go.mod" (
    set "PROJECT_TYPE=Go"
) else if exist "!WORKSPACE!\requirements.txt" (
    set "PROJECT_TYPE=Python"
) else if exist "!WORKSPACE!\pyproject.toml" (
    set "PROJECT_TYPE=Python"
) else if exist "!WORKSPACE!\Cargo.toml" (
    set "PROJECT_TYPE=Rust"
) else if exist "!WORKSPACE!\composer.json" (
    set "PROJECT_TYPE=PHP"
) else if exist "!WORKSPACE!\pom.xml" (
    set "PROJECT_TYPE=Java (Maven)"
) else if exist "!WORKSPACE!\build.gradle" (
    set "PROJECT_TYPE=Java (Gradle)"
) else if exist "!WORKSPACE!\Gemfile" (
    set "PROJECT_TYPE=Ruby"
) else (
    set "PROJECT_TYPE=Generic"
)

echo 🚀 Starting Project Argus...
echo.

REM Set environment variables
set "ARGUS_WORKSPACE=!WORKSPACE!"
set "ARGUS_PORT=!PORT!"

echo 📡 Dashboard will be available at: http://localhost:!PORT!
echo 🎨 Test Dashboard: Open websocket_test.html in your browser
echo.
echo Press Ctrl+C to stop monitoring
echo.

REM Start the server
cd /d "%~dp0"
go run main.go "!WORKSPACE!" 