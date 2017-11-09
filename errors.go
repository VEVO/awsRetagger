package main

// ErrSanityNoMapping is returned for the sanity checks in the Mapper
type ErrSanityNoMapping struct {
	message string
}

// NewErrSanityNoMapping generates a new ErrSanityNomapping
func NewErrSanityNoMapping(message string) *ErrSanityNoMapping {
	return &ErrSanityNoMapping{
		message: message,
	}
}

// Error just returns the error message, basic error interface implementation
func (e *ErrSanityNoMapping) Error() string {
	return e.message
}
