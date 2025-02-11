package bot

import (
	"fmt"
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

	t.Run("It should not create a todo from discord message when it is not a link", func(t *testing.T) {
		message := "foobar"
		emoji := "👌"
		wantedErrorMessage := fmt.Sprintf("%s for %s", ErrCouldNotRetrieveTitle.Error(), message)
		err := bot.processMessage(message, emoji)

		if err == nil {
			t.Fatal("didn't get an error but wanted one")
		}
		if err.Error() != wantedErrorMessage {
			t.Fatalf("got %s but wanted %s", err.Error(), wantedErrorMessage)
		}
	})

	t.Run("It should do nothing on unknown emoji", func(t *testing.T) {
		message := "foobar"
		emoji := "😂"
		err := bot.processMessage(message, emoji)
		if err != nil {
			t.Fatalf("got error [%s] but did not want one", err)
		}
	})
}
