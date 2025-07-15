; ISX Daily Reports Scraper - GitHub Downloader Installer
; This is a small installer that downloads the latest release from GitHub
; Company: The Iraqi Investor Group

#define AppName "ISX Daily Reports Scraper"
#define AppVersion "1.0.0"
#define AppPublisher "The Iraqi Investor Group"
#define AppURL "https://github.com/haideralmesaody/ISXDailyReportScrapper"
#define AppExeName "isxcli.exe"
#define WebExeName "web.exe"
#define AppSupportURL "https://github.com/haideralmesaody/ISXDailyReportScrapper/issues"
#define AppUpdatesURL "https://github.com/haideralmesaody/ISXDailyReportScrapper/releases"
#define GitHubReleaseAPI "https://api.github.com/repos/haideralmesaody/ISXDailyReportScrapper/releases/latest"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
AppId={{B8F5E8C1-2D4A-4F1B-9C3E-7A5D6F8E9B0C}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppSupportURL}
AppUpdatesURL={#AppUpdatesURL}
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
AllowNoIcons=yes
LicenseFile=assets\LICENSE.txt
InfoBeforeFile=assets\README.txt
InfoAfterFile=assets\AFTER_INSTALL.txt
OutputDir=output
OutputBaseFilename=ISX-Daily-Reports-Scraper-Setup-{#AppVersion}
;SetupIconFile=assets\setup-icon.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
MinVersion=10.0.19041
; Windows 11 compatibility
SetupMutex=ISXDailyReportsScraper_Setup
RestartIfNeededByRun=no
CloseApplications=yes
DirExistsWarning=no
DisableDirPage=no
DisableProgramGroupPage=no
CreateAppDir=yes
UsePreviousAppDir=yes
UsePreviousGroup=yes
AppCopyright=Copyright (C) 2024 The Iraqi Investor Group
AppModifyPath="{uninstallexe}"
UninstallDisplayIcon={app}\{#AppExeName}
UninstallDisplayName={#AppName}
ChangesAssociations=yes
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[CustomMessages]
LaunchProgram=Launch ISX Daily Reports Scraper Web Interface
CreateDesktopIcon=Create a &desktop icon
CreateQuickLaunchIcon=Create a &Quick Launch icon
ProgramOnTheWeb={#AppName} on the Web
UninstallProgram=Uninstall {#AppName}
LaunchWebInterface=Launch Web Interface
ConfigureLicense=Configure License
ViewDocumentation=View Documentation
OpenDataFolder=Open Data Folder
DownloadingFiles=Downloading application files from GitHub...

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1
Name: "associate"; Description: "Associate .xlsx files with ISX processor"; GroupDescription: "File associations"
Name: "addtopath"; Description: "Add application directory to PATH"; GroupDescription: "System integration"
Name: "autostart"; Description: "Enable automatic startup with Windows"; GroupDescription: "System integration"

[Files]
; Configuration files and templates (included in installer)
Source: "assets\license-config-template.json"; DestDir: "{app}"; DestName: "license-config.json"; Flags: ignoreversion onlyifdoesntexist
Source: "assets\app-config.json"; DestDir: "{app}"; Flags: ignoreversion onlyifdoesntexist

; Batch files and scripts (included in installer)
Source: "assets\start-web-interface.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\configure-license.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\run-scraper.bat"; DestDir: "{app}"; Flags: ignoreversion

; Downloader script (included in installer)
Source: "assets\download-github-release.ps1"; DestDir: "{tmp}"; Flags: deleteafterinstall

; Visual C++ Redistributable (downloaded if needed)
;Source: "assets\vc_redist.x64.exe"; DestDir: "{tmp}"; Flags: deleteafterinstall; Check: VCRedistNeedsInstall

[Dirs]
Name: "{app}\downloads"; Flags: uninsneveruninstall
Name: "{app}\reports"; Flags: uninsneveruninstall
Name: "{app}\logs"; Flags: uninsneveruninstall
Name: "{app}\backups"; Flags: uninsneveruninstall
Name: "{app}\temp"
Name: "{app}\tools"
Name: "{app}\docs"
Name: "{app}\web"
Name: "{app}\license-generator"

[Icons]
Name: "{group}\{#AppName} Web Interface"; Filename: "{app}\{#WebExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\web\static\images\favicon.ico"
Name: "{group}\{#AppName} CLI"; Filename: "{app}\{#AppExeName}"; WorkingDir: "{app}"; Parameters: "--help"
Name: "{group}\License Generator"; Filename: "{app}\tools\license-generator.exe"; WorkingDir: "{app}\tools"
Name: "{group}\Configure License"; Filename: "{app}\configure-license.bat"; WorkingDir: "{app}"
Name: "{group}\Documentation"; Filename: "{app}\docs\README.md"
Name: "{group}\Data Folder"; Filename: "{app}\reports"; WorkingDir: "{app}\reports"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"

Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#WebExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\web\static\images\favicon.ico"; Tasks: desktopicon

Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\{#AppName}"; Filename: "{app}\{#WebExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\web\static\images\favicon.ico"; Tasks: quicklaunchicon

[Run]
; Install Microsoft Visual C++ Redistributable if needed
;Filename: "{tmp}\vc_redist.x64.exe"; Parameters: "/quiet /norestart"; StatusMsg: "Installing Microsoft Visual C++ Redistributable..."; Check: VCRedistNeedsInstall; Flags: waituntilterminated

; Download application files from GitHub
Filename: "powershell.exe"; Parameters: "-ExecutionPolicy Bypass -File ""{tmp}\download-github-release.ps1"" ""{app}"""; StatusMsg: "{cm:DownloadingFiles}"; Flags: waituntilterminated; Check: not IsUpgrade

; Configure firewall rules
Filename: "netsh"; Parameters: "advfirewall firewall add rule name=""ISX Daily Reports Scraper Web"" dir=in action=allow protocol=TCP localport=8080"; StatusMsg: "Configuring Windows Firewall..."; Flags: waituntilterminated runhidden; Check: IsAdminInstallMode

; Post-install configuration
Filename: "{app}\configure-license.bat"; Description: "Configure license settings"; Flags: postinstall shellexec skipifsilent; Check: not IsUpgrade

; Launch application
Filename: "{app}\{#WebExeName}"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: postinstall shellexec skipifsilent nowait

[UninstallRun]
; Remove firewall rules
Filename: "netsh"; Parameters: "advfirewall firewall delete rule name=""ISX Daily Reports Scraper Web"""; Flags: waituntilterminated runhidden; Check: IsAdminInstallMode

; Stop any running processes
Filename: "taskkill"; Parameters: "/F /IM isxcli.exe"; Flags: waituntilterminated runhidden
Filename: "taskkill"; Parameters: "/F /IM web.exe"; Flags: waituntilterminated runhidden
Filename: "taskkill"; Parameters: "/F /IM web-licensed.exe"; Flags: waituntilterminated runhidden

[Registry]
; File associations
Root: HKCR; Subkey: ".isxdata"; ValueType: string; ValueName: ""; ValueData: "ISXDataFile"; Flags: uninsdeletevalue; Tasks: associate
Root: HKCR; Subkey: "ISXDataFile"; ValueType: string; ValueName: ""; ValueData: "ISX Data File"; Flags: uninsdeletekey; Tasks: associate
Root: HKCR; Subkey: "ISXDataFile\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\{#AppExeName},0"; Tasks: associate
Root: HKCR; Subkey: "ISXDataFile\shell\open\command"; ValueType: string; ValueName: ""; ValueData: """{app}\{#AppExeName}"" ""%1"""; Tasks: associate

; Add to PATH
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: NeedsAddPath('{app}'); Tasks: addtopath; Flags: uninsdeletevalue

; Auto-start registry entries
Root: HKCU; Subkey: "SOFTWARE\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "ISXDailyReportsScraper"; ValueData: """{app}\{#WebExeName}"""; Tasks: autostart; Flags: uninsdeletevalue

; Application settings
Root: HKCU; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: string; ValueName: "Version"; ValueData: "{#AppVersion}"; Flags: uninsdeletekey
Root: HKCU; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: dword; ValueName: "FirstRun"; ValueData: "1"; Flags: uninsdeletekey

[Code]
// Import Windows API functions
// function InternetCheckConnection(lpszUrl: PAnsiChar; dwFlags: DWORD; dwReserved: DWORD): BOOL; external 'InternetCheckConnectionA@wininet.dll stdcall';

// function VCRedistNeedsInstall: Boolean;
// var
//   Version: String;
// begin
//   Result := not RegQueryStringValue(HKLM, 'SOFTWARE\Microsoft\VisualStudio\14.0\VC\Runtimes\x64', 'Version', Version);
// end;

function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', OrigPath)
  then begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;

function IsUpgrade: Boolean;
begin
  Result := (GetPreviousData('AppVersion', '') <> '');
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssInstall then
  begin
    // SetProgress(0, 100);
    // if VCRedistNeedsInstall then
    //   ExtractTemporaryFile('vc_redist.x64.exe');
    ExtractTemporaryFile('download-github-release.ps1');
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    // Clean up any remaining files
    DelTree(ExpandConstant('{app}\logs'), True, True, True);
    DelTree(ExpandConstant('{app}\temp'), True, True, True);
    DelTree(ExpandConstant('{app}\backups'), True, True, True);
  end;
end;

function InitializeSetup(): Boolean;
var
  Version: TWindowsVersion;
begin
  GetWindowsVersionEx(Version);
  
  // Check Windows version
  if Version.Major < 6 then
  begin
    MsgBox('This application requires Windows 7 or later.', mbCriticalError, MB_OK);
    Result := False;
    Exit;
  end;
  
  // Check internet connection
  // if not InternetCheckConnection(PAnsiChar('http://www.google.com'), 1, 0) then
  // begin
  //   if MsgBox('Internet connection is required to download application files. Continue anyway?', mbConfirmation, MB_YESNO) = IDNO then
  //   begin
  //     Result := False;
  //     Exit;
  //   end;
  // end;
  
  Result := True;
end;

function InternetCheckConnection(lpszUrl: String; dwFlags: DWORD; dwReserved: DWORD): Boolean;
external 'InternetCheckConnectionA@wininet.dll stdcall';

[Messages]
BeveledLabel=The Iraqi Investor Group
SetupAppTitle=Setup - {#AppName}
SetupWindowTitle=Setup - {#AppName} 