// Package api holds the business-logic layer for the generic, curl-like
// passthrough to the Bitrise API exposed by the `api` command.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/bitrise-io/bitrise/v2/internal/bitriseapi"
)

// KeyValue is one --field key=value pair.
type KeyValue struct {
	Key   string
	Value string
}

// Request describes one `api` command invocation.
type Request struct {
	Method   string
	Path     string
	Fields   []KeyValue
	Headers  http.Header
	Body     io.Reader
	Paginate bool
}

// Response is the result of a Request, ready for the cmd layer to print.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

// Service exposes the raw API passthrough to the cmd layer.
type Service struct {
	client *bitriseapi.Client
}

// NewService returns a Service backed by the given API client.
func NewService(client *bitriseapi.Client) *Service {
	return &Service{client: client}
}

// Do executes req. Fields become a query string for GET requests, or a JSON
// body for anything else. Every non-GET body — whether built from Fields or
// passed through from --input — gets a JSON Content-Type unless the caller set
// one, since the API won't parse a body sent without it. When Paginate is set,
// the request must be a GET; pages are followed via the "paging.next" cursor
// and their "data" arrays merged into one response.
func (s *Service) Do(ctx context.Context, req Request) (Response, error) {
	header := cloneHeader(req.Headers)
	isGet := req.Method == http.MethodGet

	var query url.Values
	body := req.Body

	if len(req.Fields) > 0 {
		if isGet {
			query = fieldsToQuery(req.Fields)
		} else {
			data, err := json.Marshal(fieldsToMap(req.Fields))
			if err != nil {
				return Response{}, fmt.Errorf("encode fields: %w", err)
			}
			body = bytes.NewReader(data)
		}
	}
	if body != nil && !isGet && header.Get("Content-Type") == "" {
		header.Set("Content-Type", "application/json")
	}

	if req.Paginate {
		if !isGet {
			return Response{}, fmt.Errorf("--all is only supported for GET requests")
		}
		if body != nil {
			return Response{}, fmt.Errorf("--all cannot be combined with --input")
		}
		return s.paginate(ctx, req.Method, req.Path, query, header)
	}

	raw, err := s.client.RawRequest(ctx, req.Method, req.Path, query, header, body)
	if err != nil {
		return Response{}, err
	}
	return asResponse(raw), nil
}

// pagedBody is the minimal shape of a Bitrise cursor-paginated list response.
type pagedBody struct {
	Data   json.RawMessage `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

// paginate follows the "paging.next" cursor, merging every page's "data"
// array into one {"data": [...]} envelope. A page that isn't a paginated
// list (no "data" array) or is non-2xx stops the walk, and that page's raw
// response is returned unchanged rather than the merge — the caller can't
// tell the difference between "no more data" and "this wasn't a list" from
// a merged envelope alone.
//
// A cursor that repeats means the walk isn't advancing, so it stops with an
// error instead of looping forever: the API may be echoing the cursor back, or
// PATH already carried a "next" query parameter, which is merged alongside the
// one set here and takes precedence server-side.
//
// The merged response carries the last page's status and headers, so --include
// reports what the API actually sent (rate limits, request IDs) rather than a
// synthesized status line — minus the headers that describe that one page's
// body, which the merged body would contradict.
func (s *Service) paginate(ctx context.Context, method, path string, query url.Values, header http.Header) (Response, error) {
	merged := []json.RawMessage{}
	seen := map[string]bool{}
	nextQuery := cloneValues(query)
	var lastPage *bitriseapi.RawResponse

	for {
		raw, err := s.client.RawRequest(ctx, method, path, nextQuery, header, nil)
		if err != nil {
			return Response{}, err
		}
		lastPage = raw

		if raw.StatusCode < 200 || raw.StatusCode >= 300 {
			return asResponse(raw), nil
		}

		var page pagedBody
		if err := json.Unmarshal(raw.Body, &page); err != nil || page.Data == nil {
			return asResponse(raw), nil
		}
		var items []json.RawMessage
		if err := json.Unmarshal(page.Data, &items); err != nil {
			return asResponse(raw), nil
		}
		merged = append(merged, items...)

		if page.Paging.Next == "" {
			break
		}
		if seen[page.Paging.Next] {
			return Response{}, fmt.Errorf("pagination stalled: the API returned the cursor %q twice", page.Paging.Next)
		}
		seen[page.Paging.Next] = true

		nextQuery = cloneValues(query)
		nextQuery.Set("next", page.Paging.Next)
	}

	body, err := json.Marshal(map[string]any{"data": merged})
	if err != nil {
		return Response{}, fmt.Errorf("encode paginated response: %w", err)
	}
	mergedHeader := cloneHeader(lastPage.Header)
	for _, name := range []string{"Content-Length", "Content-Encoding", "Etag"} {
		mergedHeader.Del(name)
	}
	return Response{StatusCode: lastPage.StatusCode, Header: mergedHeader, Body: body}, nil
}

func fieldsToQuery(fields []KeyValue) url.Values {
	q := url.Values{}
	for _, kv := range fields {
		q.Add(kv.Key, kv.Value)
	}
	return q
}

func fieldsToMap(fields []KeyValue) map[string]string {
	m := make(map[string]string, len(fields))
	for _, kv := range fields {
		m[kv.Key] = kv.Value
	}
	return m
}

func cloneHeader(h http.Header) http.Header {
	if h == nil {
		return http.Header{}
	}
	return h.Clone()
}

func cloneValues(v url.Values) url.Values {
	if v == nil {
		return url.Values{}
	}
	out := url.Values{}
	for k, vs := range v {
		out[k] = append([]string(nil), vs...)
	}
	return out
}

func asResponse(raw *bitriseapi.RawResponse) Response {
	return Response{StatusCode: raw.StatusCode, Header: raw.Header, Body: raw.Body}
}
