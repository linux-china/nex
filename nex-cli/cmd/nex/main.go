package main

import (
	"fmt"
	"os"

	cli "github.com/ConnectEverything/nex/nex-cli"
	"github.com/choria-io/fisk"
	"github.com/fatih/color"
)

func main() {
	blue := color.New(color.FgBlue).SprintFunc()
	help := fmt.Sprintf("%s\nNATS Execution Engine CLI Version %s\n", blue(cli.Banner), cli.VERSION)

	ncli := fisk.New("nex", help)
	ncli.Author("Synadia Communications")
	ncli.UsageWriter(os.Stdout)
	ncli.Version(cli.VERSION)
	ncli.HelpFlag.Short('h')
	ncli.WithCheats().CheatCommand.Hidden()

	ncli.Flag("server", "NATS server urls").Short('s').Envar("NATS_URL").PlaceHolder("URL").StringVar(&cli.Opts.Servers)
	ncli.Flag("user", "Username or Token").Envar("NATS_USER").PlaceHolder("USER").StringVar(&cli.Opts.Username)
	ncli.Flag("password", "Password").Envar("NATS_PASSWORD").PlaceHolder("PASSWORD").StringVar(&cli.Opts.Password)
	ncli.Flag("creds", "User credentials file (JWT authentication)").Envar("NATS_CREDS").PlaceHolder("FILE").StringVar(&cli.Opts.Creds)
	ncli.Flag("nkey", "User NKEY file for single-key auth").Envar("NATS_NKEY").PlaceHolder("FILE").StringVar(&cli.Opts.Nkey)
	ncli.Flag("tlscert", "TLS public certificate file").Envar("NATS_CERT").PlaceHolder("FILE").ExistingFileVar(&cli.Opts.TlsCert)
	ncli.Flag("tlskey", "TLS private key file").Envar("NATS_KEY").PlaceHolder("FILE").ExistingFileVar(&cli.Opts.TlsKey)
	ncli.Flag("tlsca", "TLS certificate authority chain file").Envar("NATS_CA").PlaceHolder("FILE").ExistingFileVar(&cli.Opts.TlsCA)
	ncli.Flag("tlsfirst", "Perform TLS handshake before expecting the server greeting").BoolVar(&cli.Opts.TlsFirst)
	ncli.Flag("timeout", "Time to wait on responses from NATS").Default("2s").Envar("NATS_TIMEOUT").PlaceHolder("DURATION").DurationVar(&cli.Opts.Timeout)

	nodes := ncli.Command("node", "Interact with execution engine nodes")
	nodes_ls := nodes.Command("ls", "List nodes")
	nodes_ls.Action(cli.ListNodes)

	nodes_info := nodes.Command("info", "Get information for an engine node")
	nodes_info.Arg("id", "Public key of the node you're interested in").Required().String()
	nodes_info.Action(cli.NodeInfo)

	run := ncli.Command("run", "Run a workload on a target node")
	run.Arg("id", "Public key of the node to run the workload").Required().String()
	run.Arg("file", "Path to local file to upload and run").File()
	run.Arg("url", "URL pointing to the file to run").URL()
	run.Action(cli.RunWorkload)

	ncli.MustParseWithUsage(os.Args[1:])
}