package testservice

import "errors"

// Validate returns an error if the message is invalid.
func (m *Input) Validate() error {
	if m.GetData() == "" {
		return errors.New("input data must not be empty")
	}

	return nil
}

// Validate returns an error if the message is invalid.
func (m *Output) Validate() error {
	if m.GetData() == "" {
		return errors.New("output data must not be empty")
	}

	return nil
}
