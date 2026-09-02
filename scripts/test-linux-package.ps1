[CmdletBinding()]
param(
    [string]$Ref = "HEAD"
)

$ErrorActionPreference = "Stop"
$temporaryDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ("kiosk-client-package-" + [guid]::NewGuid())
[System.IO.Directory]::CreateDirectory($temporaryDirectory) | Out-Null

try {
    $archivePath = Join-Path $temporaryDirectory "source.zip"
    & (Join-Path $PSScriptRoot "package-linux-source.ps1") -Ref $Ref -OutputPath $archivePath -Verify
    if ($LASTEXITCODE -ne 0) {
        throw "linux package verification failed"
    }
}
finally {
    Remove-Item -LiteralPath $temporaryDirectory -Recurse -Force -ErrorAction SilentlyContinue
}
