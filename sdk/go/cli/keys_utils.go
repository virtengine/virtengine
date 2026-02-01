package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"

	cryptokeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"

	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
)

type bechKeyOutFn func(k *cryptokeyring.Record) (KeyOutput, error)

func printKeyringRecord(w io.Writer, k *cryptokeyring.Record, bechKeyOut bechKeyOutFn, output string) error {
	ko, err := bechKeyOut(k)
	if err != nil {
		return err
	}

	switch output {
	case cflags.OutputFormatText:
		if err := printTextRecords(w, []KeyOutput{ko}); err != nil {
			return err
		}

	case cflags.OutputFormatJSON:
		out, err := json.Marshal(ko)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(w, string(out)); err != nil {
			return err
		}
	}

	return nil
}

func printKeyringRecords(w io.Writer, records []*cryptokeyring.Record, output string) error {
	kos, err := MkAccKeysOutput(records)
	if err != nil {
		return err
	}

	switch output {
	case cflags.OutputFormatText:
		if err := printTextRecords(w, kos); err != nil {
			return err
		}

	case cflags.OutputFormatJSON:
		out, err := json.Marshal(kos)
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s", out); err != nil {
			return err
		}
	}

	return nil
}

func printTextRecords(w io.Writer, kos []KeyOutput) error {
	out, err := yaml.Marshal(&kos)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, string(out)); err != nil {
		return err
	}

	return nil
}

