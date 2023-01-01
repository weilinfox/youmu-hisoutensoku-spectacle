package lib

import (
	"fmt"
	"net"
	"sync"
	"time"

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

var successChan = make(chan int)
var spectacleChan = make(chan int)
var startChan = make(chan int)

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

		if buf[5] == 0x10 {
		} else if buf[5] == 0x11 {
			spectacleChan <- 1
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
	case HOST_GAME, CLIENT_GAME:
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

				successChan <- 1
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
			logger.Info("GAME_MATCH")
			if len(buf) == 59 {

			} else if len(buf) == 99 {
				startChan <- 1
			} else {
				logger.Info("GAME_MATCH package length is not 59/99? ", len(buf))
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

			//   127.0.0.1-34756-watch.pcapng line 9431 不知道为啥抓包 0x0e 0x0b 包出现在 0x0d 0x09 包后面
			//   0x0e ,0x0b ,0xd0 ,0x25 ,0x00 ,0x00 ,0x06 [frame id 208 37 0 0] [match id 0x06]
			//   0x0d ,0x09 ,0x18 ,0x78 ,0x9c ,0xbb ,0xa0 ,0xca ,0x00 ,0x06 ,0x6c ,0x5c ,0x1c ,0x0c ,0x30 ,0xa8 ,0x01 ,0xc6 ,0x0c ,0x0c ,0x9a ,0x0c ,0x00 ,0x23 ,0x97 ,0x01 ,0xaf
			//   decompress [208 37 0 0 0 0 0 0 6 10 8 0 8 0 8 0 8 0 8 0 40 0 8 0 40 0 0 0 41 0]

			//   127.0.0.1-34756-watch.pcapng line 9433
			//   0x0e ,0x0b ,0xd8 ,0x25 ,0x00 ,0x00 ,0x06 [frame id 216 37 0 0] [match id 0x06]
			//   0x0d ,0x09 ,0x15 ,0x78 ,0x9c ,0xbb ,0xa1 ,0xca ,0x00 ,0x06 ,0x6c ,0x1c ,0x1c ,0x0c ,0x5a ,0x0c ,0x1c ,0x48 ,0x10 ,0x00 ,0x1e ,0xb7 ,0x01 ,0x6e
			//   decompress [216 37 0 0 0 0 0 0 6 8 8 0 42 0 8 0 8 0 8 0 8 0 8 0 8 0]

			//   frame_id [216 37 0 0] end_frame_id [0 0 0 0] match_id [6] game_inputs_count [8] replay_inputs [8 0] [42 0] [8 0] [8 0] [8 0] [8 0] [8 0] [8 0]
			//   >> quote touhou-protocol-docs here （问就是没看懂）
			//   >> replay_inputs is a list of the last replay_inputs_count inputs of both players, stored in descending frame order.
			//   >> The first replay_input is the input at frame frame_id, the next one is the input at frame frame_id - 1, and so on.

			//   10800与52513对战，51390从1487开始观战.pcapng line 1720
			//   0x0e ,0x0b ,0xe0 ,0x01 ,0x00 ,0x00 ,0x01   frame id [ 000001e0 ]
			//   0x0d ,0x09 ,0x11 ,0x78 ,0x9c ,0x7b ,0xc0 ,0xc8 ,0x00 ,0x06 ,0x8c ,0x36 ,0x0c ,0x03 ,0x04 ,0x00 ,0x8f ,0x99 ,0x01 ,0x1f
			//   [224 1 0 0 0 0 0 0 1 60 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
			//   全是 0 大失败

			//   10800与54015对战，39965从2171开始观战.pcapng line 4687   不知道为啥抓包 0x0e 0x0b 包出现在 0x0d 0x09 包后面， wireshark 里包时间的顺序并不是实际的主机发送的顺序
			//    frame id [ 00000bde ]
			//   0e 0b d6 0b 00 00 01       这里因为包顺序不对，忽略就好了，实际对应应该是 0e 0b de 0b 00 00 01
			//   0x0d ,0x09 ,0x13 ,0x78 ,0x9c ,0xbb ,0xc7 ,0xcd ,0x00 ,0x06 ,0x8c ,0x1c ,0x0c ,0x28 ,0x80 ,0x85 ,0x01 ,0x00 ,0x18 ,0x5b ,0x00 ,0xf7
			//   [222 11 0 0 0 0 0 0 1 8 0 0 0 0 0 0 0 0 0 0 0 0 0 0 4 0]
			//    de  b
			//    bde>>1==5ef            |  5ef  |  5f0  |  5f1  |  5f2  |  // 不知道 frame id 对不对但是不重要了 client input + host input 一组 4 bytes
			//                                             line 4522
			//   0d 03 ef 05 00 00 05 02 00 00 00 00
			//   0e 03 ef 05 00 00 05 01 00 00
			//   0d 03 f0 05 00 00 05 02 00 00 00 00
			//   0e 03 f0 05 00 00 05 01 00 00
			//   0d 03 f1 05 00 00 05 02 04 00 00 00
			//   0e 03 f1 05 00 00 05 01 00 00
			//   0d 03 f2 05 00 00 05 02 04 00 00 00
			//   0e 03 f2 05 00 00 05 01 00 00
			//   0d 03 f3 05 00 00 05 02 04 00 00 00
			//   0e 03 f3 05 00 00 05 01 00 00
			//   0d 03 f4 05 00 00 05 02 04 00 00 00
			//   ...
			//   0e 03 fc 05 00 00 05 01 00 00
			//   0d 03 fd 05 00 00 05 02 04 00 00 00
			//   0e 03 fd 05 00 00 05 01 00 00
			//   0d 03 fe 05 00 00 05 02 14 00 00 00
			//   0e 03 fe 05 00 00 05 01 00 00
			//   0d 03 ff 05 00 00 05 02 14 00 00 00
			//   0e 03 ff 05 00 00 05 01 00 00
			//   0d 03 00 06 00 00 05 02 14 00 00 00
			//   0e 03 ff 05 00 00 05 01 00 00
			//   0e 03 00 06 00 00 05 01 00 00
			//   0d 03 01 06 00 00 05 02 14 00 00 00
			//   0d 03 02 06 00 00 05 04 14 00 00 00 14 00 00 00
			//   0e 03 02 06 00 00 05 01 00 00
			//   0d 03 03 06 00 00 05 04 04 00 00 00 14 00 00 00
			//   0e 03 03 06 00 00 05 01 00 00
			//   0d 03 04 06 00 00 05 04 04 00 00 00 04 00 00 00
			//   0e 03 04 06 00 00 05 01 00 00
			//   0d 03 05 06 00 00 05 02 04 00 00 00
			//   0e 03 05 06 00 00 05 01 00 00
			//   0d 03 06 06 00 00 05 04 04 00 00 00 04 00 00 00
			//   0e 03 06 06 00 00 05 01 00 00
			//   0d 03 07 06 00 00 05 04 14 00 00 00 04 00 00 00
			//   0e 03 07 06 00 00 05 01 00 00
			//   0d 03 08 06 00 00 05 02 14 00 00 00
			//                                             line 4698
			//   0e 0b de 0b 00 00 01
			//   0x0d ,0x09 ,0x11 ,0x78 ,0x9c ,0x7b ,0xc6 ,0xcd ,0x00 ,0x06 ,0x8c ,0x1c ,0x0c ,0x68 ,0x00 ,0x00 ,0x19 ,0x23 ,0x00 ,0xfb
			//   [230 11 0 0 0 0 0 0 1 8 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
			//   0e 0b e6 0b 00 00 01
			//   0x0d ,0x09 ,0x11 ,0x78 ,0x9c ,0x7b ,0xc7 ,0xcd ,0x00 ,0x06 ,0x8c ,0x1c ,0x0c ,0x68 ,0x00 ,0x00 ,0x19 ,0xf3 ,0x01 ,0x03
			//   [238 11 0 0 0 0 0 0 1 8 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
			//   0e 0b ee 0b 00 00 01
			//   [246 11 0 0 0 0 0 0 1 8 0 0 132 0 0 0 132 0 0 0 132 0 0 0 4 0]  11,246->bf6 bf6>>1==5fb 132->0x84

			//  replay 发送的是机体的状态而不是玩家的操作
			//  replay 中的 frame id 是这个操作在 0e03/0d03 中 frame id << 1
			//  >> quote touhou-protocol-docs here （说法略有不同）
			//  >> replay_input is a pair of a client_input and a host_input, always sent in pairs and in that order.
			//  >> There are game_inputs_count game_input, which means that there are actually replay_inputs_count = game_inputs_count / 2 pairs of host and client inputs.

			/*	b := bytes.NewBuffer([]byte{0x78, 0x9c, 0xbb, 0xa1, 0xca, 0x00, 0x06, 0x6c, 0x1c, 0x1c, 0x0c, 0x5a, 0x0c, 0x1c, 0x48, 0x10, 0x00, 0x1e, 0xb7, 0x01, 0x6e})
				r, err := zlib.NewReader(b)
				if err != nil {
					fmt.Println(err)
				} else {
					ans := make([]byte, 2048)
					n, err := r.Read(ans)
					r.Close()
					fmt.Println(err == io.EOF, err, ans[:n])
				}*/
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

	wg.Add(3)
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

	go func() {
		defer wg.Done()

		logger.Warn("Wait for spectacle init")
		<-successChan

		logger.Warn("Send spectacle init")
		slave.WriteToUDP([]byte{5, 110, 115, 101, 217, 255, 196, 110, 72, 141, 124, 161, 146, 49, 52, 114, 149, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, slaveAddr)

		<-spectacleChan
		// slave.WriteToUDP([]byte{0x04, 0x04, 0x00, 0x00, 0x00}, slaveAddr)
		logger.Warn("Get match info")
		slave.WriteToUDP([]byte{0x0e, 0x0b, 0xff, 0xff, 0xff, 0xff, 0x00}, slaveAddr)

		<-startChan
		logger.Warn("Get some replay package")
		for i := 0; i < 15*5; i += 1 {
			// 如果不存在的时刻 得到的会是空
			// 0e 0b 34 05 00 00 01
			// [13, 9 ,15, 120, 156, 99 ,224 ,231 ,103, 0 ,1 ,70, 6 ,0 ,1 ,11, 0, 32]  [0 15 15 0 0 0 0 0 1 0]
			// [13, 9 ,15, 120, 156, 19, 224, 231, 103, 0 ,1 ,70, 6 ,0 ,1 ,171, 0, 48] [16 15 15 0 0 0 0 0 1 0]
			// 注意前 1s 是没有数据的 毕竟在放第一战动画
			// 抓包计算
			// 对于 frame id 60f/s
			// 对于 replay frame id 120f/s
			// 对于 0x0e, 0x0b 15f/s
			slave.WriteToUDP([]byte{0x0e, 0x0b, 0x00, 0x0, 0x0, 0x00, 0x01}, slaveAddr)
			time.Sleep(time.Millisecond * 66)
		}

	}()

	wg.Wait()
}
