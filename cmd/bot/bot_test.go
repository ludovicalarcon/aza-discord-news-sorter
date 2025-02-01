package bot

import (
	"testing"
)

func TestBot(t *testing.T) {
	assertError := func(t testing.TB, got, want error) {
		t.Helper()
		if got == nil {
			t.Fatal("didn't get an error but wanted one")
		}

		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}

	bot := Bot{}

	t.Run("It should handle error on token not provided", func(t *testing.T) {
		err := bot.Start()
		assertError(t, err, ErrTokenNotProvided)
	})

	t.Run("It should handle error on api key not provided", func(t *testing.T) {
		t.Setenv(DISCORD_TOKEN, "XXXX")
		err := bot.Start()
		assertError(t, err, ErrApiKeyNotProvided)
	})

	t.Run("It should handle error on invalid token", func(t *testing.T) {
		err := bot.Start()

		if err == nil {
			t.Fatal("didn't get an error but wanted one")
		}
	})
}
