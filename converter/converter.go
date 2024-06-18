package converter

import (
	_ "embed"

	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/schollz/progressbar/v3"
)

//go:embed schema/krakend.tmpl
var krakendSchema string

//go:embed schema/endpoint.tmpl
var endpointSchema string

var title string
var tmpDir string

type Converter struct {
	count     int
	Spec      string
	Includes  string
	Excludes  string
	statItems StatItems
	templates []string
	OutputDir string
}

func (c *Converter) generate(path string, method string, parameters map[string][]string) error {

	if slices.Contains([]string{"/", "/health"}, path) {
		return nil
	}

	parts := strings.Split(path, "/")
	tag := parts[1]

	if c.statItems == nil {
		c.statItems = make(StatItems)

	}

	c.statItems[tag]++

	template := fmt.Sprintf("endpoints-%s.tmpl", tag)

	c.templates = append(c.templates, template)

	templatesDir := fmt.Sprintf("%s/templates", c.OutputDir)

	if !isDirectory(templatesDir) {
		_ = os.MkdirAll(templatesDir, os.ModePerm)
	}

	filename := fmt.Sprintf("%s/%s", templatesDir, template)

	endpoint := path
	url_pattern := path

	if slices.Contains([]string{"post", "put"}, method) {

		parameters["header"] = append(parameters["header"], "content-type")
		parameters["header"] = append(parameters["header"], "content-length")

	}

	var input_query_strings string

	input_headers := fmt.Sprintf("\"input_headers\": [%s],", strings.Join(Map(parameters["header"], func(s string) string {
		return fmt.Sprintf("\"%s\"", strings.Title(s))
	}), ", "))

	if len(parameters["query"]) > 0 {
		input_query_strings = fmt.Sprintf("\n  \"input_query_strings\": [%s],", strings.Join(Map(parameters["query"], func(s string) string {
			return fmt.Sprintf("\"%s\"", s)
		}), ", "))
	}

	method = strings.ToUpper(method)

	templateContent := fmt.Sprintf(endpointSchema, endpoint, method, input_headers, input_query_strings, url_pattern, method)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	if isFile(filename) {
		file.WriteString(",\n")
	}

	file.WriteString(templateContent)

	return nil

}

func (c *Converter) Convert() error {

	bar := progressbar.New(4)
	bar.Add(1)
	bar.Describe(fmt.Sprintf("Retrieving spec: %s", c.Spec))

	data, err := loadSpec(c.Spec)

	if err != nil {
		return err

	}

	if data.Info.Title == "" {
		return fmt.Errorf("no data found")

	}

	bar.Add(1)
	bar.Describe("Converting OpenAPI to KrakenD Configs")

	err = <-func() <-chan error {

		ch := make(chan error)

		go c.reset(ch)

		return ch

	}()

	if err != nil {

		return err

	}

	title = data.Info.Title

	for path, pathItem := range data.Paths {

		for method, resource := range pathItem {

			parameters := getParameters(resource.Parameters)

			tag := strings.Split(path, "/")[1]

			if c.Excludes != "" {

				if contains(strings.Split(c.Excludes, ","), tag) {
					continue
				}

			}

			if c.Includes != "" {

				if contains(strings.Split(c.Includes, ","), tag) {

					err := c.generate(path, method, parameters)
					if err != nil {
						return err
					}

					c.count++
				}

				continue

			}

			err := c.generate(path, method, parameters)
			if err != nil {
				return err
			}

			c.count++

		}

	}

	err = c.update(bar)

	if err != nil {
		return err
	}

	return nil
}

func (c *Converter) update(bar *progressbar.ProgressBar) error {

	defer c.stats()
	defer func() {

		bar.Add(1)
		bar.Describe("Updated krakend.yml")
		bar.Finish()

	}()

	var endpoint string
	endpoints := []string{}

	sort.Strings(c.templates)

	for i, tmpl := range uniqueItems(c.templates) {

		if i == 0 {
			endpoint = fmt.Sprintf("{{ template \"%s\" . }}", tmpl)
		} else {
			endpoint = fmt.Sprintf("\t\t{{ template \"%s\" . }}", tmpl)
		}

		endpoints = append(endpoints, endpoint)

	}

	krakend_tmpl := fmt.Sprintf("%s/krakend.tmpl", c.OutputDir)

	if isFile(krakend_tmpl) {

		fileContent, err := os.ReadFile(krakend_tmpl)

		if err != nil {
			return err
		}

		pattern := `"endpoints":\s*\[([^]]+)]`
		replacement := `"endpoints": [
		%s
	]`
		re := regexp.MustCompile(pattern)

		if re.Match(fileContent) {
			krakendSchema = re.ReplaceAllString(string(fileContent), replacement)
		}

	}

	krakendSchema = fmt.Sprintf(krakendSchema, strings.Join(endpoints, ",\n"))

	os.WriteFile(krakend_tmpl, []byte(krakendSchema), 0644)

	return nil

}

func getParameters(parameters []Parameter) ParameterItems {

	items := ParameterItems{}

	for _, parameter := range parameters {

		items[parameter.In] = append(items[parameter.In], parameter.Name)

	}

	return items
}

func (c *Converter) reset(ch chan<- error) {

	defer close(ch)

	var err error

	tmpDir, err = os.MkdirTemp("", "oapi-krakend")

	if err != nil {

		ch <- fmt.Errorf("error creating temporary directory: %s", err.Error())
		return

	}

	// fmt.Println("Temporary directory:", tmpDir)

	files, err := filepath.Glob(c.OutputDir + "/templates/*")

	if err != nil {

		ch <- fmt.Errorf("error reading directory: %s", err.Error())
		return

	}

	for _, file := range files {

		if strings.Contains(file, "endpoints-") {

			err = copyFile(file, fmt.Sprintf("%s/%s", tmpDir, filepath.Base(file)))

			if err != nil {

				ch <- fmt.Errorf("error copying file: %s", err.Error())
				return
			}

			err = os.Remove(file)

			if err != nil {

				ch <- fmt.Errorf("error deleting file: %s", err.Error())
				return
			}

		}

	}

	ch <- nil

}

func (c *Converter) stats() {

	items := []map[string]interface{}{}

	for tag, qty := range c.statItems {

		items = append(items, map[string]interface{}{"tag": tag, "qty": qty})

	}

	sort.Slice(items, func(i, j int) bool {
		return items[i]["tag"].(string) < items[j]["tag"].(string)
	})

	data := struct {
		StatItems []map[string]interface{}
		Count     int
		Title     string
	}{
		StatItems: items,
		Count:     c.count,
		Title:     title,
	}

	str := `
-----------------------	
{{ .Title }} Endpoints
-----------------------
{{ range $index, $item := .StatItems }}
{{ $item.tag }}, {{ $item.qty }}
{{ end }}
-----------------------
Total Endpoints = {{ .Count }}
-----------------------
`

	renderTemplate(str, data)
}

func (c *Converter) Restore() {

	files, err := filepath.Glob(tmpDir + "/*")

	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	for _, file := range files {

		fmt.Println("Restoring file:", file)

		err = os.Rename(file, c.OutputDir+"/templates/"+filepath.Base(file))

		if err != nil {
			fmt.Println("Error renaming file:", err)

		}

	}

	err = os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Println("Error removing temporary directory:", err)
	}
}
