package server

import "strconv"

type ErrClientAlreadyExists struct {
	Name string
}

func (e *ErrClientAlreadyExists) Error() string {
	return "A client with the name '" + e.Name + "' already exists"
}

type ErrClientNotExists struct {
	Name string
}

func (e *ErrClientNotExists) Error() string {
	return "A client with the name '" + e.Name + "' does not exist"
}

type ErrHookAlreadyExists struct {
	Identifier string
}

func (e *ErrHookAlreadyExists) Error() string {
	return "Hook with identifier '" + e.Identifier + "' already exists"
}

type ErrHookNotExists struct {
	Identifier string
}

func (e *ErrHookNotExists) Error() string {
	return "Hook '" + e.Identifier + "' not found"
}

type ErrInvalidClientName struct {
	Name string
}

func (e *ErrInvalidClientName) Error() string {
	return "Clientname '" + e.Name + "' contains '" + delimeter + "' and is invalid"
}

type ErrSecretGenerationFailed struct {
	Message string
}

func (e *ErrSecretGenerationFailed) Error() string {
	return "Could not generate client secret: " + e.Message
}

type ErrSecretTooShort struct {
	n int
}

func (e *ErrSecretTooShort) Error() string {
	return "Secret only has " + strconv.Itoa(e.n) + "byte instead of " + strconv.Itoa(secretByteLength)
}

type ErrCreatingUUID struct {
	Message string
}

func (e *ErrCreatingUUID) Error() string {
	return "Could not generate UUID: " + e.Message
}

type ErrCouldNotConnect struct {
	Message string
}

func (e *ErrCouldNotConnect) Error() string {
	return "Could not connect to server: " + e.Message
}

type ErrInvalidLocation struct {
	Message string
}

func (e *ErrInvalidLocation) Error() string {
	return "Invalid location: " + e.Message
}

type ErrUnknownServerError struct {
	Message string
}

func (e *ErrUnknownServerError) Error() string {
	return "Server responded with error: " + e.Message
}
