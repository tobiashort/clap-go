package main

import (
	"fmt"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Command any `clap:"command,mandatory,description='The command to run'"`

	List struct {
	} `clap:"description='List all members'"`

	Add struct {
		Name string `clap:"positional,mandatory"`
	} `clap:"description='Adds a member'"`

	Remove struct {
		Name string `clap:"positional,mandatory"`
	} `clap:"description='Removes a member'"`

	Foo any `clap:"description='The foo command'"`
	Bar any `clap:"description='The bar command'"`
}

func main() {
	args := Args{}
	clap.Parse(&args)

	switch args.Command {
	case &args.List:
		fmt.Println("1: Alice")
		fmt.Println("2: Bob")
		fmt.Println("3: Chris")
	case &args.Add:
		fmt.Println("Added " + args.Add.Name)
	case &args.Remove:
		fmt.Println("Removed " + args.Remove.Name)
	case &args.Foo:
		fmt.Println("foo")
	case &args.Bar:
		fmt.Println("bar")
	default:
		panic("unreachable")
	}
}
