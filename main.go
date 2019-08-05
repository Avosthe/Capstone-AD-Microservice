package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/mux"
)

//Parameters needed to create a new route
type route struct {
	method      string
	path        string
	handlerFunc http.HandlerFunc
}

//Used to declare a list of desired routes
type routeList []route

var routes = routeList{
	route{
		"GET",
		"/command",
		commandHandler,
	},
	route{
		"GET",
		"/test",
		func(writer http.ResponseWriter, req *http.Request) {
		},
	},
}

var microServiceIP = os.Getenv("microServiceIP")                         // used for authentication
var microServiceSecretPassword = os.Getenv("microServiceSecretPassword") // used for authentication

func commandHandler(writer http.ResponseWriter, req *http.Request) {

	// Authentication
	remoteIPAddress := strings.Split(req.RemoteAddr, ":")[0]
	queryParams := req.URL.Query()
	secretKey := queryParams.Get("secretKey")
	command := queryParams.Get("command")

	fmt.Println(microServiceIP)
	fmt.Println(microServiceSecretPassword)

	if secretKey != microServiceSecretPassword || remoteIPAddress != microServiceIP {
		writer.WriteHeader(403) // Return HTTP 403 Forbidden
		writer.Write([]byte("Access Forbidden."))
		return
	}

	if command == "" {
		writer.WriteHeader(400) // Return HTTP 400 Bad Request
		writer.Write([]byte("Bad Request."))
		return
	}

	// Switch on command query
	var cmd *exec.Cmd
	switch command {
	case "remote_shutdown":
		targetIPAddress := queryParams.Get("targetIPAddress")
		waitSeconds := queryParams.Get("waitSeconds")
		message := queryParams.Get("message")
		cmdName := "shutdown"
		cmd = exec.Command(cmdName, "/s", "/m", "\\\\"+targetIPAddress, "/t", waitSeconds, "/c", message)
		break
	default:
		break
	}

	// Execute Command
	output, err := cmd.Output()
	if err != nil {
		errStr := fmt.Sprintf("Failed to execute command: %s\n", err.Error())
		writer.Write([]byte(errStr))
	} else {
		success := fmt.Sprintf("Successfully executed command:\n%s\n", output)
		writer.Write([]byte(success))
	}

}

//NewRouter creates a mux router will all routes specified in routes
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.HandleFunc(route.path, route.handlerFunc).Methods(route.method)
	}
	return router
}

func main() {
	router := NewRouter()
	fmt.Println("Server listening on :5020")
	log.Fatal(http.ListenAndServe(":5020", router))
}
