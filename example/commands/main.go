package main

import (
	"fmt"

	"github.com/tobiashort/clap-go"
)

type Args struct {
	Command any `clap:"cmd,mandatory,desc='The cmd to run'"`

	List any `clap:"cmdopt,desc='List all members'"`

	Add struct {
		Name string `clap:"positional,mandatory"`
	} `clap:"cmdopt,desc='Adds a member'"`

	Remove struct {
		Name string `clap:"positional,mandatory"`
	} `clap:"cmdopt,desc='Removes a member'"`

	Foo any `clap:"cmdopt,desc='The foo cmd'"`
	Bar any `clap:"cmdopt,desc='The bar cmd'"`
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
