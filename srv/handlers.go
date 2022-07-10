package srv

import "net/http"

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func AddHandler(w http.ResponseWriter, r *http.Request) {

}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	// need to figure out the form the query will come in...
	// e.g. GetWhere() takes a []godoo.UserQuery and a godoo.TodoItem
	//http://192.168.0.1?name1=value1&name2=value2 is apparantly the standard

	/*
	* think I might need a different implementation of IRepository which
	* can convert a []godoo.UserQuery and a godoo.TodoItem into
	* a proper http query string? ... or maybe you can put an array into
	* an http query string, in which case I could just have a list of string
	* values and match that to []godoo.UserQuery, and then various keyVal pairs
	* for todo item attributes to search by... too tired to figure it out now...
	 */

}

func EditHandler(w http.ResponseWriter, r *http.Request) {

}
