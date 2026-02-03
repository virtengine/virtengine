package specs

import (
	"embed"
	"os"
)

//go:embed jwt-schema.json
var schemas embed.FS

func GetSpecFile(name string) ([]byte, error) {
	dirEntries, err := schemas.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() && dirEntry.Name() == name {
			schemaContent, err := schemas.ReadFile(dirEntry.Name())
			if err != nil {
				return nil, err
			}

			return schemaContent, nil
		}
	}

	return nil, os.ErrNotExist
}
