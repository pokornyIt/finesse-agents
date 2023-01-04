package main

import (
	"fmt"
	api "github.com/pokornyIt/finesse-api"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	server    *api.FinesseServer
	Version   string // for build data
	Revision  string // for build data
	Branch    string // for build data
	BuildUser string // for build data
	BuildDate string // for build data
)

func setupLog() *os.File {
	var f *os.File
	var err error
	errLog := logConfig.Validate()
	if errLog == nil && len(logConfig.Folder) > 0 {
		logFile := filepath.Join(logConfig.Folder, fmt.Sprintf("finesse-agents-%s.log", time.Now().Format("20060102")))
		//logFile := filepath.Join(logConfig.Folder, fmt.Sprintf("finesse-agents-%s.log", time.Now().Format("20060102-150405")))
		f, err = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("Failed to create logfile" + logFile)
			panic(err)
		}
		log.SetOutput(f)
	} else {
		f = nil
	}
	log.SetLevel(logConfig.GetLevel())
	return f
}

func printActualConfig(cmd string) {

	a := fmt.Sprintf("%s\r\n\r\nActual setup from line", versionToString())
	a = fmt.Sprintf("%s\r\n%s", a, logConfig.sprint())
	a = fmt.Sprintf("%s\r\n%s", a, serverConfig.sprint())
	if cmd == agApp.FullCommand() {
		a = fmt.Sprintf("%s\r\n%s", a, agentConfig.sprint())
	}
	if cmd == srvApp.FullCommand() {
		a = fmt.Sprintf("%s\r\n%s", a, agentGroupConfig.sprint())
	}
	fmt.Println(a)
}

func validateSetup() {
	if len(Version) < 1 {
		Version = "0.0.0"
	}
	if len(Branch) < 1 {
		Branch = "unknown"
	}
	if len(Revision) < 1 {
		Revision = "unknown"
	}
	if len(BuildUser) < 1 {
		BuildUser = "unknown"
	}
	if len(BuildDate) < 1 {
		BuildDate = time.Now().Format("20060102-15:04:05")
	}
}

func versionToString() string {
	a := fmt.Sprintf("%s, version %s (branch: %s, revision: %s)", filepath.Base(os.Args[0]), Version, Branch, Revision)
	a = fmt.Sprintf("%s\r\n  build user: %s", a, BuildUser)
	a = fmt.Sprintf("%s\r\n  build date: %s", a, BuildDate)
	a = fmt.Sprintf("%s\r\n  go version: %s", a, runtime.Version())
	a = fmt.Sprintf("%s\r\n  platform  : %s", a, runtime.GOOS)
	return a
}

func main() {
	var err error
	retCode := 0
	ts := time.Now()
	defer func() {
		fmt.Printf("\n\rprogram run %v\n\r", Timespan(time.Since(ts)).Format("15:04:05.000"))
		os.Exit(retCode)
	}()
	validateSetup()

	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Version(versionToString())
	kingpin.HelpFlag.Short('h')
	command := kingpin.Parse()

	// Only show actual data
	if logConfig.Dry {
		printActualConfig(command)
		return
	}

	// setup log
	f := setupLog()
	if f != nil {
		defer func() { _ = f.Close() }()
	}

	err = Validate(command)
	if err != nil {
		fmt.Printf("problem with application configuration\r\n\t%s", err)
		retCode = 1
		return
	}

	log.Infof("Program start (version: %s, branch: %s, build date: %s", Version, Branch, BuildDate)
	server = api.NewFinesseServer(serverConfig.FinesseServer, serverConfig.FinessePort)
	if command == agApp.FullCommand() {
		agentSelectOperation()
	} else if command == srvApp.FullCommand() {
		finesseSelectCmd()
	}
	// await before ends to made time finish all output data
	time.Sleep(10)
}
