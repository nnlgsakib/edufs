package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/urfave/cli/v2"
)

var shellInstance *shell.Shell
var gatewayURL = "http://api-ipfs.web3twenty.com:3002/ipfs/"

func main() {
	app := &cli.App{
		Name:  "eduFs",
		Usage: "IPFS file and folder management application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "node",
				Value: "http://api-ipfs.web3twenty.com:3001",
				Usage: "IPFS node API URL",
			},
		},
		EnableBashCompletion: true,
		HideHelp:             false,
	}

	app.Before = func(c *cli.Context) error {
		nodeURL := c.String("node")
		shellInstance = shell.NewShell(nodeURL)
		return nil
	}

	app.Commands = []*cli.Command{
		{
			Name:  "status",
			Usage: "Get the status of your IPFS node",
			Action: func(c *cli.Context) error {
				return getStatus()
			},
		},
		{
			Name:  "add",
			Usage: "Add a file or folder to IPFS and pin it to the node",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the file or folder to add",
				},
			},
			Action: func(c *cli.Context) error {
				return addFileOrFolder(c.String("path"))
			},
		},
		{
			Name:  "cat",
			Usage: "Retrieve a file from IPFS by CID",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "cid",
					Usage: "Content Identifier (CID) of the file to retrieve",
				},
			},
			Action: func(c *cli.Context) error {
				return retrieveFile(c.String("cid"))
			},
		},
		{
			Name:  "publish",
			Usage: "Publish a directory on IPFS and IPNS",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the directory to publish",
				},
				&cli.StringFlag{
					Name:  "ipns-key",
					Usage: "IPNS key for the website (optional)",
				},
			},
			Action: func(c *cli.Context) error {
				return publishDirectory(c.String("path"), c.String("ipns-key"))
			},
		},
		{
			Name:  "download",
			Usage: "Download a file from IPFS",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "cid",
					Usage: "Content Identifier (CID) of the file to download",
				},
				&cli.StringFlag{
					Name:  "output",
					Usage: "Path to save the downloaded file",
				},
			},
			Action: func(c *cli.Context) error {
				return downloadFile(c.String("cid"), c.String("output"))
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getStatus() error {
	nodeInfo, err := shellInstance.ID()
	if err != nil {
		return fmt.Errorf("error getting node status: %v", err)
	}
	fmt.Println("Connected to IPFS node:")
	fmt.Println("Node ID:", nodeInfo.ID)
	fmt.Println("Agent Version:", nodeInfo.AgentVersion)
	fmt.Println("Protocol Version:", nodeInfo.ProtocolVersion)
	return nil
}

func addFileOrFolder(path string) error {
	if path == "" {
		return fmt.Errorf("please provide a file or folder path using --path")
	}

	cid, err := addPathToIPFS(path)
	if err != nil {
		return fmt.Errorf("error adding file or folder to IPFS: %v", err)
	}

	fmt.Printf("File or folder added and pinned to IPFS.\nCID: %s\nURL: %s%s\n", cid, gatewayURL, cid)

	return nil
}

func retrieveFile(cid string) error {
	if cid == "" {
		return fmt.Errorf("please provide a CID using --cid")
	}

	reader, err := shellInstance.Cat(cid)
	if err != nil {
		return fmt.Errorf("error retrieving file from IPFS: %v", err)
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading data: %v", err)
	}

	fmt.Print(string(data))
	return nil
}

func publishDirectory(dirPath, ipnsKey string) error {
	if dirPath == "" {
		return fmt.Errorf("please provide a directory path using --path")
	}

	// Add the entire directory to IPFS
	cid, err := addFolderToIPFS(dirPath)
	if err != nil {
		return fmt.Errorf("error adding directory to IPFS: %v", err)
	}

	// If an IPNS key is provided, publish via IPNS
	if ipnsKey != "" {
		err := shellInstance.Publish(ipnsKey, cid)
		if err != nil {
			return fmt.Errorf("error publishing to IPNS: %v", err)
		}
		fmt.Printf("Directory published successfully via IPNS!\nIPNS Key: %s\nIPNS Link: %s%s\n", ipnsKey, gatewayURL, ipnsKey)
		return nil
	}

	fmt.Printf("Directory published successfully via IPFS!\nIPFS Link: %s%s\n", gatewayURL, cid)
	return nil
}

func addFolderToIPFS(folderPath string) (string, error) {
	res, err := shellInstance.AddDir(folderPath)
	if err != nil {
		return "", err
	}
	return res, nil
}

func downloadFile(cid, outputFilePath string) error {
	if cid == "" {
		return fmt.Errorf("please provide a CID using --cid")
	}
	if outputFilePath == "" {
		return fmt.Errorf("please provide an output file path using --output")
	}

	reader, err := shellInstance.Cat(cid)
	if err != nil {
		return fmt.Errorf("error retrieving file from IPFS: %v", err)
	}
	defer reader.Close()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading data: %v", err)
	}

	err = ioutil.WriteFile(outputFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("error saving file: %v", err)
	}

	fmt.Printf("File downloaded and saved to %s\n", outputFilePath)
	return nil
}
func addPathToIPFS(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return addFolderToIPFS(path)
	}

	return addFileToIPFS(path)
}

// func addFolderToIPFS(folderPath string) (string, error) {
// 	res, err := shellInstance.AddDir(folderPath)
// 	if err != nil {
// 		return "", err
// 	}
// 	return res, nil
// }

func addFileToIPFS(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	// Create a progress bar
	bar := pb.Full.Start64(fileInfo.Size())
	bar.Set(pb.Bytes, true)

	// Create a reader with a proxy reader which updates the progress bar
	reader := bar.NewProxyReader(file)

	res, err := shellInstance.Add(reader)
	if err != nil {
		return "", err
	}

	// Pin the file to the IPFS node
	err = shellInstance.Pin(res)
	if err != nil {
		return "", err
	}

	bar.Finish()
	return res, nil
}
