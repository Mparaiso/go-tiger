//    Copyright (C) 2016  mparaiso <mparaiso@online.fr>
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//      http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package matcher_test

import (
	"fmt"
	r "github.com/mparaiso/go-tiger/matcher"
	"net/http"
	re "regexp"
	"testing"
)

func createRequest(url string) *http.Request {
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	return r
}

func BenchmarkRegexMatcher(b *testing.B) {
	matcher := r.Pattern("/document/:title/name/:name/id/:id/:*filepath", "")
	request := createRequest("https://example.com/document/something/name/someone/id/the%20id/documents/file.pdf")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.Match(request)
	}
}
func BenchmarkFastMatcher(b *testing.B) {
	matcher := r.NewFastMatcher("/document/:part1/name/:part2/id/:part3/:*filepath/", ":")
	request := createRequest("https://example.com/document/something/name/someone/id/the%20id/documents/file.pdf")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matcher.Match(request)
	}
}

func ExampleFastMatcher() {
	// match a route with 2 variables. The last / will be trimmed automatically
	matcher := r.NewFastMatcher("/documents/:foo/:bar/", ":")
	request := createRequest("https://example.com/documents/books/340340340")
	match := matcher.Match(request)
	fmt.Println(match, request.URL.Query().Get(":foo"), request.URL.Query().Get(":bar"))
	request = createRequest("https://example.com/documents/books/340340340/")
	match = matcher.Match(request)
	fmt.Println(match, request.URL.Query().Get(":foo"), request.URL.Query().Get(":bar"))
	request = createRequest("https://example.com/documents/books/340340340/something")
	match = matcher.Match(request)
	fmt.Println(match)
	matcher = r.NewFastMatcher("/public_assets/:*filepath", "@")
	request = createRequest("https://example.com/public_assets/css/stylesheets.css")
	match = matcher.Match(request)
	fmt.Println(match, request.URL.Query().Get("@filepath"))
	matcher = r.NewFastMatcher("/public_assets/:filepath", ":")
	request = createRequest("https://example.com/public_assets/css/stylesheets.css")
	match = matcher.Match(request)
	fmt.Println(match)

	matcher = r.NewFastMatcher("/user/:id/name/:name/", ":")
	request = createRequest("https://example.com/user/x2i/name/john%20doe")
	match = matcher.Match(request)
	fmt.Println(match, request.URL.Query().Get(":id"), request.URL.Query().Get(":name"))

	matcher = r.NewFastMatcher("/user/:path", ":")
	request = createRequest("https://example.com/user/some%2Fpath")
	match = matcher.Match(request)
	// fmt.Println(match, request.URL.Query().Get(":path"))

	// Output:
	// true books 340340340
	// true books 340340340
	// false
	// true css/stylesheets.css
	// false
	// true x2i john doe
}

func ExamplePattern() {
	matcher := r.Pattern("/:foo/:bar", "/root/")
	fmt.Println(matcher.Regexp.String())
	r, err := http.NewRequest("GET", "https://acme.com/root/users/22a39b6", nil)
	if err != nil {
		fmt.Println(err)
	} else {
		match := matcher.Match(r)
		fmt.Println(match, r.URL.Query().Get(":foo"), r.URL.Query().Get(":bar"))
	}
	// Output:
	// ^/root/(?P<foo>[^/]+)/(?P<bar>[^/]+)/?$
	// true users 22a39b6

}

func ExamplePattern_Second() {
	matcher := r.Pattern("/:foo/:*bar", "/root")
	fmt.Println(matcher.Regexp.String())

	r, err := http.NewRequest("GET", "http://example.com/root/static-assets/some/path/to/file/image.jpg", nil)
	if err != nil {
		fmt.Println(err)
	} else {
		isMatched := matcher.Match(r)
		fmt.Println(isMatched, r.URL.Query().Get(":foo"), r.URL.Query().Get(":bar"))
	}
	// Output:
	// ^/root/(?P<foo>[^/]+)/(?P<bar>.+)/?$
	// true static-assets some/path/to/file/image.jpg

}

func ExamplePattern_Third() {
	matcher := r.Pattern("/document/:title/name/:name/id/:id/:*filepath", "")
	fmt.Println(matcher.Regexp.String())
	request := createRequest("https://example.com/document/something/name/someone/id/the%20id/documents/file.pdf")
	match := matcher.Match(request)
	fmt.Println(match, request.URL.Query().Get(":title"), request.URL.Query().Get(":name"), request.URL.Query().Get(":id"), request.URL.Query().Get(":filepath"))
	// Output:
	// ^/document/(?P<title>[^/]+)/name/(?P<name>[^/]+)/id/(?P<id>[^/]+)/(?P<filepath>.+)/?$
	// true something someone the id documents/file.pdf
}

func ExampleRouter() {
	approuter := &r.RequestMatcher{
		[]r.MatcherProvider{
			r.DefaultMatcherProvider{
				r.Matchers{r.NewRegexMatcher(re.MustCompile(`^/$`))},
			},
			r.DefaultMatcherProvider{
				r.Matchers{
					r.NewMethodMatcher("PUT"),
					r.NewRegexMatcher(re.MustCompile(`^/resource/(?P<resource_id>[0-9]+)/?$`)),
				},
			},
		},
	}
	request, err := http.NewRequest("PUT", "http://some-url/resource/12", nil)
	if err != nil {
		fmt.Print(err)
		return
	}
	route := approuter.Match(request)
	if route == nil {
		fmt.Print("route is nil")
		return
	}
	fmt.Println("matchers in current matcher provider :", len(route.GetMatchers()))
	fmt.Println("resource id :", request.URL.Query().Get(":resource_id"))
	// Output:
	// matchers in current matcher provider : 2
	// resource id : 12
}
