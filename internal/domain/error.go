package domain

import (
	"bytes"
)

const (
	ErrPrefix   ErrCode = "APP-"
	ErrInternal         = ErrPrefix + "0001"
	ErrNotFound         = ErrPrefix + "0002"
)

type ErrCode string
type ErrOp string
type Error struct {
	Code ErrCode `json:"code"`
	Op   ErrOp   `json:"op"`
	Msg  string  `json:"msg"`
	Err  error   `json:"err"`
}

func (e *Error) Error() string {
	var buf bytes.Buffer
	buf.WriteString(string(e.Op))
	if e.Code != "" {
		buf.WriteString("|")
		buf.WriteString(string(e.Code))
	}
	if e.Msg != "" {
		buf.WriteString("|")
		buf.WriteString(e.Msg)
	}
	if e.Err != nil {
		buf.WriteString(", ")
		buf.WriteString(e.Err.Error())
	}
	return buf.String()
}

func E(op ErrOp, args ...interface{}) error {
	e := &Error{Op: op}
	for _, arg := range args {
		switch arg := arg.(type) {
		case error:
			e.Err = arg
		case ErrCode:
			e.Code = arg
		case string:
			e.Msg = arg
		}
	}
	return e
}

func ECode(err error) ErrCode {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Err != nil {
		return ECode(e.Err)
	}
	return ErrInternal
}

func EOps(err error) []ErrOp {
	if e, ok := err.(*Error); ok {
		result := []ErrOp{e.Op}
		nextOps := EOps(e.Err)
		if nextOps != nil {
			return append(result, nextOps...)
		}
		return result
	}
	return nil
}

func EMsgs(err error) []string {
	if e, ok := err.(*Error); ok {
		var result []string
		msg := e.Msg
		if msg != "" {
			result = append(result, msg)
		}
		if e.Err != nil {
			next := EMsgs(e.Err)
			if next != nil {
				result = append(result, next...)
			}
		}
		return result
	}
	return []string{err.Error()}
}
