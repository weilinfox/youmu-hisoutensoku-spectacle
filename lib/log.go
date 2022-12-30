package lib

import (
	"fmt"
	"net"
	"sync"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("log", "main")

type type123pkg byte

const (
	HELLO type123pkg = iota + 1
	PUNCH
	OLLEH
	CHAIN
	INIT_REQUEST
	INIT_SUCCESS
	INIT_ERROR
	REDIRECT
	QUIT      = iota + 3
	HOST_GAME = iota + 4
	CLIENT_GAME
	SOKUROLL_TIME = iota + 5
	SOKUROLL_TIME_ACK
	SOKUROLL_STATE
	SOKUROLL_SETTINGS
	SOKUROLL_SETTINGS_ACK
)

func sockaddrIn2String(addr []byte) string {
	return fmt.Sprintf("%d.%d.%d.%d:%d", addr[2], addr[3], addr[4], addr[5], int(addr[0])<<8+int(addr[1]))
}

func detect(buf []byte) {

	logger.Info(buf)

	switch type123pkg(buf[0]) {
	case HELLO:
		// [1 2 0 18 38 127 0 0 1 108 144 210 229 152 3 32 103 2 0 18 38 127 0 0 1 108 144 210 229 152 3 32 103 0 0 0 0]
		logger.Info("HELLO peer ", sockaddrIn2String(buf[3:9]), " target ", sockaddrIn2String(buf[19:25]))
	case PUNCH:
		// [2 2 0 198 97 127 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0]
		logger.Info("PUNCH for ", sockaddrIn2String(buf[3:9]))
	case OLLEH:
		// [3]
		logger.Info("OLLEL")
	case CHAIN:
		// [4 1 0 0 0]
		// [4 4 0 0 0]
		logger.Info("CHAIN ", buf[1:])
	case INIT_REQUEST:
		// [5   110 115 101 217 255 196 110 72 141 124 161 146 49 52 114 149   179 83 1 0 40 0 0 89   1   5 121 111 117 109 117 0 32 0 0 0 0 0 0 0 177 47 68 0 75 2 84 2 0 0 0 0 0 0 0 0 51 1 0 0 18 0 0 0]
		// [5   110 115 101 217 255 196 110 72 141 124 161 146 49 52 114 149   5 0 1 248 40 0 13 80   1   5 121 111 117 109 117 0 32 0 0 0 0 0 0 0 177 47 68 0 75 2 97 5 0 0 0 0 0 0 0 0 51 1 0 0 18 0 0 0]
		// [5   110 115 101 217 255 196 110 72 141 124 161 146 49 52 114 149   5 0 1 0 40 0 0 3       0   0 16 63 18 0 0 0 32 0 0 0 0 0 0 0 177 47 68 0 75 2 99 5 0 0 0 0 0 0 0 0 51 1 0 0 18 0 0 0]
		logger.Info("INIT_REQUEST with game id ", buf[1:17], " | type ", buf[25])
	case INIT_SUCCESS:
		// no spectate [6 0 0 0 0 0 0 0 0 68 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
		//    spectate [6 0 0 0 0 16 0 0 0 68 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
		//             [6 0 0 0 0 17 0 0 0 68 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
		logger.Info("INIT_SUCCESS with spectate info ", buf[5], " data size ", buf[9], " host_profile ", buf[13:45], " client_profile ", buf[45:77], " swr_disabled ", buf[77:81])
	case INIT_ERROR:
		logger.Info("INIT_ERROR with reason", buf[1])
	case REDIRECT:
		// [8 1 0 0 0 2 0 205 33 127 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
		logger.Info("REDIRECT with child id ", buf[1], " addr ", sockaddrIn2String(buf[5:11]))
	case QUIT:
		// [11]
		logger.Info("QUIT")
	case HOST_GAME:
		logger.Info("HOST_GAME data")
	case CLIENT_GAME:
		logger.Info("CLIENT_GAME data")
	case SOKUROLL_TIME:
		logger.Info("SOKUROLL_TIME")
	case SOKUROLL_TIME_ACK:
		logger.Info("SOKUROLL_TIME_ACK")
	case SOKUROLL_STATE:
		logger.Info("SOKUROLL_STATE")
	case SOKUROLL_SETTINGS:
		logger.Info("SOKUROLL_SETTINGS")
	case SOKUROLL_SETTINGS_ACK:
		logger.Info("SOKUROLL_SETTINGS_ACK")
	}

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
