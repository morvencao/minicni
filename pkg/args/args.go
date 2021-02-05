package args

import (
	"fmt"
	"io/ioutil"
	"os"
)

const (
	CommandEnvKey     string = "CNI_COMMAND"
	ContainerIDEnvKey string = "CNI_CONTAINERID"
	NetnsEnvKey       string = "CNI_NETNS"
	IfNameEnvKey      string = "CNI_IFNAME"
	PathEnvKey        string = "CNI_PATH"
	ArgsEnvKey        string = "CNI_ARGS"
)

const (
	AddCmd     string = "ADD"
	DelCmd     string = "DEL"
	CheckCmd   string = "CHECK"
	VersionCmd string = "VERSION"
)

type CmdEnv struct {
	CmdArgKey   string
	CmdArgValue string
	ReqForCmd   map[string]bool
}

type CmdArgs struct {
	ContainerID string
	Netns       string
	IfName      string
	Path        string
	Args        string
	StdinData   []byte
}

type CNIConfiguration struct {
	CniVersion string `json:"cniVersion"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Bridge     string `json:"bridge"`
	MTU        int    `json:"mtu"`
	Network    string `json:"network"`
	Subnet     string `json:"subnet"`
}

func GetArgsFromEnv() (string, *CmdArgs, error) {
	var cmd, conID, netns, ifName, path, args string
	cmd = os.Getenv(CommandEnvKey)
	if cmd == "" {
		fmt.Fprintf(os.Stderr, "%s environment variable is missing!", CommandEnvKey)
		return "", nil, fmt.Errorf("%s environment variable is missing!", CommandEnvKey)
	}
	var cmdEnvs = []CmdEnv{
		{
			ContainerIDEnvKey,
			conID,
			map[string]bool{
				AddCmd:     true,
				DelCmd:     true,
				CheckCmd:   true,
				VersionCmd: false,
			},
		},
		{
			NetnsEnvKey,
			netns,
			map[string]bool{
				AddCmd:     true,
				DelCmd:     false,
				CheckCmd:   true,
				VersionCmd: false,
			},
		},
		{
			IfNameEnvKey,
			ifName,
			map[string]bool{
				AddCmd:     true,
				DelCmd:     false,
				CheckCmd:   true,
				VersionCmd: false,
			},
		},
		{
			PathEnvKey,
			path,
			map[string]bool{
				AddCmd:     false,
				DelCmd:     false,
				CheckCmd:   false,
				VersionCmd: false,
			},
		},
		{
			ArgsEnvKey,
			args,
			map[string]bool{
				AddCmd:     false,
				DelCmd:     false,
				CheckCmd:   false,
				VersionCmd: false,
			},
		},
	}
	argsMissing := false
	for _, v := range cmdEnvs {
		v.CmdArgValue = os.Getenv(v.CmdArgKey)
		if v.CmdArgValue == "" && v.ReqForCmd[cmd] == true {
			fmt.Fprintf(os.Stderr, "%s environment variable is missing!", v.CmdArgKey)
			argsMissing = true
		}
	}
	if argsMissing {
		return "", nil, fmt.Errorf("required environment variable is missing!")
	}

	stdinData, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", nil, fmt.Errorf("error reading from stdin: %v", err)
	}
	cmdArgs := &CmdArgs{
		ContainerID: conID,
		Netns:       netns,
		IfName:      ifName,
		Path:        path,
		Args:        args,
		StdinData:   stdinData,
	}

	return cmd, cmdArgs, nil
}
