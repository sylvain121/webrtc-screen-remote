package lib

import (
	"encoding/json"
	"fmt"
	"github.com/pion/webrtc/v3"
	"net/http"
	"remote_linux/lib/Video"
	"remote_linux/lib/WebRTC"
)

func handleError(w http.ResponseWriter, err error) {
	fmt.Printf("Error: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
}

func Init(x int, y int, width int, height int) {
	fmt.Println("init web server")
	server := http.NewServeMux()
	fmt.Println("Handle /api")
	server.Handle("/api/", http.StripPrefix("/api", apiHandler(x, y, width, height)))
	fmt.Println("Handle /static")
	server.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./web"))))
	server.HandleFunc("*", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "./web/index.html")
	})

	fmt.Println("starting web server on port 8080")
	http.ListenAndServe(":8080", server)
}

func apiHandler(x int, y int, width int, height int) http.Handler {
	endpoint := http.NewServeMux()
	endpoint.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handle connection on /api/connect")
		if r.Method != http.MethodPost {
			fmt.Println("not a POST method")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		offer := webrtc.SessionDescription{}
		err := json.NewDecoder(r.Body).Decode(&offer)

		if err != nil {
			fmt.Println("%s", err.Error())
			panic(err)
		}

		webrtcEndpoint := WebRTC.WebRtcEndpoint{}
		w.Header().Set("Content-Type", "application/json")
		answer, connectionContext := webrtcEndpoint.NewConnection(offer)
		json.NewEncoder(w).Encode(answer)

		go Video.Stream(&webrtcEndpoint, connectionContext, x, y, width, height)
	})

	return endpoint
}
