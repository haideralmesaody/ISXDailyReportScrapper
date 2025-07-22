; ISX Daily Reports Scrapper - Professional Windows Installer (x64)
; Creates a standard Windows installer with GitHub download capability
; Version 1.0.0 - Professional Edition

#define MyAppName "ISX Daily Reports"
#define MyAppNameLong "ISX Daily Reports Scrapper"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "The Iraqi Investor Group"
#define MyAppURL "https://github.com/haideralmesaody/ISXDailyReportScrapper"
#define MyAppExeName "start-web-interface.exe"
#define MyAppDescription "Professional Iraqi Stock Exchange Analytics Platform"
#define MyAppId "ISXDailyReports"

[Setup]
; Unique identifier for this application
AppId={{B8C15D2C-7A4E-4F47-9E1D-890123456789}
AppName={#MyAppNameLong}
AppVersion={#MyAppVersion}
AppVerName={#MyAppName} v{#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}/issues
AppUpdatesURL={#MyAppURL}/releases
AppCopyright=Copyright © 2024 {#MyAppPublisher}
AppComments={#MyAppDescription}

; Installation directories
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
AllowNoIcons=yes

; Files and documentation
LicenseFile=assets\LICENSE.txt
InfoBeforeFile=assets\README.txt

; Output configuration
OutputDir=..\
OutputBaseFilename=ISX-Daily-Reports-Professional-x64-Installer
SetupIconFile=assets\isx-app-icon.ico
UninstallDisplayIcon={app}\{#MyAppExeName}

; Compression and modern options
Compression=lzma/ultra64
SolidCompression=yes
LZMAUseSeparateProcess=yes

; Architecture and privileges
ArchitecturesInstallIn64BitMode=x64
ArchitecturesAllowed=x64
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog

; Modern Windows features
UsePreviousAppDir=yes
UsePreviousGroup=yes
DisableProgramGroupPage=auto
DisableReadyPage=no
DisableFinishedPage=no
DisableWelcomePage=no
ShowLanguageDialog=auto
SetupLogging=yes

; Visual appearance
WizardStyle=modern
WizardSizePercent=100
WizardResizable=yes

; Uninstaller
UninstallDisplayName={#MyAppName}
UninstallFilesDir={app}\Uninstall
CreateUninstallRegKey=yes

; Version info
VersionInfoVersion={#MyAppVersion}
VersionInfoCompany={#MyAppPublisher}
VersionInfoDescription={#MyAppDescription}
VersionInfoCopyright=Copyright © 2024 {#MyAppPublisher}
VersionInfoProductName={#MyAppNameLong}
VersionInfoProductVersion={#MyAppVersion}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Messages]
WelcomeLabel1=Welcome to the [name] Setup Wizard
WelcomeLabel2=This will install [name/ver] on your computer.%n%nThis professional installer will download the latest release from GitHub and set up ISX Daily Reports with desktop shortcuts and Start Menu entries.%n%nIt is recommended that you close all other applications before continuing.

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop shortcut"; GroupDescription: "Additional shortcuts:"; Flags: unchecked
Name: "quicklaunchicon"; Description: "Create a &Quick Launch shortcut"; GroupDescription: "Additional shortcuts:"; Flags: unchecked; OnlyBelowVersion: 6.1

[Files]
; Core installer files
Source: "assets\download-github-release.ps1"; DestDir: "{tmp}"; Flags: ignoreversion deleteafterinstall
Source: "assets\favicon.svg"; DestDir: "{app}"; DestName: "app-icon.svg"; Flags: ignoreversion
Source: "assets\LICENSE.txt"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\README.txt"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu shortcuts
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Comment: "{#MyAppDescription}"; IconFilename: "{app}\app-icon.svg"
Name: "{group}\{#MyAppName} Documentation"; Filename: "{app}\README.txt"; Comment: "Read the documentation"
Name: "{group}\Visit {#MyAppName} Website"; Filename: "{#MyAppURL}"; Comment: "Visit the official website"
Name: "{group}\Uninstall {#MyAppName}"; Filename: "{uninstallexe}"; Comment: "Uninstall {#MyAppName}"

; Desktop shortcut (optional)
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Comment: "{#MyAppDescription}"; Tasks: desktopicon; IconFilename: "{app}\app-icon.svg"

; Quick Launch shortcut (optional, Windows XP/Vista/7)
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: quicklaunchicon; IconFilename: "{app}\app-icon.svg"

[Run]
; Download and install the application
Filename: "powershell.exe"; Parameters: "-ExecutionPolicy Bypass -File ""{tmp}\download-github-release.ps1"" -InstallPath ""{app}"" -AppName ""{#MyAppName}"""; StatusMsg: "Downloading and installing {#MyAppName}..."; Flags: waituntilterminated

; Option to launch the application
Filename: "{app}\{#MyAppExeName}"; Description: "Launch {#MyAppName} now"; Flags: nowait postinstall skipifsilent

[UninstallDelete]
; Clean up downloaded files and logs
Type: filesandordirs; Name: "{app}\web"
Type: filesandordirs; Name: "{app}\downloads"
Type: filesandordirs; Name: "{app}\logs"
Type: filesandordirs; Name: "{app}\reports"
Type: filesandordirs; Name: "{app}\bin"
Type: filesandordirs; Name: "{app}\tools"

[Code]
// Modern Inno Setup code for enhanced functionality

function GetUninstallString(): String;
var
  sUnInstPath: String;
  sUnInstallString: String;
begin
  sUnInstPath := ExpandConstant('Software\Microsoft\Windows\CurrentVersion\Uninstall\{#emit SetupSetting("AppId")}_is1');
  sUnInstallString := '';
  if not RegQueryStringValue(HKLM, sUnInstPath, 'UninstallString', sUnInstallString) then
    RegQueryStringValue(HKCU, sUnInstPath, 'UninstallString', sUnInstallString);
  Result := sUnInstallString;
end;

function IsUpgrade(): Boolean;
begin
  Result := (GetUninstallString() <> '');
end;

function InitializeSetup(): Boolean;
var
  V: Integer;
  iResultCode: Integer;
  sUnInstallString: String;
begin
  Result := True; // if false, then cancel installation

  // Check if application is already installed
  if RegValueExists(HKEY_LOCAL_MACHINE,'Software\Microsoft\Windows\CurrentVersion\Uninstall\{#emit SetupSetting("AppId")}_is1', 'UninstallString') then
  begin
    V := MsgBox(ExpandConstant('An existing installation of {#MyAppName} was detected. Do you want to uninstall it first?'), mbInformation, MB_YESNO);
    if V = IDYES then
    begin
      sUnInstallString := GetUninstallString();
      sUnInstallString := RemoveQuotes(sUnInstallString);
      if Exec(sUnInstallString, '/SILENT', '', SW_HIDE, ewWaitUntilTerminated, iResultCode) then
        Result := True
      else
        Result := False;
    end
    else
      Result := False;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    // Additional post-installation steps can be added here
    // For example, registering file associations, etc.
  end;
end; 