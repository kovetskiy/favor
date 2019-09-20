package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/reconquest/karma-go"
)

func pick(cmdline []string, trees []*Tree, items []*ScanItem) (*ScanItem, error) {
	var args []string
	if len(cmdline) > 1 {
		args = cmdline[1:]
	}

	cmd := exec.Command(cmdline[0], args...)

	cmd.Stderr = os.Stderr
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to pipe stdin for picker",
		)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to start picker",
		)
	}

	for _, item := range items {
		_, err := stdin.Write(
			[]byte(
				item.tree.Name + ": " + item.dir + "\n",
			),
		)
		if err != nil {
			break
		}
	}

	stdin.Close()

	contents, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to read picker stdout",
		)
	}

	result := strings.TrimSpace((string(contents)))

	err = func() error {
		err = cmd.Wait()
		if exit, ok := err.(*exec.ExitError); ok {
			if status, ok := exit.Sys().(syscall.WaitStatus); ok {
				// fzf returns 130 when user interrupts without choice
				if status.ExitStatus() == 130 {
					return nil
				}
			}
		}

		return err
	}()
	if err != nil {
		return nil, karma.Format(
			err,
			"picker execution failed",
		)
	}

	if len(result) == 0 {
		return nil, nil
	}

	parts := strings.Split(result, ":")
	if len(parts) != 2 {
		return nil, karma.Describe("output", result).Format(
			err,
			"picker %q returned invalid output, "+
				"expected to get format 'name: dir'",
			cmdline,
		)
	}

	name := strings.TrimSpace(parts[0])
	dir := strings.TrimSpace(parts[1])

	for _, item := range items {
		if item.tree.Name == name && item.dir == dir {
			return item, nil
		}
	}

	return nil, karma.Describe("name", name).
		Describe("dir", dir).
		Format(
			nil,
			"invalid picker output: unexpected choose",
		)
}
