; ISX Daily Reports Scraper - Smart Installer
; Downloads latest release from GitHub automatically
; Company: The Iraqi Investor Group

#define AppName "ISX Daily Reports Scraper"
#define AppVersion "1.0-Alpha"
#define AppPublisher "The Iraqi Investor Group"
#define AppURL "https://github.com/haideralmesaody/ISXDailyReportScrapper"
#define AppExeName "isx-web-interface.exe"
#define CliExeName "isxcli.exe"

; GitHub Release URLs - Update these when you create releases
#define GitHubReleasesAPI "https://api.github.com/repos/haideralmesaody/ISXDailyReportScrapper/releases/latest"
#define GitHubDownloadBase "https://github.com/haideralmesaody/ISXDailyReportScrapper/releases/latest/download"

[Setup]
AppId={{A1B2C3D4-5E6F-7890-ABCD-123456789ABC}
AppName={#AppName} (Alpha)
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
DefaultDirName={autopf}\ISX
DefaultGroupName=ISX Daily Reports Scraper (Alpha)
OutputDir=..\release
OutputBaseFilename=ISX-Smart-Installer
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
AppCopyright=Copyright (C) 2025 The Iraqi Investor Group
UninstallDisplayIcon={app}\bin\{#AppExeName}
UninstallDisplayName={#AppName} (Alpha)
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: checked
Name: "addtopath"; Description: "Add ISX commands to PATH environment variable"; GroupDescription: "System Integration"; Flags: checked
Name: "installchrome"; Description: "Download and install Chrome browser (if not found)"; GroupDescription: "Dependencies"; Flags: checked

[Files]
; Only include essential files in installer - everything else downloaded
Source: "assets\setup-icon.ico"; DestDir: "{tmp}"; Flags: dontcopy noencryption
Source: "assets\chrome-installer.exe"; DestDir: "{tmp}"; Flags: dontcopy noencryption external skipifnotexists

[Icons]
; Start Menu
Name: "{group}\ISX Web Interface"; Filename: "{app}\bin\{#AppExeName}"; WorkingDir: "{app}"
Name: "{group}\ISX Command Line"; Filename: "{app}\bin\{#CliExeName}"; WorkingDir: "{app}"; Parameters: "--help"
Name: "{group}\User Guide"; Filename: "https://github.com/haideralmesaody/ISXDailyReportScrapper/blob/main/release/docs/ALPHA-USER-GUIDE.md"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"

; Desktop
Name: "{autodesktop}\ISX Web Interface"; Filename: "{app}\bin\{#AppExeName}"; WorkingDir: "{app}"; Tasks: desktopicon

[Registry]
; Application registration
Root: HKLM; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: string; ValueName: "InstallPath"; ValueData: "{app}"; Flags: uninsdeletevalue uninsdeletekeyifempty
Root: HKLM; Subkey: "SOFTWARE\{#AppPublisher}\{#AppName}"; ValueType: string; ValueName: "Version"; ValueData: "{#AppVersion}"; Flags: uninsdeletevalue uninsdeletekeyifempty

[Run]
; Start the application after installation
Filename: "{app}\bin\{#AppExeName}"; Description: "Launch ISX Web Interface"; Flags: postinstall nowait skipifsilent unchecked

[Code]
const
  EnvironmentKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';
  
var
  DownloadPage: TDownloadWizardPage;

function OnDownloadProgress(const Url, FileName: String; const Progress, ProgressMax: Int64): Boolean;
begin
  if Progress = ProgressMax then
    Log(Format('Successfully downloaded %s', [FileName]));
  Result := True;
end;

procedure InitializeWizard;
begin
  // Create the download page
  DownloadPage := CreateDownloadPage(SetupMessage(msgWizardPreparing), SetupMessage(msgPreparingDesc), @OnDownloadProgress);
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  
  if CurPageID = wpReady then begin
    // Check internet connection first
    if not CheckForInternetConnection then begin
      MsgBox('Internet connection required to download ISX components. Please check your connection and try again.', mbError, MB_OK);
      Result := False;
      Exit;
    end;
    
    // Check for Chrome browser
    if WizardIsTaskSelected('installchrome') then begin
      if not CheckChromeInstalled then begin
        if MsgBox('Chrome browser not found. Download and install Chrome?', mbConfirmation, MB_YESNO) = IDNO then begin
          Result := False;
          Exit;
        end;
      end;
    end;
    
    DownloadPage.Clear;
    
    // Add downloads to the queue
    DownloadPage.Add('{#GitHubDownloadBase}/isx-web-interface.exe', 'isx-web-interface.exe', '');
    DownloadPage.Add('{#GitHubDownloadBase}/isxcli.exe', 'isxcli.exe', '');
    DownloadPage.Add('{#GitHubDownloadBase}/web-assets.zip', 'web-assets.zip', '');
    DownloadPage.Add('{#GitHubDownloadBase}/docs.zip', 'docs.zip', '');
    
    // Add Chrome installer if needed
    if WizardIsTaskSelected('installchrome') and not CheckChromeInstalled then begin
      DownloadPage.Add('https://dl.google.com/chrome/install/latest/chrome_installer.exe', 'chrome_installer.exe', '');
    end;
    
    DownloadPage.Show;
    try
      try
        DownloadPage.Download; // This downloads the files to {tmp}
        Result := True;
      except
        if DownloadPage.AbortedByUser then
          Log('Aborted by user.')
        else
          SuppressibleMsgBox(AddPeriod(GetExceptionMessage), mbCriticalError, MB_OK, IDOK);
        Result := False;
      end;
    finally
      DownloadPage.Hide;
    end;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ChromeInstaller: String;
  ResultCode: Integer;
begin
  if CurStep = ssInstall then begin
    SetCurrentDir(ExpandConstant('{tmp}'));
    
    // Create directory structure
    CreateDir(ExpandConstant('{app}\bin'));
    CreateDir(ExpandConstant('{app}\web'));
    CreateDir(ExpandConstant('{app}\docs'));
    CreateDir(ExpandConstant('{app}\data'));
    CreateDir(ExpandConstant('{app}\data\downloads'));
    CreateDir(ExpandConstant('{app}\data\reports'));
    
    // Copy downloaded executables
    if FileExists(ExpandConstant('{tmp}\isx-web-interface.exe')) then
      FileCopy(ExpandConstant('{tmp}\isx-web-interface.exe'), ExpandConstant('{app}\bin\isx-web-interface.exe'), False);
    
    if FileExists(ExpandConstant('{tmp}\isxcli.exe')) then
      FileCopy(ExpandConstant('{tmp}\isxcli.exe'), ExpandConstant('{app}\bin\isxcli.exe'), False);
    
    // Extract downloaded archives
    if FileExists(ExpandConstant('{tmp}\web-assets.zip')) then
      Exec('powershell.exe', '-Command "Expand-Archive -Path ''' + ExpandConstant('{tmp}\web-assets.zip') + ''' -DestinationPath ''' + ExpandConstant('{app}\web') + ''' -Force"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    
    if FileExists(ExpandConstant('{tmp}\docs.zip')) then
      Exec('powershell.exe', '-Command "Expand-Archive -Path ''' + ExpandConstant('{tmp}\docs.zip') + ''' -DestinationPath ''' + ExpandConstant('{app}\docs') + ''' -Force"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    
    // Install Chrome if downloaded
    ChromeInstaller := ExpandConstant('{tmp}\chrome_installer.exe');
    if FileExists(ChromeInstaller) then begin
      Log('Installing Chrome browser...');
      Exec(ChromeInstaller, '/silent /install', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    end;
  end;
  
  if (CurStep = ssPostInstall) and WizardIsTaskSelected('addtopath') then
    EnvAddPath(ExpandConstant('{app}\bin'));
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
    EnvRemovePath(ExpandConstant('{app}\bin'));
end;

// Utility functions
function CheckForInternetConnection: Boolean;
var
  ResultCode: Integer;
begin
  Result := Exec('ping', 'google.com -n 1 -w 3000', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) and (ResultCode = 0);
end;

function CheckChromeInstalled: Boolean;
begin
  Result := FileExists(ExpandConstant('{pf}\Google\Chrome\Application\chrome.exe')) or 
            FileExists(ExpandConstant('{pf32}\Google\Chrome\Application\chrome.exe')) or
            FileExists(ExpandConstant('{localappdata}\Google\Chrome\Application\chrome.exe'));
end;

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