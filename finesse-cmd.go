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
	var opError []api.OperationError

	ret := 0
	agents, err := collectAgents()
	if agents == nil && err != nil {
		log.Errorf("problem collect agents - %s", err)
		return ExitNotCollectAgentGroup
	}
	if agents != nil && len(agents.Agents) > 0 && err != nil {
		log.Warnf("system continue with limited set of agents - %s", err)
	}
	if agents != nil && len(agents.Agents) == 0 && err != nil {
		log.Errorf("system ends no agent to operate - %s", err)
		return ExitAgentGroupEmpty
	}

	msg := fmt.Sprintf("finesse start process operation [%s] for %d agents", agentGroupConfig.Operation, len(agents.Agents))
	fmt.Println(msg)
	log.Infof(msg)
	switch agentGroupConfig.Operation {
	case "status":
		printStates("Agent statuses:", agents)
		finalFinesseMessage(agents)
		break
	case "login":
		opError = agents.Login()
		err = operationErrorList(opError)
		if err != nil {
			log.Error(err)
			ret = ExitAgentGroupOperation
		} else {
			finalFinesseMessage(agents)
		}
		break
	case "ready":
		opError = agents.Ready(serverConfig.Force)
		err = operationErrorList(opError)
		if err != nil {
			log.Error(err)
			ret = ExitAgentGroupOperation
		} else {
			finalFinesseMessage(agents)
		}
		break
	case "not-ready":
		opError = agents.NotReady()
		err = operationErrorList(opError)
		if err != nil {
			log.Error(err)
			ret = ExitAgentGroupOperation
		} else {
			finalFinesseMessage(agents)
		}
		break
	case "logout":
		opError = agents.Logout(serverConfig.Force)
		err = operationErrorList(opError)
		if err != nil {
			log.Error(err)
			ret = ExitAgentGroupOperation
		} else {
			finalFinesseMessage(agents)
		}
		break
	case "list":
		ListAgent(agents)
		break
	default:
		ret = agentErrorMessage(fmt.Sprintf("unknown operation [%s]", agentGroupConfig.Operation))
	}

	return ret
}

func collectAgents() (*api.AgentGroup, error) {
	ag := api.NewAgentGroup()

	file, err := os.ReadFile(agentGroupConfig.FileName)
	if err != nil {
		return nil, err
	}
	file = bytes.Trim(file, "\xef\xbb\xbf")
	reader := csv.NewReader(bytes.NewReader(file))
	reader.Comma = agentGroupConfig.GetDivider()

	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var agents []api.BulkAgent
	problemInLoad := 0
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
			agents = append(agents, api.BulkAgent{
				Name:     c.Name,
				Password: c.Password,
				Line:     c.Line,
			})
		} else {
			_ = agentErrorMessage(fmt.Sprintf("line [%d] has one of filed [%s, %s, %s]", i, c.Name, c.Password, c.Line))
			problemInLoad++
		}
	}
	if problemInLoad > 0 {
		log.Errorf("from file not loaded [%d] agents", problemInLoad)
	}
	oe := ag.AddBulkAgents(agents, server)
	for _, operationError := range oe {
		if operationError.Type != api.TypeErrorNoError {
			_ = agentErrorMessage(fmt.Sprintf("%s", operationError.Error))
			problemInLoad++
		}
	}
	if problemInLoad > 0 {
		return ag, fmt.Errorf("problem collects [%d] agents into group ", problemInLoad)
	}
	return ag, nil
}

func (a CsvAgent) Valid() bool {
	return len(a.Name) > 0 && len(a.Password) > 0 && len(a.Line) > 0
}

func finalFinesseMessage(ag *api.AgentGroup) {
	o := fmt.Sprintf("%s operation [%s] finish sucess for all [%d] agents", srvApp.FullCommand(), agentGroupConfig.Operation, len(ag.Agents))
	log.Info(o)
	fmt.Println(o)

}

func printStates(msg string, ag *api.AgentGroup, expect ...map[string]string) {
	a := msg
	var show bool
	for _, agent := range ag.Agents {
		state := agent.GetLastStatus()
		if len(expect) > 0 {
			_, show = expect[0][state.State]
			show = !show
		} else {
			show = true
		}
		if show {
			a = fmt.Sprintf("%s\r\n%s", a, ToStringSimple(state))
		}
	}
	fmt.Println(a)
}

func ToStringSimple(u *api.XmppUser) string {
	return fmt.Sprintf("Agent: %30s Actual: %20s Pending: %s", u.LoginName, u.State, u.PendingState)
}

func operationErrorList(oe []api.OperationError) error {
	cnt := 0
	for _, operationError := range oe {
		if operationError.Type != api.TypeErrorNoError {
			fmt.Printf(" >>> %s\r\n", operationError.Error)
			cnt++
		}
	}
	if cnt > 0 {
		return fmt.Errorf("agents problems durion operation [%d]", cnt)
	}
	return nil
}

func ListAgent(group *api.AgentGroup) {
	m := "Agent list:"
	for _, agent := range group.Agents {
		m = fmt.Sprintf("%s\r\n%s", m, agent)
	}
	fmt.Printf("%s\r\n", m)
}
