# 🏴 flag-go

A declarative, struct-based argument parser for Go.

---

## ✨ Features

- Define CLI flags using Go structs
- Support for:
  - `short` and `long` flag names
  - `mandatory` flags
  - `positional` arguments
  - `description` (for help text)
  - `conflicts-with` (mutually exclusive flags)
- Parses booleans, strings, ints, slices, etc.
- Automatically generates usage messages

---

## 📦 Installation

```bash
go get github.com/tobiashort/flag-go
```

Import it in your project:

```go
import "github.com/tobiashort/flag-go"
```

## 🚀 Quick Start

```go
package main

import (
	"fmt"
	"github.com/tobiashort/flag-go"
)

type Args struct {
	Name  string `flag:"mandatory,long=name,description='Full name'"`
	Debug bool   `flag:"short=d,long=debug,description='Enable debug mode'"`
}

func main() {
	var args Args
	flag.Parse(&args)

	fmt.Println("Name:", args.Name)
	fmt.Println("Debug mode:", args.Debug)
}
```

```shell
$ go run main.go --name "Alice" -d
```

## 🧠 Supported Tag Options

The flag struct tag supports the following options:

|Option          |Type   |Description                                   |
|----------------|-------|----------------------------------------------|
|mandatory       |keyword|Flag is required; parser will error if missing|
|short=x         |string |Single-letter short flag (e.g. -x)            |
|long=name       |string |Full-length flag (e.g. --name)                |
|description=... |string |Help/usage description                        |
|conflicts-with=x|string |Mutually exclusive with another field         |
|positional      |keyword|Argument must be passed in a specific position|

