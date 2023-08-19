package generator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

var (
	ADJECTIVES []string
	NOUNS      []string
)

//go:embed adjectives.json
var a_file []byte

//go:embed nouns.json
var n_file []byte

func readList(data []byte, list *[]string) error {
	if err := json.Unmarshal(data, list); err != nil {
		return err
	}

	return nil
}

func InitGenerator() (err error) {
	err = readList(a_file, &ADJECTIVES)
	if err != nil {
		return
	}
	err = readList(n_file, &NOUNS)
	if err != nil {
		return
	}

	rand.Seed(time.Now().Unix())

	return
}

func GenerateID() string {
	noun := NOUNS[rand.Intn(len(NOUNS))]
	adjective := ADJECTIVES[rand.Intn(len(ADJECTIVES))]
	num := rand.Intn(100)

	return fmt.Sprintf("%s-%s-%d", adjective, noun, num)
}
