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
	code ErrCode `json:"code"`
	op   ErrOp   `json:"op"`
	msg  string  `json:"msg"`
	err  error   `json:"err"`
}

func (e *Error) Error() string {
	var buf bytes.Buffer
	buf.WriteString(string(e.op))
	if e.code != "" {
		buf.WriteString("|")
		buf.WriteString(string(e.code))
	}
	if e.msg != "" {
		buf.WriteString("|")
		buf.WriteString(e.msg)
	}
	if e.err != nil {
		buf.WriteString(", ")
		buf.WriteString(e.err.Error())
	}
	return buf.String()
}

func E(op ErrOp, args ...interface{}) error {
	e := &Error{op: op}
	for _, arg := range args {
		switch arg := arg.(type) {
		case error:
			e.err = arg
		case ErrCode:
			e.code = arg
		case string:
			e.msg = arg
		}
	}
	return e
}

func ECode(err error) ErrCode {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.code != "" {
		return e.code
	} else if ok && e.err != nil {
		return ECode(e.err)
	}
	return ErrInternal
}

func EOps(err error) []ErrOp {
	if e, ok := err.(*Error); ok {
		result := []ErrOp{e.op}
		nextOps := EOps(e.err)
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
		msg := e.msg
		if msg != "" {
			result = append(result, msg)
		}
		if e.err != nil {
			next := EMsgs(e.err)
			if next != nil {
				result = append(result, next...)
			}
		}
		return result
	}
	return []string{err.Error()}
}
