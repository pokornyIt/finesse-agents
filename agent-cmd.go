package main

import (
	"fmt"
	api "github.com/pokornyIt/finesse-api"
	log "github.com/sirupsen/logrus"
)

func agentSelectOperation() int {
	var status *api.UserDetailResponse
	var err error
	ret := 0

	agent := api.NewAgentName(agentConfig.Name, agentConfig.Password, agentConfig.Line)
	server.AddAgent(agent)

	switch agentConfig.Command {
	case "status":
		log.Debugf("collect status for agent [%s]", agentConfig.Name)
		status, err = server.GetAgentStatusDetail(agentConfig.Name)
		if err != nil {
			ret = agentErrorMessage(fmt.Sprintf("problem collect agent actual status for agent [%s]", agentConfig.Name))
			break
		}
		agentFinalState("", "")
		break
	case "login":
		log.Debugf("log in agent [%s]", agentConfig.Name)
		state := server.LoginAgent(agentConfig.Name)
		if state {
			agentFinalState("login success, but ", "")
		} else {
			ret = agentErrorMessage(fmt.Sprintf("problem log in agent [%s]", agentConfig.Name), "agent not logged in")
		}
		break
	case "ready":
		log.Debugf("set READY agent [%s]", agentConfig.Name)
		status, err = server.GetAgentStatusDetail(agentConfig.Name)
		if err != nil {
			ret = agentErrorMessage(fmt.Sprintf("problem collect agent actual status"))
			break
		}
		if !serverConfig.Force && (status.State == api.AgentStateUnknown || status.State == api.AgentStateLogout) {
			ret = agentErrorMessage(fmt.Sprintf("agent [%s] is not in state for set to ready, actual state [%s]", agentConfig.Name, status.State))
			break
		}
		if !serverConfig.Force && (status.State == api.AgentStateNotReady || status.PendingState == api.AgentStateWorkNotReady) {
			if server.ReadyAgent(agentConfig.Name) {
				agentFinalState("agent ready success, but ", "")
			} else {
				ret = agentErrorMessage(fmt.Sprintf("problem set agent [%s] to ready state", agentConfig.Name))
			}
			break
		}
		if serverConfig.Force {
			_ = server.LoginAgent(agentConfig.Name)
		}
		read := server.ReadyAgent(agentConfig.Name)
		if read {
			agentFinalState("ready state success, but ", "")
		} else {
			ret = agentErrorMessage(fmt.Sprintf("problem log in agent [%s]", agentConfig.Name))
		}
		break
	case "not-ready":
		log.Debugf("set NOT-READY agent [%s]", agentConfig.Name)
		nReady := server.NotReadyAgent(agentConfig.Name)
		if nReady {
			agentFinalState("not-ready success, but ", "")
		} else {
			ret = agentErrorMessage(fmt.Sprintf("problem set agent to not-ready state [%s]", agentConfig.Name))
		}
		break
	case "logout":
		log.Debugf("logout agent [%s]", agentConfig.Name)
		if serverConfig.Force {
			_ = server.NotReadyAgent(agentConfig.Name)
		}

		logout := server.LogoutAgent(agentConfig.Name)
		if logout {
			agentFinalState("logout success, but ", "")
		} else {
			ret = agentErrorMessage(fmt.Sprintf("problem logout agent [%s]", agentConfig.Name), "agent not logged in")
		}
		break
	default:
		log.Errorf("unknown operation <%s>", agentConfig.Command)
		ret = 1
	}

	return ret
}

func agentFinalState(prefixError string, suffixError string) {
	status, err := server.GetAgentStatusDetail(agentConfig.Name)
	if err != nil {
		fmt.Printf("%sproblem collect status for agent [%s]: %s%s\r\n", prefixError, agentConfig.Name, err, suffixError)
	} else {
		o := fmt.Sprintf("agent [%s] actual status [%s], pending state [%s]", agentConfig.Name, status.State, status.PendingState)
		log.Info(o)
		fmt.Println(o)
	}
}

func agentErrorMessage(msg string, o ...string) int {

	log.Errorf(msg)
	if len(o) > 0 {
		fmt.Println(o[0])
	} else {
		fmt.Println(msg)
	}
	return 1
}
