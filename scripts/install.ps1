# install.ps1
$ErrorActionPreference = "Stop"

# Configuration
$BinaryName = "kata"
$Repo = "phantompunk/kata"
$InstallPath = if ($env:INSTALL_PATH) { $env:INSTALL_PATH } else { "$env:USERPROFILE\.local\bin" }

# Helper functions
function Write-Info($msg) {
    Write-Host "==> " -NoNewline -ForegroundColor Blue
    Write-Host $msg
}

function Write-Success($msg) {
    Write-Host "✓ " -NoNewline -ForegroundColor Green
    Write-Host $msg
}

function Write-ErrorMsg($msg) {
    Write-Host "✗ " -NoNewline -ForegroundColor Red
    Write-Host $msg
}

function Write-Warn($msg) {
    Write-Host "! " -NoNewline -ForegroundColor Yellow
    Write-Host $msg
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { 
            Write-ErrorMsg "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Get latest release version from GitHub
function Get-LatestVersion {
    $url = "https://api.github.com/repos/$Repo/releases/latest"
    
    try {
        $response = Invoke-RestMethod -Uri $url -Method Get -ErrorAction Stop
        $version = $response.tag_name
        
        if ([string]::IsNullOrEmpty($version)) {
            Write-ErrorMsg "Could not determine latest version from $url"
            exit 1
        }
        
        return $version
    }
    catch {
        Write-ErrorMsg "Could not fetch latest version: $_"
        exit 1
    }
}

# Download file
function Get-File {
    param(
        [string]$Url,
        [string]$OutFile
    )
    
    try {
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -ErrorAction Stop
    }
    catch {
        Write-ErrorMsg "Could not download from $Url"
        Write-ErrorMsg "Error: $_"
        exit 1
    }
}

# Main installation
Write-Info "Installing $BinaryName..."

# Detect system
$OS = "windows"
$Arch = Get-Architecture
Write-Info "Detected: $OS/$Arch"

# Get version
Write-Info "Fetching latest release..."
if ($env:VERSION) {
    $Version = $env:VERSION
}
else {
    $Version = Get-LatestVersion
}
Write-Info "Found version: $Version"

# Construct download URL
$Archive = "${BinaryName}_${Version}_${OS}_${Arch}.zip"
$Binary = "${BinaryName}.exe"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Archive"

Write-Info "Downloading $BinaryName..."

# Create temp directory
$TempDir = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "kata_$(Get-Random)"))

try {
    # Download archive
    $ArchivePath = Join-Path $TempDir.FullName $Archive
    Get-File -Url $DownloadUrl -OutFile $ArchivePath
    
    # Extract archive
    Write-Info "Extracting..."
    try {
        Expand-Archive -Path $ArchivePath -DestinationPath $TempDir.FullName -Force -ErrorAction Stop
    }
    catch {
        Write-ErrorMsg "Failed to extract archive"
        exit 1
    }
    
    # Verify binary exists
    $BinaryPath = Join-Path $TempDir.FullName $Binary
    if (-not (Test-Path $BinaryPath)) {
        Write-ErrorMsg "Binary $Binary not found in archive"
        exit 1
    }
    
    # Install binary
    Write-Info "Installing to $InstallPath..."
    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }
    
    $DestPath = Join-Path $InstallPath $Binary
    try {
        Copy-Item -Path $BinaryPath -Destination $DestPath -Force -ErrorAction Stop
    }
    catch {
        Write-ErrorMsg "Failed to copy binary to $InstallPath"
        exit 1
    }
    
    Write-Success "Installed $BinaryName to $DestPath"
    
    # Check if install path is in PATH
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $MachinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $CurrentPath = "$UserPath;$MachinePath"
    
    if ($CurrentPath -like "*$InstallPath*") {
        Write-Success "Installation complete!"
        Write-Info "Run: $BinaryName --help"
    }
    else {
        Write-Success "Installation complete!"
        Write-Warn "$InstallPath is not in your PATH"
        Write-Info "Add it by running (as Administrator):"
        Write-Host '    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")' -ForegroundColor Gray
        Write-Host "    [Environment]::SetEnvironmentVariable(`"Path`", `"`$currentPath;$InstallPath`", `"User`")" -ForegroundColor Gray
        Write-Host ""
        Write-Info "Or restart your terminal and it may be available"
    }
}
finally {
    # Cleanup temp directory
    if (Test-Path $TempDir.FullName) {
        Remove-Item -Path $TempDir.FullName -Recurse -Force -ErrorAction SilentlyContinue
    }
}
