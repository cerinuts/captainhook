package server

import "strconv"

// ErrClientAlreadyExists occurs if someone tries to add a client with an existing name
type ErrClientAlreadyExists struct {
	Name string
}

func (e *ErrClientAlreadyExists) Error() string {
	return "A client with the name '" + e.Name + "' already exists"
}

// ErrClientNotExists occurs if someone tries to edit a client that does not exist
type ErrClientNotExists struct {
	Name string
}

func (e *ErrClientNotExists) Error() string {
	return "A client with the name '" + e.Name + "' does not exist"
}

// ErrHookAlreadyExists occurs if someone tries to add a hook with an existing identifier
type ErrHookAlreadyExists struct {
	Identifier string
}

func (e *ErrHookAlreadyExists) Error() string {
	return "Hook with identifier '" + e.Identifier + "' already exists"
}

// ErrHookNotExists occurs if someone tries to edit a hook that does not exist
type ErrHookNotExists struct {
	Identifier string
}

func (e *ErrHookNotExists) Error() string {
	return "Hook '" + e.Identifier + "' not found"
}

// ErrInvalidClientName occurs if someone tries to add a client with an invalid name
type ErrInvalidClientName struct {
	Name string
}

func (e *ErrInvalidClientName) Error() string {
	return "Clientname '" + e.Name + "' contains '" + delimeter + "' and is invalid"
}

// ErrSecretGenerationFailed occurs if something went wrong while generating the client secret
type ErrSecretGenerationFailed struct {
	Message string
}

func (e *ErrSecretGenerationFailed) Error() string {
	return "Could not generate client secret: " + e.Message
}

// ErrSecretTooShort occurs if a client sent a secret that is too short
type ErrSecretTooShort struct {
	n int
}

func (e *ErrSecretTooShort) Error() string {
	return "Secret only has " + strconv.Itoa(e.n) + "byte instead of " + strconv.Itoa(secretByteLength)
}

//ErrCreatingUUID occurs if something went wrong while generating the uuid
type ErrCreatingUUID struct {
	Message string
}

func (e *ErrCreatingUUID) Error() string {
	return "Could not generate UUID: " + e.Message
}

// ErrCouldNotConnect occurs if the client could not connect to the server
type ErrCouldNotConnect struct {
	Message string
}

func (e *ErrCouldNotConnect) Error() string {
	return "Could not connect to server: " + e.Message
}

// ErrInvalidLocation occurs if the redirect for a client is invalid
type ErrInvalidLocation struct {
	Message string
}

func (e *ErrInvalidLocation) Error() string {
	return "Invalid location: " + e.Message
}

// ErrUnknownServerError occurs on clients if any unknown error occurs
type ErrUnknownServerError struct {
	Message string
}

func (e *ErrUnknownServerError) Error() string {
	return "Server responded with error: " + e.Message
}
