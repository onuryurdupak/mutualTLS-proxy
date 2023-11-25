package program

import (
	"fmt"

	"github.com/onuryurdupak/gomod/v2/stdout"
)

const (
	stamp_build_date  = "${build_date}"
	stamp_commit_hash = "${commit_hash}"
	stamp_source      = "${source}"

	ErrSuccess  = 0
	ErrInput    = 1
	ErrInternal = 2
	ErrUnknown  = 3

	helpPrompt = `Run 'middleman -h' for help.`

	helpMessage = `
Readme is availabile at: https://github.com/onuryurdupak/mutualTLS-proxy#readme
`
)

func versionInfo() string {
	return fmt.Sprintf(`Build Date: %s | Commit: %s
Source: %s`, stamp_build_date, stamp_commit_hash, stamp_source)
}

func helpMessageStyled() string {
	msg, _ := stdout.ProcessStyle(helpMessage)
	return msg
}

func helpMessageUnstyled() string {
	return stdout.RemoveStyle(helpMessage)
}
