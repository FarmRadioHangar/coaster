package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	yaml "gopkg.in/yaml.v2"
)

// Manitest defines the components installed with the voxbox playbook.
type Manitest struct {
	Version    string       `json:"version"`
	Components []*Component `json:"components"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	Error      struct {
		Version string `json:"version"`
		Message string `json:"message"`
	} `json:"error"`
}

// Greater return true when v1 is greater than v2
func Greater(v1, v2 string) (bool, error) {
	if v1 == "" && v2 != "" {
		return false, nil
	}
	if v1 != "" && v2 == "" {
		return false, nil
	}
	sv1, err := semver.Make(v1)
	if err != nil {
		return false, err
	}
	sv2, err := semver.Make(v2)
	if err != nil {
		return false, err
	}
	return sv1.GTE(sv2), nil
}

// Component represent apps that are part of voxbox
type Component struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// LoadFromPlaybook loads the manifest by traversing the playbook
func LoadFromPlaybook(playbookPath string) (*Manitest, error) {
	m := &Manitest{}
	v, err := ioutil.ReadFile(filepath.Join(playbookPath, "VERSION"))
	if err != nil {
		return nil, err
	}
	m.Version = string(v)
	p := filepath.Join(playbookPath, "roles")
	ferr := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.Contains(path, "vars") {
			c, err := component(p, path)
			if err != nil {
				return err
			}
			m.Components = append(m.Components, c)
		}
		return nil
	})
	if ferr != nil {
		return nil, ferr
	}
	return m, nil
}

func component(base, p string) (*Component, error) {
	bp := strings.TrimPrefix(p, base)
	bp = strings.TrimPrefix(bp, string(filepath.Separator))
	pre := filepath.Join(string(filepath.Separator), "vars", filepath.Base(bp))
	name := strings.TrimSuffix(bp, pre)
	o := make(map[string]interface{})
	d, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(d, &o)
	if err != nil {
		return nil, err
	}
	c := &Component{Name: name}
	vkey := fmt.Sprintf("%s_version", name)
	for k, v := range o {
		if vkey == k {
			c.Version = fmt.Sprint(v)
		}
	}
	return c, nil
}
