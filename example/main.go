package main

import (
	"fmt"

	"github.com/tobiashort/flag-go"
)

type Args struct {
	Name     string
	Age      int
	Sex      string
	Employed bool
	Job      string
	Salary   int `flag:"short=m"`
}

func main() {
	args := Args{}
	flag.Parse(&args)

	fmt.Println("Name:", args.Name)
	fmt.Println("Age", args.Age)
	fmt.Println("Sex:", args.Sex)
	fmt.Println("Employed:", args.Employed)
	fmt.Println("Job:", args.Job)
	fmt.Println("Salary:", args.Salary)
}
