package pogo_test

import (
	"bytes"
	"context"
	"log"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/golden"

	"github.com/vporoshok/pogo"
)

func TestTranslator(t *testing.T) {
	suite.Run(t, new(TranslatorSuite))
}

type TranslatorSuite struct {
	suite.Suite
}

func (s *TranslatorSuite) TestSimple() {
	tr := pogo.NewTranslator("ru_RU", pogo.FileLoader(s.Pattern()))
	ctx := pogo.ContextWithLanguage(context.Background(), "es_ES")
	msg := tr.Translate(ctx, "Let’s make the web multilingual.")
	s.Equal("Hagamos la web multilingüe.", msg)
}

func (s *TranslatorSuite) TestGoFormat() {
	tr := pogo.NewTranslator("ru_RU", pogo.FileLoader(s.Pattern()))
	ctx := pogo.ContextWithLanguage(context.Background(), "es_ES")
	msg := tr.Translate(ctx, "Welcome back, %s! Your last visit was on %s",
		pogo.WithContext("header"), pogo.WithGoFormat("John", "viernes"),
	)
	s.Equal("¡Bienvenido, John! Su última visita fue el viernes", msg)
}

func (s *TranslatorSuite) TestGoTemplate() {
	tr := pogo.NewTranslator("ru_RU", pogo.FileLoader(s.Pattern()))
	ctx := context.Background()
	msg := tr.Translate(ctx, "Hello {{ . }}",
		pogo.WithDomain("domain"), pogo.WithGoTemplate("Женя"),
	)
	s.Equal("Привет, Женя", msg)
}

func (s *TranslatorSuite) TestPlural() {
	tr := pogo.NewTranslator("ru_RU", pogo.FileLoader(s.Pattern()))
	ctx := context.Background()
	n := 22
	msg := tr.Translate(ctx, "%d page read.",
		pogo.WithPlural("%d pages read.", n), pogo.WithGoFormat(n),
	)
	s.Equal("22 страницы прочитаны.", msg)
}

func (s *TranslatorSuite) TestLogger() {
	buf := new(bytes.Buffer)
	logger := log.New(buf, "", 0)
	tr := pogo.NewTranslator("de_DE",
		pogo.FileLoader(s.Pattern()),
		pogo.WithLogger(logger))
	ctx := context.Background()
	msg := tr.Translate(ctx, "Let’s make the web multilingual.")
	s.Equal("Let’s make the web multilingual.", msg)
	s.Equal("error on load locale \"de_DE/default\": open testdata/locales/de/default.po: no such file or directory\n", buf.String())
}

func (s *TranslatorSuite) Pattern() string {
	relative := filepath.Join("locales", "{{ language }}", "{{ domain }}.{{ ext }}")
	return golden.Path(relative)
}
