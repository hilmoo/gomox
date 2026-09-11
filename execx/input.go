package execx

import (
	"strings"
)

// GetSpamYOpts returns [ExecOptions] whose stdin answers "y" to a command's
// interactive confirmation prompt.
func GetSpamYOpts() ExecOptions {
	return ExecOptions{
		Stdin: strings.NewReader("y\n"),
	}
}
