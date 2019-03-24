package v2

import (
	"flag"
	"fmt"
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

// viper will by default read config.[ext] in Folders directory.
type config struct {
	Folders   []string
	EnvPrefix string
	cfgStruct interface{}
}

//NewConfig constructor for config.
func NewConfig(envPrefix string, folders ...string) *config {
	return &config{Folders: folders, EnvPrefix: envPrefix}
}

//Load will setup the config object passed by reading configurations from different sources like env, cmd line flag, config file.
func (c *config) Load(cfg interface{}) error {
	if err := c.setupViper(); err != nil {
		return errors.Wrap(err, "error in setting up viper")
	}

	c.cfgStruct = cfg

	iter, err := structutils.NewIterator(c.cfgStruct, []string{"json", "required", "usage"})
	if err != nil {
		return err
	}

	for {
		field := iter.Next()
		if field == nil {
			break
		}

		c.setDefaultConfig(field)
		err = c.setFlag(field)
		if err != nil {
			return err
		}
	}

	c.bindFlagsViper()
	return c.setConfig()
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

	switch field.Type {
	case "string":
		flag.String(jt, typeutils.ToStr(field.Value), ut)
	case "int":
		flag.Int(jt, typeutils.ToInt(field.Value), ut)
	case "float64":
		flag.Float64(jt, typeutils.ToFloat64(field.Value), ut)
	case "uint64":
		flag.Uint64(jt, typeutils.ToUint64(field.Value), ut)
	case "bool":
		flag.Bool(jt, typeutils.ToBool(field.Value), ut)
	}
	return nil
}

func (c *config) setDefaultConfig(field *structutils.Field) {
	if typeutils.Blank(field.Value) {
		return
	}

	key := c.getTag(field, "json")
	if key == "" {
		return
	}

	viper.SetDefault(key, field.Value)
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
	viper.SetEnvPrefix(c.EnvPrefix)
	viper.AutomaticEnv()
	return c.setConfigPath()
}

func (c *config) setConfigPath() error {
	if len(c.Folders) == 0 {
		return nil
	}

	for _, folder := range c.Folders {
		viper.AddConfigPath(folder)
	}

	return viper.ReadInConfig()
}

func (c *config) bindFlagsViper() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}

func (c *config) setConfig() error {
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

		value := viper.Get(key)
		if typeutils.Blank(value) && c.isRequired(field) {
			return fmt.Errorf("%s is required config. Either pass as --%s cmd-line args or set %s_%s as environment variable", field.Name, key, strings.ToUpper(c.EnvPrefix), strings.ToUpper(key))
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
