package main

import (
	"context"
	"fmt"
	api "github.com/pokornyIt/finesse-api"
	log "github.com/sirupsen/logrus"
	"time"
)

func agentSelectOperation() int {
	var opErr api.OperationError
	ret := 0

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		log.Debugf("ends background processes for agent [%s] and wait one second for finish", agentConfig.Name)
		cancelFunc()
		time.Sleep(1 * time.Second)
	}()

	agent, err := server.CreateAgent(ctx, agentConfig.Name, agentConfig.Password, agentConfig.Line)
	if err != nil {
		log.Errorf("problem create agent object for agent [%s] with error - %s", agent.LoginName, err)
		return ExitAgentCreate
	}

	err = agent.StartXmpp()
	if err != nil {
		log.Errorf("problem start XMPP listener for agent [%s] with error - %s", agent.LoginName, err)
		return ExitXmppProblem
	}

	log.Tracef("start process operation for agent [%s]", agent.LoginName)

	switch agentConfig.Command {
	case "status":
		log.Debugf("collect status for agent [%s]", agentConfig.Name)
		_, err = agent.GetStatus()
		if err != nil {
			ret = agentErrorMessage(fmt.Sprintf("problem collect agent actual status for agent [%s]", agentConfig.Name))
			break
		}
		agentFinalState(agent, "", "")
		break
	case "login":
		log.Debugf("LOGIN agent [%s]", agentConfig.Name)
		opErr = agent.Login()
		if opErr.Error != nil {
			ret = agentErrorMessage(fmt.Sprintf("problem log in agent [%s]", agentConfig.Name), "agent not logged in")
		} else {
			agentFinalState(agent, "login success, but ", "")
		}
		break
	case "ready":
		log.Debugf("set READY agent [%s]", agentConfig.Name)
		opErr = agent.Ready(serverConfig.Force)
		if opErr.Error == nil {
			agentFinalState(agent, "agent ready success, but ", "")
		} else {
			ret = agentErrorMessage(fmt.Sprintf("problem set agent [%s] to ready state", agentConfig.Name))
		}
		break
	case "not-ready":
		log.Debugf("set NOT-READY agent [%s]", agentConfig.Name)
		opErr = agent.NotReady()
		if opErr.Error == nil {
			agentFinalState(agent, "NOT-READY success, but ", "")
		} else {
			ret = agentErrorMessage(fmt.Sprintf("problem set agent to not-ready state [%s]", agentConfig.Name))
		}
		break
	case "logout":
		log.Debugf("logout agent [%s]", agentConfig.Name)
		opErr = agent.Logout(serverConfig.Force)
		if opErr.Error == nil {
			agentFinalState(agent, "LOGOUT success, but ", "")
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

func agentFinalState(agent *api.Agent, prefixError string, suffixError string) {
	status := agent.GetLastStatus()
	o := fmt.Sprintf("agent [%s] actual status [%s], pending state [%s]", agentConfig.Name, status.State, status.PendingState)
	log.Info(o)
	fmt.Println(o)
}

func agentErrorMessage(msg string, o ...string) int {
	log.Errorf(msg)
	if len(o) > 0 {
		fmt.Println(o[0])
	} else {
		fmt.Println(msg)
	}
	return ExitAgentCommandProblem
}
