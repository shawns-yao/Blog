package rag

import (
	"sync"

	"github.com/tiktoken-go/tokenizer"
)

// A fixed reference encoding keeps experiments comparable across provider aliases.
// These counts are exact for cl100k_base, not the billing count of BGE/GPT aliases.
const TokenEncoding = "cl100k_base"

var tokenCodec = sync.OnceValues(func() (tokenizer.Codec, error) {
	return tokenizer.Get(tokenizer.Cl100kBase)
})

func CountTokens(source string) int {
	codec, err := tokenCodec()
	if err == nil {
		count, encodeErr := codec.Count(source)
		if encodeErr == nil {
			return count
		}
	}
	// Bytes are a conservative bound if the reference encoder cannot process input.
	return len(source)
}
