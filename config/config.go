package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/alokic/gopkg/structutils"
	"github.com/alokic/gopkg/typeutils"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Logger interface for config.
type Logger interface {
	Log(...interface{}) error
}

// viper will by default read config.[ext] in folders directory.
type config struct {
	folders   []string
	envPrefix string
	cfgStruct interface{}
	flags     *pflag.FlagSet
	v         *viper.Viper
}

// New constructor for config.
func New(cfg interface{}, envPrefix string, flags *pflag.FlagSet, folders ...string) *config {
	return &config{cfgStruct: cfg, folders: folders, envPrefix: envPrefix, flags: flags, v: viper.New()}
}

//Load will setup the config object passed by reading configurations from different sources like env, cmd line flag, config file.
func (c *config) Load() error {
	if err := c.setupViper(); err != nil {
		return errors.Wrap(err, "error in setting up viper")
	}

	iter, err := structutils.NewIterator(c.cfgStruct, []string{"json", "required", "usage"})
	if err != nil {
		return err
	}

	for {
		field := iter.Next()
		if field == nil {
			break
		}

		c.setDefaultValue(field)
		err = c.setFlag(field)
		if err != nil {
			return err
		}
	}

	c.parseFlags()
	return c.populateStruct()
}

// Print config.
func (c *config) Print(logger Logger) error {
	iter, err := structutils.NewIterator(c.cfgStruct, []string{"print"})
	if err != nil {
		return err
	}
	for {
		field := iter.Next()
		if field == nil {
			break
		}

		val := field.Value
		if c.getTag(field, "print") == "false" {
			val = "MASKED"
		}
		logger.Log(field.Name, val)
	}
	return nil
}

func (c *config) setFlag(field *structutils.Field) error {
	jt := c.getTag(field, "json")
	if jt == "" {
		return fmt.Errorf("missing json tag in %s", field.Name)
	}

	ut := c.getTag(field, "usage")
	if jt == "" {
		return fmt.Errorf("missing usage tag in %s", field.Name)
	}

	if c.isRequired(field) {
		ut += " (mandatory)"
	} else {
		ut += " (optional)"
	}

	// get default or environment value
	defaultVal := c.v.Get(jt)

	switch field.Type {
	case "string":
		flag.String(jt, typeutils.ToStr(defaultVal), ut)
	case "int":
		flag.Int(jt, typeutils.ToInt(defaultVal), ut)
	case "float64":
		flag.Float64(jt, typeutils.ToFloat64(defaultVal), ut)
	case "uint64":
		flag.Uint64(jt, typeutils.ToUint64(defaultVal), ut)
	case "bool":
		flag.Bool(jt, typeutils.ToBool(defaultVal), ut)
	}

	return nil
}

func (c *config) setDefaultValue(field *structutils.Field) {
	if typeutils.Blank(field.Value) {
		return
	}

	key := c.getTag(field, "json")
	if key == "" {
		return
	}

	c.v.SetDefault(key, field.Value)
}

func (c *config) getTag(field *structutils.Field, tag string) string {
	for _, v := range field.Tags {
		if v.Name == tag {
			return v.Value
		}
	}
	return ""
}

func (c *config) setupViper() error {
	if c.envPrefix != "" {
		viper.SetEnvPrefix(c.envPrefix)
	}

	c.v.AutomaticEnv()
	return c.setConfigPath()
}

func (c *config) setConfigPath() error {
	if len(c.folders) == 0 {
		return nil
	}

	for _, folder := range c.folders {
		c.v.AddConfigPath(folder)
	}

	return c.v.ReadInConfig()
}

func (c *config) parseFlags() {
	c.flags.AddGoFlagSet(flag.CommandLine)
	c.v.BindPFlags(c.flags)
	c.flags.Parse(os.Args[1:])
}

func (c *config) populateStruct() error {
	iter, err := structutils.NewIterator(c.cfgStruct, []string{"json", "required"})
	if err != nil {
		return err
	}

	for {
		field := iter.Next()

		if field == nil {
			break
		}

		key := c.getTag(field, "json")

		value := c.v.Get(key)
		if typeutils.Blank(value) && c.isRequired(field) {
			return fmt.Errorf("%s is required config. Either pass as --%s cmd-line args or set %s as environment variable", field.Name, key, c.envKey(key))
		}

		err := c.setStructField(field, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *config) setStructField(field *structutils.Field, value interface{}) error {
	// pointer to struct
	s := reflect.ValueOf(c.cfgStruct)

	if s.Kind() == reflect.Ptr {
		// s is a pointer, indirect it to get the
		// underlying value, and make sure it is a struct.
		// struct
		s = s.Elem()
	}

	f := s.FieldByName(field.Name)
	if !f.IsValid() || !f.CanSet() {
		return fmt.Errorf("cannot set %s", field.Name)
	}

	switch field.Type {
	case "string":
		f.SetString(typeutils.ToStr(value))
	case "int":
		f.SetInt(typeutils.ToInt64(value))
	case "float64":
		f.SetFloat(typeutils.ToFloat64(value))
	case "uint64":
		f.SetUint(typeutils.ToUint64(value))
	case "bool":
		f.SetBool(typeutils.ToBool(value))
	}
	return nil
}

func (c *config) isRequired(field *structutils.Field) bool {
	return c.getTag(field, "required") == "true"
}

func (c *config) envKey(key string) string {
	if c.envPrefix != "" {
		return strings.ToUpper(c.envPrefix) + "_" + strings.ToUpper(key)
	}
	return strings.ToUpper(key)
}
