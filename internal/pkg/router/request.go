package router

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/shandysiswandi/gobite/internal/pkg/goerror"
)

// Request wraps http.Request with helpers for inbound handlers.
type Request struct {
	// Request is the underlying http.Request.
	*http.Request
}

// GetParam reads a path parameter from the request context (as stored by httprouter).
func (r *Request) GetParam(key string) string {
	return httprouter.ParamsFromContext(r.Context()).ByName(key)
}

func (r *Request) GetParamInt64(key string) (int64, error) {
	paramValue := r.GetParam(key)
	value, err := strconv.ParseInt(paramValue, 10, 64)
	if err != nil {
		return 0, goerror.NewInvalidFormat("param must integer value")
	}
	return value, nil
}

func (r *Request) GetQuery(key string) string {
	return strings.TrimSpace(r.URL.Query().Get(key))
}

func (r *Request) GetQueries(key string) []string {
	return r.URL.Query()[key]
}

func (r *Request) GetQueryInt32(key string) (int32, error) {
	queryValue := r.GetQuery(key)
	if queryValue == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(queryValue, 10, 32)
	if err != nil {
		return 0, goerror.NewInvalidFormat()
	}

	return int32(value), nil
}

func (r *Request) GetQueryInt16(key string) (int16, error) {
	queryValue := r.GetQuery(key)
	if queryValue == "" {
		return 0, nil
	}

	value, err := strconv.ParseInt(queryValue, 10, 16)
	if err != nil {
		return 0, goerror.NewInvalidFormat()
	}

	return int16(value), nil
}

func (r *Request) GetQueryDate(key, format string) (time.Time, error) {
	queryValue := r.GetQuery(key)
	if queryValue == "" {
		return time.Time{}, nil
	}

	value, err := time.Parse(format, queryValue)
	if err != nil {
		return time.Time{}, goerror.NewInvalidFormat("Invalid query " + key)
	}

	return value, nil
}

// DecodeBody decodes the JSON body into dst.
func (r *Request) DecodeBody(dst any) error {
	if r == nil || r.Body == nil {
		return goerror.NewInvalidFormat()
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return goerror.NewInvalidFormat()
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return goerror.NewInvalidFormat()
	}

	return nil
}

// StreamSingleFile returns the first multipart file matching the form field name.
func (r *Request) StreamSingleFile(name string) (io.ReadCloser, error) {
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		return nil, goerror.NewInvalidFormat("Invalid request content-type")
	}

	mr, err := r.MultipartReader()
	if err != nil {
		return nil, goerror.NewInvalidFormat()
	}

	var file io.ReadCloser
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, goerror.NewInvalidFormat()
		}

		if part.FormName() == name {
			file = part
			break
		}

		if _, errCopy := io.Copy(io.Discard, part); errCopy != nil {
			if err := part.Close(); err != nil {
				return nil, goerror.NewInvalidFormat(err.Error())
			}
			return nil, goerror.NewInvalidFormat(errCopy.Error())
		}
		if err := part.Close(); err != nil {
			return nil, goerror.NewInvalidFormat(err.Error())
		}
	}

	if file == nil {
		return nil, goerror.NewInvalidFormat()
	}

	return file, nil
}
