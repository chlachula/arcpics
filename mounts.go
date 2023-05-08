package arcpics

import "fmt"

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
		s += fmt.Sprintf("<button onclick=\"mountLabel('%s')\">mount label %s</button>", label, label)
	}
	return s
}
func MountLabeledDirectory(dir string) {
	if !DirExists(dir) {
		fmt.Printf("-m %s :directory doesn't exist", dir)
		return
	}
	label, err := getLabel(dir)
	if err != nil {
		fmt.Printf("-m %s :error %s", dir, err.Error())
		return
	}
	LabelMounts.Set(label, dir)
}
