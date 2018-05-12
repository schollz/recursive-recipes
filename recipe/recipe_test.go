package recipe

import (
	"testing"

	log "github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func BenchmarkGetRecipe(b *testing.B) {
	SetLogLevel("info")
	for i := 0; i < b.N; i++ {
		GetRecipe("chocolate chip cookies", 100000)
	}
}

func TestOpen2(t *testing.T) {
	defer log.Flush()
	err := SetLogLevel("info")
	assert.Nil(t, err)

	err = GetRecipe("chocolate chip cookies", 1)
	assert.Nil(t, err)

}
