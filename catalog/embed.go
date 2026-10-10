// Package catalog embute os dados dos bancos suportados (databases.yaml e as migrations de cada um).
package catalog

import "embed"

//go:embed databases.yaml migrations
var FS embed.FS
