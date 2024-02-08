# eduFs - IPFS File and Folder Management Application

![GitHub](https://img.shields.io/github/license/nnlgsakib/eduFs)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/nnlgsakib/eduFs)

eduFs is an IPFS-based file and folder management application written in Go. It allows you to interact with the InterPlanetary File System (IPFS) to add, retrieve, and manage files and directories.

## Features

- Get the status of your IPFS node.
- Add files or folders to IPFS and pin them to your node.
- Retrieve files from IPFS using Content Identifiers (CIDs).
- Publish directories on IPFS and IPNS.
- Download files from IPFS.



Usage
To use eduFs, install it by clicking this  [installation url](https://ipfs.io/ipfs/QmebZ46pvJfSgG4AMTNnun3qZqpZxco2EzdAqRdSjZ85yd?filename=edufs.exe)

# Get the status of your IPFS node:

```
edufs status 
```

# Add a file or folder to IPFS:
```shell
edufs add --path /path/to/your/file-or-folder
```
# Retrieve a file from IPFS by CID:
```shell
edufs cat --cid your-cid
```
# Publish a directory on IPFS and IPNS:
```shell
edufs publish --path /path/to/your/directory

```

# Download a file from IPFS:

```shell
edufs download --cid your-cid --output /path/to/save/file

```


License
This project is licensed under the MIT License - see the LICENSE file for details.

Acknowledgments
This project uses the go-ipfs-api library for interacting with IPFS.
It also utilizes the urfave/cli library for command-line interface functionality.
Feel free to contribute to this project or report any issues on GitHub!

Author
@nnlgsakib
GitHub: Your GitHub Profile
Happy file and folder management with eduFS!
