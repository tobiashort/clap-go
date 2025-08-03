# 🏴 clap-go

A command line argument parser in go. Inspired by [clap-rs/clap](https://github.com/clap-rs/clap).

---

## ✨ Features

- Define CLI arguments using Go structs
- Support for:
  - `short` and `long` argument names
  - `mandatory` arguments
  - `positional` arguments
  - `description` (for help text)
  - `conflicts-with` (mutually exclusive arguments)
- Parses booleans, strings, ints, slices, etc.
- Automatically generates usage messages

---

## 📦 Installation

```bash
go get github.com/tobiashort/clap-go
```

Import it in your project:

```go
import "github.com/tobiashort/clap-go"
```

## 🚀 Quick Start

```go
TODO
```

```shell
$ go run ./example --name "John Doe" --email john@company.com -t "Designer" -F --salary 85000 -N "#design" -N "#it" D12345 Design
=== New Employee Onboarding ===
Name:        John Doe
Email:       john@company.com
Position:    Designer
Type:        Full-Time
Salary:      $85000
Department:  Design
Employee ID: D12345
Notify:      [#design #it]
```

```shell
TODO
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
|default-value   |string |Default value                                 |

