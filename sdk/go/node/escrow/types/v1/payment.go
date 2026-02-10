package v1

type Payments []Payment

func (m *Payment) ValidateBasic() error {
	err := m.ID.ValidateBasic()
	if err != nil {
		return err
	}

	err = m.State.ValidateBasic()
	if err != nil {
		return err
	}

	return nil
}
