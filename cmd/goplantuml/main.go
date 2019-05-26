package main

import (
	"errors"
	"fmt"
	"os"

	goplantuml "github.com/jfeliu007/goplantuml/parser"
)

func main() {
	dir, err := getDirectory()

	if err != nil {
		fmt.Println("ussage:\ngoplantum <DIR>\nDIR Must be a valid directory")
		fmt.Println(err.Error())
		os.Exit(1)
	}
	result, _ := goplantuml.NewClassDiagram(dir)
	fmt.Print(result.Render())
}

func getDirectory() (string, error) {

	if len(os.Args) < 2 {
		return "", errors.New("DIR missing")
	}
	dir := os.Args[1]
	fi, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("could not find directory %s", dir)
	}
	if !fi.Mode().IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return os.Args[1], nil
}
