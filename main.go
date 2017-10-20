//go:generate go-bindata -pkg main -o credits.go vendor/CREDITS
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/yuuki/binrep/pkg/command"
)

const (
	defaultKeepReleases    int = 5
	defaultSyncConcurrency int = 4
)

var (
	creditsText string = string(MustAsset("vendor/CREDITS"))
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

	switch cmd := args[1]; cmd {
	case "list":
		err = cli.doList(args[2:])
	case "show":
		err = cli.doShow(args[2:])
	case "push":
		err = cli.doPush(args[2:])
	case "pull":
		err = cli.doPull(args[2:])
	case "sync":
		err = cli.doSync(args[2:])
	case "-v", "--version":
		fmt.Fprintf(cli.errStream, "%s version %s, build %s, date %s \n", name, version, commit, date)
		return 0
	case "--credits":
		fmt.Println(creditsText)
		return 0
	case "-h", "--help":
		fmt.Fprint(cli.errStream, helpText)
	default:
		fmt.Fprintf(cli.errStream, "%s is undefined subcommand or option", cmd)
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

  The static binary repository manager.

Commands:
  list		show releases on remote repository
  show          show binary information.
  push		push binary.
  pull		pull binary.
  sync          sync remote repository to local directory.

Options:
  --version, -v		print version
  --help, -h            print help
`

func (cli *CLI) prepareFlags(help string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)
	flags.Usage = func() {
		fmt.Fprint(cli.errStream, help)
	}
	return flags
}

var listHelpText = `
Usage: binrep list [options]

show releases on remote repository

Options:
  --endpoint, -e   s3 URI
`

func (cli *CLI) doList(args []string) error {
	var param command.ListParam
	flags := cli.prepareFlags(listHelpText)
	flags.StringVar(&param.Endpoint, "e", "", "")
	flags.StringVar(&param.Endpoint, "endpoint", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if param.Endpoint == "" {
		fmt.Fprint(cli.errStream, listHelpText)
		return errors.Errorf("--endpoint required")
	}
	if len(flags.Args()) != 0 {
		fmt.Fprint(cli.errStream, listHelpText)
		return errors.Errorf("extra arguments")
	}
	return command.List(&param)
}

var showHelpText = `
Usage: binrep show [options] <host>/<user>/<project>

show binary information.

Options:
  --endpoint, -e	s3 uri
  --timestamp, -t       binary timestamp
`

func (cli *CLI) doShow(args []string) error {
	var param command.ShowParam
	flags := cli.prepareFlags(showHelpText)
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
	if len(flags.Args()) < 1 {
		fmt.Fprint(cli.errStream, pushHelpText)
		return errors.Errorf("too few arguments")
	}
	return command.Show(&param, flags.Arg(0))
}

var pushHelpText = `
Usage: binrep push [options] <host>/<user>/<project> /path/to/binary ...

push binary.

Options:
  --endpoint, -e	s3 uri
  --timestamp, -t       binary timestamp
  --keep-releases, -k	the number of releases that it keeps (default: 5)
`

func (cli *CLI) doPush(args []string) error {
	var param command.PushParam
	flags := cli.prepareFlags(pushHelpText)
	flags.StringVar(&param.Timestamp, "t", "", "")
	flags.StringVar(&param.Timestamp, "timestamp", "", "")
	flags.IntVar(&param.KeepReleases, "k", defaultKeepReleases, "")
	flags.IntVar(&param.KeepReleases, "keep-releases", defaultKeepReleases, "")
	flags.StringVar(&param.Endpoint, "e", "", "")
	flags.StringVar(&param.Endpoint, "endpoint", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if param.Endpoint == "" {
		fmt.Fprint(cli.errStream, pushHelpText)
		return errors.Errorf("--endpoint required")
	}
	argLen := len(flags.Args())
	if argLen < 2 {
		fmt.Fprint(cli.errStream, pushHelpText)
		return errors.Errorf("too few arguments")
	}
	return command.Push(&param, flags.Arg(0), flags.Args()[1:argLen])
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

var syncHelpText = `
Usage: binrep sync [options] DEST_DIR

pull binary.

Options:
  --endpoint, -e	s3 uri
  --concurrency, -c     number of multiple release fetchers to make at a time
  --max-bandwidth, -bw	max bandwidth for download binaries (Bytes/sec) eg. '1 MB', '1024 KB'
`

func (cli *CLI) doSync(args []string) error {
	var param command.SyncParam
	flags := cli.prepareFlags(syncHelpText)
	flags.StringVar(&param.Endpoint, "e", "", "")
	flags.StringVar(&param.Endpoint, "endpoint", "", "")
	flags.IntVar(&param.Concurrency, "c", defaultSyncConcurrency, "")
	flags.IntVar(&param.Concurrency, "concurrency", defaultSyncConcurrency, "")
	flags.StringVar(&param.MaxBandWidth, "bw", "", "")
	flags.StringVar(&param.MaxBandWidth, "max-bandwidth", "", "")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if param.Endpoint == "" {
		fmt.Fprint(cli.errStream, syncHelpText)
		return errors.Errorf("--endpoint required")
	}
	if len(flags.Args()) < 1 {
		fmt.Fprint(cli.errStream, syncHelpText)
		return errors.Errorf("too few arguments")
	}
	return command.Sync(&param, flags.Arg(0))
}
