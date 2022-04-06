package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/sewh/cping/config"
	"github.com/sewh/cping/icmp"
)

//go:embed usage.txt
var usage string

func Usage() {
	fmt.Printf("\n%s\n", usage)
}

func main() {
	// make sure we have arguments
	if len(os.Args) < 2 {
		Usage()
		os.Exit(0)
	}

	// parse arguments into a config
	c := config.Default()
	err := c.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// if help requested, print and exit
	if c.HelpMode {
		Usage()
		os.Exit(0)
	}

	// make sure we have an address
	if !c.ValidIP() {
		fmt.Println("malformed ip address")
		os.Exit(1)
	}

	// create the sender
	sender := icmp.NewSender(c)
	defer sender.Close()

	// print the header banner
	fmt.Printf("Sending %d, %d-byte ICMP Echos to %s, timeout is %d seconds:\n",
		c.Count, c.Size, c.DestIP, c.TimeoutSecs)

	// run the ping loop!
	for i := 0; i < c.Count; i += 1 {
		res := sender.SendAndReceive()
		switch res {
		case nil:
			fmt.Printf("!")
		case icmp.TimeoutExceeded:
			fmt.Printf(".")
		case icmp.DestUnreachable:
			fmt.Printf("U")
		case icmp.SourceQuench:
			fmt.Printf("Q")
		case icmp.CouldNotFragment:
			fmt.Printf("M")
		case icmp.UnknownPacket:
			fmt.Printf("?")
		case icmp.TTLExpired:
			fmt.Printf("&")
		default:
			fmt.Println(res)
			os.Exit(1)
		}
	}

	// print the result banner
	// Success rate is 100 percent (5/5), round-trip min/avg/max = 2/2/3 ms
	fmt.Println("\n" + sender.Stats())
}
