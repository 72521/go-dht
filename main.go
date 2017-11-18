package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/champii/go-dht/dht"
	"github.com/urfave/cli"
)

func prepareArgs() *cli.App {
	cli.AppHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}

USAGE:
	{{if .VisibleFlags}}{{.HelpName}} [options]{{end}}
	{{if len .Authors}}
AUTHOR:
	{{range .Authors}}{{ . }}{{end}}
	{{end}}{{if .Commands}}
VERSION:
	{{.Version}}

OPTIONS:
	{{range .VisibleFlags}}{{.}}
	{{end}}{{end}}{{if .Copyright }}

COPYRIGHT:
	{{.Copyright}}
	{{end}}{{if .Version}}
	{{end}}`

	cli.VersionFlag = cli.BoolFlag{
		Name:  "V, version",
		Usage: "Print version",
	}

	cli.HelpFlag = cli.BoolFlag{
		Name:  "h, help",
		Usage: "Print help",
	}

	app := cli.NewApp()

	app.Name = "DHT"
	app.Version = "0.1.1"
	app.Compiled = time.Now()

	app.Usage = "Experimental Distributed Hash Table"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c, connect",
			Usage: "Connect to bootstrap node ip:port",
		},
		cli.StringFlag{
			Name:  "l, listen",
			Usage: "Listening address and port",
			Value: "0.0.0.0:3000",
		},
		cli.BoolFlag{
			Name:  "i",
			Usage: "Interactif",
		},
		cli.BoolFlag{
			Name:  "s",
			Usage: "Stat mode",
		},
		cli.IntFlag{
			Name:  "n, network",
			Value: 0,
			Usage: "Spawn X new `nodes` in a network.",
		},
		cli.IntFlag{
			Name:  "v, verbose",
			Value: 3,
			Usage: "Verbose `level`, 0 for CRITICAL and 5 for DEBUG",
		},
	}

	app.UsageText = "dht [options]"

	return app
}

func manageArgs() {
	app := prepareArgs()

	app.Action = func(c *cli.Context) error {
		options := dht.DhtOptions{
			ListenAddr:    c.String("l"),
			BootstrapAddr: c.String("c"),
			Verbose:       c.Int("v"),
			Stats:         c.Bool("s"),
			Interactif:    c.Bool("i"),
			// OnStore:       func(dht.Packet) interface{} {},
		}

		if c.Int("n") > 0 {
			options.Stats = false
			options.Interactif = false

			cluster(c.Int("n"), options)

			return nil
		}

		if options.Interactif {
			options.Stats = false
		}

		client := dht.New(options)

		if err := client.Start(); err != nil {
			client.Logger().Critical(err)
			return err
		}

		client.Wait()

		return nil
	}

	app.Run(os.Args)
}

func main() {
	manageArgs()

}

func cluster(count int, options dht.DhtOptions) {
	network := []*dht.Dht{}
	i := 0

	if len(options.BootstrapAddr) == 0 {
		client := startOne(options)

		network = append(network, client)

		options.BootstrapAddr = options.ListenAddr

		i++
	}

	for ; i < count; i++ {
		options2 := options

		addrPort := strings.Split(options.ListenAddr, ":")

		addr := addrPort[0]

		port, _ := strconv.Atoi(addrPort[1])

		options2.ListenAddr = addr + ":" + strconv.Itoa(port+i)

		client := startOne(options2)

		network = append(network, client)
	}

	for {
		time.Sleep(time.Second)
	}
}

func startOne(options dht.DhtOptions) *dht.Dht {
	client := dht.New(options)

	if err := client.Start(); err != nil {
		client.Logger().Critical(err)
	}

	return client
}
