// Package templates embute os projetos de referência usados pelo forge.
//
// Atenção: "all:" inclui arquivos ocultos (.mvn, .gitignore), mas também qualquer target/ que
// exista localmente. Rode `./mvnw clean` (ou apague target/) no template antes de compilar o forge.
package templates

import "embed"

//go:embed all:spring-api
var FS embed.FS

// SpringAPI é o nome do diretório do template Spring dentro de FS.
const SpringAPI = "spring-api"
