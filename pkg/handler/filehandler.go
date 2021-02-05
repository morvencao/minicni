package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/morvencao/mccni/pkg/args"
	"github.com/morvencao/mccni/pkg/nettool"

	"github.com/containernetworking/plugins/pkg/ns"
)

type FileHandler struct {
	IPStore string
}

func NewFileHandler(filename string) Handler {
	return &FileHandler{IPStore: filename}
}

func (fh *FileHandler) HandleAdd(cmdArgs *args.CmdArgs) error {
	fmt.Println(cmdArgs.ContainerID)
	fmt.Println(cmdArgs.Netns)
	fmt.Println(cmdArgs.IfName)
	fmt.Println(cmdArgs.Path)
	fmt.Println(cmdArgs.Args)
	fmt.Println(cmdArgs.StdinData)
	cniConfig := args.CNIConfiguration{}
	if err := json.Unmarshal(cmdArgs.StdinData, &cniConfig); err != nil {
		return err
	}
	allIPs, err := nettool.GetAllIPs(cniConfig.Subnet)
	if err != nil {
		return err
	}
	gwIP := allIPs[0]

	// open or create the file that stores all the reserved IPs
	f, err := os.OpenFile(fh.IPStore, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file that stores reserved IPs %v", err)
	}
	defer f.Close()

	// get all the reserved IPs from file
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	reservedIPs := strings.Split(strings.TrimSpace(string(content)), "\n")

	podIP := ""
	for _, ip := range allIPs[1:] {
		reserved := false
		for _, rip := range reservedIPs {
			if ip == rip {
				reserved = true
				break
			}
		}
		if !reserved {
			podIP = ip
			reservedIPs = append(reservedIPs, podIP)
			break
		}
	}
	if podIP == "" {
		return fmt.Errorf("No IP available!")
	}

	// Create or update bridge
	brName := cniConfig.Bridge
	if brName != "" {
		// fall back to default bridge name: mccni0
		brName = "mccni0"
	}
	mtu := cniConfig.MTU
	if mtu == 0 {
		// fall back to default MTU: 1500
		mtu = 1500
	}
	br, err := nettool.CreateOrUpdateBridge(brName, gwIP, mtu)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(cmdArgs.Netns)
	if err != nil {
		return err
	}

	if err := nettool.SetupVeth(netns, br, cmdArgs.IfName, podIP, mtu); err != nil {
		return err
	}

	// write reserved IPs back into file
	if err := ioutil.WriteFile(fh.IPStore, []byte(strings.Join(reservedIPs, "\n")), 0644); err != nil {
		fmt.Errorf("failed to write reserved IPs into file: %v", err)
	}

	return nil
}

func (fh *FileHandler) HandleDel(cmdArgs *args.CmdArgs) error {
	fmt.Println(cmdArgs.ContainerID)
	fmt.Println(cmdArgs.Netns)
	fmt.Println(cmdArgs.IfName)
	fmt.Println(cmdArgs.Path)
	fmt.Println(cmdArgs.Args)
	fmt.Println(cmdArgs.StdinData)
	netns, err := ns.GetNS(cmdArgs.Netns)
	if err != nil {
		return err
	}
	ip, err := nettool.GetVethIPInNS(netns, cmdArgs.IfName)
	if err != nil {
		return err
	}

	// open or create the file that stores all the reserved IPs
	f, err := os.OpenFile(fh.IPStore, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file that stores reserved IPs %v", err)
	}
	defer f.Close()

	// get all the reserved IPs from file
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	reservedIPs := strings.Split(strings.TrimSpace(string(content)), "\n")

	for i, rip := range reservedIPs {
		if rip == ip {
			reservedIPs = append(reservedIPs[:i], reservedIPs[i+1:]...)
			break
		}
	}

	// write reserved IPs back into file
	if err := ioutil.WriteFile(fh.IPStore, []byte(strings.Join(reservedIPs, "\n")), 0644); err != nil {
		fmt.Errorf("failed to write reserved IPs into file: %v", err)
	}

	return nil
}

func (fh *FileHandler) HandleCheck(cmdArgs *args.CmdArgs) error {
	// to br implemented
	return nil
}

func (fh *FileHandler) HandleVersion(cmdArgs *args.CmdArgs) error {
	// to br implemented
	return nil
}
