package lib

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
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

var gameLoadSuccessChan = make(chan int)
var spectacleAcceptChan = make(chan int)

// var startChan = make(chan int)

var repReqStatus byte

var quitFlag = false
var gameId [16]byte     // 16 bytes
var hostInfo [45]byte   // 45 bytes
var clientInfo [45]byte // 45 bytes
var stageId byte
var musicId byte
var randomSeeds [4]byte // 4 bytes
var matchId byte

var replayData = make(map[byte][]uint16)
var replayEnd = make(map[byte]bool)

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
			copy(gameId[:], buf[1:17])
		} else {
			// spectate_request
			logger.Info("INIT_REQUEST with game id ", buf[1:17], " | ", buf[25], " is spectate_request")
			copy(gameId[:], buf[1:17])
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
			spectacleAcceptChan <- 1
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
		quitFlag = true
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

				gameLoadSuccessChan <- 1
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

				copy(hostInfo[:], buf[2:47])
				copy(clientInfo[:], buf[47:92])
				stageId = buf[92]
				musicId = buf[93]
				copy(randomSeeds[:], buf[94:98])
				matchId = buf[98]
				replayData[matchId] = make([]uint16, 1) // 填充一个 garbage
				replayEnd[matchId] = false

				repReqStatus = 0x00

				// startChan <- 1
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

			//   在观战的一开始 观战方首先发送 0x0e, 0x0b, 0xff, 0xff, 0xff, 0xff, 0x00
			//   此时 客户端 必定返回 GAME_MATCH
			//   故私以为 frame id 应该是 有符号整数 而不是 doc 所说 无符号整数

			//   127.0.0.1-34756-watch.pcapng line 9431
			//   0x0e ,0x0b ,0xd0 ,0x25 ,0x00 ,0x00 ,0x06 [frame id 208 37 0 0] [match id 0x06]
			//   0x0d ,0x09 ,0x18 ,0x78 ,0x9c ,0xbb ,0xa0 ,0xca ,0x00 ,0x06 ,0x6c ,0x5c ,0x1c ,0x0c ,0x30 ,0xa8 ,0x01 ,0xc6 ,0x0c ,0x0c ,0x9a ,0x0c ,0x00 ,0x23 ,0x97 ,0x01 ,0xaf
			//   decompress [208 37 0 0 0 0 0 0 6 10 8 0 8 0 8 0 8 0 8 0 40 0 8 0 40 0 0 0 41 0]

			//   127.0.0.1-34756-watch.pcapng line 9433
			//   0x0e ,0x0b ,0xd8 ,0x25 ,0x00 ,0x00 ,0x06 [frame id 216 37 0 0] [match id 0x06]
			//   0x0d ,0x09 ,0x15 ,0x78 ,0x9c ,0xbb ,0xa1 ,0xca ,0x00 ,0x06 ,0x6c ,0x1c ,0x1c ,0x0c ,0x5a ,0x0c ,0x1c ,0x48 ,0x10 ,0x00 ,0x1e ,0xb7 ,0x01 ,0x6e
			//   decompress [216 37 0 0 0 0 0 0 6 8 8 0 42 0 8 0 8 0 8 0 8 0 8 0 8 0]

			//   frame_id [216 37 0 0] end_frame_id [0 0 0 0] match_id [6] game_inputs_count [8] replay_inputs [8 0] [42 0] [8 0] [8 0] [8 0] [8 0] [8 0] [8 0]

			//   10800与52513对战，51390从1487开始观战.pcapng line 1720
			//   0x0e ,0x0b ,0xe0 ,0x01 ,0x00 ,0x00 ,0x01   frame id [ 000001e0 ]
			//   0x0d ,0x09 ,0x11 ,0x78 ,0x9c ,0x7b ,0xc0 ,0xc8 ,0x00 ,0x06 ,0x8c ,0x36 ,0x0c ,0x03 ,0x04 ,0x00 ,0x8f ,0x99 ,0x01 ,0x1f
			//   [224 1 0 0 0 0 0 0 1 60 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
			//   全是 0 大失败

			//   10800与54015对战，39965从2171开始观战.pcapng line 4687   不知道为啥抓包 0x0e 0x0b 包出现在 0x0d 0x09 包后面， wireshark 里包时间的顺序并不是实际的主机发送的顺序
			//    frame id [ 00000bde ]
			//   0e 0b d6 0b 00 00 01       观战方已经收到的 frame id
			//   0x0d ,0x09 ,0x13 ,0x78 ,0x9c ,0xbb ,0xc7 ,0xcd ,0x00 ,0x06 ,0x8c ,0x1c ,0x0c ,0x28 ,0x80 ,0x85 ,0x01 ,0x00 ,0x18 ,0x5b ,0x00 ,0xf7
			//   [222 11 0 0 0 0 0 0 1 8 0 0 0 0 0 0 0 0 0 0 0 0 0 0 4 0]
			//    de  b                 |bde bdd|bdc bdb|bda bd9|bd8 bd7|  // client 发送新的 frame
			//    bde>>1==5ef           |  5ef  |  5f0  |  5f1  |  5f2  |  // client input + host input 一组 4 bytes
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
			//  猜测 replay 中的 frame id 是这个操作在 0e03/0d03 中 frame id << 1
			//  由于暂时无法精确了解玩家操作如何影响机体状态 暂时无法验证

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
			// 0x0d ,0x09 ,0x11 ,0x78 ,0x9c ,0x7b ,0xc6 ,0xcd ,0x00 ,0x06 ,0x8c ,0x1c ,0x0c ,0x68 ,0x00 ,0x00 ,0x19 ,0x23 ,0x00 ,0xfb
			if len(buf) > 3 && len(buf)-3 == int(buf[2]) {
				r, err := zlib.NewReader(bytes.NewBuffer(buf[3:]))
				if err != nil {
					logger.WithError(err).Error("New zlib reader error")
				} else {
					ans := make([]byte, 2048)
					n, err := r.Read(ans)
					_ = r.Close()
					if err == io.EOF {
						// decode package
						logger.Info("Replay data decompress ", ans[:n])

						//   frame_id [216 37 0 0] end_frame_id [0 0 0 0] match_id [6] game_inputs_count [8] replay_inputs [8 0] [42 0] [8 0] [8 0] [8 0] [8 0] [8 0] [8 0]
						if n >= 10 && n-10 == int(ans[9])*2 {
							frameId := int(ans[0]) | int(ans[1])<<8 | int(ans[2])<<16 | int(ans[3])<<24
							endFrameId := int(ans[4]) | int(ans[5])<<8 | int(ans[6])<<16 | int(ans[7])<<24

							data := replayData[ans[8]]
							getDataLen := len(data) - 1
							if getDataLen == -1 {
								logger.Error("No such match")
							} else if frameId-getDataLen <= int(ans[9]) {
								newDataLen := frameId - getDataLen

								if newDataLen > 0 {
									newData := make([]uint16, newDataLen)

									for i := 0; i < newDataLen; i++ {
										newData[newDataLen-1-i] = uint16(ans[10+i*2])<<8 | uint16(ans[11+i*2])
									}

									replayData[ans[8]] = append(data, newData...)

									if len(replayData[ans[8]])-1 != frameId {
										logger.Error("Replay data not match after append new data")
									}
								}

								if endFrameId != 0 {
									logger.Info("Frame id ", frameId, " end frame id ", endFrameId)

									if endFrameId == frameId {
										logger.Info("Match end")
										replayEnd[ans[8]] = true
									}
								}

								repReqStatus = 0x00
							} else {
								logger.Warn("Replay data package drop: frame id ", frameId, " length ", ans[9])
							}
						} else {
							logger.Error("Replay data content invalid")
						}
					} else {
						logger.WithError(err).Error("Zlib decode error")
					}
					// fmt.Println(err == io.EOF, err, ans[:n])
				}
			} else {
				logger.Error("Replay data invalid")
			}

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

		logger.Warn("Wait for host and client init")
		<-gameLoadSuccessChan

		requestData := append([]byte{byte(INIT_REQUEST)}, gameId[:]...) // INIT_REQUEST and game id
		requestData = append(requestData, make([]byte, 8)...)           // garbage
		requestData = append(requestData, 0x00)                         // spectacle request
		requestData = append(requestData, 0x00)                         //  data length 0
		requestData = append(requestData, make([]byte, 38)...)          // make it 65 bytes long

		logger.Warn("Send spectacle init ", requestData)
		slave.WriteToUDP(requestData, slaveAddr)

		matchId = 0x00
		<-spectacleAcceptChan
		// slave.WriteToUDP([]byte{0x04, 0x04, 0x00, 0x00, 0x00}, slaveAddr)
		// logger.Warn("Get match info")
		// slave.WriteToUDP([]byte{CLIENT_GAME, GAME_REPLAY_REQUEST, 0xff, 0xff, 0xff, 0xff, 0x00}, slaveAddr) // client will return GAME_MATCH package

		for !quitFlag {
			// <-startChan
			logger.Warn("Get some replay package")
			repReqStatus = 0x00
			for !quitFlag {

				// 如果不存在的时刻 得到的会是空
				// 0e 0b 34 05 00 00 01
				// [13, 9 ,15, 120, 156, 99 ,224 ,231 ,103, 0 ,1 ,70, 6 ,0 ,1 ,11, 0, 32]  [0 15 15 0 0 0 0 0 1 0]
				// [13, 9 ,15, 120, 156, 19, 224, 231, 103, 0 ,1 ,70, 6 ,0 ,1 ,171, 0, 48] [16 15 15 0 0 0 0 0 1 0]
				// 注意前 1s 是没有数据的 毕竟在放第一战动画
				// 抓包计算
				// 对于 frame id 60f/s
				// 对于 replay frame id 120f/s
				// 对于 0x0e, 0x0b 15f/s
				switch repReqStatus {
				case 0x00, 0x03:
					getId := len(replayData[matchId]) - 1 // at start, match id == 0 , get id == -1
					logger.Info("Send replay request ", getId, " ", matchId)
					slave.WriteToUDP([]byte{CLIENT_GAME, GAME_REPLAY_REQUEST, byte(getId), byte(getId >> 8), byte(getId >> 16), byte(getId >> 24), matchId}, slaveAddr)
					repReqStatus = 0x01
				case 0x01, 0x02:
					repReqStatus++
				}

				time.Sleep(time.Millisecond * 66)

				if replayEnd[matchId] {
					break
				}
			}

			if !quitFlag {
				logger.Info("Waiting for another match")
				<-gameLoadSuccessChan
			}
		}

	}()

	wg.Wait()
}
