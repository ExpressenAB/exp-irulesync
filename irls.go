package main

import (
	"os/exec"
	"syscall"
	"io/ioutil"
	"fmt"
	"github.com/cldmnky/f5er/f5"
	"github.com/cldmnky/irulesync/config"
	"github.com/cldmnky/irulesync/irule"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"github.com/Songmu/prompter"
)

const version string = "0.1.0"

const irls string = `
 _      _
(_)    | |
 _ _ __| |___
| | '__| / __|
| | |  | \__ \
|_|_|  |_|___/

`

var (
	appliance *f5.Device
	filename  string
	username  string
	password  string
	host      string
	debug     bool
	force     bool
)

func main() {
	app := cli.NewApp()
	app.Name = irls
	app.Usage = "IRuLeSync - Sync iRules to a remote BigIP"
	app.EnableBashCompletion = true
	app.Version = version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Value:       "./irls.yml",
			Usage:       "Path to irls config file",
			EnvVar:      "IRLS_CONFIG",
			Destination: &filename,
		},
		cli.StringFlag{
			Name:        "username, u",
			Value:       "",
			Usage:       "Set username",
			EnvVar:      "IRLS_USERNAME",
			Destination: &username,
		},
		cli.StringFlag{
			Name:        "password, p",
			Value:       "",
			Usage:       "Set password",
			EnvVar:      "IRLS_PASSWORD",
			Destination: &password,
		},
		cli.StringFlag{
			Name:        "host, n",
			Value:       "",
			Usage:       "Set host",
			EnvVar:      "IRLS_HOST",
			Destination: &host,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "enable debug",
			EnvVar:      "IRLS_DEBUG",
			Destination: &debug,
		},
		cli.BoolFlag{
			Name:        "force, f",
			Usage:       "force, don not ask for confirmation",
			EnvVar:      "IRLS_FORCE",
			Destination: &force,
		},
	}
	app.Commands = []cli.Command{
		{
			Name: "push",
			Usage: "push configuration to BigIP",
			Aliases: []string{"s"},
			Action: func(c *cli.Context) error {
				fmt.Println("hmmm", c.Args().First())
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name: "irules",
					Usage: "push irules to BigIP",
					Aliases: []string{"r"},
					Action: func(c *cli.Context) error {
						if c.Args().Present() {
							pushIrule(c.Args().First())
						}
						return nil
					},
				},
			},
		},
		{
			Name:    "pull",
			Usage:   "pull configuration and iRules from BigIP",
			Aliases: []string{"p"},
			Action: func(c *cli.Context) error {
				fmt.Println("specify config or irules to pull", c.Args())
				return nil
			},
			Subcommands: []cli.Command{
				{
					Name: "config",
					Usage: "pull configuration only",
					Aliases: []string{"c"},
					Action: func(c *cli.Context) error {
						conf, err := config.LoadConfigFile(filename)
						if err != nil {
							log.Fatalf("error: %s", err)
						}
						//log.Printf("%s", conf)
						appliance := f5.New(host, username, password, f5.TOKEN)
						if debug == true {
							appliance.SetDebug(true)
						}
						for _, vs := range conf.Vips {
							log.Printf("Checking: %s", vs.Name)
							v, err := config.GetVirtualServer(vs.Name, appliance)
							if err != nil {
								log.Fatalf("error: %s", err)
							}
							y, err := yaml.Marshal(&v)
							fmt.Printf("%s", y)
						}
						return nil
					},
				},
				{
					Name: "irules",
					Usage: "pull iRules only",
					Aliases: []string{"r"},
					Action: func(c *cli.Context) error {
						conf, err := config.LoadConfigFile(filename)
						if err != nil {
							log.Fatalf("error: %s", err)
						}
						appliance := f5.New(host, username, password, f5.TOKEN)
						if debug == true {
							appliance.SetDebug(true)
						}
						for _, vs := range conf.Vips {
							for _, r := range vs.Rules {
								log.Printf("Syncing iRule for %s - Remote: %s, Local: %s", vs.Name, r.Remote, r.Local)
								rirule := irule.GetIrule(r.Remote, appliance)
								if r.Local != "" {
									f, err := os.Create(r.Local)
									if err != nil {
										log.Fatalf("Error writing file: %s", err)
									}
									defer f.Close()
									f.WriteString(rirule)
								}
							}
						}
						return nil
					},
				},
			},
		},
	}
	app.Run(os.Args)
}

func pushIrule(localfn string) error {
	//var ack bool
	var ack = true
	diffcmd, err := exec.LookPath("diff")
	if err != nil {
		log.Fatal("could not find a diff executable in your path")
	}
	conf, err := config.LoadConfigFile(filename)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	appliance := f5.New(host, username, password, f5.TOKEN)
	if debug == true {
		appliance.SetDebug(true)
	}
	// lookup localfn in config
	for _, vs := range conf.Vips {
		for _, r := range vs.Rules {
			if r.Local == localfn {
				// first match get the irule
				tmpfile, err := ioutil.TempFile("", "rule")
				if err != nil {
					log.Fatal(err)
				}
				defer os.Remove(tmpfile.Name())
				rirule := irule.GetIrule(r.Remote, appliance)
				if _, err := tmpfile.WriteString(rirule); err != nil {
					log.Fatalf("Error writing tempfile: %s", err)
				}
				tmpfile.Sync()
				// do the diff
				cmd := exec.Command(diffcmd, "-Naur", "--ignore-all-space", tmpfile.Name(), localfn)
				var waitStatus syscall.WaitStatus
				if out, err := cmd.CombinedOutput(); err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						waitStatus = exitError.Sys().(syscall.WaitStatus)
						if waitStatus.ExitStatus() == 1 {
							// diff returns 1 when diff, 0 when nodiff and > 1 for errors
							fmt.Printf("%s\n", out)
							// ask for confirmation
							if ! force {
								ack = false
								if prompter.YN("Push changes to" + r.Remote + "?", true) {
									ack = true
								}
							}
							if ack {
								fmt.Printf("Pushing local iRule: %s to remote: %s\n", localfn, r.Remote)
								irule.UpdateIruleFile(localfn, r.Remote, appliance)
							}
						}
						return nil
					}

				}
				fmt.Printf("No changes detected\n")
				return nil

			}
		}
	}
	log.Fatalf("Error diffing files (%s)", localfn)
	return nil
}
