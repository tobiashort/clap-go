package flag

import (
	"os"
	"testing"
)

func withArgs(args []string, fn func()) {
	original := os.Args
	defer func() { os.Args = original }()
	os.Args = args
	fn()
}

func TestMandatoryFlag(t *testing.T) {
	withArgs([]string{"prog", "--name", "Alice"}, func() {
		type Args struct {
			Name string `flag:"mandatory,long=name"`
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
			Salary int `flag:"default-value=9999"`
		}

		args := Args{}
		parse(&args)

		if args.Salary != 9999 {
			t.Fatalf("expected default 9999, got %d", args.Salary)
		}
	})
}

func TestShortAndLongFlags(t *testing.T) {
	withArgs([]string{"prog", "-F"}, func() {
		type Args struct {
			FullTime bool `flag:"short=F,long=full-time"`
		}

		args := Args{}
		parse(&args)

		if !args.FullTime {
			t.Fatal("expected FullTime to be true")
		}
	})

	withArgs([]string{"prog", "--full-time"}, func() {
		type Args struct {
			FullTime bool `flag:"short=F,long=full-time"`
		}

		args := Args{}
		parse(&args)

		if !args.FullTime {
			t.Fatal("expected FullTime to be true via long flag")
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

func TestConflictingFlags(t *testing.T) {
	withArgs([]string{"prog", "-F", "-P"}, func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected conflict panic, got none")
			}
		}()

		type Args struct {
			FullTime bool `flag:"short=F,long=full-time,conflicts-with=PartTime"`
			PartTime bool `flag:"short=P"`
		}

		args := Args{}
		parse(&args)
	})
}

func TestStringFlag(t *testing.T) {
	withArgs([]string{"prog", "--email", "test@company.com"}, func() {
		type Args struct {
			Email string `flag:"long=email"`
		}

		args := Args{}
		parse(&args)

		if args.Email != "test@company.com" {
			t.Fatalf("expected email to be set, got '%s'", args.Email)
		}
	})
}

func TestStringSliceFlag(t *testing.T) {
	withArgs([]string{"prog", "-N", "#eng", "-N", "#ops"}, func() {
		type Args struct {
			Notify []string `flag:"short=N,long=notify"`
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
			EmployeeID string `flag:"positional,mandatory"`
			Department string `flag:"positional"`
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

func TestPositionalDefault(t *testing.T) {
	withArgs([]string{"prog", "EMP999"}, func() {
		type Args struct {
			EmployeeID string `flag:"positional,mandatory"`
			Department string `flag:"positional,default-value=Design"`
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
			Apprenticeship bool `flag:"short=A"`
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
			Name string `flag:"mandatory"`
		}

		args := Args{}
		parse(&args)
	})
}
