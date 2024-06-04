module github.com/tylermmorton/torque/.www/docsite

go 1.21

toolchain go1.21.6

replace github.com/tylermmorton/torque => ../../../torque

require (
	github.com/adrg/frontmatter v0.2.0
	github.com/alecthomas/chroma v0.10.0
	github.com/algolia/algoliasearch-client-go/v3 v3.31.0
	github.com/gomarkdown/markdown v0.0.0-20231222211730-1d6d20845b47
	github.com/joho/godotenv v1.5.1
	github.com/stretchr/testify v1.8.4
	github.com/tylermmorton/tmpl v0.0.0-20231025031313-5552ee818c6d
	github.com/tylermmorton/torque v1.1.0
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/go-chi/chi/v5 v5.0.11 // indirect
	github.com/gorilla/schema v1.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
