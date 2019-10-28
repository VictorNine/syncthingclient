// This package provides a ligth client for syncthing, for when you need to upload/download but don't want a full sync.
// For now this repo just specifies the API for commenting.
package syncthingclient

type Client struct {
	Remote Remote
	Cert   tls.Certificate
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
}

// Download retrives the specified file and stores it in "to"
func (c *Client) Download(from, to string) error {
}

// GetFileList retrieves a file list for the folder.
func (c *Client) GetFileList(folder string) []string {
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
