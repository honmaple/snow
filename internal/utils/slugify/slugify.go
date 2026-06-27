package slugify

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/gosimple/unidecode"
	"regexp"
)

var (
	regexpMultipleDashes = regexp.MustCompile("-+")
)

type (
	Slugify struct {
		lowercase       bool
		substitute      map[rune]string
		preserveChars   map[rune]bool
		preserveUnicode bool
	}
	Option func(*Slugify)
)

func (s *Slugify) isAllowed(r rune) bool {
	if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
		return true
	}
	if r == '-' || r == '_' {
		return true
	}
	if s.preserveUnicode && (unicode.IsLetter(r) || unicode.IsNumber(r)) {
		return true
	}
	if s.preserveChars[r] {
		return true
	}
	return false
}

func (s *Slugify) Make(str string) string {
	if !s.preserveUnicode {
		str = unidecode.Unidecode(str)
	}

	var buf bytes.Buffer

	for _, r := range str {
		if s.lowercase {
			r = unicode.ToLower(r)
		}

		if !s.isAllowed(r) {
			r = '-'
		}

		if d, ok := s.substitute[r]; ok {
			buf.WriteString(d)
		} else {
			buf.WriteRune(r)
		}
	}
	result := buf.String()
	result = regexpMultipleDashes.ReplaceAllString(result, "-")
	return strings.Trim(result, "-_")
}

func WithLowercase(lowercase bool) Option {
	return func(s *Slugify) {
		s.lowercase = lowercase
	}
}

func WithSubstitute(sub map[rune]string, override bool) Option {
	return func(s *Slugify) {
		if s.substitute == nil {
			s.substitute = make(map[rune]string)
		}
		if override {
			s.substitute = sub
			return
		}
		for k, v := range sub {
			s.substitute[k] = v
		}
	}
}

func WithPreserveChars(chars string) Option {
	return func(s *Slugify) {
		if s.preserveChars == nil {
			s.preserveChars = make(map[rune]bool)
		}
		for _, r := range chars {
			s.preserveChars[r] = true
		}
	}
}

func WithPreserveUnicode(preserve bool) Option {
	return func(s *Slugify) {
		s.preserveUnicode = preserve
	}
}

func Make(str string, opts ...Option) string {
	return New(opts...).Make(str)
}

func New(opts ...Option) *Slugify {
	s := &Slugify{
		lowercase:       true,
		substitute:      enSub,
		preserveUnicode: false,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

var defaultSub = map[rune]string{
	'"':  "",
	'\'': "",
	'’':  "",
	'‒':  "-", // figure dash
	'–':  "-", // en dash
	'—':  "-", // em dash
	'―':  "-", // horizontal bar
}

var enSub = map[rune]string{
	'&': "and",
	'@': "at",
}

func init() {
	for _, sub := range []*map[rune]string{
		&enSub,
	} {
		for key, value := range defaultSub {
			(*sub)[key] = value
		}
	}
}
