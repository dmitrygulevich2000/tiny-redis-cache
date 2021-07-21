package api

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/storage"
	
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

type Api struct {
	Data *storage.KVStorage
}

func New() *Api {
	return &Api{
		Data: storage.New(0),
	}
}

type Error struct {
	Op string
	Err string
}

func (e *Error) Error() string {
	return e.Op + ": " + e.Err
}


type SetParams struct {
	Key string
	Value interface{}
	Ttl time.Duration
}

func ValidateSetParams(p *SetParams) error {
	if p.Key == "" {
		return errors.New("key argument must be specified")
	}
	if p.Value == nil {
		return errors.New("value argument must be specified")
	}
	if p.Ttl < 0 {
		return errors.New("ttl argument must be nonegative")
	}
	return nil
}

func (api *Api) HandleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}
	
	enc := json.NewDecoder(r.Body)
	params := new(SetParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := ValidateSetParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := Error{"SET", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}

	api.Data.Set(params.Key, params.Value, params.Ttl)
	w.Write([]byte(`"OK"`))
}


type GetParams struct {
	Key string
}

func ValidateGetParams(p *GetParams) error {
	if p.Key == "" {
		return errors.New("key argument must be specified")
	}
	return nil
}

func (api *Api) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewDecoder(r.Body)
	params := new(GetParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := ValidateGetParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := Error{"GET", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}
	
	val, exists := api.Data.Get(params.Key)
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


type DelParams struct {
	Keys []string
}

func ValidateDelParams(p *DelParams) error {
	if p.Keys == nil || len(p.Keys) == 0 {
		return errors.New("at least one key must be in keys argument")
	}
	return nil
}

func (api *Api) HandleDel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Use POST method to access api", http.StatusMethodNotAllowed)
		return
	}

	enc := json.NewDecoder(r.Body)
	params := new(DelParams)
	errJSON := enc.Decode(params)
	errString := ""
	if errJSON != nil {
		errString = errJSON.Error()
	}
	if err := ValidateDelParams(params); err != nil {
		errString = err.Error()
	}
	if errString != "" {
		w.WriteHeader(http.StatusBadRequest)
		err := Error{"DEL", errString}
		resp, _ := json.Marshal(err)
		w.Write(resp)
		return
	}
	
	deleted := api.Data.Delete(params.Keys...)

	resp, err := json.Marshal(deleted)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}