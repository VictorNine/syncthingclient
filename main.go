// This package provides a ligth client for syncthing, for when you need to upload/download but don't want a full sync.
package syncthingclient

import (
	"crypto/tls"
	"errors"
	"github.com/syncthing/syncthing/lib/protocol"
	"log"
	"os"
	"strings"
)

type Client struct {
	Remote   Remote            // The remote syncthing instance
	Cert     tls.Certificate   // Your certificate
	DeviceID protocol.DeviceID // Your device ID
}

// LoadCertificate loads the syncthing certificate from file
func (c *Client) LoadCertificate(certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err == nil {
		c.Cert = cert
	}

	return err
}

// Upload uploades the "from" file to the remote "to" path specified
func (c *Client) Upload(from, to string) error {
	c.Remote.connect(c.Cert)

	return errors.New("Not implemented")
}

// Download retrieves the specified file and stores it in "to"
func (c *Client) Download(from, to string) error {
	c.Remote.Callback.IndexRecived = make(chan bool)

	root := from[:strings.Index(from, "/")]
	file := from[strings.Index(from, "/")+1:]

	c.Remote.connect(c.Cert)
	defer c.Remote.close()

	log.Println("Sending cluster config")

	cc, err := c.Remote.MakeClusterConfig(root)
	if err != nil {
		return err
	}

	err = c.Remote.Send(cc)
	if err != nil {
		return err
	}

	// Wait for the callback
	<-c.Remote.Callback.IndexRecived

	fileIndex := -1
	for i, f := range c.Remote.Index.Files {
		if f.Name == file {
			fileIndex = i
			break
		}
	}

	if fileIndex == -1 {
		return errors.New("File not found on remote")
	}

	log.Printf("Downloading %v blocks", len(c.Remote.Index.Files[fileIndex].Blocks))

	c.Remote.Callback.ResponseRecived = make(chan *protocol.Response)
	// Create the file
	f, err := os.Create(to)
	if err != nil {
		return err
	}

	var id int32 = 0
	for _, b := range c.Remote.Index.Files[fileIndex].Blocks {
		req, err := MakeRequst(id, root, file, b.Offset, b.Size, b.Hash)
		if err != nil {
			return err
		}

		err = c.Remote.Send(req)
		if err != nil {
			return err
		}

		// Wait for response
		res := <-c.Remote.Callback.ResponseRecived

		_, err = f.Write(res.Data)
		if err != nil {
			return err
		}

		id++
	}

	f.Close()

	return nil
}

type FileInfo struct {
	Name string
	Size int64
	Type string
}

// GetFileList retrieves a file list for the folder.
func (c *Client) GetFileList(folder string) ([]FileInfo, error) {
	c.Remote.Callback.IndexRecived = make(chan bool)

	c.Remote.connect(c.Cert)
	defer c.Remote.close()

	log.Println("Sending cluster config")

	cc, err := c.Remote.MakeClusterConfig(folder)
	if err != nil {
		return nil, err
	}

	err = c.Remote.Send(cc)
	if err != nil {
		return nil, err
	}

	// Wait for the callback
	<-c.Remote.Callback.IndexRecived

	fileList := make([]FileInfo, len(c.Remote.Index.Files))

	for i, f := range c.Remote.Index.Files {
		fileList[i] = FileInfo{
			Name: folder + "/" + f.Name,
			Size: f.Size,
		}

		switch f.Type {
		case 0:
			fileList[i].Type = "FILE"
		case 1:
			fileList[i].Type = "DIRECTORY"
		default:
			fileList[i].Type = "SYMLINK"
		}
	}

	// Remote the callback
	c.Remote.Callback.IndexRecived = nil

	return fileList, nil
}

// GetSharedFolders retrieves the folders that are shared with this client
func (c *Client) GetSharedFolders() ([]string, error) {
	c.Remote.Callback.CCrecived = make(chan bool)

	c.Remote.connect(c.Cert)
	defer c.Remote.close()

	log.Println("Sending cluster config")

	cc, err := c.Remote.MakeClusterConfig("")
	if err != nil {
		return nil, err
	}

	err = c.Remote.Send(cc)
	if err != nil {
		return nil, err
	}

	// Wait for the callback
	<-c.Remote.Callback.CCrecived

	folders := make([]string, len(c.Remote.ClusterConfig.Folders))

	for i, f := range c.Remote.ClusterConfig.Folders {
		folders[i] = f.ID
	}

	// Remote the callback
	c.Remote.Callback.CCrecived = nil

	return folders, nil
}

// AddRemote adds a new remote syncthing client
func (c *Client) AddRemote(deviceID string) error {
	id, err := protocol.DeviceIDFromString(deviceID)
	if err != nil {
		return err
	}

	c.Remote = Remote{
		DeviceID: id,
		Callback: Callback{},
	}

	return nil
}
