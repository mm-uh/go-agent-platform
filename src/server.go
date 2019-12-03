package core

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"
)

type ServerHttp struct {
	ip     string
	port   int
	router mux.Router
	prefix string
	pl     Platform
}

type Params struct {
	Name string `json:"Name"`
}

// [GET] Get all agents that are some
func (server ServerHttp) HandleGetSimilarAgents(w http.ResponseWriter, r *http.Request) {
	go func() {
		// If the Content-Type header is present, check that it has the value
		// application/json.
		if r.Header.Get("Content-Type") != "" {
			value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
			if value != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				return
			}
		}

		// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
		// response body.
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		// Setup the decoder and call the DisallowUnknownFields() method on it.
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var getAgent Params
		err := dec.Decode(&getAgent)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check that the request body only contained a single JSON object.
		if dec.More() {
			msg := "Request body must only contain a single JSON object"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		agent, err := server.pl.LocateAgent(getAgent.Name)
		if err != nil {
			msg := "Couldn't locate the agent"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		response, err := json.Marshal(agent)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(response)
		if err != nil {
			msg := "Couldn't marshal response"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}()
}

// [GET] Get Agent That Match some criteria
// Should return at most one Agent
func (server ServerHttp) HandleGetAgent(w http.ResponseWriter, r *http.Request) {
	go func() {
		// If the Content-Type header is present, check that it has the value
		// application/json.
		if r.Header.Get("Content-Type") != "" {
			value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
			if value != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				return
			}
		}

		// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
		// response body.
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		// Setup the decoder and call the DisallowUnknownFields() method on it.
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var getAgent Params
		err := dec.Decode(&getAgent)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check that the request body only contained a single JSON object.
		if dec.More() {
			msg := "Request body must only contain a single JSON object"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		agent, err := server.pl.LocateAgent(getAgent.Name)
		if err != nil {
			msg := "Couldn't locate the agent"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		response, err := json.Marshal(agent)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(response)
		if err != nil {
			msg := "Couldn't marshal response"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}()
}

// [GET] Get Agents currently working for some function name
// Should return a list of Agents
func (server ServerHttp) HandleGetAgentsFunctions(w http.ResponseWriter, r *http.Request) {
	go func() {
		// If the Content-Type header is present, check that it has the value
		// application/json.
		if r.Header.Get("Content-Type") != "" {
			value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
			if value != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				return
			}
		}

		// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
		// response body.
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		// Setup the decoder and call the DisallowUnknownFields() method on it.
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var params Params
		err := dec.Decode(&params)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check that the request body only contained a single JSON object.
		if dec.More() {
			msg := "Request body must only contain a single JSON object"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		agents, err := server.pl.GetAgentsByFunctions(params.Name)
		if err != nil {
			msg := "Couldn't locate the agent"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		response, err := json.Marshal(agents)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(response)
		if err != nil {
			msg := "Couldn't marshal response"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}()
}

// [GET] Get all agents registered in the platform
func (server ServerHttp) HandleAgentsNames(w http.ResponseWriter, r *http.Request) {
	go func() {

		agentsNames, err := server.pl.GetAllAgentsNames()
		if err != nil {
			msg := "Couldn't get all agents names"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(agentsNames)
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(response)
		if err != nil {
			msg := "Couldn't marshal response"
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}
	}()

}

// [POST] Register Agent
func (server ServerHttp) HandleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	go func() {
		// If the Content-Type header is present, check that it has the value
		// application/json.
		if r.Header.Get("Content-Type") != "" {
			value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
			if value != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				return
			}
		}

		// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
		// response body.
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		// Setup the decoder and call the DisallowUnknownFields() method on it.
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var agent Agent
		err := dec.Decode(&agent)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// Check that the request body only contained a single JSON object.
		if dec.More() {
			msg := "Request body must only contain a single JSON object"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		err = server.pl.Register(agent)
		if err != nil {
			msg := "Request body must only contain a single JSON object"
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

	}()
}

func NewServer(prefix string, platform Platform, addr Addr) *ServerHttp {
	server := &ServerHttp{
		ip:     addr.ip,
		port:   addr.port,
		router: *mux.NewRouter(),
		prefix: prefix,
		pl:     platform,
	}
	if len(prefix) < 1 {
		prefix = "/api/v1"
	}
	api := server.router.PathPrefix("prefix").Subrouter()
	api.HandleFunc("/getAllAgents", server.HandleGetAgent).Methods(http.MethodGet)
	api.HandleFunc("/registerAgent", server.HandleRegisterAgent).Methods(http.MethodPost)
	api.HandleFunc("/getAllAgentsNames", server.HandleAgentsNames).Methods(http.MethodPost)
	api.HandleFunc("/getAgentsForFunction", server.HandleGetAgentsFunctions).Methods(http.MethodPost)
	api.HandleFunc("/getSimilarAgents", server.HandleGetSimilarAgents).Methods(http.MethodPost)

	return server
}

func (server *ServerHttp) RunServer() {
	log.Fatal(http.ListenAndServe(server.ip+":"+strconv.FormatInt(int64(server.port), 10), &server.router))
}
