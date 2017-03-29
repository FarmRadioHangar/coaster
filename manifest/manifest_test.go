package manifest

import (
	"fmt"
	"testing"

	"github.com/kr/pretty"
)

func TestLoadFromPlaybook(t *testing.T) {
	p := "fixture/voxbox-playbook"
	m, err := LoadFromPlaybook(p)
	if err != nil {
		t.Fatal(err)
	}
	ev := "0.1.0"
	if m.Version != ev {
		t.Errorf("expected %s got %s", ev, m.Version)
	}
	for _, v := range m.Components {
		switch v.Name {
		case "fconf-node":
			err := matchVersion(v, "")
			if err != nil {
				t.Error(err)
			}
		case "fconf-polymer":
			err := matchVersion(v, "")
			if err != nil {
				t.Error(err)
			}
		case "fdevices":
			err := matchVersion(v, "0.1.9")
			if err != nil {
				t.Error(err)
			}
		case "fessbox":
			err := matchVersion(v, "0.2.0")
			if err != nil {
				t.Error(err)
			}
		case "metrics":
			err := matchVersion(v, "")
			if err != nil {
				t.Error(err)
			}
		case "redis":
			err := matchVersion(v, "3.2")
			if err != nil {
				t.Error(err)
			}
		case "voxbox-ui":
			err := matchVersion(v, "")
			if err != nil {
				t.Error(err)
			}
		case "fastc":
			err := matchVersion(v, "0.1.5")
			if err != nil {
				t.Error(err)
			}
		case "fconf":
			err := matchVersion(v, "0.4.10")
			if err != nil {
				t.Error(err)
			}
		default:
			pretty.Println(v)
		}
	}
	// m.CreatedAt = time.Now()
	// b, _ := json.MarshalIndent(m, "", "\t")
	// ioutil.WriteFile("json", b, 0600)
}

func matchVersion(c *Component, v string) error {
	if c.Version != v {
		return fmt.Errorf("expected %s got %s", v, c.Version)
	}
	return nil
}
