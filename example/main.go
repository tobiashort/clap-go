package main

import (
	"fmt"

	"github.com/tobiashort/flag-go"
)

type Args struct {
	Name    string `flag:"mandatory,long=full-name"`
	Age     int
	Male    bool `flag:"conflicts-with=Female"`
	Female  bool
	Job     string
	Salary  int      `flag:"short=$"`
	Friends []string `flag:"short=F"`
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

	fmt.Printf("Name:    %v\n", args.Name)
	fmt.Printf("Age      %v\n", args.Age)
	fmt.Printf("Sex:     %v\n", sex)
	fmt.Printf("Job:     %v\n", args.Job)
	fmt.Printf("Salary:  %v\n", args.Salary)
	fmt.Printf("Friends: %v\n", args.Friends)
}
