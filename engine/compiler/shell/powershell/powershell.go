// Copyright (c) 2021 Meano

package powershell

import (
	"bytes"
	"fmt"
	"strings"
)

// Script converts a slice of individual shell commands to a powershell script.
func Script(commands []string) string {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, startScript)
	for _, command := range commands {
		escaped := "+ " + command
		escaped = `"` + strings.Replace(strings.Replace(strings.Replace(escaped, "`", "``", -1), "\"", "`\"", -1), "$", "`$", -1) + `"`
		buf.WriteString(fmt.Sprintf(
			traceScript,
			escaped,
			command,
		))
	}

	fmt.Fprint(buf, endScript)
	return buf.String()
}

// Start Script:
// 1. Set the default Encoding
// 2. Setup the NETRC_FILE
// 3. Read CI_SHARE_* env vars from env file
// 4. Remove share env file
const startScript = `$OutputEncoding = [console]::InputEncoding = [console]::OutputEncoding = New-Object Text.UTF8Encoding
if ($Env:CI_NETRC_MACHINE) {
    echo $Env:CI_NETRC_FILE > (Join-Path $Env:USERPROFILE '_netrc');
    $Env:CI_NETRC_USERNAME = $Env:CI_NETRC_PASSWORD = $Env:CI_NETRC_USERNAME = $Env:CI_NETRC_FILE = $null;
}
$Env:CI_SCRIPT = $null;
$erroractionpreference = "stop";

$shareEnvPath = "$env:USERPROFILE/.ci_share_env"
if (Test-Path $shareEnvPath) {
    foreach ($shareEnvLine in Get-content $shareEnvPath) {
        $shareEnv = $shareEnvLine.Split("=", 2)
        if ($shareEnv.Count -eq 2 -and $shareEnv[0].StartsWith("CI_SHARE_")) {
            set-item "env:$($shareEnv[0])" "$($shareEnv[1])"
        }
    }
    Remove-Item $shareEnvPath
}
`

// End Script:
// 1. Save all CI_SHARE_* env vars to env file
const endScript = `$shareEnvs = (Get-ChildItem "env:CI_SHARE_*")
foreach ($shareEnv in $shareEnvs) {
    Write-Output "$($shareEnv.Name)=$($shareEnv.Value)" >> "$shareEnvPath"
}
`

// Trace Script:
// 1. Trace the $LastExitCode
const traceScript = `
echo %s
%s
if ($LastExitCode -gt 0) { exit $LastExitCode }
`
