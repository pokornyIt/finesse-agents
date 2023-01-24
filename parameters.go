package main

import (
	"errors"
	"fmt"
	api "github.com/pokornyIt/finesse-api"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	DefaultFileName = "agents.csv"
)

type FinesseServerConfig struct {
	FinesseServer            string
	Port                     int
	Force                    bool
	IgnoreCertificateProblem bool
	InsecureConnect          bool
}

type AgentConfig struct {
	Name     string
	Password string
	Line     string
	Command  string
}

type AgentGroupConfig struct {
	FileName  string
	Operation string
	Divider   string
}

type LogConfig struct {
	Level  string
	Folder string
	Dry    bool
}

type errMsg struct {
	Msg string
}

var (
	//app = kingpin.New("agents", "command line for manipulate with finesse agents")

	agApp  = kingpin.Command("agent", "operation with one specific agent")
	srvApp = kingpin.Command("finesse", "operation with group of agents")

	serverConfig     = FinesseServerConfig{}
	agentConfig      = AgentConfig{}
	agentGroupConfig = AgentGroupConfig{}
	logConfig        = LogConfig{
		Level:  "e",
		Folder: "",
		Dry:    false,
	}
)

func init() {
	kingpin.Flag("server", "finesse server name or IP address").Short('s').PlaceHolder("finesse.server.local").Default("").StringVar(&serverConfig.FinesseServer)
	kingpin.Flag("force", "force operation").Short('f').Default("false").BoolVar(&serverConfig.Force)
	kingpin.Flag("port", "port for API connection").Short('p').Default(strconv.Itoa(api.DefaultServerHttpsPort)).HintOptions("443", "8443", strconv.Itoa(api.DefaultServerHttpsPort)).IntVar(&serverConfig.Port)
	kingpin.Flag("ignore-security-check", "ignore HTTPS security check").Short('i').Default("false").BoolVar(&serverConfig.IgnoreCertificateProblem)
	kingpin.Flag("insecure-xmpp", "use insecure connection to XMPP, need change XMPP port to 5222").Default("false").BoolVar(&serverConfig.InsecureConnect)
	kingpin.Flag("level", "define logger level (error, warning, info, debug, trace)").Short('l').Default("error").
		EnumVar(&logConfig.Level, "error", "err", "e", "warning", "warn", "w", "info", "i", "debug", "deb", "d", "trace", "trc", "t")
	kingpin.Flag("directory", "define alternative directory for store logs").Short('D').Default("").StringVar(&logConfig.Folder)
	kingpin.Flag("show", "only show actual configuration").Short('S').Default("false").BoolVar(&logConfig.Dry)

	agApp.Flag("name", "agent login name").Short('a').PlaceHolder("agent").Default("").StringVar(&agentConfig.Name)
	agApp.Flag("pwd", "agent password").Short('P').PlaceHolder("password").Default("").StringVar(&agentConfig.Password)
	agApp.Flag("number", "line number used by agent").Short('n').PlaceHolder("1234").Default().StringVar(&agentConfig.Line)
	agApp.Flag("operation", "operation with agent (status, login, ready, not-ready, logout)").Short('o').Default("status").
		EnumVar(&agentConfig.Command, "status", "login", "ready", "not-ready", "logout")

	srvApp.Flag("agents", "agents group definition CSV file").Short('a').PlaceHolder(DefaultFileName).Default("").ExistingFileVar(&agentGroupConfig.FileName)
	srvApp.Flag("operation", "operation with agents group (status, login, ready, not-ready, logout, list)").Short('o').Default("status").
		EnumVar(&agentGroupConfig.Operation, "status", "login", "ready", "not-ready", "logout", "list")
	srvApp.Flag("divider", "CSV file separator").Short('d').Default(",").StringVar(&agentGroupConfig.Divider)
}

func (f *FinesseServerConfig) Validate() error {
	err := f.serverValid()
	if err == nil {
		err = f.portValid()
	}
	return err
}

func (f *FinesseServerConfig) serverValid() error {
	if len(f.FinesseServer) < 1 {
		return errors.New("server name not defined")
	}
	if !api.ValidServerNameIp(f.FinesseServer) {
		return fmt.Errorf("server name is not valid FQDN or IP address <%s>", f.FinesseServer)
	}
	return nil
}

func (f *FinesseServerConfig) portValid() error {
	if (f.Port > 1023 && f.Port < 65536) || f.Port == 443 {
		return nil
	}
	return fmt.Errorf("finesse port is out of valid range 1024 - 65536 or different from standard HTTPS port 443 <%d> ", f.Port)
}

func (f *FinesseServerConfig) sprint() string {
	m := errMsg{Msg: ""}
	a := "Server setup"
	a = fmt.Sprintf("%s\r\n\tServer                        [%s]%s", a, f.FinesseServer, m.marker(f.serverValid()))
	a = fmt.Sprintf("%s\r\n\tPort                          [%d]%s", a, f.Port, m.marker(f.portValid()))
	a = fmt.Sprintf("%s\r\n\tXMPP Port                     [%d]", a, api.DefaultServerXmppPort)
	a = fmt.Sprintf("%s\r\n\tForce                         [%t]", a, f.Force)
	a = fmt.Sprintf("%s\r\n\tIgnore Certificate problem    [%t]", a, f.IgnoreCertificateProblem)
	a = fmt.Sprintf("%s\r\n\tInsecure XMPP                 [%t]", a, f.InsecureConnect)

	return a + m.message()
}

func (a *AgentConfig) Validate() error {
	err := a.agentValid()
	if err == nil {
		err = a.passwordValid()
	}
	if err == nil {
		err = a.lineValid()
	}
	return err
}

func (a *AgentConfig) agentValid() error {
	rx := regexp.MustCompile(`^\S+$`)
	if !rx.MatchString(a.Name) {
		return fmt.Errorf("agent name is empty or contains space <%s>", a.Name)
	}
	return nil
}

func (a *AgentConfig) passwordValid() error {
	if len(a.Password) < 1 {
		return errors.New("agent password not defined")
	}
	return nil
}

func (a *AgentConfig) lineValid() error {
	rx := regexp.MustCompile(`^\d+$`)
	if !rx.MatchString(a.Line) {
		return fmt.Errorf("line not valid expect only numbers <%s>", a.Line)
	}
	return nil
}

func (a *AgentConfig) sprint() string {
	e := errMsg{Msg: ""}
	m := "Agent setup"
	m = fmt.Sprintf("%s\r\n\tAgent name     [%s]%s", m, a.Name, e.marker(a.agentValid()))
	m = fmt.Sprintf("%s\r\n\tPassword       [%t]%s", m, len(a.Password) > 0, e.marker(a.passwordValid()))
	m = fmt.Sprintf("%s\r\n\tLine           [%s]%s", m, a.Line, e.marker(a.lineValid()))
	m = fmt.Sprintf("%s\r\n\tCommand        [%s]", m, a.Command)

	return m + e.message()
}

func (l *LogConfig) Validate() error {
	if len(l.Folder) < 1 {
		return nil
	}
	err := IsWritable(l.Folder)
	if err != nil {
		return err
	}
	return nil
}

func (l *LogConfig) fileOrErr() string {
	if l.Validate() == nil && len(l.Folder) == 0 {
		return "stderr"
	}
	return l.Folder
}

func (l *LogConfig) sprint() string {
	e := errMsg{Msg: ""}
	m := "Loge setup"
	m = fmt.Sprintf("%s\r\n\tLevel          [%s]", m, l.GetLevel())
	m = fmt.Sprintf("%s\r\n\tDirectory      [%s]%s", m, l.fileOrErr(), e.marker(l.Validate()))
	return m + e.message()
}

func (a *AgentGroupConfig) Validate() error {
	r, l := utf8.DecodeRuneInString(a.Divider)
	if l < 1 {
		a.Divider = ","
	}
	if l > 1 || r == '\r' || r == '\n' {
		a.Divider = string(r)
	}
	return nil
}

func (a *AgentGroupConfig) GetDivider() rune {
	r, _ := utf8.DecodeRuneInString(a.Divider)
	return r
}

func (a *AgentGroupConfig) sprint() string {
	m := "Agent Group setup"
	m = fmt.Sprintf("%s\r\n\tFile           [%s]", m, a.FileName)
	m = fmt.Sprintf("%s\r\n\tCommand        [%s]", m, a.Operation)
	m = fmt.Sprintf("%s\r\n\tDivider        [%s]", m, a.Divider)

	return m
}

func (l *LogConfig) GetLevel() log.Level {
	switch strings.ToLower(l.Level[0:1]) {
	case "w":
		return log.WarnLevel
	case "i":
		return log.InfoLevel
	case "d":
		return log.DebugLevel
	case "t":
		return log.TraceLevel
	}
	return log.ErrorLevel
}

func Validate(cmd string) error {
	err := serverConfig.Validate()
	if err != nil {
		return err
	}
	if cmd == agApp.FullCommand() {
		return agentConfig.Validate()
	}
	if cmd == srvApp.FullCommand() {
		return agentGroupConfig.Validate()
	}
	return nil
}

func (m *errMsg) message() string {
	if len(m.Msg) > 0 {
		return fmt.Sprintf("\r\n   * Errors:\r\n%s", m.Msg)
	}
	return ""
}

func (m *errMsg) marker(err error) string {
	if err != nil {
		if len(m.Msg) > 0 {
			m.Msg = fmt.Sprintf("%s\r\n\t%s", m.Msg, err)
		} else {
			m.Msg = fmt.Sprintf("%s\t%s", m.Msg, err)
		}
		return " *"
	}
	return ""
}
