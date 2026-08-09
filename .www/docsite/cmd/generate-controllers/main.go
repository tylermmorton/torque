package main

import (
	"log"

	"github.com/tylermmorton/torque/pkg/pageobjects"

	"github.com/tylermmorton/torque/.www/docsite/components"
	"github.com/tylermmorton/torque/.www/docsite/layouts"
	"github.com/tylermmorton/torque/.www/docsite/routes"
)

func main() {
	gen := pageobjects.NewGenerator(pageobjects.Config{
		OutputDir: ".www/docsite/testutils/pageobjects",
		Package:   "pageobjects",
	})
	pageobjects.Generate[*routes.Document](gen)
	pageobjects.Generate[*layouts.DocsLayout](gen)
	pageobjects.Generate[*components.SearchDialog](gen)
	pageobjects.Generate[*components.Stargazers](gen)
	if err := gen.Run(); err != nil {
		log.Fatal(err)
	}
}
