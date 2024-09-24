package assets

import "embed"

//go:embed styles.css
var Css string

//go:embed logo.png
var Logo []byte

//go:embed public
var Assets embed.FS
