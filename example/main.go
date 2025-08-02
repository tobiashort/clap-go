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
	Salary         int      `flag:"description='Starting salary in USD'"`
	TeamsChannel   []string `flag:"long=notify,short=N,description='Slack team channels to notify (e.g., #eng, #ops)'"`
	EmployeeID     string   `flag:"positional,mandatory,description='Unique employee ID'"`
	Department     string   `flag:"positional,description='Department name (e.g., Engineering, HR)'"`
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
