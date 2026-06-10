package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunInRootDirUsesProjectRootAndRestoresWorkingDirectory(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "config.yaml"), []byte("title: Root Site\n"), 0644))
	realRoot, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)

	var cwdDuring string
	var title string
	err = runInRootDir(root, func() error {
		var err error
		cwdDuring, err = os.Getwd()
		if err != nil {
			return err
		}

		conf := core.DefaultConfig()
		if err := conf.LoadFromFile(""); err != nil {
			return err
		}
		title = conf.GetString("title")
		return nil
	})
	require.NoError(t, err)

	cwdAfter, err := os.Getwd()
	require.NoError(t, err)
	assert.Equal(t, realRoot, cwdDuring)
	assert.Equal(t, "Root Site", title)
	assert.Equal(t, wd, cwdAfter)
}
