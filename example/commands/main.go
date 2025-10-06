package main

import (
	"fmt"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Command string `clap:"command,mandatory"`

	List struct {
	}

	Add struct {
		Name string `clap:"positional,mandatory"`
	}

	Remove struct {
		Name string `clap:"positional,mandatory"`
	}
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
