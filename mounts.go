package arcpics

import (
	"fmt"
	"path/filepath"
)

func (lm *LabelMountsType) Set(label string, value string) {
	(*lm)[label] = value
}

func (lm *LabelMountsType) Get(label string) string {
	s := (*lm)[label]
	return s
}

func (lm *LabelMountsType) Html(label string) string {
	mountPoint := lm.Get(label)
	s := ""
	if mountPoint != "" {
		s += fmt.Sprintf("<b>mounted at</b> %s", mountPoint)
	} else {
		s += fmt.Sprintf("<button onclick=\"mountLabelByBrowser('%s')\">mount label by browser %s</button> &nbsp; ", label, label)
		s += fmt.Sprintf("<button onclick=\"mountLabel('%s')\">mount label %s</button>", label, label)
	}
	return s
}
func MountLabeledDirectory(dir string) error {
	if !DirExists(dir) {
		return fmt.Errorf("-m %s :directory doesn't exist", dir)

	}
	label, err := getLabel(dir)
	if err != nil {
		return fmt.Errorf("-m %s :error %s", dir, err.Error())
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("error at MountLabeledDirectory converting to abs path dir: '%s'", dir)
	}

	LabelMounts.Set(label, absDir)
	return nil
}
