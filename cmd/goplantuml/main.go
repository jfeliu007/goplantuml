package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

func main() {
	recursive := flag.Bool("recursive", false, "walk all directories recursively")
	flag.Parse()
	dirs, err := getDirectories()

	if err != nil {
		fmt.Println("usage:\ngoplantum <DIR>\nDIR Must be a valid directory")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	result, err := goplantuml.NewClassDiagram(dirs, *recursive)
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
		dir, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("could not find directory %s", dir)
		}
		dirs = append(dirs, dir)
	}
	return dirs, nil
}
