package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/yuuki/binrep/pkg/command"
)

// CLI is the command line object.
type CLI struct {
	// outStream and errStream are the stdout and stderr
	// to write message from the CLI.
	outStream, errStream io.Writer
}

func main() {
	cli := &CLI{outStream: os.Stdout, errStream: os.Stderr}
	os.Exit(cli.Run(os.Args))
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	if len(args) <= 1 {
		fmt.Fprint(cli.errStream, helpText)
		return 2
	}

	var err error

	switch args[1] {
	case "push":
		err = cli.doPush(args[2:])
	case "pull":
		err = cli.doPull(args[2:])
	// case "sync":
	// 	err = cli.doSync(args[2:])
	case "-v", "--version":
		fmt.Fprintf(cli.errStream, "%s version %s, build %s \n", Name, Version, GitCommit)
		return 0
	case "-h", "--help":
		fmt.Fprint(cli.errStream, helpText)
	default:
		fmt.Fprint(cli.errStream, helpText)
		return 1
	}

	if err != nil {
		fmt.Fprintln(cli.errStream, err)
		return 2
	}

	return 0
}

var helpText = `
Usage: binrep [options]

  static binary repository.

Commands:
  push		push binary.
  pull		pull binary.
  sync          sync remote repository to local directory.

Options:
  --version, -v		print version
  --help, -h            print help
`

func (cli *CLI) prepareFlags(help string) *flag.FlagSet {
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() {
		fmt.Fprint(cli.errStream, help)
	}
	return flags
}

var pushHelpText = `
Usage: binrep push [options] <host>/<user>/<project> /path/to/binary

push binary.

Options:
  --endpoint, -e	s3 uri
  --timestamp, -t       binary timestamp
`

func (cli *CLI) doPush(args []string) error {
	var param command.PushParam
	flags := cli.prepareFlags(pushHelpText)
	flags.StringVar(&param.Timestamp, "t", "", "")
	flags.StringVar(&param.Timestamp, "timestamp", "", "")
	flags.StringVar(&param.Endpoint, "e", "", "")
	flags.StringVar(&param.Endpoint, "endpoint", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if param.Endpoint == "" {
		fmt.Fprint(cli.errStream, pushHelpText)
		return errors.Errorf("--endpoint required")
	}
	if len(flags.Args()) != 2 {
		fmt.Fprint(cli.errStream, pushHelpText)
		return errors.Errorf("too few or many arguments")
	}
	return command.Push(&param, flags.Arg(0), flags.Arg(1))
}

var pullHelpText = `
Usage: binrep pull [options] <host>/<user>/<project> /path/to/binary

pull binary.

Options:
  --endpoint, -e	s3 uri
  --timestamp, -t       binary timestamp
`

func (cli *CLI) doPull(args []string) error {
	var param command.PullParam
	flags := cli.prepareFlags(pullHelpText)
	flags.StringVar(&param.Timestamp, "t", "", "")
	flags.StringVar(&param.Timestamp, "timestamp", "", "")
	flags.StringVar(&param.Endpoint, "e", "", "")
	flags.StringVar(&param.Endpoint, "endpoint", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if param.Endpoint == "" {
		fmt.Fprint(cli.errStream, pullHelpText)
		return errors.Errorf("--endpoint required")
	}
	if len(flags.Args()) != 2 {
		fmt.Fprint(cli.errStream, pullHelpText)
		return errors.Errorf("too few or many arguments")
	}
	return command.Pull(&param, flags.Arg(0), flags.Arg(1))
}
