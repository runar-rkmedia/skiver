package interpolator

import (
	"fmt"
	"time"
)

func mustParseDate(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Errorf("Failed to parse date %s: %w", s, err))
	}
	return t
}

var DefaultInterpolationExamples = map[string]interface{}{
	"count":        42,
	"total":        42,
	"limit":        1000,
	"maxLimit":     5000,
	"value":        83,
	"year":         `2006`,
	"color":        "blue",
	"colour":       "green",
	"error":        "Simulated error",
	"errormessage": "Simulated error",
	"errormsg":     "Simulated error",
	"regionname":   "Gigantis",
	"region":       "Gigantis",
	"country":      "Japan",
	"countryname":  "Japan",
	"companyname":  "Wily",
	"name":         "Douglas Dagurasu",
	"firstName":    "Axl",
	"lastName":     "Akuseru",
	"month":        "August",
	"price":        "5000",
	"email":        "roll@example.com",
	"date":         mustParseDate("1987-12-17T06:00:00-09:00"),
	"expires":      mustParseDate("1995-03-24T06:00:00-09:00"),
	"days":         6,
}

type InterpolationOptions struct {
	// Used to sepearte format from Interpolation value. Default: ","
	// Not yet supported
	FormatSeperator string
	// Prefix for Interpolation. Default: "{{"
	// Not yet supported
	Prefix string
	// Suffix for Interpolation. Default: "{{"
	// Not yet supported
	Suffix string

	// Prefix for nesting of Interpolation. Default: "$t("
	// Not yet supported
	NestingPrefix string
	// Suffix for nesting of Interpolation. Default: ")"
	// Not yet supported
	NestingSuffix string
	// Seperates the options from nesting key. Default: ","
	// Not yet supported
	NestingOptionsSeperator string

	// global variables to use in interpolation replacements: Default: null, but examples uses defaultInterpolationExamples
	// Not yet supported
	DefaultVariables map[string]interface{}
	// After How many interpolation runs to break out before throwing a stack overflow: Default: 1000
	// Not yet supported
	MaxReplaces int

	// Will skip to interpolate the variables, example:
	// Not yet supported
	SkipOnVariables bool
}

type Interpolator struct {
	Options InterpolationOptions
}

func NewInterpolator() Interpolator {
	ip := Interpolator{}
	ip.Options.FormatSeperator = ","
	ip.Options.Prefix = "{{"
	ip.Options.Suffix = "}}"
	ip.Options.NestingPrefix = "$t("
	ip.Options.NestingSuffix = "$t("
	ip.Options.NestingOptionsSeperator = ","
	ip.Options.MaxReplaces = 1000
	ip.Options.SkipOnVariables = true
	return ip
}

type AstKind int

const (
	AstKindLiteral AstKind = iota
)

type InterpolationAST struct {
	Kind AstKind
}

func (ip *Interpolator) Infer(s string) InterpolationAST {

	return InterpolationAST{}
}
