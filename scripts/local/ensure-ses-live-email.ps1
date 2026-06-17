param(
    [string]$Email = "varathu09@gmail.com",
    [string]$Region = "ap-south-1",
    [string]$OutputPath = "data/local/cloud/ses-live-email-status.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "ensure-ses-live-email" -LogRoot "logs/local/cloud" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)
    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Get-TraceDeckObjectProperty {
    param(
        [object]$InputObject,
        [string]$Name,
        [object]$DefaultValue = $null
    )

    if ($null -eq $InputObject) {
        return $DefaultValue
    }
    $property = $InputObject.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $DefaultValue
    }
    return $property.Value
}

try {
    $resolvedOutputPath = Resolve-TraceDeckPath -PathValue $OutputPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedOutputPath) | Out-Null

    $account = aws sesv2 get-account --region $Region --output json | ConvertFrom-Json
    $identities = aws sesv2 list-email-identities --region $Region --output json | ConvertFrom-Json
    $identity = $identities.EmailIdentities | Where-Object { $_.IdentityName -eq $Email } | Select-Object -First 1

    if (-not $identity) {
        Invoke-TraceDeckLoggedCommand -Label "Request SES email identity verification" -Command {
            aws sesv2 create-email-identity --email-identity $Email --region $Region --output json
        }
        $identities = aws sesv2 list-email-identities --region $Region --output json | ConvertFrom-Json
        $identity = $identities.EmailIdentities | Where-Object { $_.IdentityName -eq $Email } | Select-Object -First 1
    }

    $verifiedForSending = [bool](Get-TraceDeckObjectProperty -InputObject $identity -Name "VerifiedForSendingStatus" -DefaultValue $false)
    $identityType = [string](Get-TraceDeckObjectProperty -InputObject $identity -Name "IdentityType" -DefaultValue "")
    $status = [pscustomobject]@{
        email = $Email
        region = $Region
        production_access_enabled = [bool]$account.ProductionAccessEnabled
        sending_enabled = [bool]$account.SendingEnabled
        max_24_hour_send = $account.SendQuota.Max24HourSend
        max_send_rate = $account.SendQuota.MaxSendRate
        sent_last_24_hours = $account.SendQuota.SentLast24Hours
        identity_present = [bool]$identity
        identity_type = $identityType
        verified_for_sending = $verifiedForSending
        can_send_to_requested_email = $verifiedForSending
        live_email_ready = ([bool]$account.SendingEnabled -and [bool]$identity -and $verifiedForSending)
        next_action = if ($identity -and $verifiedForSending) {
            "SES identity is verified. Live SES alert email can be enabled."
        } else {
            "Open the SES verification email sent to this address and click the verification link, then rerun this script."
        }
        checked_at = (Get-Date).ToUniversalTime().ToString("o")
    }

    $status | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $resolvedOutputPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "SES live email status saved: $resolvedOutputPath"
    $status | ConvertTo-Json -Depth 5
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
