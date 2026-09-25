// Copyright (c) 2026 Bart Venter <72999113+bartventer@users.noreply.github.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"net/url"

	"github.com/bartventer/httpcache/pkg/urlkey"
)

// URLKeyer describes the interface implemented by types that can generate a
// normalized cache key from a URL, following rules specified in RFC 3986 §6.
type URLKeyer interface {
	URLKey(u *url.URL) string
}

type URLKeyerFunc func(u *url.URL) string

func (f URLKeyerFunc) URLKey(u *url.URL) string {
	return f(u)
}

func NewURLKeyer() URLKeyer { return URLKeyerFunc(urlkey.Normalize) }
