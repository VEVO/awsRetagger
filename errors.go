package main

// ErrSanityNoMapping is returned for the sanity checks in the Mapper
type ErrSanityNoMapping struct {
	message  string
	TagName  string
	TagValue string
}

// NewErrSanityNoMapping generates a new ErrSanityNomapping
func NewErrSanityNoMapping(message, tagName, tagValue string) *ErrSanityNoMapping {
	return &ErrSanityNoMapping{
		message:  message,
		TagName:  tagName,
		TagValue: tagValue,
	}
}

// Error just returns the error message, basic error interface implementation
func (e *ErrSanityNoMapping) Error() string {
	return e.message
}

// ErrSanityConfig is returned for the sanity checks in the Mapper
type ErrSanityConfig struct {
	message string
	TagName string
}

// NewErrSanityConfig generates a new ErrSanityNomapping
func NewErrSanityConfig(message, tagName string) *ErrSanityConfig {
	return &ErrSanityConfig{
		message: message,
		TagName: tagName,
	}
}

// Error just returns the error message, basic error interface implementation
func (e *ErrSanityConfig) Error() string {
	return e.message
}
