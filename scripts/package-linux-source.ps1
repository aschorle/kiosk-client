[CmdletBinding()]
param(
    [string]$Ref = "HEAD",
    [Parameter(Mandatory = $true)]
    [string]$OutputPath,
    [switch]$Verify
)

$ErrorActionPreference = "Stop"

function Invoke-Git {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)

    $result = & git @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Arguments -join ' ') failed"
    }
    return $result
}

function Test-LinuxSourceArchive {
    param(
        [Parameter(Mandatory = $true)][string]$ArchivePath,
        [Parameter(Mandatory = $true)][string]$Commit
    )

    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $archive = [System.IO.Compression.ZipFile]::OpenRead($ArchivePath)
    try {
        $entries = @($archive.Entries | Where-Object { -not $_.FullName.EndsWith("/") })
        $actual = @($entries | ForEach-Object FullName | Sort-Object)
        $expected = @(Invoke-Git ls-tree -r --name-only $Commit | Sort-Object)
        $difference = Compare-Object -ReferenceObject $expected -DifferenceObject $actual
        if ($difference) {
            throw "archive contents do not match git commit $Commit"
        }

        $shellEntries = @($entries | Where-Object { $_.FullName.EndsWith(".sh") })
        if ($shellEntries.Count -eq 0) {
            throw "archive contains no shell scripts"
        }

        $managementInstaller = $false
        foreach ($entry in $shellEntries) {
            if ($entry.FullName -eq "installer/install-management-only.sh") {
                $managementInstaller = $true
            }

            $stream = $entry.Open()
            try {
                $buffer = New-Object System.IO.MemoryStream
                $stream.CopyTo($buffer)
                [byte[]]$bytes = $buffer.ToArray()
            }
            finally {
                $stream.Dispose()
            }

            if ($bytes -contains [byte]13) {
                throw "shell script contains CR character: $($entry.FullName)"
            }
            if ($bytes.Length -lt 3 -or $bytes[0] -ne [byte][char]'#' -or $bytes[1] -ne [byte][char]'!') {
                throw "shell script has no shebang: $($entry.FullName)"
            }
            if (-not ($bytes -contains [byte]10)) {
                throw "shell script has no LF line ending: $($entry.FullName)"
            }
        }

        if (-not $managementInstaller) {
            throw "archive is missing installer/install-management-only.sh"
        }
    }
    finally {
        $archive.Dispose()
    }
}

$commit = (Invoke-Git rev-parse "$Ref^{commit}").Trim()
$archivePath = [System.IO.Path]::GetFullPath($OutputPath)
$archiveDirectory = Split-Path -Parent $archivePath
if (-not (Test-Path -LiteralPath $archiveDirectory -PathType Container)) {
    throw "output directory does not exist: $archiveDirectory"
}
if (Test-Path -LiteralPath $archivePath) {
    throw "refusing to overwrite existing archive: $archivePath"
}

Invoke-Git -c core.autocrlf=false archive --format=zip "--output=$archivePath" $commit | Out-Null
if ($Verify) {
    Test-LinuxSourceArchive -ArchivePath $archivePath -Commit $commit
}

Write-Output "Created Linux source archive from ${commit}: $archivePath"
