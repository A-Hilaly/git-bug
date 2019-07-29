package core

import "fmt"

type ImportEvent int

const (
	_ ImportEvent = iota
	ImportEventBug
	ImportEventComment
	ImportEventCommentEdition
	ImportEventStatusChange
	ImportEventTitleEdition
	ImportEventLabelChange
	ImportEventIdentity
	ImportEventNothing
)

// ImportResult is an event that is emitted during the import process, to
// allow calling code to report on what is happening, collect metrics or
// display meaningful errors if something went wrong.
type ImportResult struct {
	Err    error
	Event  ImportEvent
	ID     string
	Reason string
}

func (er ImportResult) String() string {
	switch er.Event {
	case ImportEventBug:
		return "new issue"
	case ImportEventComment:
		return "new comment"
	case ImportEventCommentEdition:
		return "updated comment"
	case ImportEventStatusChange:
		return "changed status"
	case ImportEventTitleEdition:
		return "changed title"
	case ImportEventLabelChange:
		return "changed label"
	case ImportEventIdentity:
		return "new identity"
	case ImportEventNothing:
		return fmt.Sprintf("no event: %v", er.Reason)
	default:
		panic("unknown import result")
	}
}

func NewImportError(err error, reason string) ImportResult {
	return ImportResult{
		Err:    err,
		Reason: reason,
	}
}

func NewImportNothing(id string, reason string) ImportResult {
	return ImportResult{
		ID:     id,
		Reason: reason,
		Event:  ImportEventNothing,
	}
}

func NewImportBug(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventBug,
	}
}

func NewImportComment(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventComment,
	}
}

func NewImportCommentEdition(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventCommentEdition,
	}
}

func NewImportStatusChange(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventStatusChange,
	}
}

func NewImportLabelChange(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventLabelChange,
	}
}

func NewImportTitleEdition(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventTitleEdition,
	}
}

func NewImportIdentity(id string) ImportResult {
	return ImportResult{
		ID:    id,
		Event: ImportEventIdentity,
	}
}
