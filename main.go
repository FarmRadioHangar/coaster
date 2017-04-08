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

	"bytes"

	"strings"

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
		cli.BoolFlag{
			Name:  "ensure",
			Usage: "this will rerun the play forever until it succeeds",
		},
		cli.DurationFlag{
			Name:  "interval",
			Usage: "the interval to retry running the playbook in minutes",
		},
		cli.DurationFlag{
			Name:  "timeout",
			Usage: "the time  to stop retrying running the playbook in minutes",
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
	fmt.Println("executing  ", strings.Join(cmd.Args, " "))
	cmd.Dir = filepath.Join(cfg.PlaybookPath, book)
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func playCMD(ctx *cli.Context, book string, cfg *config, stdout, stderr io.Writer) error {
	cfg.Stdout = stdout
	cfg.Stderr = stderr
	return play(book, cfg)
}

func playService(ctx *cli.Context) error {
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
	e := ctx.Bool("ensure")
	var stdout bytes.Buffer
	if e {
		i := ctx.Duration("interval")
		timeout := ctx.Duration("timeout")
		fmt.Printf("interval %s timeout %s", i, timeout)
		now := time.Now()
		e := now.Add(timeout)
		err := playCMD(ctx, book, cfg, &stdout, os.Stderr)
		if err == nil {
			return updateManifest(mp, &m, pm, nil)
		}

	stop:
		for {
			time.Sleep(i)
			n := time.Now()
			if n.After(e) {
				return updateManifest(mp, &m, pm, fmt.Errorf("%s\n%v", &stdout, err))
			}
			stdout.Reset()
			err = playCMD(ctx, book, cfg, &stdout, os.Stderr)
			if err == nil {
				break stop
			}
		}
		return err
	}
	return playCMD(ctx, book, cfg, os.Stdout, os.Stderr)
}

func updateManifest(p string, m, pm *manifest.Manitest, err error) error {
	now := time.Now()
	if m.Version == "" {
		m.CreatedAt = now
	} else {
		m.UpdatedAt = now
	}
	if err != nil {
		m.Error.Version = pm.Version
		m.Error.Message = err.Error()
	}
	m.Components = pm.Components
	m.Version = pm.Version
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(p, b, 0600)
}
