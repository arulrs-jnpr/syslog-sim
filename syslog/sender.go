package syslog

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"
)

func GetTLSServerConfig(serverCa []string, clientCa []string, clientCert string, clientKey string) (*tls.Config, error) {
	serverCACerts := x509.NewCertPool()
	for _, c := range serverCa {
		ok := serverCACerts.AppendCertsFromPEM([]byte(c))
		if !ok {
			return nil, errors.New("cannot create client CA cert pool")
		}
	}
	clientCACerts := x509.NewCertPool()
	for _, c := range clientCa {
		ok := clientCACerts.AppendCertsFromPEM([]byte(c))
		if !ok {
			return nil, errors.New("cannot create client CA cert pool")
		}
	}
	clientCerts, err := tls.X509KeyPair([]byte(clientCert), []byte(clientKey)) // Load Server private key and load certificates to send it to client.
	if err != nil {
		return nil, fmt.Errorf("cannot load server server key/certificate: %s", err)
	}
	return &tls.Config{
		ServerName:            "4c9614c95000",
		VerifyConnection:      func(cs tls.ConnectionState) error { return nil },
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error { return nil },
		Certificates:          []tls.Certificate{clientCerts},
		ClientAuth:            tls.RequireAnyClientCert,
		InsecureSkipVerify:    true,
		ClientCAs:             clientCACerts,
		RootCAs:               serverCACerts,
		MinVersion:            tls.VersionTLS12,
		CipherSuites: []uint16{
			/* TLS 1.2 */
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}, nil
}

func Connect(hostPort string, tlsConf *tls.Config) (*tls.Conn, error) {
	conn, err := tls.Dial("tcp4", hostPort, tlsConf)
	if err != nil {
		log.Printf("TLS conn error %v", err)
		return nil, err
	}
	return conn, nil
}

const (
	bufSize     = 4096
	readBufSize = 8192
)

func setupMsg() []byte {
	// This is real message sent by my SRX300
	s := `
		33 34 33 20 0b c0 0a 62 15 00 04 85 62 21 47 2e
		00 13 43 6c 6f 73 65 64 20 62 79 20 6a 75 6e 6f
		73 2d 61 6c 67 00 00 c0 a8 0a 02 00 00 e7 39 00
		01 01 01 01 00 00 00 35 00 0d 6a 75 6e 6f 73 2d
		64 6e 73 2d 75 64 70 00 00 03 44 4e 53 00 00 07
		55 4e 4b 4e 4f 57 4e 00 00 c0 a8 02 b3 00 00 6d
		ab 00 01 01 01 01 00 00 00 35 00 0f 73 6f 75 72
		63 65 2d 6e 61 74 2d 72 75 6c 65 00 00 03 4e 2f
		41 00 00 00 00 11 00 10 74 72 75 73 74 2d 74 6f
		2d 75 6e 74 72 75 73 74 00 00 05 74 72 75 73 74
		00 00 07 75 6e 74 72 75 73 74 00 00 00 00 02 00
		00 43 2c 00 01 31 00 00 00 00 00 00 00 00 4a 00
		01 31 00 00 00 00 00 00 00 00 4a 00 00 00 04 00
		03 4e 2f 41 00 00 03 4e 2f 41 00 00 02 4e 6f 00
		00 03 4e 2f 41 00 00 03 4e 2f 41 00 00 07 64 65
		66 61 75 6c 74 00 00 0a 67 65 2d 30 2f 30 2f 30
		2e 30 00 00 03 4e 2f 41 00 00 00 00 00 00 00 00
		00 00 00 00 00 00 00 00 00 00 0e 49 6e 66 72 61
		73 74 72 75 63 74 75 72 65 00 00 0a 4e 65 74 77
		6f 72 6b 69 6e 67 00 00 03 4e 2f 41 00 00 03 4e
		2f 41 00 00 03 4e 2f 41 00 00 03 4e 2f 41 00 00
		03 4e 2f 41 00 00 03 4e 2f 41 00`
	reg, _ := regexp.Compile("[^0-9a-fA-F]+")
	msg, _ := hex.DecodeString(reg.ReplaceAllString(s, ""))
	return msg
}

func Send(conn *tls.Conn) error {
	msg := setupMsg()
	size := (readBufSize/len(msg) + 1) * len(msg)
	msgs := make([]byte, size)
	for offset := 0; offset < size; offset += len(msg) {
		copy(msgs[offset:], msg)
	}

	for j := 0; j < size/len(msg); j++ {
		i := 0
		for ; i < 4; i++ {
			if _, err := conn.Write(msg[i : i+1]); err != nil {
				return err
			}
			time.Sleep(1 * time.Microsecond)
		}
		if _, err := conn.Write(msg[i : len(msg)-1]); err != nil {
			return err
		}
		time.Sleep(1 * time.Microsecond)
		if _, err := conn.Write(msg[len(msg)-1:]); err != nil {
			return err
		}
		time.Sleep(1 * time.Microsecond)
	}

	if _, err := conn.Write(msgs); err != nil {
		return err
	}
	return nil
}

func SendEncoded(encodedData string, conn *tls.Conn) (int, error) {
	encodeStr, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return 0, err
	}
	l, err := conn.Write(encodeStr)
	return l, err
}
