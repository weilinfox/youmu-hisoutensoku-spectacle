package lib

import (
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("log", "main")

func detect(buf []byte) {

	logger.Info(buf)

}

func Sync(master, slave *net.UDPConn) {
	var slaveAddr *net.UDPAddr
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()

		buf := make([]byte, 2048)

		for {
			n, err := master.Read(buf)
			if err != nil {
				logger.WithError(err).Error("master read error")
				break
			}

			if slaveAddr != nil {
				detect(buf[:n])

				_, err = slave.WriteToUDP(buf[:n], slaveAddr)
				if err != nil {
					logger.WithError(err).Error("slave write error")
					break
				}
			}
		}

	}()

	go func() {
		defer wg.Done()

		var n int
		var err error
		buf := make([]byte, 2048)

		for {
			n, slaveAddr, err = slave.ReadFromUDP(buf)
			if err != nil {
				logger.WithError(err).Error("slave read error")
				break
			}

			detect(buf[:n])
			_, err = master.Write(buf[:n])
			if err != nil {
				logger.WithError(err).Error("master write error")
				break
			}
		}

	}()

	wg.Wait()
}
