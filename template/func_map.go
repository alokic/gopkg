package template

import (
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/alokic/gopkg/stringutils"
)

func New(file string) (*template.Template, error) {
	return template.New(path.Base(file)).Funcs(funcMap()).ParseFiles(file)
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"toSnakeCase": func(s string) string {
			return stringutils.ToLowerSnakeCase(s)
		},
		"toUpperFirstCamelCase": func(s string) string {
			return stringutils.ToUpperFirstCamelCase(s)
		},
		"toLowerFirstCamelCase": func(s string) string {
			return stringutils.ToLowerFirstCamelCase(s)
		},
		"toUpperFirst": func(s string) string {
			return stringutils.ToUpperFirst(s)
		},
		"fileSeparator": func() string {
			if filepath.Separator == '\\' {
				return "\\\\"
			}
			return string(filepath.Separator)
		},
		"toCamelCase": func(s string) string {
			return stringutils.ToCamelCase(s)
		},

		"env": func(s string) string {
			return os.Getenv(strings.ToTitle(s))
		},
	}
}
