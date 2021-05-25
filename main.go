package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	githubOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	randomString = "random"
)

func main() {

	http.Handle("/", IsAuthenticated(handleHome))
	http.HandleFunc("/signup", handleSignup)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/callback", handleCallback)
	http.ListenAndServe(":8080", nil)
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	var html = `
	<html>
		<body>
			<a href="/login">
				Github Log In
			</a>
		</body>
	</html>
	`
	fmt.Fprint(w, html)
}

func handleHome(token string, w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "token "+token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("could not create get request: %s\n", err.Error())
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}
	defer res.Body.Close()
	// fmt.Println(res.Header)
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("could not parse response: %s\n", err.Error())
		http.Redirect(w, r, "/signup", http.StatusFound)
		return
	}

	fmt.Fprintf(w, "Response: %s", content)

	// var html = `
	// <html>
	// 	<body>
	// 		<h1>Hello %s</h1>
	// 	</body>
	// </html>
	// `
	// fmt.Fprintf(w, html)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL(randomString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != randomString {
		fmt.Println("State is not valid")
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := githubOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("could not get the token: %s\n", err.Error())
		http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		return
	}

	c := &http.Cookie{Name: "token",
		Value: token.AccessToken,
	}
	http.SetCookie(w, c)
	http.Redirect(w, r, "/", http.StatusFound)

}

func IsAuthenticated(endpoint func(token string, w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("token")
		if err != nil {
			fmt.Printf("could not read cookies \n")
			http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		}
		// fmt.Println("Cookies token", cookie.Value, err)
		if cookie != nil {
			endpoint(cookie.Value, w, r)
		} else {
			fmt.Printf("could not get the code from header \n")
			http.Redirect(w, r, "/signup", http.StatusTemporaryRedirect)
		}

	})
}
