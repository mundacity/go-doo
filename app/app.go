package app

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	fp "github.com/mundacity/flag-parser"
	godoo "github.com/mundacity/go-doo"
	"github.com/mundacity/go-doo/cli"
	"github.com/mundacity/go-doo/util"
	lg "github.com/mundacity/quick-logger"
	"github.com/spf13/viper"
)

// AppContext encapsulates user inputs and is
// passed around to execute desired operations
type AppContext struct {
	Config  godoo.ConfigVals
	cmdName string
}

func (ac *AppContext) SetupCliContext(args []string) {

	ac.Config = godoo.ConfigVals{}
	ac.cmdName = args[0]
	ac.Config.Args = args[1:]

	SetConfigVals()
	ac.Config.MaxLen = viper.GetInt("MAX_LENGTH")
	ac.Config.IntDigits = viper.GetInt("MAX_INT_DIGITS")
	ac.Config.TagDelim = viper.GetString("TAG_DELIMITER")
	ac.Config.Instance = godoo.InstanceType(viper.GetInt("INSTANCE_TYPE"))
	ac.Config.DateLayout = viper.GetString("DATETIME_FORMAT")
	ac.Config.NowString = util.StringFromDate(time.Now())

	startLogger("cli application started...")
	tolog := []any{ac.Config.MaxLen, ac.Config.IntDigits, ac.Config.TagDelim, ac.Config.Instance, ac.Config.DateLayout}
	s := "[MaxLen: %v, IntDigits: %v, TagDelim: %v, InstanceType: %v, DateLayout: %v]"

	if ac.Config.Instance != 0 {
		ac.Config.Client = *http.DefaultClient
		ac.Config.RemoteUrl = fmt.Sprintf("%v:%v", viper.GetString("BASE_URL"), viper.GetInt("SERVER_PORT"))

		tolog = append(tolog, ac.Config.RemoteUrl)
		s = s[:len(s)-1] + ", RemoteUrl: %v]"
		lg.Logger.Logf(lg.Info, s, tolog...)
		return
	}

	// only runs in local mode
	ac.Config.Conn = getConn()
	ac.Config.TodoRepo = getRepo(getDbKind(viper.GetString("DB_TYPE")), ac.Config.Conn, ac.Config.DateLayout, 0)

	tolog = append(tolog, ac.Config.Conn)
	s += ac.Config.Conn
	lg.Logger.Logf(lg.Info, s, tolog...)
}

func (ac *AppContext) GetCommand() (godoo.ICommand, error) {

	var cmd godoo.ICommand
	var err error

	switch ac.cmdName {
	case "add":
		cmd = cli.NewAddCommand(&ac.Config)
	case "get":
		cmd = cli.NewGetCommand(&ac.Config)
	case "edit":
		cmd = cli.NewEditCommand(&ac.Config)
	default:
		return nil, errors.New("invalid command")
	}
	return cmd, err
}

func (ac *AppContext) SetupFlagParser() {
	canonicalFlags := ac.getValidFlags()
	ac.Config.Parser = fp.NewParser(canonicalFlags, ac.Config.Args, ac.Config.NowString, ac.Config.DateLayout)
}

func (ac *AppContext) getValidFlags() []fp.FlagInfo {
	switch ac.cmdName {
	case "add":
		return ac.getAddFlags()
	case "get":
		return ac.getGetFlags()
	case "edit":
		return ac.getEditFlags()
	default:
		return nil
	}
}

func (ac *AppContext) getAddFlags() []fp.FlagInfo {
	var ret []fp.FlagInfo

	lenMax := ac.Config.MaxLen
	maxIntDigits := ac.Config.IntDigits

	f2 := fp.FlagInfo{FlagName: string(godoo.Body), FlagType: fp.Str, MaxLen: lenMax}
	f3 := fp.FlagInfo{FlagName: string(godoo.Mode), FlagType: fp.Str, MaxLen: 1}
	f4 := fp.FlagInfo{FlagName: string(godoo.Tag), FlagType: fp.Str, MaxLen: lenMax}
	f5 := fp.FlagInfo{FlagName: string(godoo.Child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f6 := fp.FlagInfo{FlagName: string(godoo.Parent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f7 := fp.FlagInfo{FlagName: string(godoo.Date), FlagType: fp.DateTime, MaxLen: 20}

	ret = append(ret, f2, f3, f4, f5, f6, f7)
	return ret
}

func (ac *AppContext) getGetFlags() []fp.FlagInfo {
	var ret []fp.FlagInfo

	maxIntDigits := ac.Config.IntDigits

	lenMax := ac.Config.MaxLen

	f8 := fp.FlagInfo{FlagName: string(godoo.Body), FlagType: fp.Str, MaxLen: lenMax}
	f2 := fp.FlagInfo{FlagName: string(godoo.ItmId), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f3 := fp.FlagInfo{FlagName: string(godoo.Next), FlagType: fp.Boolean, Standalone: true}
	f13 := fp.FlagInfo{FlagName: string(godoo.DateMode), FlagType: fp.Boolean, Standalone: true}
	f4 := fp.FlagInfo{FlagName: string(godoo.Date), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f5 := fp.FlagInfo{FlagName: string(godoo.Tag), FlagType: fp.Str, MaxLen: lenMax}
	f6 := fp.FlagInfo{FlagName: string(godoo.Child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f7 := fp.FlagInfo{FlagName: string(godoo.Parent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f9 := fp.FlagInfo{FlagName: string(godoo.Creation), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f10 := fp.FlagInfo{FlagName: string(godoo.All), FlagType: fp.Boolean, Standalone: true}
	f11 := fp.FlagInfo{FlagName: string(godoo.Finished), FlagType: fp.Boolean, Standalone: true}
	f12 := fp.FlagInfo{FlagName: string(godoo.MarkComplete), FlagType: fp.Boolean, Standalone: true}

	ret = append(ret, f8, f2, f3, f4, f5, f6, f7, f9, f10, f11, f12, f13)
	return ret
}

func (ac *AppContext) getEditFlags() []fp.FlagInfo {
	var ret []fp.FlagInfo

	maxIntDigits := ac.Config.IntDigits
	lenMax := ac.Config.MaxLen

	f1 := fp.FlagInfo{FlagName: string(godoo.Body), FlagType: fp.Str, MaxLen: lenMax}
	f2 := fp.FlagInfo{FlagName: string(godoo.ItmId), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f3 := fp.FlagInfo{FlagName: string(godoo.Date), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f4 := fp.FlagInfo{FlagName: string(godoo.Tag), FlagType: fp.Str, MaxLen: lenMax}
	f5 := fp.FlagInfo{FlagName: string(godoo.Child), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f6 := fp.FlagInfo{FlagName: string(godoo.Creation), FlagType: fp.DateTime, MaxLen: 21, AllowDateRange: true}
	f14 := fp.FlagInfo{FlagName: string(godoo.Finished), FlagType: fp.Boolean, Standalone: true}

	f7 := fp.FlagInfo{FlagName: string(godoo.AppendMode), FlagType: fp.Boolean, Standalone: true}
	f8 := fp.FlagInfo{FlagName: string(godoo.ReplaceMode), FlagType: fp.Boolean, Standalone: true}

	f9 := fp.FlagInfo{FlagName: string(godoo.ChangeBody), FlagType: fp.Str, MaxLen: lenMax}
	f10 := fp.FlagInfo{FlagName: string(godoo.ChangeTag), FlagType: fp.Str, MaxLen: lenMax}
	f11 := fp.FlagInfo{FlagName: string(godoo.ChangeParent), FlagType: fp.Integer, MaxLen: maxIntDigits}
	f12 := fp.FlagInfo{FlagName: string(godoo.ChangedDeadline), FlagType: fp.DateTime, MaxLen: 20}
	f13 := fp.FlagInfo{FlagName: string(godoo.MarkComplete), FlagType: fp.Boolean, Standalone: true}
	f15 := fp.FlagInfo{FlagName: string(godoo.ChangeMode), FlagType: fp.Str, MaxLen: 1}

	ret = append(ret, f1, f2, f3, f4, f5, f6, f7, f8, f9, f10, f11, f12, f13, f14, f15)
	return ret
}
