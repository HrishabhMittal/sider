package main

import (
	"bufio"
	"fmt"
	"strconv"
)

type DecodingError struct {
	message string
}

func NewDecodingError(message string) *DecodingError {
	return &DecodingError{
		message,
	}
}
func (e *DecodingError) Error() string {
	return fmt.Sprintf("decoding error: %s", e.message)
}

const (
	SIMPLE_STR      = '+'
	SIMPLE_ERR      = '-'
	INTEGER         = ':'
	BULK_STRING     = '$'
	ARRAY           = '*'
	NULL            = '_'
	BOOLEAN         = '#'
	DOUBLE          = ','
	BIGNUM          = '('
	BULK_ERROR      = '!'
	VERBATIM_STRING = '='
	MAP             = '%'
	ATTRIBUTE       = '|'
	SET             = '~'
	PUSH            = '>'
)

type RespReader struct {
	*bufio.Reader
}

func NewRespReader(Reader *bufio.Reader) RespReader {
	return RespReader{
		Reader,
	}
}

type RespWriter struct {
	*bufio.Writer
}

func NewRespWriter(Writer *bufio.Writer) RespWriter {
	return RespWriter{
		Writer,
	}
}
func (r *RespReader) decodeLine() (string, error) {
	line, err := r.Reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 3 {
		return "", NewDecodingError("string not long enough")
	}
	end := line[len(line)-2:]
	if end != "\r\n" {
		return "", NewDecodingError("string doesnt end in \\r\\n")
	}
	return line[:len(line)-2], nil
}
func (r *RespReader) decodeObj() (any, error) {
	line, err := r.decodeLine()
	if err != nil {
		return nil, err
	}
	switch line[0] {
	case SIMPLE_STR:
		return line[1:], nil
	case ARRAY:
		num, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		return r.decodeArray(num)
	case BULK_STRING:
		num, err := strconv.Atoi(line[1:])
		if err != nil {
			return nil, err
		}
		return r.decodeBulkString(num)
	}
	return nil, nil
}

func (r *RespReader) decodeBulkString(length int) (string, error) {
	line, err := r.decodeLine()
	if err != nil {
		return "", err
	}
	if len(line) == length {
		return line, nil
	}
	return "", NewDecodingError("bulk string length doesn't match")
}
func (r *RespReader) decodeArray(length int) ([]any, error) {
	ret := make([]any, 0, length)
	for range length {
		obj, err := r.decodeObj()
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}
	return ret, nil
}
