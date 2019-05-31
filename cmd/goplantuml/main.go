package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

type NewParser struct {
	goplantuml.ClassParser
}

func main() {
	recursive := flag.Bool("recursive", false, "walk all directories recursively")
	flag.Parse()
	dir, err := getDirectory()

	if err != nil {
		fmt.Println("usage:\ngoplantum <DIR>\nDIR Must be a valid directory")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		fmt.Println("usage:\ngoplantum <DIR>\nDIR Must be a valid directory")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	result, err := goplantuml.NewClassDiagram(absPath, *recursive)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Print(result.Render())
}

func getDirectory() (string, error) {

	args := flag.Args()
	if len(args) < 1 {
		return "", errors.New("DIR missing")
	}
	dir := args[0]
	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("could not find directory %s", dir)
	}
	if !fi.Mode().IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return dir, nil
}
