package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sim/mist"
	"sim/syslog"
	"sync"
	"syscall"
	"time"
)

var (
	deviceConf = flag.String("c", "./conf/4c9614c95000.json", ".conf files with mist internal config response content")
	mac        = flag.String("m", "", "device mac for config to be fetched from api internal/devices/<mac>/config")
	parallel   = flag.Int("p", 1, "Number of parallel syslog stream connections to create")
	interval   = flag.Int("i", 10, "Interval between syslogs")
)

func main() {
	var mistCert mist.SecurityLog
	var err error
	if *mac != "" {
		mistCert, err = mist.GetDeviceCerts(*mac)
		if err != nil {
			log.Printf("Get Cert Config %s error  %v", *mac, err)
			return
		}
	} else {
		f, err := os.Open(*deviceConf)
		if err != nil {
			log.Printf("Error opening Config file %s error  %v", *deviceConf, err)
			return
		}
		defer f.Close()
		content, err := io.ReadAll(f)
		if err != nil {
			log.Printf("Error reading Config file %s error  %v", *deviceConf, err)
			return
		}
		var papiconfig mist.PAPIConfigResponse
		err = json.Unmarshal(content, &papiconfig)
		if err != nil {
			log.Printf("Error parsing Config file %s error  %v", *deviceConf, err)
			return
		}
		mac = &papiconfig.ID
		mistCert = papiconfig.SecurityLog
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-sigs
		log.Printf("🐼 EXITING on signal %s :frowning:", s)
		os.Exit(1)
	}()

	var wg sync.WaitGroup
	for i := 0; i < *parallel; i++ {
		log.Printf("Strating thread %d", i)
		wg.Add(1)
		go func(count int) {
			defer wg.Done()
			run(mistCert, *mac, count)
		}(i)
	}
	wg.Wait()
}

func run(mistCert mist.SecurityLog, mac string, t int) {
start:
	log.Printf("Connecting %s", mac)
	tlsConf, err := syslog.GetTLSServerConfig(mistCert.ServerCACerts, mistCert.ClientCACerts, mistCert.ClientCert, mistCert.ClientKey)
	if err != nil {
		log.Printf("Get TLS Config %s error  %v", mac, err)
		return
	}
	hostPort := net.JoinHostPort(mistCert.Host, fmt.Sprint(mistCert.Port))
	//hostPort := net.JoinHostPort("localhost", fmt.Sprint(55514))
	conn, err := syslog.Connect(hostPort, tlsConf)
	if err != nil {
		log.Printf("Connect %s error  %v", mac, err)
		return
	}
	defer conn.Close()
	//Contains APPTRACK_SESSION_CLOSE and RT_FLOW_SESSION_CLOSE events
	rawData := "NDMyIAzgCmIVAASBZOSe1wAWQ2xvc2VkIGJ5IGp1bm9zLWR5bmFwcAAArBADAaJuABfWrSgAUAAAAAAACmp1bm9zLWh0dHAAAAoNAQIAAAX3ABfWrSgAAABQAAAAAAALc291cmNlIHJ1bGUAABNzcG9rZS1ndWVzdF90b193YW4xAAADTi9BAAADTi9BAAAAAAYAFzAyX2ludGVybmV0LWd1ZXN0LWJsb2NrAAALc3Bva2UtZ3Vlc3QAAAR3YW4xAAAAAAMAAUgIAAAAAAAAAAkAAAAAAAACZQAAAAAAAAABAAAAAAAAADwAAAADAARIVFRQAAALU1RFQU0tU1RPUkUAAANOL0EAAANOL0EAAApnZS0wLzAvMy4wAAACTm8AAAZHYW1pbmcAAAlQcm90b2NvbHMAAAAAAwAoTG9zcyBvZiBQcm9kdWN0aXZpdHk7QmFuZHdpZHRoIENvbnN1bWVyOwAAAk5BAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAVOQSBOQQAAA04vQQAAA04vQQAAA09mZgAABHJvb3QAAAAAAAADTi9BAAADTi9BADM3MyALwApiFQAEgWTkntcAFkNsb3NlZCBieSBqdW5vcy1keW5hcHAAAKwQAwEAAKJuABfWrSgAAABQAApqdW5vcy1odHRwAAAESFRUUAAAC1NURUFNLVNUT1JFAAAKDQECAAAF9wAX1q0oAAAAUAATc3Bva2UtZ3Vlc3RfdG9fd2FuMQAAA04vQQAAAAAGABcwMl9pbnRlcm5ldC1ndWVzdC1ibG9jawAAC3Nwb2tlLWd1ZXN0AAAEd2FuMQAAAAADAAFICAABOQAAAAAAAAACZQABMQAAAAAAAAAAPAAAAAMAA04vQQAAA04vQQAAAk5vAAAPc3Bva2UtZ3Vlc3RfMDAxAAACMDEAAA1hcGJyX3ByaS13YW4xAAAKZ2UtMC8wLzEuMAAAA04vQQAAAAAAAAAAAAAAAAAAAAAAAAZHYW1pbmcAAAlQcm90b2NvbHMAAAIwMgAAA04vQQAAA04vQQAAA04vQQAAA04vQQAAB2RlZmF1bHQA"
	for {
		log.Printf("Sending data for %s thread %d", mac, t)
		//err := syslog.Send(conn)
		_, err := syslog.SendEncoded(rawData, conn)
		timeout := *interval
		if err != nil {
			log.Printf("Send %s error  %v", mac, err)
			goto start
		}
		time.Sleep(time.Duration(int(time.Second) * timeout))
	}

}
