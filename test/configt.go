package configt

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const EserverConf = "eserver.conf"

// Personal extension to XDG scheme. Supports vendor application scoped
// configuration settings. Should be read only like application executable
// and only changed by vendor during application updates.
const XDG_CONFIG_EXEC = "XDG_CONFIG_EXEC"

type Link struct {
	dirType string
	dir     string
	v       interface{}
}

func LinkNew(dirType string, v interface{}) (c *Link) {
	c = new(Link)
	c.dirType = dirType
	c.v = v
	return
}
func Chain(
	chain ...*Link, // one or more ordered input config structs with values.  Config Load function merges starting array elements with those that appear later, resulting in later struct values dominating those that appear sooner in the array's order.

) (cleanup func(), // cleanup environment variables and config files
	err error,
) {
	var fRemove func()
	if fRemove, err = createTempConfigDir(chain); err != nil {
		return nil, err
	}
	cleanup = func() {
		fRemove()
	}
	var envUnset func()
	if envUnset, err = setenv(chain); err != nil {
		cleanup()
		return nil, err
	}
	cleanup = func() {
		envUnset()
		fRemove()
	}
	if err = create(chain); err != nil {
		cleanup()
		return nil, err
	}
	return
}
func createTempConfigDir(cfgs []*Link) (dfn func(), err error) {
	dirs := make([]string, 0, 10)
	dfn = func() {
		for _, dir := range dirs {
			os.RemoveAll(dir)
		}
	}
	for id, cfg := range cfgs {
		var cnfPath string
		if cfg.dirType == XDG_CONFIG_EXEC {
			cnfPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		} else {
			cnfPath = path.Join(os.TempDir(), cfg.dirType+"TestConfig"+strconv.Itoa(id))
			if err = os.MkdirAll(cnfPath, (os.ModeDir | os.ModePerm)); err != nil {
				return
			}
			dirs = append(dirs, cnfPath)
		}
		cfg.dir = cnfPath
		fmt.Fprintf(os.Stderr, "cnfPath '%s'", cfg.dir)
	}
	return
}
func setenv(cfgs []*Link) (df func(), err error) {
	dmap := make(map[string]string)
	envs := make([]string, 0, 1)
	df = func() {
		for _, env := range envs {
			os.Unsetenv(env)
		}
	}
	for _, cfg := range cfgs {
		if dir := dmap[cfg.dirType]; dir == "" {
			dmap[cfg.dirType] = cfg.dir
		} else {
			// reverse order path as the last directory in the array is the most
			// important XDG directory and last chain element is most important.
			dmap[cfg.dirType] = cfg.dir + string(filepath.ListSeparator) + dir
		}
	}
	for dType := range dmap {
		if !strings.HasPrefix(dType, "XDG") {
			err = fmt.Errorf("Unsupported directory type: %s. See XDG_...", dType)
			break
		}
		if err = os.Setenv(dType, dmap[dType]); err != nil {
			break
		}
		envs = append(envs, dType)
	}
	return
}
func create(cfgs []*Link) error {

	for _, cfg := range cfgs {
		eserverFpth := path.Join(cfg.dir, EserverConf)

		var config *os.File
		if iof, err := os.Create(eserverFpth); err != nil {
			return err
		} else {
			config = iof
		}
		encdr := toml.NewEncoder(config)
		if err := encdr.Encode(cfg.v); err != nil {
			fmt.Fprintf(os.Stderr, "here\n")
			return err
		}
		fmt.Fprintf(os.Stderr, "here2 '%s' \n", eserverFpth)
		config.Close()
		Copy("/tmp/eserver.conf", eserverFpth)
	}
	return nil
}
func Copy(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}
