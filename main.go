package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"io/ioutil"

	"os"

	"github.com/urfave/cli"
)

type config struct {
	PlaybookPath  string    `json:"playbookPath"`
	ManifestFile  string    `json:"manifestFile"`
	InventoryFile string    `json:"inventoryFile"`
	Stderr        io.Writer `json:"-"`
	Stdout        io.Writer `json:"-"`
	Tags          []string  `json:"-"`
}

func (c *config) tagsString() string {
	var t string
	for i, v := range c.Tags {
		if i == 0 {
			t += v

		} else {
			t += "," + v
		}
	}
	return t
}

func main() {
	a := cli.NewApp()
	a.Version = "0.1.0"
	a.Name = "coaster"
	a.Usage = "manages ansible playbooks on a host machine"
}

const ansible = "ansible-playbook"

func play(book string, cfg *config) error {
	cmd := exec.Command(ansible,
		"-i", cfg.InventoryFile, fmt.Sprintf("--tags=%s", cfg.tagsString()),
	)
	cmd.Dir = filepath.Join(cfg.PlaybookPath, book)
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func playCMD(ctx *cli.Context, stdout, stderr io.Writer) error {
	cfgFile := ctx.String("config")
	if cfgFile == "" {
		return errors.New("missing configuration file")
	}
	book := ctx.String("book")
	if book == "" {
		return errors.New("missing name of playbook to run")
	}
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	cfg := &config{}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		return err
	}
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	return play(book, cfg)
}

func playService(ctx *cli.Context) error {
	return playCMD(ctx, os.Stdout, os.Stderr)
}
