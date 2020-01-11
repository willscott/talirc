package services

import (
	"reflect"

	"github.com/privacylab/talek/libtalek"
)

type participant struct {
	Handle *libtalek.Handle
	Nick   string

	selector *reflect.SelectCase
}
