package v2

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	mobilePattern       = "^(?:(?:\\+|0{0,2})91(\\s*[\\-]\\s*)?|[0]?)?[789]\\d{9}$"
	emailPattern        = "^[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?$"
	urlPattern          = "^(?:(?:https?|ftp):\\/\\/)(?:\\S+(?::\\S*)?@)?(?:(?:(?:[a-z\\x00a1-\\xffff0-9]-*)*[a-z\\x00a1-\\xffff0-9]+)(?:\\.(?:[a-z\\x00a1-\\xffff0-9]-*)*[a-z\\x00a1-\\xffff0-9]+)*(?:\\.(?:[a-z\\x00a1-\\xffff]{2,}))\\.?)(?::\\d{2,5})?(?:[/?#]\\S*)?$"
	numericPattern      = "^[0-9]*$"
	alphaPattern        = "^[a-zA-Z]*$"
	alphaNumericPattern = "^[a-zA-Z0-9]*$"
)

const (
	dbTagName   = "db"
	dbIgnoreTag = "-"
)

var (
	strValidatorRequired      = "required"
	strValidatorErrBlankField = "'%s' field can not be empty"
	strValidatorErrPattern    = "'%s' field's pattern is not proper"

	validationHandler = map[string]validFn{
		"regex":        regex,
		"required":     isPresent,
		"mobile":       isMobile,
		"email":        isEmail,
		"username":     isUsername,
		"password":     isPassword,
		"url":          isUrl,
		"alpha":        isAlpha,
		"numeric":      isNumeric,
		"alphanumeric": isAlphanumeric,
		"len":          validateLength,
		"maxlen":       validateMaxLength,
		"minlen":       validateMinLength,
	}
)

// FieldInfo of struct to be used in SQL
type FieldInfo struct {
	NameBindVars string   // ":first_name, :last_name, :email"
	EnumBindVars string   // "$1, $2, $3"
	QBindVars    string   // "?, ? , ?""
	DBTags       string   // "first_name, last_name, email"
	DBTagsArr    []string // ["first_name", "last_name", "email"]
	Names        string   // "FirstName, LastName, Email"
	Valids       map[string]*validInfo
	FieldCount   int
	DriverName   string // postgres, sqllite etc
}

type validFn func(reflect.Value, string) bool

type tag struct {
	name  string
	param string
}

type credStringInfo struct {
	number  bool
	special bool
	length  int
	upper   bool
	unknown bool
}

type validInfo struct {
	Tags     []tag
	Required bool
}

// GenFieldInfo generates info about struct like bindvars, tags, validation info (used for validating struct instance).
// For fields to store in db, it looks for "db" tags in struct
func GenFieldInfo(driverName string, b interface{}) *FieldInfo {
	var namevars, enumvars, qvars, dbtags, names []string
	var js, fn string

	val := reflect.ValueOf(b)
	f := &FieldInfo{Valids: make(map[string]*validInfo)}

	for i := 0; i < val.Type().NumField(); i++ {
		field := val.Type().Field(i)
		js = field.Tag.Get(dbTagName)
		if js != dbIgnoreTag {
			fn = field.Name

			namevars = append(namevars, fmt.Sprintf(":%s", js))
			enumvars = append(enumvars, fmt.Sprintf("$%d", i+1))
			qvars = append(qvars, "?")
			dbtags = append(dbtags, js)
			names = append(names, fn)

			vinfo := parseTags(field.Tag.Get("valid"))
			if vinfo != nil {
				f.Valids[fn] = vinfo
			}
		}
	}
	f.NameBindVars = strings.Join(namevars, ", ")
	f.EnumBindVars = strings.Join(enumvars, ", ")
	f.QBindVars = strings.Join(qvars, ", ")
	f.DBTags = strings.Join(dbtags, ", ")
	f.DBTagsArr = dbtags
	f.Names = strings.Join(names, ", ")
	f.FieldCount = val.Type().NumField()
	f.DriverName = driverName
	return f
}

func (f *FieldInfo) Print() {
	if f == nil {
		return
	}
	fmt.Println(f.NameBindVars)
	fmt.Println(f.EnumBindVars)
	fmt.Println(f.QBindVars)
	fmt.Println(f.DBTags)
	fmt.Println(f.Names)
}

// Validate struct against conttraints.
func (f *FieldInfo) Validate(s interface{}, validateOnlyPresent ...bool) (string, error) {
	if s == nil {
		return "", fmt.Errorf("Struct is nil")
	}

	val := reflect.Indirect(reflect.ValueOf(s))

	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("Expecting struct. Got %T", val)
	}

	t := val.Type()

	checkPresent := (len(validateOnlyPresent) > 0) && validateOnlyPresent[0]

	for i := 0; i < val.NumField(); i++ {
		// check if valid: clause is applied on field, skip otherwise
		vi, ok := f.Valids[t.Field(i).Name]
		if !ok {
			continue
		}

		//fmt.Println(t.Field(i).Name, vi.Tags, vi.Required, val.Field(i).Interface())
		present := isPresent(val.Field(i), "")

		if !present {
			// check if only fields which are present
			if checkPresent {
				continue
			}

			if vi.Required {
				return t.Field(i).Name, fmt.Errorf(strValidatorErrBlankField, t.Field(i).Name)
			}
		}

		if present {
			err := f.validateTags(val.Field(i), t.Field(i).Name, vi.Tags)
			if err != nil {
				return t.Field(i).Name, err
			}
		}
	}
	return "", nil
}

// DollarBindVar generates string for batch insert
// example : INSERT INTO payees (id, name) VALUES ($1, $2), ($3, $4), ($5, $6)
// $ bindvars are used in postgres
func (f *FieldInfo) DollarBindVar(numrecord int) string {
	b := []string{}

	for i := 0; i < numrecord; i++ {
		arr := make([]string, f.FieldCount)
		for j := 0; j < f.FieldCount; j++ {
			arr[j] = fmt.Sprintf("$%d", f.FieldCount*i+j+1)
		}

		b = append(b, fmt.Sprintf("(%s)", strings.Join(arr, ",")))
	}
	return strings.Join(b, ",")
}

// QuestionBindVar generates string for batch insert
// example : INSERT INTO payees (id, name) VALUES (?, ?), (?, ?), (?, ?)
// ? bindvars are used in mysql
func (f *FieldInfo) QuestionBindVar(numrecord int) string {
	b := []string{}
	for i := 0; i < numrecord; i++ {
		b = append(b, fmt.Sprintf("(%s)", f.QBindVars))
	}
	return strings.Join(b, ",")
}

func (f *FieldInfo) validateTags(v reflect.Value, fieldName string, tags []tag) error {
	for _, tag := range tags {
		fn, b := validationHandler[tag.name]
		if !b {
			return fmt.Errorf("'%s' is not validate function", tag.name)
		}

		res := fn(v, tag.param)
		if !res {
			return fmt.Errorf(strValidatorErrPattern, fieldName)
		}

	}
	return nil
}

func parseTags(t string) *validInfo {
	if t == "" {
		return nil
	}

	vd := &validInfo{}

	tl := strings.Split(t, ";")

	for _, i := range tl {
		tg := tag{}
		v := strings.SplitN(i, ":", 2)

		tg.name = strings.Trim(v[0], " ")
		if tg.name == "" {
			return nil
		}

		// skip required
		if tg.name == "required" {
			vd.Required = true
			continue
		}

		if len(v) > 1 {
			tg.param = strings.Trim(v[1], " ")
		}

		vd.Tags = append(vd.Tags, tg)
	}
	return vd
}

func regex(v reflect.Value, s string) bool {
	if v.Kind() != reflect.String {
		return false
	}

	str := v.Interface().(string)
	//return true if value is not set
	if str == "" {
		return true
	}
	b, _ := regexp.MatchString(s, str)
	return b
}

func isPresent(v reflect.Value, s string) bool {
	if v.Interface() == reflect.Zero(v.Type()).Interface() {
		return false
	}
	return true
}

func isEmail(v reflect.Value, s string) bool {
	return regex(v, emailPattern)
}

func isMobile(v reflect.Value, s string) bool {
	return regex(v, mobilePattern)
}

func isPassword(v reflect.Value, s string) bool {
	p := &credStringInfo{}
	str, _ := v.Interface().(string)
	p.parseCreds(str)
	return (p.length >= 8) && !p.unknown
}

func isUsername(v reflect.Value, s string) bool {
	p := &credStringInfo{}
	str, _ := v.Interface().(string)
	p.parseCreds(str)
	return (p.length >= 8) && !p.unknown
}

func isUrl(v reflect.Value, s string) bool {
	return regex(v, urlPattern)
}

func isAlpha(v reflect.Value, s string) bool {
	return regex(v, alphaPattern)
}

func isNumeric(v reflect.Value, s string) bool {
	return regex(v, numericPattern)
}

func isAlphanumeric(v reflect.Value, s string) bool {
	return regex(v, alphaNumericPattern)
}

func validateLength(v reflect.Value, s string) bool {
	str, _ := v.Interface().(string)
	size, _ := strconv.Atoi(s)

	if len(str) == size {
		return true
	}
	return false
}

func validateMaxLength(v reflect.Value, s string) bool {
	str, _ := v.Interface().(string)
	size, _ := strconv.Atoi(s)

	if len(str) <= size {
		return true
	}
	return false
}

func validateMinLength(v reflect.Value, s string) bool {
	str, _ := v.Interface().(string)
	size, _ := strconv.Atoi(s)

	if len(str) >= size {
		return true
	}
	return false
}

func (p *credStringInfo) parseCreds(s string) {
	p.length = len(s)

	for _, s := range s {

		switch {
		case unicode.IsNumber(s):
			p.number = true
		case unicode.IsUpper(s):
			p.upper = true
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			p.special = true
		case unicode.IsLetter(s):
		default:
			p.unknown = true
		}
	}
}
