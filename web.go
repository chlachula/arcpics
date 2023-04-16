package arcpics

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

var webCounter int
var mutex = &sync.Mutex{}

func homeString(w http.ResponseWriter, r *http.Request) {
	htmlPage := `<html><title>home</title>
<body style="text-align:center">home <hr/>
try <a href="/increment">increment</a> page! <br/>or<br/>
<a href="/hello">Hello</a>.
</body></html>`
	fmt.Fprintf(w, "%s", htmlPage)
}

func incrementCounter(w http.ResponseWriter, r *http.Request) {
	htmlPage := `<html><title>home</title>
<body style="text-align:center"> <a href="/">home</a><hr/>
<a href="/increment">increment</a> webCounter: %d
</body></html>`
	mutex.Lock()
	webCounter++
	fmt.Fprintf(w, htmlPage, webCounter)
	mutex.Unlock()
}

func Web(port int) {
	http.HandleFunc("/", homeString)

	http.HandleFunc("/increment", incrementCounter)

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		hello := `<html><title>hello</title>
		<body style="text-align:center"><a href="/">home</a><hr/>
		<h1>Hello World!</h1>
		</body></html>`
		fmt.Fprintf(w, "%s", hello)
	})
	colonPort := fmt.Sprintf(":%d", port)
	fmt.Printf("... listening at port %d", port)
	log.Fatal(http.ListenAndServe(colonPort, nil))

}
