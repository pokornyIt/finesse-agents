package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	api "github.com/pokornyIt/finesse-api"
	log "github.com/sirupsen/logrus"
	"os"
)

type CsvAgent struct {
	Name     string
	Password string
	Line     string
}

func finesseSelectCmd() int {
	var states []*api.UserDetailResponse
	ret := 0
	err := collectAgents()
	if err != nil {
		return agentErrorMessage(fmt.Sprintf("problem create agents list %s", err))
	}
	switch agentGroupConfig.Operation {
	case "status":
		log.Debugf("collect status for agents [%d]", len(server.GetAgentsList()))
		states, err = server.GetStateAgentsParallel()
		printStates("Agent statuses:", states)
		finalFinesseMessage()
		break
	case "login":
		log.Debugf("login agents [%d]", len(server.GetAgentsList()))
		states, err = server.LoginAgentsParallelWithStatus()
		if err != nil {
			printStates("Agents with invalid state:", states, api.AgentLoginStates)
			ret = agentErrorMessage(err.Error())
		} else {
			finalFinesseMessage()
		}
		break
	case "ready":
		log.Debugf("set ready agents [%d]", len(server.GetAgentsList()))
		states, err = server.ReadyAgentsParallelWithStatus(serverConfig.Force)
		if err != nil {
			printStates("Agents with invalid state:", states, api.AgentReadyStates)
			ret = agentErrorMessage(err.Error())
		} else {
			finalFinesseMessage()
		}
		break
	case "not-ready":
		log.Debugf("set not-ready agents [%d]", len(server.GetAgentsList()))
		states, err = server.NotReadyAgentsParallelWithStatus(serverConfig.Force)
		if err != nil {
			printStates("Agents with invalid state:", states, api.AgentLoginStates)
			ret = agentErrorMessage(err.Error())
		} else {
			finalFinesseMessage()
		}
		break
	case "logout":
		log.Debugf("logout agents [%d]", len(server.GetAgentsList()))
		states, err = server.LogoutAgentsParallelWithStatus(serverConfig.Force)
		if err != nil {
			printStates("Agents with invalid state:", states, api.AgentLogoutState)
			ret = agentErrorMessage(err.Error())
		} else {
			finalFinesseMessage()
		}
		break
	case "list":
		log.Debugf("list agents [%d]", len(server.GetAgentsList()))
		break
	default:
		ret = agentErrorMessage(fmt.Sprintf("unknown operation [%s]", agentGroupConfig.Operation))
	}

	return ret
}

func collectAgents() error {
	file, err := os.ReadFile(agentGroupConfig.FileName)
	if err != nil {
		return err
	}
	file = bytes.Trim(file, "\xef\xbb\xbf")
	reader := csv.NewReader(bytes.NewReader(file))
	reader.Comma = agentGroupConfig.GetDivider()

	data, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for i, row := range data {
		if i == 0 {
			if row[0] == "name" && row[1] == "password" && row[2] == "line" {
				continue
			}
		}
		c := CsvAgent{
			Name:     row[0],
			Password: row[1],
			Line:     row[2],
		}
		if c.Valid() {
			agent := api.NewAgentName(c.Name, c.Password, c.Line)
			server.AddAgent(agent)

		} else {
			_ = agentErrorMessage(fmt.Sprintf("line [%d] has one of filed [%s, %s, %s]", i, c.Name, c.Password, c.Line))
		}
	}

	return nil
}

func (a CsvAgent) Valid() bool {
	return len(a.Name) > 0 && len(a.Password) > 0 && len(a.Line) > 0
}

func finalFinesseMessage() {
	o := fmt.Sprintf("%s operation [%s] finish sucess for all [%d] agents", srvApp.FullCommand(), agentGroupConfig.Operation, len(server.GetAgentsList()))
	log.Info(o)
	fmt.Println(o)

}

func printStates(msg string, states []*api.UserDetailResponse, expect ...map[string]string) {
	a := msg
	var show bool
	for _, state := range states {
		if len(expect) > 0 {
			_, show = expect[0][state.State]
			show = !show
		} else {
			show = true
		}
		if show {
			a = fmt.Sprintf("%s\r\n%s", a, state.ToStingSimple())
		}
	}
	fmt.Println(a)
}
