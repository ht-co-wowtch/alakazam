# 聊天室
- [快速開始使用服務](#quick-reference)
- [快速建置聊天室服務](#quick-start)
  - [編譯](#build)
  - [運行](#run)
- [依賴工具](#dependencies)
- [架構圖](#architecture)
- [功能](#features)
- [聊天室Web Socket協定](#protocol-body)
- [Web Socket](#web-socket)
- [會員身份權限](#member-permissions)
- [訊息規則](#message-rule)
- [錯誤訊息](#system-message)
- [聊天室版本](#tag)

## Quick Start

### Build
```
    make build
```

### Run
```
    make run
    make stop
```

## Dependencies
[Redis](#)

[Kafka](https://kafka.apache.org/quickstart)

    cd cmd/kafka
    docker-compose up -d
    
## Architecture

![arch](./doc/architecture.png)


`comet` : 負責管理web端連線與將訊息推送給web

`logic` : 負責接收各種商業邏輯訊息推送

`job` : 負責告知comet要推送什麼訊息以及房間訊息聚合

## Features

- [x] 用戶禁言
- [x] 用戶封鎖
- [ ] 用戶搶紅包
- [ ] 用戶發紅包
- [ ] 用戶發跟注
- [ ] 用戶跟注
- [x] 用戶切換房間
- [ ] 管理員訊息
- [x] 單房間推送訊息
- [x] 多房間推送訊息
- [ ] 密語
- [x] 房間禁言
- [x] 房間聊天限制(打碼量&充值)
- [x] 房間在線人數
- [ ] 歷史紀錄
- [ ] 訊息置頂
- [ ] 房間在線名單
- [x] 設定房間權限(禁言,紅包,跟注,聊天限制)
- [ ] 反垃圾訊息

TODO
- [ ] json改用msgp
- [ ] 評估kafka or nats
- [ ] 分析要放的log(非存DB)
- [ ] Metrics監控
- [ ] 註冊中心(etcd)
- [ ] docker file
- [ ] 整合測試 bash shell
- [ ] database,redis connection pool  maxConn 優化
- [ ] 限制訊息長度
- [ ] Test Case 用BDD改寫

## Quick Reference

前台：
1. 如何跟進入聊天室 [答案](#room)
2. 如何進入聊天室有失敗會怎樣 `答案:失敗會直接close連線`
3. 如何在聊天室發訊息 [請看前台訊息推送API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649)
4. 如何接收聊天室訊息 [答案](#message)
5. 封鎖狀態下進入聊天室會怎樣 [答案](#response)
6. 禁言狀態下聊天會怎樣 [請看前台訊息推送API範例](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649)
7. 如何知道用戶在聊天室相關權限，如聊天，發紅包等等動作 [答案](#response)
8. 如何在聊天室發紅包 [答案](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#f6c5fb74-cd42-40fb-bb66-1c2bde3419af)
9. 如何搶紅包
10. 如何在聊天室發跟注
11. 如何跟注
12. 如何切換聊天室房間[答案](#change-room)
13. 如何拿到歷史紀錄
14. 如何跟聊天室做心跳 [答案](#heartbeat)
15. 聊天室心跳週期是多少 `答案:每分鐘心跳一次`
16. 如何產生一個跟websocket溝通的Protocol [答案](#buffer)
17. 如何拿到房間在線人數 [答案](#heartbeat-reply)

後台：
1. 如何以管理員身份廣播多個聊天室 [請看後台訊息推送API範例](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#c14d247c-4210-446e-aa0d-97989e4fd03c)
2. 如何以系統公告身份廣播多個聊天室
3. 如何得到線上所有聊天室清單與在線人數
4. 如何得到某聊天室歷史紀錄
5. 如何禁言某會員 [答案](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#1f05bf12-bceb-431c-b50a-cda72b61804e)
6. 如何封鎖某會員 [答案](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#c84c8a03-1e15-412d-889d-642c537d0d45)
7. 如何解禁言某會員 [答案](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#b39e8934-9ca3-4dae-8d23-d1931e1bf5ee)
8. 如何解封鎖某會員 [答案](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#db157cff-a26b-4d07-a109-ff7d676f9ecb)
9. 如何在聊天室發紅包
10. 如何在聊天室發跟注
11. 如何將訊息置頂
12. 如何得到禁言名單
13. 如何得到封鎖名單
14. 如何得到某聊天室名單
15. 如何設定一個房間權限(比如`充值`&`打碼量`多少才能聊天) [答案](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#6bdffede-2646-4663-8c13-54cb161a2125)

## Protocol Body格式

採用websocket做binary傳輸，聊天室推給client訊息內容如下格式

name | length | remork |說明
-----|--------|--------|-----
Package|4 bytes|header + body length| 整個Protocol bytes長度
Header|2 bytes|protocol header length| Package  - Boyd 
Operation|4 bytes| [Operation](#operation)|Protocol的動作
Body |不固定|傳送的資料16bytes之後就是Body|json格式

![arch](./doc/protocol.png)


### Package
用於表示本次傳輸的binary內容總長度是多少(header + body)

### Header
用來說明Protocol Header binary 長度是多少

### Operation
不同的Operation說明本次Protocol資料是什麼，如心跳回覆,訊息等等

value | 說明 | body type |
-----|-----|-------|
1|[要求連線到某一個房間](#room)  | json |
2|[連線到某一個房間結果回覆](#room-reply) | json |
3|[發送心跳](#heartbeat) | 無Body |
4|[回覆心跳結果](#heartbeat-reply) | int32 |
5|[聊天室批次訊息](#message) | json |
6|[聊天室訊息](#message-raw) | json |
7|[更換房間](#change-room) | json |
8|[回覆更換房間結果](#change-room-reply) | json |

### Body
聊天室的訊息內容

###  Buffer
如何在前端產生一個Protocol Body，以進入聊天室為例子
> 本例子js只是解釋如何產生一個Protocol，實際寫法請自行實作

```
// websocket傳輸模式要設定成binary
ws.binaryType = 'arraybuffer'

// 準備好body內容
var token = '{"token":"efk350dmdh20kdf", "room_id":"123"}'

// 一個長度為10的ArrayBuffer，因為如何在前端產生一個Protocol Header長度為10
// 請看本章解上圖4+2+4 = 10
var headerBuf = new ArrayBuffer(10)

// 產生一個用於header DataView
var headerView = new DataView(headerBuf, 0)

// encoder body
var bodyBuf = textEncoder.encode(token)

// 根據set Protocol Package欄位內容，原因請看本章解上述Package解釋
// 從第0個byte開始放資料，由於是in32所以會佔4byte
headerView.setInt32(0, 10 + bodyBuf.byteLength)

// 根據set Protocol Header欄位內容，原因請看本章解上述Header解釋
// 從第4個byte開始放資料，由於是int16所以會佔2byte
headerView.setInt16(4, rawHeaderLen)

// 根據set Protocol Operation欄位內容，原因請看本章解上述Operation解釋
// 從第6個byte開始放資料，由於是int32所以會佔4byte
// 參數1代表此Protocol表示進入某房間
headerView.setInt32(6, 1)

// 將headerView與bodyBuf兩個轉成Uint8Array在做合併
var u81 = new Uint8Array(headerBuf)
var u82 = new Uint8Array(bodyBuf)
res = new Uint8Array(headerBuf.byteLength + bodyBuf.byteLength)
res.set(u81, 0)
res.set(u82, headerBuf.byteLength)

// 輸出成binary傳給websocket
ws.send(res.buffer)
```

接收到回覆

```
ws.onmessage = function (evt) {
	var data = evt.data
	var dataView = new DataView(data, 0)
	
    // Protocol Package欄位內容
	var packetLen = dataView.getInt32(0)
	
    // Protocol Header欄位內容
	var headerLen = dataView.getInt16(4)
	
    // Protocol Operation欄位內容
	var op = dataView.getInt32(6)
    
	switch (op) {
   	    // 回覆進入房間結果
		case 2:
  		    // 擷取出body內容
			var json = textDecoder.decode(data.slice(headerLen, packetLen))
			var msgBody = JSON.parse(json)
			console.log(msgBody)
			break
		// 回覆心跳結果，心跳body為空所以不做事	
		case 4:
			console.log("receive: heartbeat")
			break
	}
}
```

### Message

聊天室一次推送的訊息不一定只有一筆，依照當下該房間發話的速度決定(房間訊息聚合)，以一個簡單js做說明

![arch](./doc/message.png)

```
ws.onmessage = function (evt) {
    // 取出header
    var data = evt.data;
    var dataView = new DataView(data, 0);
    var packetLen = dataView.getInt32(0);
    var headerLen = dataView.getInt16(4);
    var op = dataView.getInt32(6);

    switch (op) {
        // 判斷是否為訊息
        case 5:
            // 一開始offset為本次Protocol Header長度，因為Body內有多筆訊息
            // 所以需要for將body內所有訊息取出來，每一則訊息還是依照解析 Protocol
            // 的方式從header取到足夠資訊來找出Body位置
            for (var offset = headerLen; offset < data.byteLength; offset += packetLen) {
                var packetLen = dataView.getInt32(offset);
                var headerLen = dataView.getInt16(offset + 4);
                var op = dataView.getInt32(offset + 6);
                
                // 確認是否為訊息
                if op == 6 {
                  var json = textDecoder.decode(data.slice(offset + headerLen, offset + packetLen));
                
                  var msgBody = JSON.parse(json)
                  console.log(msgBody)
                }
            }
            break
    }
}
```

### Response

#### Room  Reply
Operation = `2`=> 連線到某一個房間結果回覆Body

```
{
    "room_id":"a1b4bbec3f624ecf84858632a730c688",
    "uid": "82ea16cd2d6a49d887440066ef739669",
    "key": "defb108d-3d51-475a-b266-4a7f459e7a59",
    "permission": {
        "message": true,
        "send_bonus": true,
        "get_bonus": true,
        "send_follow": true,
        "get_follow": true
    }
}
```
name|說明|
----|-----|
uid|user uid，發送訊息會用到
key|這次web socket連線id，發送訊息會用到
room_id|房間id|
permission.message|是否可以聊天
permission.send_bonus|是否可以發紅包
permission.get_bonus|是否可以搶紅包
permission.send_follow|是否可以發跟注
permission.get_follow|是否可以跟注

被封鎖Body內容如下

```
{
    "code": "10024011",
    "message": "您在封鎖状态，无法进入聊天室"
}
```

#### Heartbeat Reply
Operation = `4`=> 回覆心跳結果
```
body是內容是該房間在線人數，是一個int32
```
![arch](./doc/heartbeatReply.png)

#### Message Raw
Operation = `6`=> 單筆訊息

![arch](./doc/message_raw.png)

```
{
    "uid":"7f547901b02041ab8d7d3381d0703137",
    "name": "sam78",
    "avatar": "",
    "message": "測試",
    "time": "17:00:51"
}
```

name|說明|格式
----|-----|-----|
uid|訊息人uid|string
name|訊息人名稱|string
avatar|頭像path info| string
message|訊息|string
time|發送時間|string

#### Change Room Reply
Operation = `8`=> 回覆更換房間結果

![arch](./doc/changeRoomReply.png)

name|說明|格式
----|-----|-----|
room_id|新房間id|string

```
body是新房間id
```

## Web Socket

### Room

跟websocket建立完連線後將以下json包裝成[Protocol](#protocol-body)發送至websocket，Protocol Operation[參考](#operation)

```
  {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTcyMTE2NTAsIm5iZiI6MTU1NzIxMTY1MCwiaXNzIjoibG9naW4iLCJzZXNzaW9uX3Rva2VuIjoiZjc2OTYyM2Y0YTNlNDE4MWE4NzAwYWNkYTE3NzE1MmIiLCJkYXRhIjp7InVpZCI6IjEyNTdlN2Q5ZTFjOTQ0ZWY5YTZmMTI5Y2I5NDk1ZDAyIiwidXNlcm5hbWUiOiJyb290In19.7VJxH3tQpnJqWTlPbId7f0Rt7eQoaVvaJmbWxtHTqRU,
      "room_id": "a7d20d4133c14a62b617ee38e793cf55"
  }
```
name|說明|
----|-----|
token|請自行透過[paras服務](https://jetfueltw.postman.co/collections/2999246-70e3e838-def8-4d53-ba12-e1b2797eec57?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#13ab4886-ab04-4e38-90b1-a0e73c4c8e0f)登入後拿到token
room_id|想要進入的房間id，透過`paras服務`取得

結果|說明|
----|-----|
成功|[Response](#response)
失敗(非封鎖造成)|server會把websocket close

### Heartbeat
進入房間成功後websocket需要每分鐘做一次心跳，讓server確保websocket健康狀況，請利用送一個body為空的[Protocol](#protocol-body)，以下是一個簡單的js範例，至於為什麼這樣寫[請看](#buffer)

```
var headerBuf = new ArrayBuffer(rawHeaderLen);
var headerView = new DataView(headerBuf, 0);
headerView.setInt32(packetOffset, rawHeaderLen);
headerView.setInt16(headerOffset, rawHeaderLen);
headerView.setInt32(opOffset, 3);
```
![arch](./doc/heartbeat.png)

結果|說明|
----|-----|
成功|[Response](#response)
失敗|失敗就會close連線

### Change Room

Boyd內容帶想要切換的房間Id即可

name|說明|格式
----|-----|-----|
room_id|新房間id|string

![arch](./doc/changeRoom.png)

結果|說明|
----|-----|
成功|[Response](#change-room)
失敗|失敗就會close連線

## Member Permissions
會員權限與身份

狀態 |進入聊天室 |查看聊天 | 聊天 |發紅包|搶紅包|跟注|發跟注|
-----|-----|-----|-----|-----|-----|-----|-----|
禁言|yes|yes|no|yes|yes|yes|no|
不足打碼&充值量|yes|yes|no|yes|no|yes|no|
封鎖|no|no|no|no|no|no|no|

身份 |進入聊天室 |查看聊天 | 聊天 |發紅包|搶紅包|跟注|發跟注|訊息置頂|充值限制聊天|打碼量聊天|
-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|
帶玩帳號|yes|yes|yes|yes|yes|yes|yes|no|no|no
一般帳號|yes|yes|yes|yes|yes|yes|yes|no|yes|yes
試玩帳號|yes|yes|no|no|no|no|no|no|no|no
未登入|no|no|no|no|no|no|no|no|no|no

> 帶玩帳號,後台,假人都不算帳

## Message Rule
訊息內容

  - 名稱
  - 頭像
  - 訊息
  - 發送時間

訊息種類

- 會員
- 管理員
- 發紅包
- 搶紅包
- 系統公告

歷史訊息，以10分鐘當1個區段，回滑能看到1小時訊息為止，以下是訊息種類

- 會員
- 管理員
- 發紅包
- 搶紅包


系統自動禁言10分鐘or永久禁言：
  - 不雅訊息1次
  - 10秒內相同訊息3次
  - 自動禁言達5次會永久禁言
  - 只限定`一般帳號`會有此功能

能訊息頂置身份
  - 管理員

發送訊息限制
  - 下限1字元
  - 上限100字元
  - 1秒1則

## System Message

```
{
    "code": "10024011",
    "message": "您在封鎖状态，无法进入聊天室"
}
```

Error Code|異常狀況| 訊息
-----|-----|-----
|10024000|請求進入聊天室資料錯誤|进入聊天室失败|
|10024001|用戶不在此聊天室且要發話|没有在此房间
|10024011|被封鎖|您在封鎖状态，无法进入聊天室|
|10024012|未登入|请先登入会员|
|10024013|被禁言者發言|您在禁言状态，无法发言|
|10024220|聊天室發話資料有誤|资料验证错误|
|10025000|預料之外的錯誤|应用程序错误|
||後台執行禁言|[会员名称] 被禁言
||後台執行封鎖|[会员名称] 被封鎖
||不符後台條件發言|您无法发言，当前发言条件：前两天充值不少于[  ]元；打码量不少于[  ]元

## Tag
tag|說明|
---|----|
v0.1.0|config改為yml
v0.1.1|移除Discovery
v0.1.2|module更名
v0.1.3|module移除沒用到的package
v0.2.0|更改protobuf目錄結構
v0.3.0|protobuf移除不必要參數與method
v0.4.0|移除有關於Operation推送限制
v0.5.0|移除room type推送限制
v0.6.0|refactor log or name
v0.7.0|移除單人訊息推送 
v0.7.1|訊息推送內容改json格式且包含user name, avatar,time，推送認證改用uid & key當pk
v0.7.2|封鎖 and 禁言 api
v0.7.3|前台房間限制,房間設定api, user三方接口認證