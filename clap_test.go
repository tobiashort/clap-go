package clap

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func withArgs(args []string, fn func()) {
	original := os.Args
	defer func() { os.Args = original }()
	os.Args = args
	fn()
}

func TestMandatoryArg(t *testing.T) {
	withArgs([]string{"prog", "--name", "Alice"}, func() {
		type Args struct {
			Name string `clap:"mandatory,long=name"`
		}

		args := Args{}
		parse(&args)

		if args.Name != "Alice" {
			t.Fatalf("expected 'Alice', got '%s'", args.Name)
		}
	})
}

func TestDefaultValue(t *testing.T) {
	withArgs([]string{"prog"}, func() {
		type Args struct {
			Salary int `clap:"default-value=9999"`
		}

		args := Args{}
		parse(&args)

		if args.Salary != 9999 {
			t.Fatalf("expected default 9999, got %d", args.Salary)
		}
	})
}

func TestShortAndLongArgs(t *testing.T) {
	withArgs([]string{"prog", "-F"}, func() {
		type Args struct {
			FullTime bool `clap:"short=F,long=full-time"`
		}

		args := Args{}
		parse(&args)

		if !args.FullTime {
			t.Fatal("expected FullTime to be true")
		}
	})

	withArgs([]string{"prog", "--full-time"}, func() {
		type Args struct {
			FullTime bool `clap:"short=F,long=full-time"`
		}

		args := Args{}
		parse(&args)

		if !args.FullTime {
			t.Fatal("expected FullTime to be true via long argument")
		}
	})
}

func TestShortGrouped(t *testing.T) {
	withArgs([]string{"prog", "-abc"}, func() {
		type Args struct {
			A bool
			B bool
			C bool
		}

		args := Args{}
		parse(&args)

		if !args.A || !args.B || !args.C {
			t.Fatal("expected all to be true")
		}
	})
}

func TestConflictingArgs(t *testing.T) {
	withArgs([]string{"prog", "-F", "-P"}, func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected conflict panic, got none")
			}
		}()

		type Args struct {
			FullTime bool `clap:"short=F,long=full-time,conflicts-with=PartTime"`
			PartTime bool `clap:"short=P"`
		}

		args := Args{}
		parse(&args)
	})
}

func TestStringArg(t *testing.T) {
	withArgs([]string{"prog", "--email", "test@company.com"}, func() {
		type Args struct {
			Email string `clap:"long=email"`
		}

		args := Args{}
		parse(&args)

		if args.Email != "test@company.com" {
			t.Fatalf("expected email to be set, got '%s'", args.Email)
		}
	})
}

func TestStringSliceArg(t *testing.T) {
	withArgs([]string{"prog", "-N", "#eng", "-N", "#ops"}, func() {
		type Args struct {
			Notify []string `clap:"short=N,long=notify"`
		}

		args := Args{}
		parse(&args)

		if len(args.Notify) != 2 || args.Notify[0] != "#eng" || args.Notify[1] != "#ops" {
			t.Fatalf("unexpected slice values: %+v", args.Notify)
		}
	})
}

func TestPositionalArgs(t *testing.T) {
	withArgs([]string{"prog", "EMP123", "Engineering"}, func() {
		type Args struct {
			EmployeeID string `clap:"positional,mandatory"`
			Department string `clap:"positional"`
		}

		args := Args{}
		parse(&args)

		if args.EmployeeID != "EMP123" {
			t.Fatalf("expected EmployeeID 'EMP123', got '%s'", args.EmployeeID)
		}
		if args.Department != "Engineering" {
			t.Fatalf("expected Department 'Engineering', got '%s'", args.Department)
		}
	})
}

func TestPositionalArgsDoubleDash(t *testing.T) {
	withArgs([]string{"prog", "--", "-EMP123-", "--Engineering--"}, func() {
		type Args struct {
			EmployeeID string `clap:"positional,mandatory"`
			Department string `clap:"positional"`
		}

		args := Args{}
		parse(&args)

		if args.EmployeeID != "-EMP123-" {
			t.Fatalf("expected EmployeeID 'EMP123', got '%s'", args.EmployeeID)
		}
		if args.Department != "--Engineering--" {
			t.Fatalf("expected Department 'Engineering', got '%s'", args.Department)
		}
	})
}

func TestPositionalSliceArgs(t *testing.T) {
	withArgs([]string{"prog", "EMP123", "Marketing", "Engineering"}, func() {
		type Args struct {
			EmployeeID  string   `clap:"positional,mandatory"`
			Departments []string `clap:"positional"`
		}

		args := Args{}
		parse(&args)

		if args.EmployeeID != "EMP123" {
			t.Fatalf("expected EmployeeID 'EMP123', got '%s'", args.EmployeeID)
		}

		if !reflect.DeepEqual(args.Departments, []string{"Marketing", "Engineering"}) {
			t.Fatalf("expected Departments [Marketing, Engineering], got '%s'", args.Departments)
		}
	})
}

func TestPositionalDefault(t *testing.T) {
	withArgs([]string{"prog", "EMP999"}, func() {
		type Args struct {
			EmployeeID string `clap:"positional,mandatory"`
			Department string `clap:"positional,default-value=Design"`
		}

		args := Args{}
		parse(&args)

		if args.Department != "Design" {
			t.Fatalf("expected default Department 'Design', got '%s'", args.Department)
		}
	})
}

func TestBoolDefaultFalse(t *testing.T) {
	withArgs([]string{"prog"}, func() {
		type Args struct {
			Apprenticeship bool `clap:"short=A"`
		}

		args := Args{}
		parse(&args)

		if args.Apprenticeship {
			t.Fatal("expected Apprenticeship to be false by default")
		}
	})
}

func TestMissingMandatoryPanics(t *testing.T) {
	withArgs([]string{"prog"}, func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for missing mandatory argument")
			}
		}()

		type Args struct {
			Name string `clap:"mandatory"`
		}

		args := Args{}
		parse(&args)
	})
}

func TestMissingShortAndLongPanics(t *testing.T) {
	withArgs([]string{"prog"}, func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for missing short and long name")
			}
		}()

		type Args struct {
			Name string `clap:"mandatory,short=,long="`
		}

		args := Args{}
		parse(&args)
	})
}

func TestDurationArg(t *testing.T) {
	withArgs([]string{"prog", "--duration", "01h12m02s"}, func() {
		type Args struct {
			Duration time.Duration
		}

		args := Args{}
		parse(&args)

		if fmt.Sprintf("%v", args.Duration) != "1h12m2s" {
			t.FailNow()
		}
	})
}

func TestDefaultValueWithBackslash(t *testing.T) {
	withArgs([]string{"prog"}, func() {
		type Args struct {
			Path string `clap:"default-value='C:\\\\Users\\\\user\\\\My Documents\\\\',description='The path'"`
		}

		args := Args{}
		parse(&args)

		if fmt.Sprintf("%v", args.Path) != "C:\\Users\\user\\My Documents\\" {
			t.Fatal(args.Path)
		}
	})
}
