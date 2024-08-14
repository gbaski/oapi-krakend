package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gbaski/oapi-krakend/converter"
)

func main() {

	dir, _ := os.Getwd()

	spec := flag.String("spec", "", "The URL or filepath of the OpenAPI specification.")
	includes := flag.String("includes", "", "A comma-separated list of URI resources to include from the OpenAPI specification paths.")
	excludes := flag.String("excludes", "", "A comma-separated list of URI resources to exclude from the OpenAPI specification paths.")
	outputDir := flag.String("outputDir", dir, "The output directory for the generated files.")

	flag.Parse()

	if *spec == "" {

		fmt.Print("\nError: --spec argument is missing\n\nSee the usage below:\n\n")
		flag.PrintDefaults()
		return
	}

	c := converter.NewConverter(*spec, converter.Options{
		Includes:  *includes,
		Excludes:  *excludes,
		OutputDir: *outputDir,
	})

	err := c.Convert()

	if err != nil {
		fmt.Printf("Err: %s", err.Error())
	}

}
