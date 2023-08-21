package main

import (
	"log"
	"sim/mist"
	"sim/syslog"
	"sync"
	"time"
)

func main() {
	mac := "4c:96:14:c9:50:00"
	mistCert, err := mist.GetDeviceCerts(mac)
	if err != nil {
		log.Printf("Get Cert Config %s error  %v", mac, err)
		return
	}
	var wg sync.WaitGroup
	for i := range [500]int{} {
		log.Printf("Strating thread %d", i)
		wg.Add(1)
		go run(mistCert, mac, i)
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

	conn, err := syslog.Connect(tlsConf)
	if err != nil {
		log.Printf("Connect %s error  %v", mac, err)
		return
	}
	defer conn.Close()
	//rawData := "/gTHCgD+BMIzNTYgC8AKYhUABIFk4i5mAAxpZGxlIFRpbWVvdXQAAKwQAwEAALuOAAgICAgAAAA1AA1qdW5vcy1kbnMtdWRwAAADRE5TAAAHVU5LTk9XTgAACg0BAgAACCsACAgICAAAADUAE3Nwb2tlLWd1ZXN0X3RvX3dhbjEAAANOL0EAAAAAEQANMDFfcHVibGljLWRucwAAC3Nwb2tlLWd1ZXN0AAAEd2FuMQAAAAACAAOQRQABMQAAAAAAAAAASQABMQAAAAAAAAAAdgAAADwAA04vQQAAA04vQQAAAk5vAAALc3Bva2UtZ3Vlc3QAAAIwMQAADWFwYnJfcHJpLXdhbjEAAApnZS0wLzAvMS4wAAADTi9BAAAAAAAAAAAAAAAAAAAAAAAADkluZnJhc3RydWN0dXJlAAAKTmV0d29ya2luZwAAAjAxAAADTi9BAAADTi9BAAADTi9BAAADTi9BAAAHZGVmYXVsdAAzNjcgC8AKYhUABIFk4i5mAAxpZGxlIFRpbWVvdXQAAAoKAwIAAKanAAgICAgAAAA1AA1qdW5vcy1kbnMtdWRwAAADRE5TAAAHVU5LTk9XTgAACg0AAgAABhMACAgICAAAADUAEnNwb2tlLWNvcnBfdG9fd2FuMAAAA04vQQAAAAARABcwMV9pbnRlcm5ldC1jb3JwLW5vLWlkcAAACnNwb2tlLWNvcnAAAAR3YW4wAAAAAAIAA5BHAAExAAAAAAAAAABNAAExAAAAAAAAAACGAAAAPAADTi9BAAADTi9BAAACTm8AAA5zcG9rZS1jb3JwXzAwMQAAAjAxAAANYXBicl9wcmktd2FuMAAACmdlLTAvMC8wLjAAAANOL0EAAAAAAAAAAAAAAAAAAAAAAAAOSW5mcmFzdHJ1Y3R1cmUAAApOZXR3b3JraW5nAAACMDIAAANOL0EAAANOL0EAAANOL0EAAANOL0EAAAdkZWZhdWx0ADM1NiALwApiFQAEgWTiLmYADGlkbGUgVGltZW91dAAArBADAQAA1VIACAgICAAAADUADWp1bm9zLWRucy11ZHAAAANETlMAAAdVTktOT1dOAAAKDQECAAAO+AAICAgIAAAANQATc3Bva2UtZ3Vlc3RfdG9fd2FuMQAAA04vQQAAAAARAA0wMV9wdWJsaWMtZG5zAAALc3Bva2UtZ3Vlc3QAAAR3YW4xAAAAAAIAA5BLAAExAAAAAAAAAABJAAExAAAAAAAAAACCAAAAPAADTi9BAAADTi9BAAACTm8AAAtzcG9rZS1ndWVzdAAAAjAxAAANYXBicl9wcmktd2FuMQAACmdlLTAvMC8xLjAAAANOL0EAAAAAAAAAAAAAAAAAAAAAAAAOSW5mcmFzdHJ1Y3R1cmUAAApOZXR3b3JraW5nAAACMDEAAANOL0EAAANOL0EAAANOL0EAAANOL0EAAAdkZWZhdWx0ADM2NyALwApiFQAEgWTiLmYADGlkbGUgVGltZW91dAAACgoDAgAAx3IACAgICAAAADUADWp1bm9zLWRucy11ZHAAAANETlMAAAdVTktOT1dOAAAKDQACAAB5cwAICAgIAAAANQASc3Bva2UtY29ycF90b193YW4wAAADTi9BAAA="

	for {
		log.Printf("Sending data for %s thread %d", mac, t)
		err := syslog.Send(conn)
		time.Sleep(time.Millisecond * 1000 * 10)
		if err != nil {
			log.Printf("Send %s error  %v", mac, err)
			goto start
		}
	}

}
