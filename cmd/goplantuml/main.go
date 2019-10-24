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
	aggregation := flag.Bool("aggregation", false, "renders public aggregations")
	hideFields := flag.Bool("hide-fields", false, "hides fields")
	hideMethods := flag.Bool("hide-methods", false, "hides methods")
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

	result, err := goplantuml.NewClassDiagram(dirs, ignoredDirectories, *recursive)
	result.SetRenderingOptions(&goplantuml.RenderingOptions{
		Aggregation: *aggregation,
		Fields:      !*hideFields,
		Methods:     !*hideMethods,
	})
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
