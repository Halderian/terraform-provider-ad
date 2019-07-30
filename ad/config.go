package ad

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/ldap.v2"
	"log"
)

type Config struct {
	Domain   string
	IP       string
	Username string
	Password string
	UseSSL   bool
}

// Client() returns a connection for accessing AD services.
func (c *Config) Client() (*ldap.Conn, error) {
	var username string
	username = c.Username + "@" + c.Domain
	adConn, err := clientConnect(c.IP, username, c.Password, c.UseSSL)

	if err != nil {
		return nil, fmt.Errorf("Error while trying to connect active directory server, Check server IP address, username or password: %s", err)
	}
	log.Printf("[DEBUG] AD connection successful for user: %s", c.Username)
	return adConn, nil
}

func clientConnect(ip, username, password string, useSSL bool) (*ldap.Conn, error) {
	var adConn *ldap.Conn
	var err error
	if useSSL {
		adConn, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ip, 636), &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         ip,
		})
	} else {
		adConn, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", ip, 389))
	}

	if err != nil {
		return nil, err
	}

	err = adConn.Bind(username, password)
	if err != nil {
		return nil, err
	}
	return adConn, nil
}
