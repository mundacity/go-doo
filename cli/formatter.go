package cli

// Don't want a dependency just for colours/formatting.
// Colours & init func taken from https://twin.sh/articles/35/how-to-add-colors-to-your-console-terminal-output-in-go
import (
	"fmt"
	"io"
	"runtime"
	"strings"

	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/util"
)

var Reset = "\033[0m"
var Red = "\033[31m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Blue = "\033[34m"
var Purple = "\033[35m"
var Cyan = "\033[36m"
var Gray = "\033[37m"
var White = "\033[97m"

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Red = ""
		Green = ""
		Yellow = ""
		Blue = ""
		Purple = ""
		Cyan = ""
		Gray = ""
		White = ""
	}
}

// Runs after successfully adding a new item
func printAddMessage(id int, w io.Writer) {
	msg := fmt.Sprintf("Creation successful, ItemId: %v\n", id)
	w.Write([]byte(msg))
}

// Runs after successfully editing n items
func printEditMessage(n int, w io.Writer) {
	s := ""
	if n == 0 || n > 1 {
		s = "s"
	}
	msg := fmt.Sprintf("--> Edited %v item%v\n", n, s)
	w.Write([]byte(msg))
}

// Runs after successfully retrieving item/s. Returns a func that returns a formatted string
func getOutputGenerationFunc(itms []godoo.TodoItem) func() string {
	f := func() string {
		var str string
		for _, itm := range itms {
			str += buildOutput(itm) + "\n"
		}
		c := len(itms)
		s := ""
		if c == 0 || c > 1 {
			s = "s"
		}
		str += fmt.Sprintf("--> Returned %v item%v\n", c, s)
		return str
	}
	return f
}

func buildOutput(itm godoo.TodoItem) string {
	var retStr string
	tagOut := getTagOutput(itm.Tags)
	deadline := "n/a"
	if !itm.Deadline.IsZero() {
		deadline = util.StringFromDate(itm.Deadline)
	}
	done := Red + "Not done" + Reset
	if itm.IsComplete {
		done = Green + "Done" + Reset
	}
	retStr += fmt.Sprintf(Yellow+"-- Id:"+Reset+" [%v][%v]\n\t"+Cyan+"- Created:"+Reset+"  %v     "+Cyan+"ParentId:"+Reset+" %v     "+Cyan+"Priority:"+Reset+" %v\n\t"+Cyan+"- Deadline:"+Reset+" %v\n\t"+Cyan+"- Tags:"+Reset+"     %v\n\t"+Cyan+"- Body:"+Reset+"     %v\n", itm.Id, done, util.StringFromDate(itm.CreationDate), itm.ParentId, itm.Priority, deadline, tagOut, itm.Body)
	return retStr
}

func getTagOutput(mp map[string]struct{}) string {
	var ret string
	sep := "; "

	for v := range mp {
		if len(v) > 0 {
			ret += v + sep
		}
	}
	return strings.TrimSuffix(ret, "; ")
}
