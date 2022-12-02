package srv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/auth"
	lg "github.com/mundacity/quick-logger"
)

type Handler struct {
	Repo         godoo.IRepository
	PriorityList *godoo.PriorityList
	priorityMode bool
	path         string
	pwHash       string
}

// Returns a new http handler. If runPl is true, then the handler will
// maintain a priority queue as well.
func NewHandler(ct godoo.ServerConfigVals) *Handler {

	h := &Handler{Repo: ct.Repo}
	h.path = ct.KeyPath
	h.pwHash = ct.UserPasswordHash

	if ct.RunPriorityList {
		h.priorityMode = true
		h.PriorityList = ct.PriorityList
		h.setupPriorityList()
	}

	return h
}

func (h *Handler) setupPriorityList() { // todo: look at go routines for this (maybe just the callers)
	if h.priorityMode {

		all, _ := h.Repo.GetAll()

		if len(h.PriorityList.List.Items) > 0 {
			for _, v := range h.PriorityList.List.Items {
				h.PriorityList.Delete(v.Id)
			}
		}

		for _, v := range all {
			if !v.IsComplete {
				h.PriorityList.Add(v)
			}
		}
	}
}

func (h Handler) TestHandler(w http.ResponseWriter, r *http.Request) {
	lg.Logger.Log(lg.Info, "test handler called")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) HandleRequests(w http.ResponseWriter, r *http.Request) {

	lg.Logger.Logf(lg.Info, "%v request received from %v", r.Method, r.RemoteAddr)

	switch r.Method {
	case http.MethodGet:
		//auth.Authenticate(r, h.path, h.pwHash)
		auth.ValidateJwt(h.path, h.GetHandler)
	case http.MethodPut:
		h.EditHandler(w, r)
	case http.MethodPost:
		h.AddHandler(w, r)
	default:
		lg.Logger.LogWithCallerInfo(lg.Error, "method not allowed", runtime.Caller)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) AddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var td godoo.TodoItem
	d := json.NewDecoder(r.Body)

	d.DisallowUnknownFields()
	err := d.Decode(&td)

	// validation
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if td.CreationDate.IsZero() { // only thing that really needs to not be empty
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	i, err := h.Repo.Add(&td)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.priorityMode {
		if err = h.PriorityList.Add(td); err != nil {
			// db fine but pl out of sync --> reset pl
			h.setupPriorityList()
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
	lg.Logger.Log(lg.Info, "add handler completed execution")
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {

	var itms []godoo.TodoItem
	var err error

	w.Header().Set("content-type", "application/json")

	// parse user query
	var fq godoo.FullUserQuery
	d := json.NewDecoder(r.Body)

	d.DisallowUnknownFields()
	err = d.Decode(&fq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// handle get by priority mode/date
	if h.priorityMode && len(fq.QueryOptions) == 1 {

		itm, done, msg, err := h.handleQueueMode(fq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if done {
			itms = append(itms, itm)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(itms)
			lg.Logger.Logf(lg.Info, "get handler (%v) completed execution", msg)
			return
		}
	}

	// standard get query
	itms, err = h.Repo.GetWhere(fq)
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(itms)
	lg.Logger.Log(lg.Info, "get handler completed execution")
}

func (h *Handler) handleQueueMode(fq godoo.FullUserQuery) (itm godoo.TodoItem, done bool, logMsg string, err error) {

	done = false
	if fq.QueryOptions[0].Elem == godoo.ByNextPriority {
		itm, err = h.runGetNextByPriority(fq, true)
		done = true
		logMsg = "priority mode"
	}
	if fq.QueryOptions[0].Elem == godoo.ByNextDate {
		itm, err = h.runGetNextByDate(fq, true)
		done = true
		logMsg = "date priority mode"
	}
	return itm, done, logMsg, err
}

// Pops the next item off the queue by priority. If rePush is true, the item is
// returned to the queue. Can't fully pop until the user marks
// the item as complete - i.e. via the edit command rather
// than the get command
func (h *Handler) runGetNextByPriority(fq godoo.FullUserQuery, rePush bool) (godoo.TodoItem, error) {

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
func (h *Handler) runGetNextByDate(fq godoo.FullUserQuery, rePush bool) (godoo.TodoItem, error) {

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

func (h *Handler) EditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var fq []godoo.FullUserQuery
	d := json.NewDecoder(r.Body)

	d.DisallowUnknownFields()
	err := d.Decode(&fq)

	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("bad request: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(fq) != 2 {
		msg := "operation forbidden; two FullUserQuery structs required"
		lg.Logger.LogWithCallerInfo(lg.Error, msg, runtime.Caller)
		http.Error(w, "operation forbidden", http.StatusForbidden)
		return
	}

	i, err := h.Repo.UpdateWhere(fq[0], fq[1])
	if err != nil {
		lg.Logger.LogWithCallerInfo(lg.Error, fmt.Sprintf("server error: %v", err), runtime.Caller)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.priorityMode {
		h.setupPriorityList() //probably inefficient but won't have the full items (just bits to update), so better to just start again
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
	lg.Logger.Log(lg.Info, "edit handler completed execution")
}
