package reference

import "fmt"

type PublicationCheck func([]byte) error

type PublicationValidator struct {
	check PublicationCheck
}

func NewPublicationValidator(check PublicationCheck) *PublicationValidator {
	return &PublicationValidator{check: check}
}

func (v *PublicationValidator) Validate(data []byte) error {
	if err := v.check(data); err != nil {
		return fmt.Errorf("validate publication: %v", err)
	}
	return nil
}
