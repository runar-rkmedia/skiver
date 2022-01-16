package importexport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/jmespath/go-jmespath"
	"github.com/pelletier/go-toml"
	"github.com/runar-rkmedia/skiver/types"
	"github.com/runar-rkmedia/skiver/utils"
	"gopkg.in/yaml.v2"
)

var (
	templates = PrepareTemplates()
)

func ExportByGoTemplate(templateName string, project types.ExtendedProject, i18n I18N, w io.Writer) error {

	Locales, err := project.ByLocales()
	if err != nil {
		return err
	}
	return templates.ExecuteTemplate(
		w,
		templateName, //"typescript.tmpl",
		struct {
			Project types.ExtendedProject
			Locales map[string]types.ExtendedLocale
			I18Next I18N
		}{
			Project: project,
			// I18Next: i18n,
			Locales: Locales,
		})
}

func PrepareTemplates() *template.Template {

	templatePath := "templates/*.tmpl"

	t := template.New(filepath.Base(templatePath))
	templateFuncs := make(template.FuncMap)

	jsKeyQuoteRegex := regexp.MustCompile(`(^0|\s)`)

	for k, v := range sprig.TxtFuncMap() {
		// Prehaps we should whitelist instead...
		switch k {
		case "getHostByName", "genSignedCert", "bcrypt", "htpasswd", "genPrivateKey", "derivePassword", "buildCustomCert", "genCA", "genCAWithKey", "genSelfSignedCert", "genSelfSignedCertWithKey", "genSignedCertWithKey", "encryptAES", "decryptAES", "randBytes", "osBase", "osClean", "osDir", "osExt", "osIsAbs", "env", "expandenv":
			continue
		}
		templateFuncs[k] = v

	}
	// TODO: Look into implementing something similar to Helms alterFuncMap: https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L140

	templateFuncs["toYaml"] = func(in interface{}) (string, error) {
		b, err := yaml.Marshal(in)
		return string(b), err
	}
	// copied from: https://github.com/helm/helm/blob/8648ccf5d35d682dcd5f7a9c2082f0aaf071e817/pkg/engine/engine.go#L147-L154
	templateFuncs["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	templateFuncs["join"] = func(sep string, v ...string) string {
		return strings.Join(v, sep)
	}
	templateFuncs["toJson"] = func(in interface{}) (string, error) {
		b, err := json.MarshalIndent(in, "", "  ")
		return string(b), err
	}
	templateFuncs["toToml"] = func(in interface{}) (string, error) {
		b, err := toml.Marshal(in)
		return string(b), err
	}
	templateFuncs["jsKeyQuote"] = func(s string) string {
		if jsKeyQuoteRegex.MatchString(s) {
			return `"` + s + `"`
		}
		return s
	}
	templateFuncs["tsDoc"] = func(in string) (string, error) {
		return in, nil
	}
	templateFuncs["sindent"] = func(spaces int, prefix, v string) string {
		pad := strings.Repeat(" ", spaces)
		return pad + prefix + strings.Replace(v, "\n", "\n"+pad+prefix, -1)
	}
	templateFuncs["sortStrings"] = func(in []string) ([]string, error) {
		sort.Strings(in)
		return in, nil
	}
	templateFuncs["getLocale"] = func(in map[string]types.Locale, localeIDLike string) (types.Locale, error) {
		// sort first
		var keys = make([]string, len(in))
		i := 0
		for k := range in {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for i := 0; i < len(keys); i++ {
			v := in[keys[i]]
			if v.ID == localeIDLike {
				return v, nil
			}
			if v.IETF == localeIDLike {
				return v, nil
			}
			if v.Iso639_3 == localeIDLike {
				return v, nil
			}
			if v.Iso639_2 == localeIDLike {
				return v, nil
			}
			if v.Iso639_1 == localeIDLike {
				return v, nil
			}
		}
		return types.Locale{}, nil
	}
	templateFuncs["structKeys"] = func(in interface{}) ([]string, error) {
		val := reflect.Indirect(reflect.ValueOf(in))
		kind := val.Kind()
		if kind == reflect.Map {
			if m, ok := in.(map[string]interface{}); ok {
				keys := utils.SortedMapKeys(m)
				return keys, nil
			}
		}
		if kind != reflect.Struct {
			return nil, fmt.Errorf("Only structs can be used as input to 'structKeys'. Received type '%s'", kind.String())
		}
		// return nil, fmt.Errorf("%#v", val.Kind().String())
		length := val.NumField()
		keys := make([]string, length)
		for i := 0; i < length; i++ {
			keys[i] = val.Type().Field(i).Name
		}
		keys = sort.StringSlice(keys)
		return keys, nil
	}
	templateFuncs["jmes"] = func(path string, in interface{}) (out interface{}, err error) {
		// b, err := json.Marshal(in)
		var JSON map[string]interface{}

		switch t := in.(type) {
		case []byte:
			err = json.Unmarshal(t, &JSON)
		case string:
			err = json.Unmarshal([]byte(t), &JSON)
		case map[string]interface{}:
			JSON = t
		default:
			b, err := json.Marshal(in)
			if err != nil {
				return out, err
			}
			err = json.Unmarshal([]byte(b), &JSON)
		}
		if err != nil {
			return out, err
		}
		result, err := jmespath.Search(path, JSON)
		return result, err
	}
	t, err := t.Funcs(templateFuncs).ParseFS(Content, templatePath)
	if err != nil {
		panic(err)
	}

	t.Funcs(templateFuncs)

	return t
}
