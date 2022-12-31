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

type data123pkg byte

const (
	GAME_LOADED data123pkg = iota + 1
	GAME_LOADED_ACK
	GAME_INPUT
	GAME_MATCH
	GAME_MATCH_ACK
	GAME_MATCH_REQUEST = iota + 3
	GAME_REPLAY
	GAME_REPLAY_REQUEST = iota + 4
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
		if len(buf) != 37 {
			logger.Error("HELLO package len is not 37?", len(buf))
		}
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
		// [5   110 115 101 217 255 196 110 72 141 124 161 146 49 52 114 149   179 83 1 0 40 0 0 89   1   5   121 111 117 109 117 0 32 0 0 0 0 0 0 0 177 47 68 0 75 2 84 2 0 0 0 0 0 0 0 0 51 1 0 0 18 0 0 0]
		// [5   110 115 101 217 255 196 110 72 141 124 161 146 49 52 114 149   5 0 1 248 40 0 13 80   1   5   121 111 117 109 117 0 32 0 0 0 0 0 0 0 177 47 68 0 75 2 97 5 0 0 0 0 0 0 0 0 51 1 0 0 18 0 0 0]
		// [5   110 115 101 217 255 196 110 72 141 124 161 146 49 52 114 149   5 0 1 0 40 0 0 3       0   0   16 63 18 0 0 0 32 0 0 0 0 0 0 0 177 47 68 0 75 2 99 5 0 0 0 0 0 0 0 0 51 1 0 0 18 0 0 0]
		if buf[25] == 0x01 {
			// play_request
			logger.Info("INIT_REQUEST with game id ", buf[1:17], " | ", buf[25], " is play_request |", " profile name ", buf[27:27+buf[26]])
		} else {
			// spectate_request
			logger.Info("INIT_REQUEST with game id ", buf[1:17], " | ", buf[25], " is spectate_request")
		}
		if len(buf) != 65 {
			logger.Error("INIT_REQUEST package len is not 65?", len(buf))
		}
	case INIT_SUCCESS:
		// no spectate            [6 0 0 0 0   0    0 0 0   68   0 0 0   121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0  121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0  0 0 0 0]
		// spectate               [6 0 0 0 0   16   0 0 0   68   0 0 0   121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0  121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0  0 0 0 0]
		// spectate for spectator [6 0 0 0 0   17   0 0 0   68   0 0 0   121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0  121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0  0 0 0 0]
		logger.Info("INIT_SUCCESS with spectate info ", buf[5], " data size ", buf[9], " host_profile ", buf[13:45], " client_profile ", buf[45:77], " swr_disabled ", buf[77:81])

		logger.Info()
		logger.Info("------------------------------------------------------")
		logger.Info("初始化成功")
		logger.Info("------------------------------------------------------")
		logger.Info()
		if len(buf) != 81 {
			logger.Error("INIT_SUCCESS package len is not 81?", len(buf))
		}
	case INIT_ERROR:
		logger.Info("INIT_ERROR with reason", buf[1])

		logger.Info()
		logger.Info("------------------------------------------------------")
		logger.Info("初始化失败")
		logger.Info("------------------------------------------------------")
		logger.Info()
	case REDIRECT:
		// [8 1 0 0 0 2 0 205 33 127 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 121 111 117 109 117 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
		logger.Info("REDIRECT with child id ", buf[1], " addr ", sockaddrIn2String(buf[5:11]))
		if len(buf) != 69 {
			logger.Error("REDIRECT package len is not 69?", len(buf))
		}
	case QUIT:
		// [11]
		logger.Info("QUIT")
		if len(buf) != 1 {
			logger.Error("QUIT package len is not 1?", len(buf))
		}
	case HOST_GAME | CLIENT_GAME:
		if buf[0] == HOST_GAME {
			logger.Info("HOST_GAME data")
		} else {
			logger.Info("CLIENT_GAME data")
		}
		switch data123pkg(buf[1]) {
		case GAME_LOADED:
			if buf[2] == 0x03 {
				logger.Info("Character select page loaded")
			} else if buf[2] == 0x05 {
				logger.Info("Battle loaded")
			}
			if len(buf) != 3 {
				logger.Error("GAME_LOADED package length is not 3? ", len(buf))
			}
		case GAME_LOADED_ACK:
			if buf[2] == 0x03 {
				logger.Info("Ack client character select page loaded")
			} else if buf[2] == 0x05 {
				logger.Info("Ack client battle loaded")
			}
			if len(buf) != 3 {
				logger.Error("GAME_LOADED_ACK package length is not 3? ", len(buf))
			}
		case GAME_INPUT:
			logger.Info("GAME_INPUT")
		case GAME_MATCH:
			// 10800与52513对战，51390从1487开始观战.pcapng line 712/1627
			//              here 208  201  48            f5 0e 00     should be rubbish
			// client    14 4   {208  201  48  (0)  01} {f5 0e 00 (14 64 00 64 00 65 00 65 00 66 00 66 00 67 00 67 00 c8 00 c8 00 c8 00 c8 00 c9 00 c9 00 c9 00 c9 00 cb 00 cb 00 cb 00 cb 00) 00} 00 00 48 f6 c9 0e 00
			// host      13 4   {0  0  0  (20 200 0 200 0 200 0 200 0 201 0 201 0 208 0 208 0 208 0 100 0 100 0 101 0 101 0 102 0 102 0 103 0 103 0 1 0 1 0 1 0) 00} {0f 00 00 (00) 00}            03         03    e9 05 ab 45   00
			// spectacle 13 4   {0  0  0  (20 200 0 200 0 200 0 200 0 201 0 201 0 208 0 208 0 208 0 100 0 100 0 101 0 101 0 102 0 102 0 103 0 103 0 1 0 1 0 1 0) 00}
			//                                          {0f 00 00 (14 64 00 64 00 65 00 65 00 66 00 66 00 67 00 67 00 c8 00 c8 00 c8 00 c8 00 c9 00 c9 00 c9 00 c9 00 cb 00 cb 00 cb 00 cb 00) 00} 03         03    e9 05 ab 45   01
			// 127.0.0.1-34756-watch.pcapng line 1780
			//            0d 04 {00 00 00 (14 c8 00 c8 00 c8 00 c8 00 c9 00 c9 00 d0 00 d0 00 d0 00 64 00 64 00 65 00 65 00 66 00 66 00 67 00 67 00 01 00 01 00 01 00) 00}
			//                                          {0f 00 00 (14 64 00 64 00 65 00 65 00 66 00 66 00 67 00 67 00 c8 00 c8 00 c8 00 c8 00 c9 00 c9 00 c9 00 c9 00 cb 00 cb 00 cb 00 cb 00) 00} 11         10    e3 90 c3 16   01
			//                  {character_id skin_id deck_id (deck_size deck) disabled_simultaneous_buttons}                                                                                     stage_id music_id random_seed match_id
			logger.Info("GAME_MATCH_ACK")
			if len(buf) != 2 {
				logger.Error("GAME_MATCH_ACK package length is not 59? ", len(buf))
			}
		case GAME_MATCH_ACK:
			logger.Info("GAME_MATCH_ACK")
			if len(buf) != 2 {
				logger.Error("GAME_MATCH_ACK package length is not 2? ", len(buf))
			}
		case GAME_MATCH_REQUEST:
			logger.Info("GAME_MATCH_REQUEST")
			if len(buf) != 2 {
				logger.Error("GAME_MATCH_REQUEST package length is not 2? ", len(buf))
			}
		case GAME_REPLAY:
		case GAME_REPLAY_REQUEST:
		}
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
