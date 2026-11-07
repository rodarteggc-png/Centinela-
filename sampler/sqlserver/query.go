package sqlserver

import _ "embed"

//go:embed active_sessions.sql
var ActiveSessionsQuery string