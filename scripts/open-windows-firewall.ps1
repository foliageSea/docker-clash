param(
    [ValidateRange(1, 65535)]
    [int]$Port = 27890
)

$ErrorActionPreference = "Stop"
$principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Run this script from an elevated PowerShell window (Run as administrator)."
}

$group = "Docker Clash"
$rules = @(
    @{ Name = "Docker Clash Mixed TCP $Port"; Protocol = "TCP" },
    @{ Name = "Docker Clash Mixed UDP $Port"; Protocol = "UDP" }
)

foreach ($rule in $rules) {
    Get-NetFirewallRule -DisplayName $rule.Name -ErrorAction SilentlyContinue | Remove-NetFirewallRule
    New-NetFirewallRule `
        -DisplayName $rule.Name `
        -Group $group `
        -Direction Inbound `
        -Action Allow `
        -Enabled True `
        -Profile Any `
        -Protocol $rule.Protocol `
        -LocalPort $Port `
        -RemoteAddress LocalSubnet | Out-Null
}

Write-Output "Allowed local-subnet access to Docker Clash mixed proxy port $Port (TCP/UDP)."
