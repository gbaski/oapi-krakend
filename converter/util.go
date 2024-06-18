package converter

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"text/template"
)

func Map[T any](input []T, fn func(T) T) []T {
	result := make([]T, len(input))
	for i, v := range input {
		result[i] = fn(v)
	}
	return result
}

func uniqueItems(slice []string) []string {

	uniqueMap := make(map[string]bool)

	uniqueSlice := []string{}

	for _, item := range slice {

		if !uniqueMap[item] {
			uniqueMap[item] = true
			uniqueSlice = append(uniqueSlice, item)
		}
	}
	return uniqueSlice
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isFile(str string) bool {
	info, err := os.Stat(str)
	return err == nil && !info.IsDir()
}

func isDirectory(str string) bool {
	info, err := os.Stat(str)
	return err == nil && info.IsDir()
}

func loadSpec(spec string) (OpenAPI, error) {

	data := OpenAPI{}
	var err error

	if isURL(spec) {
		data, err = loadSpecFromUrl(spec)
	}

	if isFile(spec) {
		data, err = loadSpecFromFile(spec)
	}

	return data, err
}

func loadSpecFromUrl(url string) (OpenAPI, error) {

	data := OpenAPI{}

	resp, err := http.Get(url)

	if err != nil {
		return data, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return data, err
	}

	_ = json.Unmarshal(body, &data)

	return data, nil
}

func loadSpecFromFile(filename string) (OpenAPI, error) {

	data := OpenAPI{}

	file, err := os.Open(filename)

	if err != nil {
		return data, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)

	if err != nil {
		return data, err
	}

	return data, nil
}

func renderTemplate(str string, data interface{}) {

	funcMap := template.FuncMap{
		"add": func(x, y int) int {
			return x + y
		},
	}

	tmpl, err := template.New("template").Funcs(funcMap).Parse(str)
	if err != nil {

		log.Printf("Error parsing template: %v", err)
		return
	}

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {

		log.Printf("Error executing template: %v", err)
		return
	}

}

func copyFile(src, dst string) error {

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = destination.Sync()
	if err != nil {
		return err
	}

	return nil
}
