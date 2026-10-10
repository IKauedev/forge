// Package templates embute os projetos de referência usados pelo forge.
//
// spring-api e quarkus-api: "all:" inclui arquivos ocultos (.mvn, .gitignore), mas também qualquer target/ que
// exista localmente. Rode `./mvnw clean` (ou apague target/) no template antes de compilar o forge.
//
// angular-app e react-app: os padrões são explícitos para não embutir node_modules/, dist/ e .angular/,
// que existem depois de `npm ci` e do build. Arquivo ou diretório novo na raiz de um desses templates
// precisa entrar na lista abaixo (o teste TestFrontendTemplatesAreEmbedded acusa a falta).
package templates

import "embed"

//go:embed all:spring-api all:quarkus-api
//go:embed angular-app/*.json angular-app/*.js angular-app/*.md angular-app/*.yaml
//go:embed angular-app/Dockerfile
//go:embed angular-app/.dockerignore angular-app/.editorconfig angular-app/.gitignore
//go:embed angular-app/.prettierignore angular-app/.prettierrc
//go:embed all:angular-app/.github all:angular-app/.husky
//go:embed all:angular-app/docker all:angular-app/docs all:angular-app/infra
//go:embed all:angular-app/public all:angular-app/src
//go:embed react-app/*.json react-app/*.js react-app/*.ts react-app/*.html react-app/*.md react-app/*.yaml
//go:embed react-app/Dockerfile
//go:embed react-app/.dockerignore react-app/.editorconfig react-app/.gitignore
//go:embed react-app/.prettierignore react-app/.prettierrc
//go:embed all:react-app/.github all:react-app/.husky
//go:embed all:react-app/docker all:react-app/docs all:react-app/infra
//go:embed all:react-app/public all:react-app/src
var FS embed.FS

// Nomes dos diretórios dos templates dentro de FS.
const (
	SpringAPI  = "spring-api"
	QuarkusAPI = "quarkus-api"
	AngularApp = "angular-app"
	ReactApp   = "react-app"
)

// Names lista os templates disponíveis.
var Names = []string{SpringAPI, QuarkusAPI, AngularApp, ReactApp}
