package main

import (
	"fmt"
	"time"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Name           string        `clap:"mandatory,desc='Full name of the new employee'"`
	Email          string        `clap:"desc='Company email address to assign'"`
	Position       string        `clap:"long=title,short=t,desc='Job title (e.g., Backend Engineer)'"`
	FullTime       bool          `clap:"short=F,conflicts-with=PartTime,desc='Mark as full-time employee'"`
	PartTime       bool          `clap:"short=P,desc='Mark as part-time employee'"`
	Apprenticeship bool          `clap:"short=A,desc='Indicates the employee is joining as an apprentice'"`
	Salary         int           `clap:"default=9999,desc='Starting salary in USD'"`
	TeamsChannel   []string      `clap:"long=notify,short=N,desc='Slack team channels to notify (e.g., #eng, #ops)'"`
	EmployeeID     string        `clap:"positional,mandatory,desc='Unique employee ID'"`
	Department     []string      `clap:"positional,mandatory,desc='Department name (e.g., Engineering, HR)'"`
	LongOmitted    string        `clap:"mandatory,long=,desc='No long name for this argument'"`
	ShortOmitted   string        `clap:"mandatory,short=,desc='No short name for this argument'"`
	Duration       time.Duration `clap:"short=D,desc='Duration like 1h12m0s'"`
}

func main() {
	args := Args{}
	clap.Prog("example")
	clap.Description("This example shall demonstrate how this cmd line argument parsers works.")
	clap.Example(`example --name "John Doe" --email john@company.com -t "Designer" -F --salary 85000 -N "#design" -N "#it" D12345 Marketing Engineering`)
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
	fmt.Printf("Duration:       %v\n", args.Duration)
}
