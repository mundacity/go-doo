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
	Repo         godoo.IRepository
	PriorityList godoo.PriorityList
}

// Returns a new http handler. If runPl is true, then the handler will
// maintain a priority queue as well.
func NewHandler(runPl bool, repo godoo.IRepository) Handler {

	h := Handler{Repo: repo}
	if runPl {
		h.setupPriorityList()
	}

	return h
}

func (h Handler) setupPriorityList() { // todo look at go routines for this (maybe just the callers)
	h.PriorityList = *godoo.NewPriorityList()
	all, _ := h.Repo.GetAll()

	for _, v := range all {
		h.PriorityList.Add(v)
	}
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

	if err = h.PriorityList.Add(td); err != nil {
		// db fine but pl out of sync --> reset pl
		h.setupPriorityList()
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
	lg.Logger.Log(lg.Info, "add handler completed execution")
}

func (h Handler) GetHandler(w http.ResponseWriter, r *http.Request) {

	var itms []godoo.TodoItem
	var err error

	w.Header().Set("content-type", "application/json")

	// parse user query
	var fq godoo.FullUserQuery
	d := json.NewDecoder(r.Body)
	err = d.Decode(&fq)

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//check priority queue if necessary
	if len(fq.QueryOptions) == 1 && fq.QueryOptions[0].Elem == godoo.ByNextPriority {
		itm, err := h.runGetNextByPriority(fq, true)
		if err != nil { //already logged by this point
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		itms = append(itms, itm)
	}
	//check priority queue by date mode if necessary
	if fq.QueryOptions[0].Elem == godoo.ByNextDate {

		itm, err := h.runGetNextByDate(fq, true)
		if err != nil { //already logged by this point
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		itms = append(itms, itm)

	} else { //standard query mode

		itms, err = h.Repo.GetWhere(fq)
		if err != nil {
			lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(itms)
	lg.Logger.Log(lg.Info, "get handler completed execution")
}

// Pops the next item off the queue by priority. If rePush is true, the item is
// returned to the queue. Can't fully pop until the user marks
// the item as complete - i.e. via the edit command rather
// than the get command
func (h Handler) runGetNextByPriority(fq godoo.FullUserQuery, rePush bool) (godoo.TodoItem, error) {

	td, err := h.PriorityList.GetNext()
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("priority list error: %v", err), runtime.Caller)
		return *td, err
	}

	if rePush {
		err = h.PriorityList.Add(*td)
		if err != nil {
			lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("priority list error: %v", err), runtime.Caller)
			return *td, err
		}
	}

	return *td, nil
}

// Pops the next item off the queue by date. If rePush is true, the item is
// returned to the queue. Can't fully pop until the user marks
// the item as complete - i.e. via the edit command rather
// than the get command
func (h Handler) runGetNextByDate(fq godoo.FullUserQuery, rePush bool) (godoo.TodoItem, error) {

	h.PriorityList.DateMode = true
	td, err := h.PriorityList.GetNext()
	h.PriorityList.DateMode = false //reset

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("priority list (dateMode) error: %v", err), runtime.Caller)
		return *td, err
	}

	if rePush {
		err = h.PriorityList.Add(*td)
		if err != nil {
			lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("priority list (dateMode) error: %v", err), runtime.Caller)
			return *td, err
		}
	}

	return *td, nil
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
