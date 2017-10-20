package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRun_versionFlag(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}
	args := strings.Split("binrep --version", " ")

	status := cli.Run(args)
	if status != 0 {
		t.Errorf("expected %d to eq %d", status, 0)
	}

	expected := fmt.Sprintf("binrep version %s", version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("expected %q to eq %q", errStream.String(), expected)
	}
}

func TestRun_parseError(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}
	args := strings.Split("binrep --not-exist", " ")

	status := cli.Run(args)
	if status != 1 {
		t.Errorf("expected %d to eq %d", status, 1)
	}

	expected := "--not-exist is undefined subcommand or option"
	if !strings.Contains(errStream.String(), expected) {
		t.Fatalf("expected %q to contain %q", errStream.String(), expected)
	}
}

func TestRun_listCommand(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}
	args := strings.Split("binrep list", " ")

	status := cli.Run(args)
	if status != 1 {
		t.Errorf("expected %d to eq %d", status, 1)
	}

	expected := fmt.Sprintf("binrep version %s", version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("expected %q to eq %q", errStream.String(), expected)
	}
}
