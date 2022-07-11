package srv

import (
	"encoding/json"
	"net/http"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/sqlite"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func AddHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	w.Header().Set("content-type", "application/json")

	var td godoo.TodoItem
	d := json.NewDecoder(r.Body)
	err := d.Decode(&td)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	i, err := sqlite.AppRepo.Add(&td)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	w.Header().Set("content-type", "application/json")

	var fq godoo.FullUserQuery
	d := json.NewDecoder(r.Body)
	err := d.Decode(&fq)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	itms, err := sqlite.AppRepo.GetWhere(fq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(itms)
}

func EditHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

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

	i, err := sqlite.AppRepo.UpdateWhere(fq[0], fq[1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(i)
}
