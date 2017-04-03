package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"

	"io/ioutil"

	"os"

	"time"

	"github.com/FarmRadioHangar/coaster/manifest"
	"github.com/urfave/cli"
)

const (
	ansible      = "ansible-playbook"
	manifestFile = "voxbox-manifest.json"
)

type config struct {
	PlaybookPath  string    `json:"playbookPath"`
	ManifestDir   string    `json:"manifestDir"`
	InventoryFile string    `json:"inventoryFile"`
	Stderr        io.Writer `json:"-"`
	Stdout        io.Writer `json:"-"`
	Tags          []string  `json:"tags"`
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
func (c *config) inventoryFileStr() string {
	return fmt.Sprintf("inventory/%s", c.InventoryFile)
}

func main() {
	a := cli.NewApp()
	a.Version = "0.1.1"
	a.Name = "coaster"
	a.Usage = "manages ansible playbooks on a host machine"
	a.Action = playService
	a.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Usage: "path to the configuration file",
		},
		cli.StringFlag{
			Name:  "book",
			Usage: "the name of the playbook to run",
		},
		cli.StringSliceFlag{
			Name:  "tags",
			Usage: "a list of playbook tags to run",
		},
		cli.BoolFlag{
			Name:  "force",
			Usage: "will force the operation",
		},
	}
	err := a.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func play(book string, cfg *config) error {
	cmd := exec.Command(ansible,
		"-i", cfg.inventoryFileStr(), "infra.yml", fmt.Sprintf("--tags=%s", cfg.tagsString()),
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
	tags := ctx.StringSlice("tags")
	force := ctx.Bool("force")
	b, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}
	cfg := &config{}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		return err
	}
	if len(tags) > 0 {
		cfg.Tags = nil
		cfg.Tags = append(cfg.Tags, tags...)
	}
	_, err = os.Stat(cfg.ManifestDir)
	if os.IsNotExist(err) {
		os.MkdirAll(cfg.ManifestDir, 0755)
	}
	mp := filepath.Join(cfg.ManifestDir, manifestFile)
	var m manifest.Manitest
	b, err = ioutil.ReadFile(mp)
	if err == nil {
		err = json.Unmarshal(b, &m)
		if err != nil {
			return err
		}

	}

	pm, err := manifest.LoadFromPlaybook(filepath.Join(cfg.PlaybookPath, book))
	if err != nil {
		return err
	}
	if ok, err := manifest.Greater(m.Version, pm.Version); ok && err == nil {
		if !force {
			fmt.Printf(
				"no op current installed playbook v%s>=v%s which you want to install ",
				m.Version, pm.Version,
			)
			return nil
		}
	}
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	err = play(book, cfg)
	if err != nil {
		return err
	}
	now := time.Now()
	if m.Version == "" {
		m.CreatedAt = now
	} else {
		m.UpdatedAt = now
	}
	m.Components = pm.Components
	m.Version = pm.Version
	d, _ := json.MarshalIndent(m, "", "\t")
	return ioutil.WriteFile(mp, d, 0600)
}

func playService(ctx *cli.Context) error {
	return playCMD(ctx, os.Stdout, os.Stderr)
}
