package common

import "errors"

// Errores globales reutilizables
var (
	ErrDuplicate       = errors.New("nombre duplicado")
	ErrNotFound        = errors.New("registro no encontrado")
	ErrHasDependencies = errors.New("tiene registros asociados")
	ErrInvalidID       = errors.New("ID inválido")
)

// Helpers para identificar errores
func IsDuplicateError(err error) bool        { return errors.Is(err, ErrDuplicate) }
func IsNotFoundError(err error) bool         { return errors.Is(err, ErrNotFound) }
func IsHasDependenciesError(err error) bool  { return errors.Is(err, ErrHasDependencies) }
func IsInvalidIDError(err error) bool        { return errors.Is(err, ErrInvalidID) }
