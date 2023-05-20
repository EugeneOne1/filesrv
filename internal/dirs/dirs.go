package dirs

import (
	"log"
	"net/http"
)

type Dirs struct {
	http.Handler
}

func (h *Dirs) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		u     = r.URL.Path
		isDir = u[len(u)-1] == '/'
	)
	h.Handler.ServeHTTP(w, r)
	if isDir {
		log.Println("isDir:", u)
		// io.WriteString(w, link)
	}
}
