# table <br/> [![GoDoc](https://godoc.org/github.com/rodaine/table?status.svg)](https://godoc.org/github.com/rodaine/table)

![Example Table Output With ANSI Colors](http://res.cloudinary.com/rodaine/image/upload/v1442524799/go-table-example0.png)

Package table provides a convenient way to generate tabular output of any data, primarily useful for CLI tools.

## Features

- Accepts all data types (`string`, `int`, `interface{}`, everything!) and will use the `String() string` method of a type if available.
- Can specify custom formatting for the header and first column cells for better readability.
- Columns are left-aligned and sized to fit the data, with customizable padding.
- The printed output can be sent to any `io.Writer`, defaulting to `os.Stdout`.
- Built to an interface, so you can roll your own `Table` implementation.
- Works well with ANSI colors ([fatih/color](https://github.com/fatih/color) in the example)!
- Can provide a custom `WidthFunc` to accomodate multi- and zero-width characters (such as [runewidth](https://github.com/mattn/go-runewidth))

## Usage

**Download the package:**

```sh
go get github.com/rodaine/table
```

**Example:**

```go
package main

import (
  "fmt"
  "strings"

  "github.com/fatih/color"
  "github.com/rodaine/table"
)

func main() {
  headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
  columnFmt := color.New(color.FgYellow).SprintfFunc()

  tbl := table.New("ID", "Name", "Score", "Added")
  tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

  for _, widget := range getWidgets() {
    tbl.AddRow(widget.ID, widget.Name, widget.Cost, widget.Added)
  }

  tbl.Print()
}
```

_Consult the [documentation](https://godoc.org/github.com/rodaine/table) for further examples and usage information_

## License

table is released under the MIT License (Expat). See the [full license](https://github.com/rodaine/table/blob/master/license).
