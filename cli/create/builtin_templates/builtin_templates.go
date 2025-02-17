package builtin_templates

import (
	"embed"

	"github.com/tarantool/tt/cli/create/builtin_templates/static"
)

//go:embed templates/*
var TemplatesFs embed.FS

// FileModes contains mapping of file modes by built-in template name.
var FileModes = map[string]map[string]int{
	"cartridge": static.CartridgeFileModes,
}
