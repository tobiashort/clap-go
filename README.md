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
	Name           string   `flag:"mandatory,description='Full name of the new employee'"`
	Email          string   `flag:"description='Company email address to assign'"`
	Position       string   `flag:"long=title,short=t,description='Job title (e.g., Backend Engineer)'"`
	FullTime       bool     `flag:"short=F,long=full-time,conflicts-with=PartTime,description='Mark as full-time employee'"`
	PartTime       bool     `flag:"short=P,long=part-time,description='Mark as part-time employee'"`
	Apprenticeship bool     `flag:"short=A,description='Indicates the employee is joining as an apprentice'"`
	Salary         int      `flag:"default-value=9999,description='Starting salary in USD'"`
	TeamsChannel   []string `flag:"long=notify,short=N,description='Slack team channels to notify (e.g., #eng, #ops)'"`
	EmployeeID     string   `flag:"positional,mandatory,description='Unique employee ID'"`
	Department     string   `flag:"positional,default-value=Design,description='Department name (e.g., Engineering, HR)'"`
}

func main() {
	args := Args{}
	flag.Parse(&args)

	empType := "Contractor"
	if args.FullTime {
		empType = "Full-Time"
	} else if args.PartTime {
		empType = "Part-Time"
	}

	fmt.Println("=== New Employee Onboarding ===")
	fmt.Printf("Name:           %s\n", args.Name)
	fmt.Printf("Email:          %s\n", args.Email)
	fmt.Printf("Position:       %s\n", args.Position)
	fmt.Printf("Type:           %s\n", empType)
	fmt.Printf("Apprenticeship: %v\n", args.Apprenticeship)
	fmt.Printf("Salary:         $%d\n", args.Salary)
	fmt.Printf("Department:     %s\n", args.Department)
	fmt.Printf("Employee ID:    %s\n", args.EmployeeID)
	fmt.Printf("Notify:         %v\n", args.TeamsChannel)
}
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
Usage:
  example [OPTIONS] --name <Name> <EmployeeID> [Department]

Required options:
  -n, --name <Name>            Full name of the new employee

Options:
  -e, --email <Email>          Company email address to assign
  -t, --title <Position>       Job title (e.g., Backend Engineer)
  -F, --full-time              Mark as full-time employee
  -P, --part-time              Mark as part-time employee
  -A, --apprenticeship         Indicates the employee is joining as an apprentice
  -s, --salary <Salary>        Starting salary in USD
  -N, --notify <TeamsChannel>  Slack team channels to notify (e.g., #eng, #ops) (can be specified multiple times)
  -h, --help                   Show this help message and exit

Positional arguments:
  EmployeeID                   Unique employee ID (required)
  Department                   Department name (e.g., Engineering, HR)
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

