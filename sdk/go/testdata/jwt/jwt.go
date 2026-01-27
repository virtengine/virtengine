package jwt

import (
	"embed"
	"os"
)

//go:embed mnemonic cases_es256k.json cases_jwt.json.tmpl
var jwtTestdata embed.FS

func GetTestsFile(name string) ([]byte, error) {
	dirEntries, err := jwtTestdata.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() && dirEntry.Name() == name {
			schemaContent, err := jwtTestdata.ReadFile(dirEntry.Name())
			if err != nil {
				return nil, err
			}

			return schemaContent, nil
		}
	}

	return nil, os.ErrNotExist
}
