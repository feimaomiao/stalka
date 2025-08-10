package pandatypes

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/nbio/st"
)

func TestSafeIntToInt32(t *testing.T) {
	value := 1234567890
	expected := int32(1234567890)

	result, err := SafeIntToInt32(value)
	st.Expect(t, err, nil)
	st.Expect(t, result, expected)
	// Test overflow case
	result, err = SafeIntToInt32(2147483648)
	st.Expect(t, err, fmt.Errorf("value 2147483648 overflows int32 range [-2147483648, 2147483647]"))
	st.Expect(t, result, int32(0))
}

func TestToRow(t *testing.T) {
	// open the file static/fetch_data/videogames.json with os.readfile
	data, err := os.ReadFile("../static/fetch_data/videogames.json")
	st.Assert(t, err, nil)

	var game GameLike
	err = json.Unmarshal(data, &game)
	st.Assert(t, err, nil)

	st.Assert(t, game.ID, 34)
	st.Assert(t, game.Name, "Mobile Legends: Bang Bang")
	st.Assert(t, game.Slug, "mlbb")

	// assert that gameR can be converted to GameRow
	gameR := game.ToRow().(GameRow)
	st.Assert(t, recover(), nil)

	// Assert the row data
	st.Assert(t, gameR.ID, game.ID)
	st.Assert(t, gameR.Name, game.Name)
	st.Assert(t, gameR.Slug, game.Slug)
}
