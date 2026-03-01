package chirp

import (
	"slices"
	"strings"
)

func CleanMessage(msg string) (cleaned_msg string) {
	profane := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(msg, " ")
	for i, word := range words {
		if slices.Contains(profane, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
