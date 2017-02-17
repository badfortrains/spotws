package main

import (
	"encoding/json"
	"fmt"
	"github.com/badfortrains/spotcontrol"
	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
	"html/template"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

var controllerMap map[string]*spotcontrol.SpircController
var clientWsMap map[string]*rpc.Client

func rootHandler(w http.ResponseWriter, r *http.Request) {

	cookies := r.Cookies()
	var code string
	for _, c := range cookies {
		if c.Name == "token" {
			code = c.Name
		}
	}
	if _, ok := controllerMap[code]; ok {
		http.Redirect(w, r, "/control", 303)
		return
	}

	clientId := os.Getenv("client_id")
	urlPath := "https://accounts.spotify.com/authorize?" +
		"client_id=" + clientId +
		"&response_type=code" +
		"&redirect_uri=http://localhost:8081/callback" +
		"&scope=streaming"
	t, err := template.ParseFiles("redirect.html")
	if err != nil {
		fmt.Println("error parsing", err)
	}

	data := struct {
		LinkAdr string
	}{
		urlPath,
	}
	t.Execute(w, data)
	// if there is a token in the map, serve the app.
	// Otherwise, serve the static page with the link
}

func appHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("app.html")
	if err != nil {
		fmt.Println("error parsing", err)
	}

	data := struct {
	}{}
	t.Execute(w, data)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	clientId := os.Getenv("client_id")
	clientSecret := os.Getenv("client_secret")
	code := params.Get("code")

	token, err := spotcontrol.GetOauthAccessToken(code, "http://localhost:8081/callback", clientId, clientSecret)
	if err != nil {
		fmt.Fprintf(w, "Error getting token %q", err)
		return
	}
	controller, err := spotcontrol.LoginOauthToken(token.AccessToken, "spotcontrol", "./spotify_appkey.key")
	if err != nil {
		fmt.Fprintf(w, "Error logging in %q", err, token.AccessToken)
		return
	}
	controllerMap[code] = controller
	cookie := &http.Cookie{
		Name:   "token",
		Value:  code,
		MaxAge: 3600,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/control", 303)

}

type notify struct {
	Method string
	Params []string
}

func createNotify(jsonUpdate string) ([]byte, error) {
	update := notify{
		Method: "notify",
		Params: []string{jsonUpdate},
	}
	return json.Marshal(update)
}

func serve(ws *websocket.Conn) {
	token := ws.Config().Protocol[0]
	controller, ok := controllerMap[token]
	if !ok {
		ws.Close()
	}
	clientWs := jsonrpc.NewClient(ws)
	clientWsMap[token] = clientWs
	controller.HandleUpdatesCb(func(jsonUpdate string) {
		data, err := createNotify(jsonUpdate)
		if err != nil {
			fmt.Println("error:", err)
		}
		_, err = ws.Write(data)
		if err != nil {
			fmt.Println("error:", err)
		}
	})

	jsonrpc.ServeConn(ws)
}

func main() {
	controllerMap = make(map[string]*spotcontrol.SpircController)
	clientWsMap = make(map[string]*rpc.Client)
	client := &Client{
		controllerMap: &controllerMap,
		clientWsMap:   &clientWsMap,
	}
	rpc.Register(client)
	r := mux.NewRouter()
	s := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.PathPrefix("/static/").Handler(s)
	r.HandleFunc("/callback", callbackHandler)
	r.HandleFunc("/control", appHandler)
	r.HandleFunc("/", rootHandler)
	http.Handle("/conn", websocket.Handler(serve))
	http.Handle("/", r)
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println(err)
	}
}
