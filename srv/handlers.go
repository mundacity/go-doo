package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	godoo "github.com/mundacity/go-doo"
	lg "github.com/mundacity/quick-logger"
)

type Handler struct {
	Repo godoo.IRepository
}

func (h Handler) TestHandler(w http.ResponseWriter, r *http.Request) {
	lg.Logger.Log(lg.Info, "test handler called")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h Handler) HandleRequests(w http.ResponseWriter, r *http.Request) {

	lg.Logger.Logf(lg.Info, "%v request received from %v", r.Method, r.RemoteAddr)

	switch r.Method {
	case http.MethodGet:
		h.GetHandler(w, r)
	case http.MethodPut:
		h.EditHandler(w, r)
	case http.MethodPost:
		h.AddHandler(w, r)
	default:
		lg.Logger.LogWithCallerInfo(lg.Error, "method not allowed", runtime.Caller)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h Handler) AddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var td godoo.TodoItem
	d := json.NewDecoder(r.Body)
	err := d.Decode(&td)

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	i, err := h.Repo.Add(&td)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
	lg.Logger.Log(lg.Info, "add handler completed execution")
}

func (h Handler) GetHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var fq godoo.FullUserQuery
	d := json.NewDecoder(r.Body)
	err := d.Decode(&fq)

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itms, err := h.Repo.GetWhere(fq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(itms)
	lg.Logger.Log(lg.Info, "get handler completed execution")
}

func (h Handler) EditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var fq []godoo.FullUserQuery
	d := json.NewDecoder(r.Body)
	err := d.Decode(&fq)

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(fq) != 2 {
		msg := "operation forbidden; two FullUserQuery structs required"
		lg.Logger.LogWithCallerInfo(lg.Error, msg, runtime.Caller)
		http.Error(w, msg, http.StatusForbidden)
		return
	}

	i, err := h.Repo.UpdateWhere(fq[0], fq[1])
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
	lg.Logger.Log(lg.Info, "edit handler completed execution")
}
