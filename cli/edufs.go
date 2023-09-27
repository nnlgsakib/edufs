package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"io/ioutil"
	//"strings"
	
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/urfave/cli/v2"
)

var shellInstance *shell.Shell
var gatewayURL = "https://gate-ipfs.web3twenty.com/ipfs/"

func main() {
	app := &cli.App{
		Name:  "eduFs",
		Usage: "IPFS file and folder management application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "node",
			    Value: "https://api-ipfs.web3twenty.com", 
				Usage: "IPFS node API URL",
			},
		},
		EnableBashCompletion: true,
		HideHelp:              false,
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
				nodeInfo, err := shellInstance.ID()
				if err != nil {
					fmt.Println("Error getting node status:", err)
					return err
				}
				fmt.Println("Connected to IPFS node:")
				fmt.Println("Node ID:", nodeInfo.ID)
				fmt.Println("Agent Version:", nodeInfo.AgentVersion)
				fmt.Println("Protocol Version:", nodeInfo.ProtocolVersion)
				return nil
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
				path := c.String("path")
				if path == "" {
					fmt.Println("Please provide a file or folder path using --path")
					return nil
				}

				cid, err := addPathToIPFS(path)
				if err != nil {
					fmt.Println("Error adding file or folder to IPFS:", err)
					return err
				}

				fmt.Printf("File or folder added and pinned to IPFS.\nCID: %s\nURL: %s%s\n", cid, gatewayURL, cid)

				return nil
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
				cid := c.String("cid")
				if cid == "" {
					fmt.Println("Please provide a CID using --cid")
					return nil
				}

				reader, err := shellInstance.Cat(cid)
				if err != nil {
					fmt.Println("Error retrieving file from IPFS:", err)
					return err
				}
				defer reader.Close()

				data, err := ioutil.ReadAll(reader)
				if err != nil {
					fmt.Println("Error reading data:", err)
					return err
				}

				fmt.Print(string(data))
				return nil
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
				dirPath := c.String("path")
				if dirPath == "" {
					fmt.Println("Please provide a directory path using --path")
					return nil
				}

				var files []string
				err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						files = append(files, path)
					}
					return nil
				})
				if err != nil {
					fmt.Println("Error scanning directory:", err)
					return err
				}

				// Upload all files to IPFS
				var cids []string
				for _, filePath := range files {
					cid, err := addFileToIPFS(filePath)
					if err != nil {
						return err
					}
					fmt.Printf("File %s added and pinned to IPFS.\nCID: %s\nURL: %s%s\n", filePath, cid, gatewayURL, cid)
					cids = append(cids, cid)
				}

				ipnsKey := c.String("ipns-key")
				if ipnsKey != "" {
					err := shellInstance.Publish(ipnsKey, cids[0])
					if err != nil {
						fmt.Println("Error publishing to IPNS:", err)
						return err
					}
					fmt.Printf("Directory published successfully via IPNS!\nIPNS Key: %s\nIPNS Link: %s%s\n", ipnsKey, gatewayURL, ipnsKey)
					return nil
				} else {
					fmt.Printf("Directory published successfully via IPFS!\nIPFS Link: %s%s\n", gatewayURL, cids[0])
					return nil
				}
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
				cid := c.String("cid")
				if cid == "" {
					fmt.Println("Please provide a CID using --cid")
					return nil
				}
				outputFilePath := c.String("output")
				if outputFilePath == "" {
					fmt.Println("Please provide an output file path using --output")
					return nil
				}

				reader, err := shellInstance.Cat(cid)
				if err != nil {
					fmt.Println("Error retrieving file from IPFS:", err)
					return err
				}
				defer reader.Close()

				data, err := ioutil.ReadAll(reader)
				if err != nil {
					fmt.Println("Error reading data:", err)
					return err
				}

				err = ioutil.WriteFile(outputFilePath, data, 0644)
				if err != nil {
					fmt.Println("Error saving file:", err)
					return err
				}

				fmt.Printf("File downloaded and saved to %s\n", outputFilePath)
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
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

func addFileToIPFS(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	res, err := shellInstance.Add(file)
	if err != nil {
		return "", err
	}

	// Pin the file to the IPFS node
	err = shellInstance.Pin(res)
	if err != nil {
		return "", err
	}

	return res, nil
}

func addFolderToIPFS(folderPath string) (string, error) {
	var files []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files found in folder")
	}

	var cids []string
	for _, filePath := range files {
		cid, err := addFileToIPFS(filePath)
		if err != nil {
			return "", err
		}
		cids = append(cids, cid)
	}

	return cids[0], nil
}
