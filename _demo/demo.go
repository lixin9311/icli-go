package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/lixin9311/icli-go"
	"os"
	"time"
)

var (
	tinyErr  = errors.New("Tiny error.")
	sysFault = errors.New("System Fault.")
)

func test(args ...string) error {
	fmt.Println("In test function.")
	fmt.Println("Args:", args)
	// sub flag set.
	flagset := flag.NewFlagSet(args[0], flag.ContinueOnError)
	// os.Stdout is redirected.
	flagset.SetOutput(os.Stdout)
	fault := flagset.Bool("f", false, "System Fault.")
	if err := flagset.Parse(args[1:]); err != nil {
		fmt.Println("Failed to parse flag: ", err)
		return sysFault
	}
	if *fault {
		return sysFault
	}
	return tinyErr
}

func exit(args ...string) error {
	return icli.ExitIcli
}

// should also process nil error
func errorhandler(e error) error {
	if e == tinyErr {
		fmt.Println("It's a tiny error, nothing hurts.")
	} else if e == sysFault {
		fmt.Println("OMG! System Fault! It's going down!")
		// It will close the screen immediately after return the Exit err.
		// So let's sleep for a while and see the output
		time.Sleep(3 * time.Second)
		return icli.ExitIcli
	}
	return nil
}
func main() {
	icli.AddCmd([]icli.CommandOption{
		{"test", "test", test},
		{"exit", "exit", exit},
	})
	// utf-8 safe
	icli.SetPromt("输入 input >")
	icli.Start(errorhandler)
}
