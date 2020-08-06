package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	t.Run("should parse URL correctly", func(t *testing.T) {
		url := "/fill/300/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg"
		result := ParseURL(url)
		require.Equal(t, URLParams{
			Method:      "fill",
			Height:      200,
			Width:       300,
			Filename:    "a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg",
			ExternalURL: "www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg",
			Valid:       true,
		}, result)
	})

	t.Run("should mark URL as invalid, if URL is not matched by pattert", func(t *testing.T) {
		require.Equal(t, URLParams{
			Valid: false,
		}, ParseURL("/fill/width/200/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg"))
		require.Equal(t, URLParams{
			Valid: false,
		}, ParseURL("/something/200/300/www.audubon.org/sites/default/files/a1_1902_16_barred-owl_sandra_rothenberg_kk.jpg"))
	})
}

func TestAtoi(t *testing.T) {
	t.Run("should convert string to int correctly", func(t *testing.T) {
		require.Equal(t, 10, Atoi("10", 5))
		require.Equal(t, 5, Atoi("string", 5))
	})
}
