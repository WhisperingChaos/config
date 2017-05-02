package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/WhisperingChaos/printbuf"

	"github.com/BurntSushi/toml"
)

/*
Read one or more TOML configuration files and merge their contents.  Merge overlays
existing configuration settings and adds non-existing ones. The operation reads
instances of the configuration file in an order opposite to the XDG Base
Directory Specification to process the most "important" or higher preference
directories last, thereby, ensuring their configuration values dominate the
the resultant config file. This mechanism is analogous to static inhertance.

Notes

  *  This function generates a typed error: "LoadFail", when it's unable to
     locate a single instance of the desired configuration file.
*/
func Load(
	fileName string, // config file name only
	opts interface{}, // a structure whose definition mirrors the desired TOML configuration file.
) error {
	pList := pathDerivation()
	var isLoad bool
	var lf LoadFail
	lf.Init(fmt.Sprintf("Config file %s load failed. Searched: ", fileName))
	for _, flpath := range pList {
		// load default file from config directory
		if _, err := toml.DecodeFile(path.Join(flpath, fileName), opts); err == nil {
			isLoad = true
		} else {
			lf.Sprintf("%s  ", err.Error())
		}
	}
	if !isLoad {
		return lf
	}
	return nil
}

/*
TOML golang decoder requires extension to enable support of formatted golang
duration strings.
*/
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

/*
Generate concrete typed error to facilitate failure testing.  Avoids
parsing error message text that may change.
*/
type LoadFail struct {
	printbuf.T
}

// private --------------------------------------------------------------------

const XDG_CONFIG_HOME = "XDG_CONFIG_HOME"
const XDG_CONFIG_DIRS = "XDG_CONFIG_DIRS"

type envDirMap struct {
	variable string
	defVal   string
}

func (env envDirMap) resolve() []string {
	dlist := os.Getenv(env.variable)

	if len(dlist) < 1 {
		dlist = os.ExpandEnv(env.defVal)
	}
	if len(dlist) < 1 {
		return []string{}
	}
	return remove(filepath.SplitList(dlist), "")
}

func remove(list []string, value string) []string {
	l := make([]string, 0, len(list))
	for _, s := range list {
		if s == value {
			continue
		}
		l = append(l, s)
	}
	return l
}

func reverse(strList []string) []string {
	for i := 0; i < len(strList)/2; i++ {
		j := len(strList) - i - 1
		strList[i], strList[j] = strList[j], strList[i]
	}
	return strList
}
func pathDerivation() []string {
	execDir := func() []string {
		instDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return []string{instDir}
	}
	configDirs := func() []string {
		configDirs := envDirMap{variable: XDG_CONFIG_DIRS, defVal: "/etc/xdg"}
		return reverse(configDirs.resolve())
	}
	homeDir := func() []string {
		homeDir := envDirMap{variable: XDG_CONFIG_HOME, defVal: "$HOME/.config"}
		return homeDir.resolve()
	}
	return append(execDir(), append(configDirs(), homeDir()...)...)
}
