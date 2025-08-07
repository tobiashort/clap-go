# clap-go

A command line argument parser in go. Inspired by [clap-rs/clap](https://github.com/clap-rs/clap).

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
package main

import (
	"fmt"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Name           string   `clap:"mandatory,description='Full name of the new employee'"`
	Email          string   `clap:"description='Company email address to assign'"`
	Position       string   `clap:"long=title,short=t,description='Job title (e.g., Backend Engineer)'"`
	FullTime       bool     `clap:"short=F,conflicts-with=PartTime,description='Mark as full-time employee'"`
	PartTime       bool     `clap:"short=P,description='Mark as part-time employee'"`
	Apprenticeship bool     `clap:"short=A,description='Indicates the employee is joining as an apprentice'"`
	Salary         int      `clap:"default-value=9999,description='Starting salary in USD'"`
	TeamsChannel   []string `clap:"long=notify,short=N,description='Slack team channels to notify (e.g., #eng, #ops)'"`
	EmployeeID     string   `clap:"positional,mandatory,description='Unique employee ID'"`
	Department     []string `clap:"positional,mandatory,description='Department name (e.g., Engineering, HR)'"`
}

func main() {
	args := Args{}
	clap.Prog("example")
	clap.Description("This example shall demonstrate how this command line argument parsers works.")
	clap.Parse(&args)

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
$ go run ./example --name "John Doe" --email john@company.com -t "Designer" -F --salary 85000 -N "#design" -N "#it" D12345 Marketing Engineering
=== New Employee Onboarding ===
Name:           John Doe
Email:          john@company.com
Position:       Designer
Type:           Full-Time
Apprenticeship: false
Salary:         $85000
Department:     [Marketing Engineering]
Employee ID:    D12345
Notify:         [#design #it]
```

```shell
$ go run ./example -h
This example shall demonstrate how this command line argument parsers works.

Usage:
  example [OPTIONS] --name <Name> <EmployeeID> <Department> ...

Required options:
  -n, --name <Name>            Full name of the new employee

Options:
  -e, --email <Email>          Company email address to assign
  -t, --title <Position>       Job title (e.g., Backend Engineer)
  -F, --full-time              Mark as full-time employee
  -P, --part-time              Mark as part-time employee
  -A, --apprenticeship         Indicates the employee is joining as an apprentice
  -s, --salary <Salary>        Starting salary in USD (default: 9999)
  -N, --notify <TeamsChannel>  Slack team channels to notify (e.g., #eng, #ops) (can be specified multiple times)
  -h, --help                   Show this help message and exit

Positional arguments:
  EmployeeID                   Unique employee ID (required)
  Department                   Department name (e.g., Engineering, HR) (required, can be specified multiple times)⏎
```

## 🧠 Supported Tag Options

The clap struct tag supports the following options:

|Option          |Type   |Description                                            |
|----------------|-------|-------------------------------------------------------|
|mandatory       |keyword|Argument is required; parser will error if missing     |
|short=x         |string |Single-letter short argument name (e.g. -x)            |
|long=name       |string |Full-length argument name (e.g. --name)                |
|description=... |string |Help/usage description                                 |
|conflicts-with=x|string |Mutually exclusive with another field                  |
|positional      |keyword|Argument must be passed in a specific position         |
|default-value   |string |Default value                                          |

