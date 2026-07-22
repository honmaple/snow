package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSSMinSupportsLayerAtRule(t *testing.T) {
	filter := &MinifyFilter{name: assetFilterCSSMin}

	result, err := filter.Execute([]byte(`@layer theme, base, components, utilities;
@layer theme {
  :root, :host {
    --font-sans: ui-sans-serif, system-ui, sans-serif;
  }
}
@layer components {
  .btn {
    color: red;
  }
}
@layer utilities {
  .translate-x-1\/2 {
    --tw-translate-x: calc(1 / 2 * 100%);
    translate: var(--tw-translate-x) var(--tw-translate-y);
  }
}
@property --tw-translate-x {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
`))

	require.NoError(t, err)
	assert.Contains(t, string(result), "@layer")
	assert.Contains(t, string(result), ".btn")
}
