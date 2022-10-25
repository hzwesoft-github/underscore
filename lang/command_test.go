package lang

import (
	"fmt"
	"testing"
)

func TestExecStart(t *testing.T) {
	cmd, err := NewCommand("traceroute", "www.baidu.com")
	if err != nil {
		fmt.Println(err)
	}

	cmd.Start()

	for {
		out := cmd.Output()
		fmt.Println(string(out))

		if !cmd.Available() {
			break
		}
	}

	cmd.WaitFor()
}

func TestExecStartError(t *testing.T) {
	cmd, err := NewCommand("traceroute", "www..com")
	if err != nil {
		fmt.Println(err)
	}

	cmd.Start()

	for {
		out := cmd.Output()
		fmt.Println(string(out))

		if !cmd.Available() {
			break
		}
	}

	cmd.WaitFor()
}

func TestExecRun(t *testing.T) {
	cmd, err := NewCommand("nslookup", "www.baidu.com")
	if err != nil {
		fmt.Println(err)
	}

	out, err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(out))
}
