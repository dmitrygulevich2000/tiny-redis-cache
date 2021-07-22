package server

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/storage"
	"github.com/dmitrygulevich2000/tiny-redis-cache/api"
	
	"encoding/json"
	"net/http"
)

type CacheServer struct {
	Data storage.Storage
	Mux *http.ServeMux
}

func New() *CacheServer {
	srv := &CacheServer{
		Data: storage.New(0),
		Mux: http.NewServeMux(),
	}
	srv.Mux.HandleFunc("/set", srv.HandleSet)
	srv.Mux.HandleFunc("/get", srv.HandleGet)
	srv.Mux.HandleFunc("/del", srv.HandleDel)
	srv.Mux.HandleFunc("/keys", srv.HandleKeys)

	return srv
}

func (srv *CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.Mux.ServeHTTP(w, r)
}

func (srv *CacheServer) HandleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}
	
	enc := json.NewDecoder(r.Body)
	params := new(api.SetParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := api.ValidateSetParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := api.ErrorResponse{"SET", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}

	srv.Data.Set(params.Key, params.Value, params.Ttl)
	w.Write([]byte(`"OK"`))
}

func (srv *CacheServer) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewDecoder(r.Body)
	params := new(api.GetParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := api.ValidateGetParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := api.ErrorResponse{"GET", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}
	
	val, exists := srv.Data.Get(params.Key)
	if !exists {
		w.Write([]byte("null"))
		return
	}

	resp, err := json.Marshal(val)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (srv *CacheServer) HandleDel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewDecoder(r.Body)
	params := new(api.DelParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := api.ValidateDelParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := api.ErrorResponse{"DEL", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}
	
	deleted := srv.Data.Delete(params.Keys...)

	resp, err := json.Marshal(deleted)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

func (srv *CacheServer) HandleKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewDecoder(r.Body)
	params := new(api.KeysParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := api.ValidateKeysParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := api.ErrorResponse{"KEYS", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}
	
	val, err := srv.Data.Keys(params.Pattern)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(val)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}