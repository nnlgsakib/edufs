package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"io/ioutil"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/urfave/cli/v2"
)

var shellInstance *shell.Shell

func main() {
	app := &cli.App{
		Name:  "eduFs",
		Usage: "ipfs office file management application",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "node",
				Value: "/ip4/91.208.92.6/tcp/8000",
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
			Usage: "Add a file to IPFS and pin it to the node",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "path",
					Usage: "Path to the file to add",
				},
			},
			Action: func(c *cli.Context) error {
				filePath := c.String("path")
				if filePath == "" {
					fmt.Println("Please provide a file path using --path")
					return nil
				}

				file, err := os.Open(filePath)
				if err != nil {
					fmt.Println("Error opening file:", err)
					return err
				}
				defer file.Close()

				res, err := shellInstance.Add(file)
				if err != nil {
					fmt.Println("Error adding file to IPFS:", err)
					return err
				}
				err = shellInstance.Pin(res)
				if err != nil {
					fmt.Println("Error pinning file to IPFS:", err)
					return err
				}
				fmt.Printf("File added and pinned to IPFS.\nCID: %s\nURL: https://ipfs.io/ipfs/%s\n", res, res)

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
					file, err := os.Open(filePath)
					if err != nil {
						fmt.Println("Error opening file:", err)
						return err
					}
					defer file.Close()

					res, err := shellInstance.Add(file)
					if err != nil {
						fmt.Println("Error adding file to IPFS:", err)
						return err
					}
					// Pin the file to the IPFS node
					err = shellInstance.Pin(res)
					if err != nil {
						fmt.Println("Error pinning file to IPFS:", err)
						return err
					}
					cids = append(cids, res)
				}

				ipnsKey := c.String("ipns-key")
				if ipnsKey != "" {
					err := shellInstance.Publish(ipnsKey, cids[0])
					if err != nil {
						fmt.Println("Error publishing to IPNS:", err)
						return err
					}
					fmt.Printf("Directory published successfully via IPNS!\nIPNS Key: %s\nIPNS Link: https://ipfs.io/ipns/%s\n", ipnsKey, ipnsKey)
					return nil
				} else {
					fmt.Printf("Directory published successfully via IPFS!\nIPFS Link: https://ipfs.io/ipfs/%s\n", cids[0])
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
