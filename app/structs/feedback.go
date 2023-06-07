package structs

import (
	"errors"
	"regexp"
)

const (
	nameMask    = `^[\p{L}\s'’]{2,50}$`
	mailMask    = `^[\w-\.]+@[\w-]+\.+[\w-]{2,20}$`
	badMask     = `@[a-zA-Z0-9.-]*\.ru$`
	commentMask = `^[\p{L}\s'’\(\)?!.,\-]{1,300}$`
)

var (
	nameRegex       = regexp.MustCompile(nameMask)
	mailRegex       = regexp.MustCompile(mailMask)
	commnentRegex   = regexp.MustCompile(commentMask)
	exlcludeRegex   = regexp.MustCompile(badMask)
	errWrongName    = errors.New("wrong name")
	errWrongEmail   = errors.New("wrong email")
	errWrongComment = errors.New("wrong comment")
)

type Feedback struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Comment string `json:"comment"`
	Token   string `json:"token"`
}

func (f Feedback) Valiadate() error {
	if !nameRegex.MatchString(f.Name) {
		return errWrongName
	}

	if !mailRegex.MatchString(f.Email) || exlcludeRegex.MatchString(f.Email) {
		return errWrongEmail
	}

	if !commnentRegex.MatchString(f.Comment) {
		return errWrongComment
	}

	return nil
}