package srv

import (
	"encoding/json"
	"net/http"

	godoo "github.com/mundacity/go-doo"
)

type Handler struct {
	Repo godoo.IRepository
}

func (h Handler) TestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h Handler) HandleRequests(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		h.GetHandler(w, r)
	case http.MethodPut:
		h.EditHandler(w, r)
	case http.MethodPost:
		h.AddHandler(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h Handler) AddHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var td godoo.TodoItem
	d := json.NewDecoder(r.Body)
	err := d.Decode(&td)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	i, err := h.Repo.Add(&td)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
}

func (h Handler) GetHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var fq godoo.FullUserQuery
	d := json.NewDecoder(r.Body)
	err := d.Decode(&fq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itms, err := h.Repo.GetWhere(fq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(itms)
}

func (h Handler) EditHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	var fq []godoo.FullUserQuery
	d := json.NewDecoder(r.Body)
	err := d.Decode(&fq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(fq) != 2 {
		http.Error(w, "two FullUserQuery structs required", http.StatusForbidden)
		return
	}

	i, err := h.Repo.UpdateWhere(fq[0], fq[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
}
