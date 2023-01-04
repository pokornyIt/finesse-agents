# Finesse Agents

Program for bulk manipulation with CISCO Finesse Agent API. 
The program utilizes the library "github.com/pokornyIt/finesse_api".

## Usage

```shell
finesse-agents
usage: finesse-agents  [<flags>] <command> [<args> ...] 

Flags:                                                                        
  -h, --help          Show context-sensitive help (also try --help-long and   
                      --help-man).                                            
  -s, --server=finesse.server.local                                           
                      finesse server name or IP address                       
  -p, --port=8445     finesse API port                                        
  -f, --force         force operation                                         
  -l, --level=error   define logger level (error, warning, info, debug, trace)
  -D, --directory=""  define alternative directory for store logs             
  -S, --show          only show actual configuration                          
      --version       Show application version.                               
                                                                              
Commands:                                                                     
  help [<command>...]                                                         
    Show help.                                                                
                                                                              
  agent [<flags>]                                                             
    operation with one specific agent                                         
                                                                              
  finesse [<flags>]                                                           
    operation with group of agents 
```

## Single agent operation
Command for setting one agent to ready state with force option.  

- logging level set to **TRACE**
- logging output to **log** directory
- Finesse server **finesse.server.fqdn**
- command for singe user **agent**
- agent name **name1**
- device line **1000**
- agent password **123456**
- operation **ready**
- force

```shell
finesse-agents -l trace -D log -s finesse.server.fqdn agent -a "name1" -n 1000 -P 123456 -o ready -f
```


## Group of agent operations

Before calling, the program in finesse command mode is necessary to prepare the agent CSV file.
Format for agents file:
```csv
Name1,password,1000
Name2,password,1001
Name3,password,1002
```
Program accepting file with header too:
```csv
name,password,line
Name1,password,1000
Name2,password,1001
Name3,password,1002
```

Command for getting actual status of agents.  

- logging level set to **DEBUG**
- logging output to **log** directory
- Finesse server **finesse.server.fqdn**
- command for group of users **finesse**
- agent CSV file **data\agent01.csv**
- operation **get status**

```shell
finesse-agents -l debug -D log -s finesse.server.fqdn agent -a data\agent01.csv -o status
```
