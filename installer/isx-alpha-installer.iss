; ISX Daily Reports Scraper - Alpha Release Installer
; Created with Inno Setup 6.0+
; Company: The Iraqi Investor Group

#define AppName "ISX Daily Reports Scraper"
#define AppVersion "1.0-Alpha"
#define AppPublisher "The Iraqi Investor Group"
#define AppURL "https://github.com/haideralmesaody/ISXDailyReportScrapper"
#define AppExeName "isx-web-interface.exe"
#define CliExeName "isxcli.exe"
#define AppSupportURL "https://github.com/haideralmesaody/ISXDailyReportScrapper/issues"
#define AppUpdatesURL "https://github.com/haideralmesaody/ISXDailyReportScrapper/releases"

[Setup]
; NOTE: The value of AppId uniquely identifies this application.
AppId={{A1B2C3D4-5E6F-7890-ABCD-123456789ABC}
AppName={#AppName} (Alpha)
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppSupportURL}
AppUpdatesURL={#AppUpdatesURL}
DefaultDirName={autopf}\ISX
DefaultGroupName=ISX Daily Reports Scraper (Alpha)
AllowNoIcons=yes
InfoBeforeFile=..\release\README-ALPHA.md
InfoAfterFile=..\release\QUICK_START.md
OutputDir=..\release
OutputBaseFilename=ISX-Alpha-Installer
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
AppCopyright=Copyright (C) 2025 The Iraqi Investor Group
UninstallDisplayIcon={app}\bin\{#AppExeName}
UninstallDisplayName={#AppName} (Alpha)
ChangesAssociations=no
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: checked
Name: "quicklaunchicon"; Description: "{cm:CreateQuickLaunchIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked; OnlyBelowVersion: 6.1
Name: "addtopath"; Description: "Add ISX commands to PATH environment variable"; GroupDescription: "System Integration"; Flags: checked

[Files]
; Main executables
Source: "..\release\bin\isx-web-interface.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\release\bin\isxcli.exe"; DestDir: "{app}\bin"; Flags: ignoreversion

; Web interface files
Source: "..\release\web\*"; DestDir: "{app}\web"; Flags: ignoreversion recursesubdirs createallsubdirs

; Documentation
Source: "..\release\docs\*"; DestDir: "{app}\docs"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\release\README-ALPHA.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\QUICK_START.md"; DestDir: "{app}"; Flags: ignoreversion

; Data directories (create empty)
Source: "..\release\data\*"; DestDir: "{app}\data"; Flags: ignoreversion recursesubdirs createallsubdirs uninsneveruninstall

; Tools
Source: "..\release\tools\*"; DestDir: "{app}\tools"; Flags: ignoreversion recursesubdirs createallsubdirs

; Installation scripts for reference
Source: "..\release\install-alpha.ps1"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\release\verify-package.ps1"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu
Name: "{group}\ISX Web Interface"; Filename: "{app}\bin\{#AppExeName}"; WorkingDir: "{app}"; IconFilename: "{app}\web\static\images\favicon.ico"
Name: "{group}\ISX Command Line"; Filename: "{app}\bin\{#CliExeName}"; WorkingDir: "{app}"; Parameters: "--help"
Name: "{group}\User Guide"; Filename: "{app}\docs\ALPHA-USER-GUIDE.md"
Name: "{group}\Testing Guide"; Filename: "{app}\docs\ALPHA-TESTING-GUIDE.md"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"

; Desktop
Name: "{autodesktop}\ISX Web Interface"; Filename: "{app}\bin\{#AppExeName}"; WorkingDir: "{app}"; Tasks: desktopicon; IconFilename: "{app}\web\static\images\favicon.ico"

; Quick Launch
Name: "{userappdata}\Microsoft\Internet Explorer\Quick Launch\ISX Web Interface"; Filename: "{app}\bin\{#AppExeName}"; WorkingDir: "{app}"; Tasks: quicklaunchicon

[Registry]
; File associations (optional for future use)
Root: HKLM; Subkey: "SOFTWARE\Classes\.isx"; ValueType: string; ValueName: ""; ValueData: "ISXDataFile"; Flags: uninsdeletevalue uninsdeletekeyifempty
Root: HKLM; Subkey: "SOFTWARE\Classes\ISXDataFile"; ValueType: string; ValueName: ""; ValueData: "ISX Data File"; Flags: uninsdeletevalue uninsdeletekeyifempty
Root: HKLM; Subkey: "SOFTWARE\Classes\ISXDataFile\DefaultIcon"; ValueType: string; ValueName: ""; ValueData: "{app}\bin\{#AppExeName},0"; Flags: uninsdeletevalue uninsdeletekeyifempty

; Application registration
Root: HKLM; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletevalue uninsdeletekeyifempty
Root: HKLM; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: string; ValueName: "Version"; ValueData: "{#AppVersion}"; Flags: uninsdeletevalue uninsdeletekeyifempty

[Run]
; Open readme after installation
Filename: "{app}\README-ALPHA.md"; Description: "View Alpha Release Information"; Flags: postinstall nowait skipifsilent shellexec unchecked
; Start the application
Filename: "{app}\bin\{#AppExeName}"; Description: "Launch ISX Web Interface"; Flags: postinstall nowait skipifsilent unchecked
; Open user guide
Filename: "{app}\docs\ALPHA-USER-GUIDE.md"; Description: "Open User Guide"; Flags: postinstall nowait skipifsilent shellexec unchecked

[UninstallRun]
; Stop any running ISX processes before uninstall
Filename: "{cmd}"; Parameters: "/C taskkill /F /IM {#AppExeName} /T"; Flags: runhidden; RunOnceId: "StopISXWeb"
Filename: "{cmd}"; Parameters: "/C taskkill /F /IM {#CliExeName} /T"; Flags: runhidden; RunOnceId: "StopISXCli"

[Code]
const
  EnvironmentKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';

procedure EnvAddPath(Path: string);
var
  Paths: string;
begin
  { Retrieve current path }
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
    Paths := '';

  { Skip if string already found in path }
  if Pos(';' + Uppercase(Path) + ';', ';' + Uppercase(Paths) + ';') > 0 then exit;

  { App string to the end of the path variable }
  Paths := Paths + ';'+ Path +';'

  { Overwrite (or create if missing) path environment variable }
  if RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths)
  then Log(Format('The [%s] added to PATH: [%s]', [Path, Paths]))
  else Log(Format('Error while adding the [%s] to PATH: [%s]', [Path, Paths]));
end;

procedure EnvRemovePath(Path: string);
var
  Paths: string;
  P: Integer;
begin
  { Skip if registry entry not exists }
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths) then
    exit;

  { Skip if string not found in path }
  P := Pos(';' + Uppercase(Path) + ';', ';' + Uppercase(Paths) + ';');
  if P = 0 then exit;

  { Update path variable }
  Delete(Paths, P - 1, Length(Path) + 1);

  { Overwrite path environment variable }
  if RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', Paths)
  then Log(Format('The [%s] removed from PATH: [%s]', [Path, Paths]))
  else Log(Format('Error while removing the [%s] from PATH: [%s]', [Path, Paths]));
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if (CurStep = ssPostInstall) and WizardIsTaskSelected('addtopath')
  then EnvAddPath(ExpandConstant('{app}\bin'));
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall
  then EnvRemovePath(ExpandConstant('{app}\bin'));
end;

function GetUninstallString(): String;
var
  sUnInstPath: String;
  sUnInstallString: String;
Begin
  sUnInstPath := ExpandConstant('Software\Microsoft\Windows\CurrentVersion\Uninstall\{#emit SetupSetting("AppId")}_is1');
  sUnInstallString := '';
  if not RegQueryStringValue(HKLM, sUnInstPath, 'UninstallString', sUnInstallString) then
    RegQueryStringValue(HKCU, sUnInstPath, 'UninstallString', sUnInstallString);
  Result := sUnInstallString;
End;

function IsUpgrade(): Boolean;
Begin
  Result := (GetUninstallString() <> '');
End;

function UnInstallOldVersion(): Integer;
var
  sUnInstallString: String;
  iResultCode: Integer;
Begin
  // Return Values:
  // 1 - uninstall string is empty
  // 2 - error executing the UnInstallString
  // 3 - successfully executed the UnInstallString

  // default return value
  Result := 0;

  // get the uninstall string of the old app
  sUnInstallString := GetUninstallString();
  if sUnInstallString <> '' then begin
    sUnInstallString := RemoveQuotes(sUnInstallString);
    if Exec(sUnInstallString, '/SILENT /NORESTART /SUPPRESSMSGBOXES','', SW_HIDE, ewWaitUntilTerminated, iResultCode) then
      Result := 3
    else
      Result := 2;
  end else
    Result := 1;
End;

procedure CurPageChanged(CurPageID: Integer);
begin
  if (CurPageID = wpWelcome) then
  begin
    if (IsUpgrade()) then
    begin
      if (MsgBox('An older version of ISX Daily Reports Scraper is already installed. Would you like to uninstall it first?', mbConfirmation, MB_YESNO) = IDYES) then
      begin
        UnInstallOldVersion();
      end;
    end;
  end;
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = wpReady then begin
    // Check for Chrome browser
    if not FileExists(ExpandConstant('{pf}\Google\Chrome\Application\chrome.exe')) and 
       not FileExists(ExpandConstant('{pf32}\Google\Chrome\Application\chrome.exe')) and
       not FileExists(ExpandConstant('{localappdata}\Google\Chrome\Application\chrome.exe')) then begin
      if MsgBox('Chrome browser was not detected. ISX requires Chrome for web scraping. Continue anyway?', mbConfirmation, MB_YESNO) = IDNO then begin
        Result := False;
      end;
    end;
  end;
end; 