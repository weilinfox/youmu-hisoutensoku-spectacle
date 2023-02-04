# hisoutensoku-protocol

参照 [touhou-protocol-docs](https://github.com/delthas/touhou-protocol-docs/blob/master/protocol_123.md) 做的抓包和研判。

实现的功能：

+ 判断对战双方的状态
+ 存储对战数据
+ 独立于对战双方的观战服务器

对战方连接 ``127.0.0.1:4646`` ，观战方连接 ``127.0.0.1:4647``

已经应用在 [thlink](https://github.com/weilinfox/youmu-thlink) 联机器中
