package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestServerCommandHasRootDirFlag(t *testing.T) {
	var found bool
	for _, flag := range serverCommand.Flags {
		stringFlag, ok := flag.(*cli.StringFlag)
		if !ok {
			continue
		}
		if stringFlag.Name == "root-dir" {
			found = true
			assert.Contains(t, stringFlag.Aliases, "r")
			assert.Equal(t, ".", stringFlag.Value)
		}
	}

	assert.True(t, found, "server command should expose root-dir flag")
}
