// This package provides a ligth client for syncthing, for when you need to upload/download but don't want a full sync.
// For now this repo just specifies the API for commenting.
package syncthingclient

type Client struct {
	Remote Remote
}

// Upload uploades a stream to the remote filename specified
func (c *Client) Upload(stream io.Writer, filename string) error {
}

// Download retrives the specified filename and streams it
func (c *Client) Download(stream io.Reader, filename string) error {
}

// GetFileList retrieves a file list for the folder
func (c *Client) GetFileList(folder string) {
}

// GetSharedFolders retrieves the folders that are shared with this client
func (c *Client) GetSharedFolders() {
}

// AddRemote adds a new remote syncthing client
func (c *Client) AddRemote(deviceID string) error {
}

type Remote struct {
	conn     net.Conn
	DeviceID protocol.DeviceID
	Host     string
}

// SetHost sets the remote host name host:port format
func (r *Remote) SetHost(Host string) {
	r.Host = Host
}
