package main

import (
	"fmt"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Command string `clap:"command,mandatory,description='The command to run'"`

	List struct {
	} `clap:"description='List all members'"`

	Add struct {
		Name string `clap:"positional,mandatory"`
	} `clap:"description='Adds a member'"`

	Remove struct {
		Name string `clap:"positional,mandatory"`
	} `clap:"description='Removes a member'"`
}

func main() {
	args := Args{}
	clap.Parse(&args)

	switch args.Command {
	case "list":
		fmt.Println("1: Alice")
		fmt.Println("2: Bob")
		fmt.Println("3: Chris")
	case "add":
		fmt.Println("Added " + args.Add.Name)
	case "remove":
		fmt.Println("Removed " + args.Remove.Name)
	}
}
