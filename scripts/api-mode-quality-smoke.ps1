param(
    [string]$BackendUrl = "http://localhost:8080",
    [switch]$AllowMockFallback
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$cases = @(
    @{
        Name = "3chapters-short"
        File = Join-Path $repoRoot "docs/examples/api-smoke/novel-3chapters-short.md"
        ExpectedChapters = 3
    },
    @{
        Name = "5chapters-medium"
        File = Join-Path $repoRoot "docs/examples/api-smoke/novel-5chapters-medium.md"
        ExpectedChapters = 5
    },
    @{
        Name = "6chapters-long"
        File = Join-Path $repoRoot "docs/examples/api-smoke/novel-6chapters-long.md"
        ExpectedChapters = 6
    }
)

function Assert-True {
    param(
        [bool]$Condition,
        [string]$Message
    )

    if (-not $Condition) {
        throw $Message
    }
}

function Get-YamlFieldCount {
    param(
        [string]$Yaml,
        [string]$Pattern
    )

    return ([regex]::Matches($Yaml, $Pattern, [System.Text.RegularExpressions.RegexOptions]::Multiline)).Count
}

function Test-ScreenplayYaml {
    param(
        [string]$Yaml,
        [int]$ExpectedChapters
    )

    Assert-True ($Yaml.Trim().Length -gt 0) "screenplay_yaml is empty"
    Assert-True (-not $Yaml.Contains('```')) "screenplay_yaml contains markdown code fence"

    $requiredFragments = @(
        "schema_version:",
        "metadata:",
        "source_chapter_count:",
        "generated_by:",
        "characters:",
        "source_chapters:",
        "screenplay:",
        "acts:",
        "scenes:",
        "beats:"
    )

    foreach ($fragment in $requiredFragments) {
        Assert-True ($Yaml.Contains($fragment)) "missing YAML fragment: $fragment"
    }

    $chapterCountMatch = [regex]::Match($Yaml, 'source_chapter_count:\s*"?(\d+)"?')
    Assert-True $chapterCountMatch.Success "missing metadata.source_chapter_count"
    Assert-True ([int]$chapterCountMatch.Groups[1].Value -eq $ExpectedChapters) "metadata.source_chapter_count mismatch"

    $sourceChapterCount = Get-YamlFieldCount $Yaml '^\s+- id:\s*"?chapter_\d+'
    Assert-True ($sourceChapterCount -ge $ExpectedChapters) "source_chapters count is less than expected"

    for ($i = 1; $i -le $ExpectedChapters; $i++) {
        $chapterID = "chapter_{0:D3}" -f $i
        Assert-True ($Yaml.Contains($chapterID)) "missing source chapter reference: $chapterID"
    }

    $sceneCount = Get-YamlFieldCount $Yaml '^\s+- id:\s*"?scene_\d+'
    Assert-True ($sceneCount -ge 1) "missing scenes"

    $characterCount = Get-YamlFieldCount $Yaml '^\s+- id:\s*"?char_\d+'
    Assert-True ($characterCount -ge 1) "missing characters"
}

function Invoke-Case {
    param(
        [hashtable]$Case
    )

    $content = [string](Get-Content -LiteralPath $Case.File -Raw -Encoding UTF8)
    $bodyJson = @{
        title = $Case.Name
        input_type = "md"
        content = $content
    } | ConvertTo-Json
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($bodyJson)

    $watch = [System.Diagnostics.Stopwatch]::StartNew()
    $response = Invoke-RestMethod `
        -Uri "$BackendUrl/api/convert" `
        -Method Post `
        -ContentType "application/json; charset=utf-8" `
        -Body $bodyBytes `
        -TimeoutSec 360
    $watch.Stop()

    Assert-True ($response.chapter_count -eq $Case.ExpectedChapters) "response chapter_count mismatch"
    if (-not $AllowMockFallback) {
        Assert-True ($response.mode -eq "api") "response mode is '$($response.mode)', expected api"
    }
    Assert-True ($response.mode -in @("api", "mock")) "unexpected response mode: $($response.mode)"
    Test-ScreenplayYaml -Yaml $response.screenplay_yaml -ExpectedChapters $Case.ExpectedChapters

    return [pscustomobject]@{
        Case = $Case.Name
        ExpectedChapters = $Case.ExpectedChapters
        ResponseChapters = $response.chapter_count
        Mode = $response.mode
        DurationMs = $watch.ElapsedMilliseconds
        YamlLength = $response.screenplay_yaml.Length
        Status = "PASS"
    }
}

try {
    Invoke-RestMethod -Uri "$BackendUrl/health" -TimeoutSec 10 | Out-Null
} catch {
    Write-Error "Backend health check failed: $($_.Exception.Message)"
    exit 1
}

$results = @()
$failed = $false

foreach ($case in $cases) {
    try {
        $results += Invoke-Case -Case $case
    } catch {
        $failed = $true
        $results += [pscustomobject]@{
            Case = $case.Name
            ExpectedChapters = $case.ExpectedChapters
            ResponseChapters = "-"
            Mode = "-"
            DurationMs = "-"
            YamlLength = "-"
            Status = "FAIL: $($_.Exception.Message)"
        }
    }
}

$results | Format-Table -AutoSize

if ($failed) {
    exit 1
}
