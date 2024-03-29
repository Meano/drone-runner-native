// Copyright (c) 2021 Meano

package bash

import (
	"bytes"
	"fmt"
	"strings"
)

// Script converts a slice of individual shell commands to a bash script.
func Script(commands []string) string {
	buf := new(bytes.Buffer)
	fmt.Fprint(buf, optionScript)
	fmt.Fprint(buf, tmateScript)
	for _, command := range commands {
		escaped := fmt.Sprintf("%q", command)
		escaped = strings.Replace(escaped, "$", `\$`, -1)
		buf.WriteString(fmt.Sprintf(
			traceScript,
			escaped,
			command,
		))
	}
	return buf.String()
}

// optionScript is a helper script this is added to the build
// to set shell options, in this case, to exit on error.
const optionScript = `
if [ ! -z "${CI_NETRC_FILE}" ]; then
	echo $CI_NETRC_FILE > $HOME/.netrc
	chmod 600 $HOME/.netrc
fi

unset CI_SCRIPT CI_NETRC_MACHINE CI_NETRC_USERNAME CI_NETRC_PASSWORD CI_NETRC_FILE

set -e
`

// traceScript is a helper script that is added to
// the build script to trace a command.
const traceScript = `
echo + %s
%s
`

const tmateScript = `
remote_debug() {
	if [ "$?" -ne "0" ]; then
		/usr/drone/bin/tmate -F
	fi
}

if [ "${DRONE_BUILD_DEBUG}" = "true" ]; then
	if [ ! -z "${DRONE_TMATE_HOST}" ]; then
		echo "set -g tmate-server-host $DRONE_TMATE_HOST" >> $HOME/.tmate.conf
		echo "set -g tmate-server-port $DRONE_TMATE_PORT" >> $HOME/.tmate.conf
		echo "set -g tmate-server-rsa-fingerprint $DRONE_TMATE_FINGERPRINT_RSA" >> $HOME/.tmate.conf
		echo "set -g tmate-server-ed25519-fingerprint $DRONE_TMATE_FINGERPRINT_ED25519" >> $HOME/.tmate.conf

		if [ ! -z "${DRONE_TMATE_AUTHORIZED_KEYS}" ]; then
			echo "$DRONE_TMATE_AUTHORIZED_KEYS" > $HOME/.tmate.authorized_keys
			echo "set -g tmate-authorized-keys \"$HOME/.tmate.authorized_keys\"" >> $HOME/.tmate.conf
		fi
	fi
	trap remote_debug EXIT
fi
`
