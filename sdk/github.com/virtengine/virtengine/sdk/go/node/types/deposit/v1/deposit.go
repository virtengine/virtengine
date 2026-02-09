package v1

import (
	cerrors "cosmossdk.io/errors"
)

type Deposits []Deposit

type Sources []Source

func (m *Deposit) Validate() error {
	err := m.Amount.Validate()
	if err != nil {
		return err
	}

	if len(m.Sources) == 0 {
		return cerrors.Wrap(ErrInvalidDepositSource, "empty deposit sources")
	}

	sources := make(map[Source]int)

	for _, src := range m.Sources {
		switch src {
		case SourceBalance:
		case SourceGrant:
		default:
			return cerrors.Wrapf(ErrInvalidDepositSource, "empty deposit source type %d", src)
		}

		if _, exists := sources[src]; exists {
			return cerrors.Wrapf(ErrInvalidDepositSource, "duplicate deposit source type %d", src)
		}

		sources[src] = 0
	}

	return nil
}

func (m Deposits) Validate() error {
	for _, d := range m {
		if err := d.Validate(); err != nil {
			return err
		}
	}

	return nil
}
