; ISX Daily Reports Scrapper - GitHub Release Installer (ARM64)
; This creates a small installer that downloads ARM64 binaries or builds them locally

#define MyAppName "ISX Daily Reports Scrapper"
#define MyAppVersion "0.1.0"
#define MyAppPublisher "The Iraqi Investor Group"
#define MyAppURL "https://github.com/haideralmesaody/ISXDailyReportScrapper"
#define MyAppExeName "start-web-interface.exe"
#define MyAppDescription "Iraqi Stock Exchange Daily Reports Data Scrapper (ARM64)"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
AppId={{8BC25D3C-8B5E-4A47-9F2D-1234567890AC}
AppName={#MyAppName} (ARM64)
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion} (ARM64)
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}\ARM64
DefaultGroupName={#MyAppName} (ARM64)
AllowNoIcons=yes
LicenseFile=assets\LICENSE.txt
InfoBeforeFile=assets\README.txt
OutputDir=output
OutputBaseFilename=ISXDailyReportsInstaller-ARM64-v{#MyAppVersion}
SetupIconFile=assets\favicon.ico
Compression=lzma
SolidCompression=yes
; ARM64 architecture support
ArchitecturesInstallIn64BitMode=arm64
PrivilegesRequired=lowest
UsePreviousAppDir=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; Include the PowerShell downloader script for ARM64
Source: "assets\download-github-release.ps1"; DestDir: "{tmp}"; Flags: ignoreversion
Source: "assets\README.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\LICENSE.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName} (ARM64)"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName} (ARM64)}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName} (ARM64)"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon; IconFilename: "{app}\{#MyAppExeName}"

[Run]
Filename: "powershell.exe"; Parameters: "-ExecutionPolicy Bypass -File ""{tmp}\download-github-release.ps1"" -Architecture ARM64 -InstallPath ""{app}"""; Description: "Download and install ARM64 binaries"; Flags: runascurrentuser waituntilterminated

[Code]
function InitializeSetup(): Boolean;
var
  IsARM64: Boolean;
  ResultCode: Integer;
begin
  // Check if running on ARM64 processor
  if Exec('powershell.exe', '-Command "(Get-WmiObject Win32_Processor).Architecture -eq 9"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) then
  begin
    IsARM64 := (ResultCode = 0);
    if not IsARM64 then
    begin
      if MsgBox('This installer is optimized for ARM64 processors.' + #13#10 + 
                'Your system appears to be x86/x64.' + #13#10 + #13#10 +
                'Do you want to continue anyway?' + #13#10 +
                '(Consider using the standard x64 installer instead)', 
                mbConfirmation, MB_YESNO) = IDNO then
      begin
        Result := False;
        Exit;
      end;
    end;
  end;
  Result := True;
end; 