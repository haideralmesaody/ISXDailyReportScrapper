@echo off
echo Running Pipeline Tests...
echo.

REM Run specific failing tests
echo === Running Dependency Test ===
go test -v -run TestPipelineWithDependencyFailure
echo.

echo === Running Manager Failure Test ===
go test -v -run TestManagerExecuteWithFailure
echo.

echo === Running Registry Order Test ===
go test -v -run "TestRegistryGetDependencyOrder/No_dependencies"
echo.

echo === Running WebSocket Progress Test ===
go test -v -run TestWebSocketProgressUpdates
echo.

echo Tests complete.