package migrations

import "embed"

// FS contains the SQL migrations shipped with the application.
//go:embed *.sql
var FS embed.FS
