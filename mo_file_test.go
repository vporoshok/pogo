package pogo_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gotest.tools/golden"

	"github.com/vporoshok/pogo"
)

func TestMOFile(t *testing.T) {
	data := golden.Get(t, "example.mo")
	buf := bytes.NewBuffer(data)

	mo, err := pogo.ReadMOFile(buf)
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

func TestMOFileWrite(t *testing.T) {
	data := golden.Get(t, "example.mo")
	buf := bytes.NewBuffer(data)

	mo, err := pogo.ReadMOFile(buf)
	require.NoError(t, err)

	res := new(bytes.Buffer)
	require.NoError(t, mo.Write(res))
	golden.AssertBytes(t, res.Bytes(), "example_output.mo")

	mo, err = pogo.ReadMOFile(res)
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
