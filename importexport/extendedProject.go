package importexport

import (
	"bytes"
	"fmt"

	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/skiver/types"
)

type ExportOptions struct {
	InOrg                  string
	Project                string
	Locales                []string
	LocaleKey, Format, Tag string
	NoFlatten              bool
}

type Format struct {
	name string
}

func (f Format) String() string            { return f.name }
func (f Format) Is(comparitor Format) bool { return f.name == comparitor.name }

// Deprecated, only used for earlier enums
func (f Format) From(s string) Format { return Format{s} }

var (
	FormatI18n       = Format{"i18n"}
	FormatRaw        = Format{"raw"}
	FormatTypescript = Format{"typescript"}
)

// ExportExtendedProject exports a project into a fileformat.
// If contentType is "", the format is a marshallable format, and can therefore be further marshalled into json, toml, yaml etc.
// On the other hand, some formats are not marshallable, and therefore are already ready to be returned to the user directly
//
// for instance, typescript-format would return simply a []byte, with the contentType set to 'application/typescript'
func ExportExtendedProject(l logger.AppLogger, ep types.ExtendedProject, locales []string, localeKey LocaleKeyEnum, format Format, localeFilter []string) (out interface{}, contentType string, err error) {

	if format.Is(FormatRaw) {
		out = ep
		return
	}
	if len(ep.Locales) == 0 {
		err = fmt.Errorf("No locales were published")
		return
	}
	i18nodes, err := ExportI18N(ep, ExportI18NOptions{
		LocaleFilter: localeFilter,
		LocaleKey:    LocaleKey(localeKey.Name)})
	if err != nil {
		return
	}
	if i18nodes.Nodes == nil {
		return
	}
	i18n, err := I18NNodeToI18Next(i18nodes)
	if err != nil {
		return
	}

	if format.Is(FormatTypescript) {
		var w bytes.Buffer
		if err := ExportByGoTemplate("typescript.tmpl", ep, i18nodes, &w); err != nil {

			return out, contentType, err
		}
		out = w.Bytes()
		contentType = "application/typescript"
		return
	} else {
		if len(locales) == 1 {
			out = i18n[locales[0]]
		} else {
			out = i18n
		}
	}
	return
}
func ExportExtendedProjectToI18Next(l logger.AppLogger, ep types.ExtendedProject, locales []string, localeKey LocaleKeyEnum) (out map[string]interface{}, err error) {
	if len(ep.Locales) == 0 {
		return nil, fmt.Errorf("No locales were published")
	}
	i18nodes, err := ExportI18N(ep, ExportI18NOptions{LocaleKey: LocaleKey(localeKey.Name)})
	if err != nil {
		return nil, err
	}
	if i18nodes.Nodes == nil {
		return nil, fmt.Errorf("No content was produced")
	}
	return I18NNodeToI18Next(i18nodes)
}
