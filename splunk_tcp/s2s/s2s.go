// Package s2s is a client implementation of the Splunk to Splunk protocol in Golang
package s2s

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"encoding/binary"
)

// S2S sends data to Splunk using the Splunk to Splunk protocol
type S2S struct {
	buf                *bufio.Writer
	conn               net.Conn
	initialized        bool
	Server             string `yaml:"s2sServer"`
	closed             bool
	sent               int64
	bufferBytes        int
	Tls                bool   `yaml:"tls"`
	Cert               string `yaml:"cert"`
	ServerName         string `yaml:"serverName"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	LastSend           time.Time
	EndPoints          []EndPoint `yaml:"endPoints"` // List of EndPoints
}
type EndPoint struct {
	Alias  string         `yaml:"alias"`
	Config EndPointConfig `yaml:"config"`
}

type EndPointConfig struct {
	Index      string `yaml:"index"`
	SourceType string `yaml:"sourcetype"`
}

type splunkSignature struct {
	signature  [128]byte
	ServerName [256]byte
	mgmtPort   [16]byte
}

func NewS2S(server string, bufferBytes int, EndPoints []EndPoint) (*S2S, error) {
	return NewS2STls(server, bufferBytes, EndPoints, false, "", "", false)
}

func NewS2STls(server string, bufferBytes int, EndPoints []EndPoint, Tls bool, Cert string, ServerName string, InsecureSkipVerify bool) (*S2S, error) {
	st := new(S2S)
	st.Server = server
	st.EndPoints = EndPoints
	st.bufferBytes = bufferBytes
	st.Tls = Tls
	st.Cert = Cert
	if ServerName == "" {
		st.ServerName = "SplunkServerDefaultCert"
	} else {
		st.ServerName = ServerName
	}
	st.InsecureSkipVerify = InsecureSkipVerify

	err := st.newBuf()
	if err != nil {
		return nil, err
	}
	err = st.sendSig()
	if err != nil {
		return nil, err
	}
	st.initialized = true
	return st, nil
}
func (st *S2S) Connect() {
	if st.ServerName == "" {
		st.ServerName = "SplunkServerDefaultCert"
	}
	err := st.newBuf()
	if err != nil {
		fmt.Println(err)
	}
	err = st.sendSig()
	if err != nil {
		fmt.Println(err)
	}
	st.initialized = true
}

func (st *S2S) connect(server string) error {
	var err error
	if st.Tls {
		config := &tls.Config{
			InsecureSkipVerify: st.InsecureSkipVerify,
			ServerName:         st.ServerName,
		}
		if len(st.Cert) > 0 {
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM([]byte(st.Cert))
			if !ok {
				return fmt.Errorf("Failed to parse root Certificate")
			}
			config.RootCAs = roots
		}

		st.conn, err = tls.Dial("tcp", server, config)
		return err
	}
	st.conn, err = net.DialTimeout("tcp", server, 2*time.Second)
	return err
}

func (st *S2S) sendSig() error {
	serverParts := strings.Split(st.Server, ":")
	if len(serverParts) != 2 {
		return fmt.Errorf("server malformed.  Should look like server:port")
	}
	ServerName := serverParts[0]
	mgmtPort := serverParts[1]
	var sig splunkSignature
	copy(sig.signature[:], "--splunk-cooked-mode-v2--")
	copy(sig.ServerName[:], ServerName)
	copy(sig.mgmtPort[:], mgmtPort)
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.BigEndian, sig.signature)
	binary.Write(buf, binary.BigEndian, sig.ServerName)
	binary.Write(buf, binary.BigEndian, sig.mgmtPort)
	st.buf.Write(buf.Bytes())
	return nil
}

func (st *S2S) Send() (int64, error) {
	st.LastSend = time.Now()
	return st.Copy(BUF)
}

// Copy takes a io.Reader and copies it to Splunk, needs to be encoded by EncodeEvent
func (st *S2S) Copy(r io.Reader) (int64, error) {
	bytes, err := io.Copy(st.buf, r)
	if err != nil {
		return 0, err
	}

	st.sent += bytes
	// st.sent += bytes
	if st.sent > int64(st.bufferBytes) {
		err := st.buf.Flush()
		if err != nil {
			return 0, err
		}
		st.newBuf()
		st.sent = 0
	}
	return bytes, nil
}

// Close disconnects from Splunk
func (st *S2S) Close() error {
	if !st.closed {
		err := st.buf.Flush()
		if err != nil {
			return err
		}
		err = st.conn.Close()
		if err != nil {
			return err
		}
		st.closed = true
	}
	return nil
}

func (st *S2S) newBuf() error {
	err := st.connect(st.Server)
	if err != nil {
		return err
	}
	st.buf = bufio.NewWriter(st.conn)
	return nil
}

func (st *S2S) Add(event interface{}, endpoint string) {
	raw, _ := InterfaceToString(event)
	position := 0
	for i, endpt := range st.EndPoints {
		if endpt.Alias == endpoint {
			position = i
			break
		}
		if i == len(st.EndPoints)-1 {
			panic("Endpoint not found")
		}
	}
	eventData := map[string]string{
		"_raw":       raw,
		"index":      st.EndPoints[position].Config.Index,
		"sourcetype": st.EndPoints[position].Config.SourceType,
	}
	EncodeEvent(eventData, BUF)
	fmt.Println("Size: " + fmt.Sprint(len(BUF.Bytes())) + "\n")
}

func (st *S2S) AutoPush(During time.Duration) {
	go func() {
		for {
			currentTime := time.Now()
			if currentTime.After(st.LastSend.Add(During)) {
				st.Send()
				fmt.Println("auto pushed at: " + During.String())
			}
			// sleep 60 seconds
			time.Sleep(During)
		}
	}()

}
