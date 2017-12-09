package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if err := os.Setenv("BINREP_BACKEND_ENDPOINT", "s3://binrep-testing"); err != nil {
		panic(err)
	}
	code := m.Run()
	os.Exit(code)
}

func TestRun_global(t *testing.T) {
	tests := []struct {
		desc           string
		arg            string
		expectedStatus int
		expectedSubOut string
		expectedSubErr string
	}{
		{
			desc:           "version flag",
			arg:            "binrep --version",
			expectedStatus: 0,
			expectedSubErr: fmt.Sprintf("binrep version %s, build %s, date %s", version, commit, date),
		},
		{
			desc:           "undefined flag",
			arg:            "binrep --undefined",
			expectedStatus: 1,
			expectedSubErr: "--undefined is undefined subcommand or option",
		},
		{
			desc:           "credits flag",
			arg:            "binrep --credits",
			expectedStatus: 0,
			expectedSubOut: "= Binrep licensed under: =",
		},
	}
	for _, tc := range tests {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{outStream: outStream, errStream: errStream}
		args := strings.Split(tc.arg, " ")

		status := cli.Run(args)
		if status != tc.expectedStatus {
			t.Errorf("desc: %q, status should be %v, not %v", tc.desc, tc.expectedStatus, status)
		}

		if !strings.Contains(outStream.String(), tc.expectedSubOut) {
			t.Errorf("desc: %q, subout should contain %q, got %q", tc.desc, tc.expectedSubOut, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.expectedSubErr) {
			t.Errorf("desc: %q, subout should contain %q, got %q", tc.desc, tc.expectedSubErr, errStream.String())
		}
	}
}

func TestRun_endpoint(t *testing.T) {
	if err := os.Setenv("BINREP_BACKEND_ENDPOINT", ""); err != nil {
		panic(err)
	}

	tests := []struct {
		desc           string
		arg            string
		expectedStatus int
		expectedSubOut string
		expectedSubErr string
	}{
		{
			desc:           "--endpoint",
			arg:            "binrep --endpoint s3://binrep-testing",
			expectedStatus: 1,
			expectedSubErr: "Usage: binrep",
		},
		{
			desc:           "no endpoint value",
			arg:            "binrep --endpoint",
			expectedStatus: 1,
			expectedSubErr: "want --endpoint value",
		},
	}
	for _, tc := range tests {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{outStream: outStream, errStream: errStream}
		args := strings.Split(tc.arg, " ")

		status := cli.Run(args)
		if status != tc.expectedStatus {
			t.Errorf("desc: %q, status should be %v, not %v", tc.desc, tc.expectedStatus, status)
		}

		if !strings.Contains(outStream.String(), tc.expectedSubOut) {
			t.Errorf("desc: %q, subout should contain %q, got %q", tc.desc, tc.expectedSubOut, outStream.String())
		}
		if !strings.Contains(errStream.String(), tc.expectedSubErr) {
			t.Errorf("desc: %q, subout should contain %q, got %q", tc.desc, tc.expectedSubErr, errStream.String())
		}
	}
}

func TestRun_subCommand(t *testing.T) {
	tests := []struct {
		desc           string
		arg            string
		expectedStatus int
		expectedSubOut string
	}{
		// list
		{
			desc:           "list: display help",
			arg:            "binrep list --help",
			expectedStatus: 2,
			expectedSubOut: "Usage: binrep list",
		},
		{
			desc:           "list: extra arguments error",
			arg:            "binrep list hoge",
			expectedStatus: 2,
			expectedSubOut: "extra arguments",
		},

		// show
		{
			desc:           "show: display help",
			arg:            "binrep show --help",
			expectedStatus: 2,
			expectedSubOut: "Usage: binrep show",
		},
		{
			desc:           "show: arguments error",
			arg:            "binrep show",
			expectedStatus: 2,
			expectedSubOut: "too few arguments",
		},

		// push
		{
			desc:           "push: display help",
			arg:            "binrep push --help",
			expectedStatus: 2,
			expectedSubOut: "Usage: binrep push",
		},
		{
			desc:           "push: arguments error (len: 0)",
			arg:            "binrep push",
			expectedStatus: 2,
			expectedSubOut: "too few arguments",
		},
		{
			desc:           "push: arguments error (len: 1)",
			arg:            "binrep push hoge",
			expectedStatus: 2,
			expectedSubOut: "too few arguments",
		},

		// pull
		{
			desc:           "pull: display help",
			arg:            "binrep pull --help",
			expectedStatus: 2,
			expectedSubOut: "Usage: binrep pull",
		},
		{
			desc:           "pull: arguments error (len: 1)",
			arg:            "binrep pull hoge",
			expectedStatus: 2,
			expectedSubOut: "too few or many arguments",
		},
		{
			desc:           "pull: arguments error (len: 3)",
			arg:            "binrep pull hoge foo bar",
			expectedStatus: 2,
			expectedSubOut: "too few or many arguments",
		},
	}
	for _, tc := range tests {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{outStream: outStream, errStream: errStream}
		args := strings.Split(tc.arg, " ")

		status := cli.Run(args)
		if status != tc.expectedStatus {
			t.Errorf("desc: %q, status should be %v, not %v", tc.desc, tc.expectedStatus, status)
		}

		if !strings.Contains(errStream.String(), tc.expectedSubOut) {
			t.Errorf("desc: %q, subout should contain %q, got %q", tc.desc, tc.expectedSubOut, errStream.String())
		}
	}
}
