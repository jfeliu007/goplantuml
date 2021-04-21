package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goplantuml "github.com/jfeliu007/goplantuml/v2/parser"
)

// RenderingOptionSlice will implements the sort interface
type RenderingOptionSlice []goplantuml.RenderingOption

// Len is the number of elements in the collection.
func (as RenderingOptionSlice) Len() int {
	return len(as)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (as RenderingOptionSlice) Less(i, j int) bool {
	return as[i] < as[j]
}

// Swap swaps the elements with indexes i and j.
func (as RenderingOptionSlice) Swap(i, j int) {
	as[i], as[j] = as[j], as[i]
}

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
	title := flag.String("title", "", "Title of the generated diagram")
	notes := flag.String("notes", "", "Comma separated list of notes to be added to the diagram")
	output := flag.String("output", "", "output file path. If omitted, then this will default to standard output")
	showOptionsAsNote := flag.Bool("show-options-as-note", false, "Show a note in the diagram with the none evident options ran with this CLI")
	aggregatePrivateMembers := flag.Bool("aggregate-private-members", false, "Show aggregations for private members. Ignored if -show-aggregations is not used.")
	hidePrivateMembers := flag.Bool("hide-private-members", false, "Hide private fields and methods")
	flag.Parse()
	renderingOptions := map[goplantuml.RenderingOption]interface{}{
		goplantuml.RenderConnectionLabels:  *showConnectionLabels,
		goplantuml.RenderFields:            !*hideFields,
		goplantuml.RenderMethods:           !*hideMethods,
		goplantuml.RenderAggregations:      *showAggregations,
		goplantuml.RenderTitle:             *title,
		goplantuml.AggregatePrivateMembers: *aggregatePrivateMembers,
		goplantuml.RenderPrivateMembers:    !*hidePrivateMembers,
	}
	if *hideConnections {
		renderingOptions[goplantuml.RenderAliases] = *showAliases
		renderingOptions[goplantuml.RenderCompositions] = *showCompositions
		renderingOptions[goplantuml.RenderImplementations] = *showImplementations

	}
	noteList := []string{}
	if *showOptionsAsNote {
		legend, err := getLegend(renderingOptions)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		noteList = append(noteList, legend)
	}
	if *notes != "" {
		noteList = append(noteList, "", "<b><u>Notes</u></b>")
	}
	split := strings.Split(*notes, ",")
	for _, note := range split {
		trimmed := strings.TrimSpace(note)
		if trimmed != "" {
			noteList = append(noteList, trimmed)
		}
	}
	renderingOptions[goplantuml.RenderNotes] = strings.Join(noteList, "\n")
	dirs, err := getDirectories()

	if err != nil {
		fmt.Println("usage:\ngoplantuml <DIR>\nDIR Must be a valid directory")
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	ignoredDirectories, err := getIgnoredDirectories(*ignore)
	if err != nil {

		fmt.Println("usage:\ngoplantuml [-ignore=<DIRLIST>]\nDIRLIST Must be a valid comma separated list of existing directories")
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	result, err := goplantuml.NewClassDiagram(dirs, ignoredDirectories, *recursive)
	result.SetRenderingOptions(renderingOptions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	rendered := result.Render()
	var writer io.Writer
	if *output != "" {
		writer, err = os.Create(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	} else {
		writer = os.Stdout
	}
	fmt.Fprint(writer, rendered)
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

func getLegend(ro map[goplantuml.RenderingOption]interface{}) (string, error) {
	result := "<u><b>Legend</b></u>\n"
	orderedOptions := RenderingOptionSlice{}
	for o := range ro {
		orderedOptions = append(orderedOptions, o)
	}
	sort.Sort(orderedOptions)
	for _, option := range orderedOptions {
		val := ro[option]
		switch option {
		case goplantuml.RenderAggregations:
			result = fmt.Sprintf("%sRender Aggregations: %t\n", result, val.(bool))
		case goplantuml.RenderAliases:
			result = fmt.Sprintf("%sRender Connections: %t\n", result, val.(bool))
		case goplantuml.RenderCompositions:
			result = fmt.Sprintf("%sRender Compositions: %t\n", result, val.(bool))
		case goplantuml.RenderFields:
			result = fmt.Sprintf("%sRender Fields: %t\n", result, val.(bool))
		case goplantuml.RenderImplementations:
			result = fmt.Sprintf("%sRender Implementations: %t\n", result, val.(bool))
		case goplantuml.RenderMethods:
			result = fmt.Sprintf("%sRender Methods: %t\n", result, val.(bool))
		case goplantuml.AggregatePrivateMembers:
			result = fmt.Sprintf("%sPritave Aggregations: %t\n", result, val.(bool))
		}
	}
	return strings.TrimSpace(result), nil
}
