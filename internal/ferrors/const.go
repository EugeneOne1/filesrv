package ferrors

// Str is a string [error].  It should usually be defined constant.
type Str string

func (e Str) Error() string { return string(e) }
