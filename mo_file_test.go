package pogo_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vporoshok/pogo"
)

func TestMOFile(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "example.mo"))
	require.NoError(t, err)
	defer f.Close()

	mo, err := pogo.ReadMOFile(f)
	require.NoError(t, err)
	assert.Equal(t, "Сделаем интернет многоязычным.",
		mo.Get("Let’s make the web multilingual."))
	assert.Equal(t, "Добро пожаловать? %s! Ваш последний визит был %s",
		mo.GetCtxt("Welcome back, %s! Your last visit was on %s", "header"))
	assert.Equal(t, "%d страница прочитана.",
		mo.GetN("%d page read.", "%d pages read.", 101))
	assert.Equal(t, "%d страниц прочитано.",
		mo.GetN("%d page read.", "%d pages read.", 12))
	assert.Equal(t, "%d страницы прочитаны.",
		mo.GetN("%d page read.", "%d pages read.", 22))
}
