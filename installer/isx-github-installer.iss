; ISX Daily Reports Scrapper - GitHub Release Installer
; This creates a small installer that downloads the latest release from GitHub

#define MyAppName "ISX Daily Reports Scrapper"
#define MyAppVersion "0.1.0"
#define MyAppPublisher "The Iraqi Investor Group"
#define MyAppURL "https://github.com/haideralmesaody/ISXDailyReportScrapper"
#define MyAppExeName "start-web-interface.exe"
#define MyAppDescription "Iraqi Stock Exchange Daily Reports Data Scrapper"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
AppId={{8BC25D3C-8B5E-4A47-9F2D-1234567890AB}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} {#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes
LicenseFile=assets\LICENSE.txt
InfoBeforeFile=assets\README.txt
OutputDir=output
OutputBaseFilename=ISXDailyReportsInstaller-v{#MyAppVersion}
SetupIconFile=assets\setup-icon.ico
Compression=lzma
SolidCompression=yes
ArchitecturesInstallIn64BitMode=x64
PrivilegesRequired=lowest
UsePreviousAppDir=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[Files]
; Include the PowerShell downloader script
Source: "assets\download-github-release.ps1"; DestDir: "{tmp}"; Flags: ignoreversion
Source: "assets\README.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\LICENSE.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\{#MyAppExeName}"
Name: "{group}\{cm:UninstallProgram,{#MyAppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon; IconFilename: "{app}\{#MyAppExeName}"

[Run]
; Run the PowerShell script to download from GitHub during installation
Filename: "powershell.exe"; Parameters: "-ExecutionPolicy Bypass -File ""{tmp}\download-github-release.ps1"" -InstallPath ""{app}"" -RepoOwner ""haideralmesaody"" -RepoName ""ISXDailyReportScrapper"""; Flags: runhidden waituntilterminated; StatusMsg: "Downloading latest release from GitHub..."
Filename: "{app}\{#MyAppExeName}"; Description: "{cm:LaunchProgram,{#MyAppName}}"; Flags: nowait postinstall skipifsilent

[Code]
procedure InitializeWizard;
begin
  WizardForm.WelcomeLabel1.Caption := 'Welcome to the ISX Daily Reports Scrapper Setup';
  WizardForm.WelcomeLabel2.Caption := 
    'This installer will download and install the latest version of ISX Daily Reports Scrapper from GitHub.' + #13#10 + #13#10 +
    'The installer is small (1-2 MB) and will download the application files during installation.' + #13#10 + #13#10 +
    'Make sure you have an internet connection before proceeding.' + #13#10 + #13#10 +
    'Click Next to continue, or Cancel to exit.';
end;

function InitializeSetup(): Boolean;
begin
  Result := True;
  if not CheckForInternetConnection() then
  begin
    MsgBox('Internet connection is required to download the application files from GitHub.' + #13#10 + 
           'Please check your internet connection and try again.', mbError, MB_OK);
    Result := False;
  end;
end;

function CheckForInternetConnection(): Boolean;
var
  WinHttpReq: Variant;
begin
  Result := True;
  try
    WinHttpReq := CreateOleObject('WinHttp.WinHttpRequest.5.1');
    WinHttpReq.Open('GET', 'https://api.github.com', False);
    WinHttpReq.SetTimeouts(5000, 5000, 5000, 5000);
    WinHttpReq.Send();
    Result := (WinHttpReq.Status = 200);
  except
    Result := False;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // Installation completed successfully
    // The PowerShell script has already downloaded and extracted the files
  end;
end; 