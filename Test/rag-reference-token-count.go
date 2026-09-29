// Measures stored output with the reference encoder; it does not run RAG modules.
package main

import (
	"encoding/json"
	"os"

	"github.com/tiktoken-go/tokenizer"
)

func main() {
	var texts []string
	if json.NewDecoder(os.Stdin).Decode(&texts) != nil {
		os.Exit(1)
	}
	codec, err := tokenizer.Get(tokenizer.Cl100kBase)
	if err != nil {
		os.Exit(1)
	}
	counts := make([]int, len(texts))
	for i, text := range texts {
		counts[i], err = codec.Count(text)
		if err != nil {
			os.Exit(1)
		}
	}
	if json.NewEncoder(os.Stdout).Encode(counts) != nil {
		os.Exit(1)
	}
}
