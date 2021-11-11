package pogo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"text/template"
)

type langCtxKey struct{}

// LanguageFromContext extract language name from context
func LanguageFromContext(ctx context.Context) (string, bool) {
	lang, ok := ctx.Value(langCtxKey{}).(string)

	return lang, ok
}

// ContextWithLanguage store language name in context
func ContextWithLanguage(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, langCtxKey{}, lang)
}

// Locale is an abstract to get translate by parameters
type Locale interface {
	Get(msg string) string
	GetN(msg, plural string, n int) string
	GetCtxt(msg, ctxt string) string
	GetCtxtN(msg, plural, ctxt string, n int) string
}

// Loader is a locale factory
type Loader interface {
	Load(lang, domain string) (Locale, error)
}

type fileLoader struct {
	pattern string
}

func (fl fileLoader) Load(lang, domain string) (Locale, error) {
	loc, err := fl.load(lang, domain)
	if err != nil {
		langs := strings.SplitN(lang, "_", 2)
		if len(langs) == 2 {
			return fl.load(langs[0], domain)
		}
	}
	return loc, err
}
func (fl fileLoader) load(lang, domain string) (Locale, error) {
	if moFile, err := fl.getFile(lang, domain, "mo"); err == nil {
		defer func() {
			_ = moFile.Close()
		}()
		return ReadMOFile(moFile)
	}
	poFile, err := fl.getFile(lang, domain, "po")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = poFile.Close()
	}()
	po, err := ReadPOFile(poFile)
	if err != nil {
		return nil, err
	}
	return po.MO(), nil
}

func (fl fileLoader) getFile(lang, domain, ext string) (io.ReadCloser, error) {
	filepath := fl.pattern
	filepath = strings.Replace(filepath, "{{ language }}", lang, -1) //nolint:gocritic // backward compatibility
	filepath = strings.Replace(filepath, "{{ domain }}", domain, -1) //nolint:gocritic // backward compatibility
	filepath = strings.Replace(filepath, "{{ ext }}", ext, -1)       //nolint:gocritic // backward compatibility

	return os.Open(filepath) //nolint:gosec // it is just library
}

// FileLoader standard loader po and mo files from disk
//
// Pattern should contain next three parts:
// - `{{ language }}` – language in form 'ru_RU' or 'ru';
// - `{{ domain }}` – resources domain (default is 'default');
// - `{{ ext }}` – file extension (mo | po).
//
// For example "./data/locales/{{ language }}/{{ domain }}.{{ ext }}".
//
// File will be loaded on first request with given language and domain. First
// loader try to get mo file with full language name. If any error occurred,
// try to get po file with full language name. On error loader try to get mo
// and po files with short language name. So for default domain and language
// ru_RU sequence of tries in example patter will be next:
// 1. "./data/locales/ru_RU/default.mo";
// 2. "./data/locales/ru_RU/default.po";
// 3. "./data/locales/ru/default.mo";
// 4. "./data/locales/ru/default.po";
func FileLoader(pattern string) Loader {
	return fileLoader{pattern}
}

// Logger is an interface to log errors
type Logger interface {
	Printf(string, ...interface{})
}

// Translator is an object to translate strings
type Translator struct {
	lang    string
	loader  Loader
	logger  Logger
	locales syncLoader
}

// TranslatorOption is an additional configuration to translator object
type TranslatorOption interface {
	apply(*Translator)
}

type fnTranslatorOption func(*Translator)

func (fn fnTranslatorOption) apply(t *Translator) {
	fn(t)
}

// WithLogger add logger to print errors
func WithLogger(logger Logger) TranslatorOption {
	return fnTranslatorOption(func(t *Translator) {
		t.logger = logger
	})
}

// NewTranslator with given language and loader
func NewTranslator(lang string, loader Loader, opts ...TranslatorOption) *Translator {
	t := &Translator{
		lang:   lang,
		loader: loader,
		logger: log.New(os.Stderr, "pogo", log.LstdFlags),
	}
	for _, opt := range opts {
		opt.apply(t)
	}

	return t
}

type translateConfig struct {
	domain    string
	ctxt      string
	pluralN   int
	pluralID  string
	formatter func(string) (string, error)
}

func makeDefaultTranslateConfig() translateConfig {
	return translateConfig{
		domain:  "default",
		pluralN: -1,
	}
}

// TranslateOption options to get translate
type TranslateOption interface {
	apply(translateConfig) translateConfig
}

type fnTranslateOption func(translateConfig) translateConfig

func (fn fnTranslateOption) apply(cfg translateConfig) translateConfig {
	return fn(cfg)
}

// WithDomain add domain to search msg
func WithDomain(domain string) TranslateOption {
	return fnTranslateOption(func(cfg translateConfig) translateConfig {
		cfg.domain = domain
		return cfg
	})
}

// WithContext add context to search msg
func WithContext(ctxt string) TranslateOption {
	return fnTranslateOption(func(cfg translateConfig) translateConfig {
		cfg.ctxt = ctxt
		return cfg
	})
}

// WithPlural add pluralization to search msg
func WithPlural(pluralForm string, n int) TranslateOption {
	return fnTranslateOption(func(cfg translateConfig) translateConfig {
		cfg.pluralN = n
		cfg.pluralID = pluralForm
		return cfg
	})
}

// WithGoFormat format message as fmt.Sprintf
func WithGoFormat(args ...interface{}) TranslateOption {
	return fnTranslateOption(func(cfg translateConfig) translateConfig {
		cfg.formatter = func(msg string) (string, error) {
			return fmt.Sprintf(msg, args...), nil
		}
		return cfg
	})
}

// WithGoTemplate format message as text/template
func WithGoTemplate(data interface{}) TranslateOption {
	return fnTranslateOption(func(cfg translateConfig) translateConfig {
		cfg.formatter = func(msg string) (string, error) {
			tmpl, err := template.New("").Parse(msg)
			if err != nil {
				return "", err
			}
			res := new(bytes.Buffer)
			if err := tmpl.Execute(res, data); err != nil {
				return "", err
			}
			return res.String(), nil
		}
		return cfg
	})
}

// Translate string
//
// Try extract language to translate from context or use default language.
// If there is no locale for given language and domain pair, or if msg is not
// found, return message as is.
func (t *Translator) Translate(ctx context.Context, msg string, opts ...TranslateOption) string {
	cfg := makeDefaultTranslateConfig()
	for _, opt := range opts {
		cfg = opt.apply(cfg)
	}
	lang, ok := LanguageFromContext(ctx)
	if !ok {
		lang = t.lang
	}
	str := t.getMessage(lang, msg, cfg)
	if cfg.formatter == nil {
		return str
	}
	res, err := cfg.formatter(str)
	if err != nil {
		t.logger.Printf("error on format message %q: %+v", str, err)
		return str
	}

	return res
}

func (t *Translator) getMessage(lang, msg string, cfg translateConfig) string {
	loc := t.getLocale(lang, cfg.domain)
	if loc == nil {
		return msg
	}
	if cfg.ctxt == "" {
		if cfg.pluralN < 0 {
			return loc.Get(msg)
		}
		return loc.GetN(msg, cfg.pluralID, cfg.pluralN)
	}
	if cfg.pluralN < 0 {
		return loc.GetCtxt(msg, cfg.ctxt)
	}
	return loc.GetCtxtN(msg, cfg.pluralID, cfg.ctxt, cfg.pluralN)
}

func (t *Translator) getLocale(lang, domain string) Locale {
	loc, _ := t.locales.Load(path.Join(lang, domain), func() interface{} {
		loc, err := t.loader.Load(lang, domain)
		if err != nil {
			t.logger.Printf("error on load locale %q: %+v", path.Join(lang, domain), err)
			return nil
		}
		return loc
	}).(Locale)

	return loc
}
