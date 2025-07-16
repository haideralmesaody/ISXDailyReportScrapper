; ISX Daily Reports Scraper - Windows Installer Script
; Created with Inno Setup 6.0+
; Company: The Iraqi Investor Group

#define AppName "ISX Daily Reports Scraper"
#define AppVersion "1.0.0"
#define AppPublisher "The Iraqi Investor Group"
#define AppURL "https://github.com/haideralmesaody/ISXDailyReportScraper"
#define AppExeName "isxcli.exe"
#define WebExeName "web-licensed.exe"
#define AppSupportURL "https://github.com/haideralmesaody/ISXDailyReportScraper/issues"
#define AppUpdatesURL "https://github.com/haideralmesaody/ISXDailyReportScraper/releases"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
; Do not use the same AppId value in installers for other applications.
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
SetupIconFile=assets\setup-icon.ico
Compression=lzma
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=admin
MinVersion=10.0.19041
DisableDirPage=no
DisableProgramGroupPage=no
CreateAppDir=yes
UsePreviousAppDir=yes
UsePreviousGroup=yes
AppCopyright=Copyright (C) 2024 The Iraqi Investor Group
AppModifyPath="{uninstallexe}"
UninstallDisplayIcon={app}\{#WebExeName}
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

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1
Name: "associate"; Description: "Associate .xlsx files with ISX processor"; GroupDescription: "File associations"
Name: "addtopath"; Description: "Add application directory to PATH"; GroupDescription: "System integration"
Name: "autostart"; Description: "Enable automatic startup with Windows"; GroupDescription: "System integration"

[Files]
; Main executables
Source: "..\isxcli.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\cmd\web-licensed\web-licensed.exe"; DestDir: "{app}"; Flags: ignoreversion; DestName: "web.exe"
Source: "..\cmd\process\process.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\cmd\indexcsv\indexcsv.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\cmd\license-generator\license-generator.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\cmd\bulk-license-generator\bulk-license-generator.exe"; DestDir: "{app}\tools"; Flags: ignoreversion

; Additional tools
Source: "..\cmd\marketscan\marketscan.exe"; DestDir: "{app}\tools"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\cmd\combine\combine.exe"; DestDir: "{app}\tools"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\cmd\inspect\inspect.exe"; DestDir: "{app}\tools"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\cmd\identifyformats\identifyformats.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\cmd\sampleformats\sampleformats.exe"; DestDir: "{app}\tools"; Flags: ignoreversion
Source: "..\cmd\debugindices\debugindices.exe"; DestDir: "{app}\tools"; Flags: ignoreversion

; Web interface files
Source: "..\web\*"; DestDir: "{app}\web"; Flags: ignoreversion recursesubdirs createallsubdirs

; License generator app
Source: "..\license-generator-app\*"; DestDir: "{app}\license-generator"; Flags: ignoreversion recursesubdirs createallsubdirs

; Configuration files
Source: "assets\license-config-template.json"; DestDir: "{app}"; DestName: "license-config.json"; Flags: ignoreversion onlyifdoesntexist
Source: "assets\app-config.json"; DestDir: "{app}"; Flags: ignoreversion onlyifdoesntexist

; Application icon
Source: "assets\setup-icon.ico"; DestDir: "{app}"; DestName: "app-icon.ico"; Flags: ignoreversion

; Documentation
Source: "..\README.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\WEB_README.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\WEB_INTERFACE_GUIDE.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\LICENSING_SETUP.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\EXPIRE_STATUS_SETUP.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\GOOGLE_SHEETS_SETUP.md"; DestDir: "{app}\docs"; Flags: ignoreversion
Source: "..\RECHARGE_CARD_SETUP.md"; DestDir: "{app}\docs"; Flags: ignoreversion

; Batch files and scripts
Source: "..\build-web.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\start-web-interface.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\configure-license.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "assets\run-scraper.bat"; DestDir: "{app}"; Flags: ignoreversion

; Dependencies and runtime files
Source: "..\go.mod"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\go.sum"; DestDir: "{app}"; Flags: ignoreversion

; NOTE: Don't use "Flags: ignoreversion" on any shared system files

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
Name: "{group}\{#AppName} Web Interface"; Filename: "{app}\{#WebExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\app-icon.ico"
Name: "{group}\{#AppName} CLI"; Filename: "{app}\{#AppExeName}"; WorkingDir: "{app}"; Parameters: "--help"
Name: "{group}\License Generator"; Filename: "{app}\tools\license-generator.exe"; WorkingDir: "{app}\tools"
Name: "{group}\Configure License"; Filename: "{app}\configure-license.bat"; WorkingDir: "{app}"
Name: "{group}\Documentation"; Filename: "{app}\docs\README.md"
Name: "{group}\Data Folder"; Filename: "{app}\reports"; WorkingDir: "{app}\reports"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"

Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#WebExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\app-icon.ico"; Tasks: desktopicon

Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\{#AppName}"; Filename: "{app}\{#WebExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\app-icon.ico"; Tasks: quicklaunchicon

[Run]
; Install Microsoft Visual C++ Redistributable if needed
Filename: "{tmp}\vc_redist.x64.exe"; Parameters: "/quiet /norestart"; StatusMsg: "Installing Microsoft Visual C++ Redistributable..."; Check: VCRedistNeedsInstall; Flags: waituntilterminated

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
function VCRedistNeedsInstall: Boolean;
var
  Version: String;
begin
  Result := not RegQueryStringValue(HKLM, 'SOFTWARE\Microsoft\VisualStudio\14.0\VC\Runtimes\x64', 'Version', Version);
end;

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
    // SetProgress(0, 100);  // Function not available in Inno Setup
    ExtractTemporaryFile('vc_redist.x64.exe');
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
  
  // Check for .NET Framework (if needed)
  // Add additional checks here if needed
  
  Result := True;
end;

procedure InitializeWizard();
begin
  // Custom initialization code
end;

[Messages]
BeveledLabel=The Iraqi Investor Group
SetupAppTitle=Setup - {#AppName}
SetupWindowTitle=Setup - {#AppName} 