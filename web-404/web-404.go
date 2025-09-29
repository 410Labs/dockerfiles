package main

import (
	"net/http"
	"fmt"
)

var page = `
<html>
 <head></head>
 <body>
 <h1>Chuck Pro</h1>
 <p>Path not found.</p>
 </body>
</html>
`

func handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, page)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
