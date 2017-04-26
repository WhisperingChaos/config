package config_test

import (
	"entropy/config"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

const eserverConf = "eserver.conf"

// Personal extension to XDG scheme. Supports vendor application scoped
// configuration settings. Should be read only like application executable
// and only changed by vendor during application updates.
const XDG_CONFIG_EXEC = "XDG_CONFIG_EXEC"

func TestLoadFail(t *testing.T) {
	type Opts struct {
		Path  string
		Depth int
	}
	var opts Opts
	err := config.Load(eserverConf, &opts)
	if _, ok := err.(config.LoadFail); !ok {
		t.Fatalf("Unexpected error: %s\n", err.Error())
	}
}

func TestLoadOneConfigEqual(t *testing.T) {
	type Opts struct {
		Path  string
		Depth int
	}
	var proto Opts
	var expected Opts
	expected.Path = "/path"
	expected.Depth = 2
	in := expected
	chain := testCfgChainNew(config.XDG_CONFIG_HOME, in)
	testInherit(t, &proto, &expected, true, chain)
}
func TestLoadTwoConfigEqual(t *testing.T) {
	type Opts struct {
		Path  string
		Str1  string
		Depth int
		Str2  string
		//str1 string
		//	Depth int
		//	str2  string
	}
	type Opts_1 struct {
		Path string
		Str1 string
		//str1 string
	}
	var in_1 Opts_1
	in_1.Path = "/Path/in_1"
	in_1.Str1 = "in_1"

	type Opts_2 struct {
		Path  string
		Depth int
		Str2  string
	}
	var in_2 Opts_2
	in_2.Path = "/path/in_2"
	in_2.Depth = 2
	in_2.Str2 = "in_2"

	var proto Opts
	var expected Opts
	expected.Path = "/path/in_2"
	expected.Str1 = "in_1"
	expected.Depth = 2
	expected.Str2 = "in_2"

	chain_1 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_1)
	chain_2 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_2)
	testInherit(t, &proto, &expected, true, chain_1, chain_2)
}
func TestLoadThreeConfigEqual(t *testing.T) {
	type Opts struct {
		Path  string
		Str1  string
		Depth int
		Str2  string
		Str3  string
	}
	type Opts_1 struct {
		Path string
		Str1 string
	}
	var in_1 Opts_1
	in_1.Path = "/Path/in_1"
	in_1.Str1 = "in_1"

	type Opts_2 struct {
		Path  string
		Depth int
		Str2  string
	}
	var in_2 Opts_2
	in_2.Path = "/path/in_2"
	in_2.Depth = 2
	in_2.Str2 = "in_2"

	type Opts_3 struct {
		Path  string
		Depth int
		Str3  string
	}
	var in_3 Opts_3
	in_3.Path = "/path/in_3"
	in_3.Depth = 3
	in_3.Str3 = "in_3"

	var proto Opts
	var expected Opts
	expected.Path = "/path/in_3"
	expected.Str1 = "in_1"
	expected.Depth = 3
	expected.Str2 = "in_2"
	expected.Str3 = "in_3"

	chain_1 := testCfgChainNew(config.XDG_CONFIG_HOME, in_3)
	chain_2 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_1)
	chain_3 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_2)
	testInherit(t, &proto, &expected, true, chain_1, chain_2, chain_3)
}
func TestLoadFourConfigEqual(t *testing.T) {
	type Opts struct {
		Path  string
		Str1  string
		Depth int
		Str2  string
		Str3  string
		Str4  string
	}
	type Opts_1 struct {
		Path string
		Str1 string
	}
	var in_1 Opts_1
	in_1.Path = "/Path/in_1"
	in_1.Str1 = "in_1"

	type Opts_2 struct {
		Path  string
		Depth int
		Str2  string
	}
	var in_2 Opts_2
	in_2.Path = "/path/in_2"
	in_2.Depth = 2
	in_2.Str2 = "in_2"

	type Opts_3 struct {
		Path  string
		Depth int
		Str3  string
	}
	var in_3 Opts_3
	in_3.Path = "/path/in_3"
	in_3.Depth = 3
	in_3.Str3 = "in_3"

	type Opts_4 struct {
		Path  string
		Depth int
		Str4  string
	}
	var in_4 Opts_4
	in_4.Path = "/path/in_4"
	in_4.Depth = 4
	in_4.Str4 = "in_4"

	var proto Opts
	var expected Opts
	expected.Path = "/path/in_3"
	expected.Str1 = "in_1"
	expected.Depth = 3
	expected.Str2 = "in_2"
	expected.Str3 = "in_3"
	expected.Str4 = "in_4"

	chain_1 := testCfgChainNew(config.XDG_CONFIG_HOME, in_3)
	chain_2 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_1)
	chain_3 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_2)
	chain_4 := testCfgChainNew(XDG_CONFIG_EXEC, in_4)

	testInherit(t, &proto, &expected, true, chain_4, chain_1, chain_2, chain_3)
}
func TestXDG_CONFIG_DIRSdominateEXEequal(t *testing.T) {
	type Opts struct {
		Path  string
		Str1  string
		Depth int
		Str4  string
	}
	type Opts_1 struct {
		Path  string
		Depth int
		Str1  string
	}
	var in_1 Opts_1
	in_1.Path = "/Path/in_1"
	in_1.Depth = 1
	in_1.Str1 = "in_1"

	type Opts_4 struct {
		Path  string
		Depth int
		Str4  string
	}
	var in_4 Opts_4
	in_4.Path = "/path/in_4"
	in_4.Depth = 4
	in_4.Str4 = "in_4"

	var proto Opts
	var expected Opts
	expected.Path = "/Path/in_1"
	expected.Str1 = "in_1"
	expected.Depth = 1
	expected.Str4 = "in_4"

	chain_1 := testCfgChainNew(config.XDG_CONFIG_DIRS, in_1)
	chain_2 := testCfgChainNew(XDG_CONFIG_EXEC, in_4)

	testInherit(t, &proto, &expected, true, chain_1, chain_2)
}

func testInherit(
	t *testing.T,
	proto interface{}, // golang initialized version of the resultant configuration file.
	expected interface{}, // reflects the form of the proto struct and includes the anticipated values.
	assertIdentical bool, // when true, expected must == computed.  when false, expected must != computed.
	//	exec interface{},
	//	homeDir interface{},

	chain ...*testCfgChain, // one or more ordered input config structs with values.  Config Load function merges starting array elements with those that appear later, resulting in later struct values dominating those that appear sooner in the array's order.

) {
	var err error
	var fRemove func()
	fRemove, err = testCreateTempConfigDir(chain)
	defer fRemove()
	if err != nil {
		t.Fatal(err.Error())
	}
	var envUnset func()
	envUnset, err = testSetenv(chain)
	defer envUnset()
	if err != nil {
		t.Fatal(err.Error())
	}
	if err := testConfigCreate(chain); err != nil {
		t.Fatal(err.Error())
	}
	if err := config.Load(eserverConf, proto); err != nil {
		t.Fatal(err.Error())
		return
	}
	identical := reflect.DeepEqual(proto, expected)
	if !identical && assertIdentical {
		t.Fatal("Expected config values don't match computed ones.")
	} else if identical && !assertIdentical {
		t.Fatal("Expected config values match computed ones when they shouldn't.")
	}
}

type testCfgChain struct {
	dirType string
	dir     string
	v       interface{}
}

func testCfgChainNew(dirType string, v interface{}) (c *testCfgChain) {
	c = new(testCfgChain)
	c.dirType = dirType
	c.v = v
	return
}
func testCreateTempConfigDir(cfgs []*testCfgChain) (dfn func(), err error) {
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
	}
	return
}
func testSetenv(cfgs []*testCfgChain) (df func(), err error) {
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
func testConfigCreate(cfgs []*testCfgChain) error {

	for _, cfg := range cfgs {
		eserverFpth := path.Join(cfg.dir, eserverConf)

		var config *os.File
		if iof, err := os.Create(eserverFpth); err != nil {
			return err
		} else {
			config = iof
		}
		encdr := toml.NewEncoder(config)
		if err := encdr.Encode(cfg.v); err != nil {
			return err
		}
		config.Close()
	}
	return nil
}
