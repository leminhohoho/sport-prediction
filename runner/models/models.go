package models

type ErrorHandler interface {
	Handle(error)
}
