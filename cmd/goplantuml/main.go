package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

func main() {
	recursive := flag.Bool("recursive", false, "walk all directories recursively")
	ignore := flag.String("ignore", "", "comma separated list of folders to ignore")
	showAggregations := flag.Bool("show-aggregations", false, "renders public aggregations even when -hide-connections is used (do not render by default)")
	hideFields := flag.Bool("hide-fields", false, "hides fields")
	hideMethods := flag.Bool("hide-methods", false, "hides methods")
	hideConnections := flag.Bool("hide-connections", false, "hides all connections in the diagram")
	showCompositions := flag.Bool("show-compositions", false, "Shows compositions even when -hide-connections is used")
	showImplementations := flag.Bool("show-implementations", false, "Shows implementations even when -hide-connections is used")
	showAliases := flag.Bool("show-aliases", false, "Shows aliases even when -hide-connections is used")
	showConnectionLabels := flag.Bool("show-connection-labels", false, "Shows labels in the connections to identify the connections types (e.g. extends, implements, aggregates, alias of")

	flag.Parse()
	dirs, err := getDirectories()

	if err != nil {
		fmt.Println("usage:\ngoplantum <DIR>\nDIR Must be a valid directory")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	ignoredDirectories, err := getIgnoredDirectories(*ignore)
	if err != nil {

		fmt.Println("usage:\ngoplantum [-ignore=<DIRLIST>]\nDIRLIST Must be a valid comma separated list of existing directories")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	renderingOptions := map[goplantuml.RenderingOption]bool{
		goplantuml.RenderConnectionLabels: *showConnectionLabels,
		goplantuml.RenderFields:           !*hideFields,
		goplantuml.RenderMethods:          !*hideMethods,
		goplantuml.RenderAggregations:     *showAggregations,
	}
	if *hideConnections {
		renderingOptions[goplantuml.RenderAliases] = *showAliases
		renderingOptions[goplantuml.RenderCompositions] = *showCompositions
		renderingOptions[goplantuml.RenderImplementations] = *showImplementations

	}

	result, err := goplantuml.NewClassDiagram(dirs, ignoredDirectories, *recursive)
	result.SetRenderingOptions(renderingOptions)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Print(result.Render())
}

func getDirectories() ([]string, error) {

	args := flag.Args()
	if len(args) < 1 {
		return nil, errors.New("DIR missing")
	}
	dirs := []string{}
	for _, dir := range args {
		fi, err := os.Stat(dir)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		if !fi.Mode().IsDir() {
			return nil, fmt.Errorf("%s is not a directory", dir)
		}
		dirAbs, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		dirs = append(dirs, dirAbs)
	}
	return dirs, nil
}

func getIgnoredDirectories(list string) ([]string, error) {
	result := []string{}
	list = strings.TrimSpace(list)
	if list == "" {
		return result, nil
	}
	split := strings.Split(list, ",")
	for _, dir := range split {
		dirAbs, err := filepath.Abs(strings.TrimSpace(dir))
		if err != nil {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		result = append(result, dirAbs)
	}
	return result, nil
}
