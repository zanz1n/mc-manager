package webstatic

import "embed"

//go:embed build/*
var Build embed.FS
