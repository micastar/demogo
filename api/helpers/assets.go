package helpers

import (
	"time"

	"github.com/jessevdk/go-assets"
)

var (
	_Assets3737a75b5254ed1f6d588b40a3449721f9ea86c2 = `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<form method="post" action="/api">
				<input type="url" name="url" placeholder="Enter a URL" required>
				<input type="submit" value="Shorten">
			</form>
		</body>
		</html>
	`
)
var Assets = assets.NewFileSystem(map[string][]string{"/": {"tmpl"}, "/tmpl": {"index.tmpl"}}, map[string]*assets.File{
	"/": {
		Path:     "/",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1524365738, 1524365738517125470),
		Data:     nil,
	},
	"/tmpl": {
		Path:     "/tmpl",
		FileMode: 0x800001ed,
		Mtime:    time.Unix(1524365491, 1524365491289799093),
		Data:     nil,
	},

	"/tmpl/index.tmpl": {
		Path:     "/tmpl/index.tmpl",
		FileMode: 0x1a4,
		Mtime:    time.Unix(1524365491, 1524365491289995821),
		Data:     []byte(_Assets3737a75b5254ed1f6d588b40a3449721f9ea86c2),
	},
}, "")
