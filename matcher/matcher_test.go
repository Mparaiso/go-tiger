package matcher_test

import (
	"fmt"
	"net/http"
	re "regexp"

	r "github.com/mparaiso/tiger-go-framework/matcher"
)

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
	// ^/root/(?P<foo>[^\s /]+)/(?P<bar>[^\s /]+)/?$
	// true users 22a39b6

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
