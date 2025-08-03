package main

import (
	"fmt"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Name           string   `clap:"mandatory,description='Full name of the new employee'"`
	Email          string   `clap:"description='Company email address to assign'"`
	Position       string   `clap:"long=title,short=t,description='Job title (e.g., Backend Engineer)'"`
	FullTime       bool     `clap:"short=F,long=full-time,conflicts-with=PartTime,description='Mark as full-time employee'"`
	PartTime       bool     `clap:"short=P,long=part-time,description='Mark as part-time employee'"`
	Apprenticeship bool     `clap:"short=A,description='Indicates the employee is joining as an apprentice'"`
	Salary         int      `clap:"default-value=9999,description='Starting salary in USD'"`
	TeamsChannel   []string `clap:"long=notify,short=N,description='Slack team channels to notify (e.g., #eng, #ops)'"`
	EmployeeID     string   `clap:"positional,mandatory,description='Unique employee ID'"`
	Department     string   `clap:"positional,default-value=Design,description='Department name (e.g., Engineering, HR)'"`
}

func main() {
	args := Args{}
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
