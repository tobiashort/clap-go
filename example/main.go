package main

import (
	"fmt"

	"github.com/tobiashort/flag-go"
)

type Args struct {
	Name   string `flag:"long=full-name"`
	Age    int
	Male   bool `flag:"conflicts-with=Female"`
	Female bool
	Job    string
	Salary int `flag:"short=$"`
}

func main() {
	args := Args{}
	flag.Parse(&args)

	sex := ""
	if args.Female {
		sex = "Female"
	} else if args.Male {
		sex = "Male"
	} else {
		sex = "Other"
	}

	fmt.Println("Name:", args.Name)
	fmt.Println("Age", args.Age)
	fmt.Println("Sex:", sex)
	fmt.Println("Job:", args.Job)
	fmt.Println("Salary:", args.Salary)
}
