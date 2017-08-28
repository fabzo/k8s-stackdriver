package translator

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestCreateMatchQueryString(t *testing.T) {
	testData := [...]struct {
		input     []string
		output    string
	}{
		{
			input: []string{},
			output: "",
		},
		{
			input: nil,
			output: "",
		},
		{
			input: []string{"{job=\"prometheus\"}"},
			output: "?match[]={job=\"prometheus\"}",
		},
		{
			input: []string{"{job=\"prometheus\"}", "{__name__=~\"job:.*\"}"},
			output: "?match[]={job=\"prometheus\"}&match[]={__name__=~\"job:.*\"}",
		},
	}

	for _, c := range testData {
		result := createMatchQueryString(c.input)
		assert.Equal(t, result, c.output)
	}
}