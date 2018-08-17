package main

import (
	"errors"
	"os"

	"github.com/codegangsta/cli"
	"golang.org/x/sys/unix"
)

var networkCommand = cli.Command{
	Name:   "network",
	Hidden: true,
	Subcommands: []cli.Command{
		networkCreateCommand,
		networkDeleteCommand,
	},
}

var networkCreateCommand = cli.Command{
	Name: "create",
	Action: func(c *cli.Context) error {
		path := c.Args().First()
		if path == "" {
			return errors.New("netns path required")
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
		return unix.Mount("/proc/self/ns/net", path, "none", unix.MS_BIND, "")
	},
}

var networkDeleteCommand = cli.Command{
	Name: "delete",
	Action: func(c *cli.Context) error {
		path := c.Args().First()
		if path == "" {
			return errors.New("netns path required")
		}
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if err := unix.Unmount(path, 0); err != nil {
			return err
		}

		return os.RemoveAll(path)
	},
}
