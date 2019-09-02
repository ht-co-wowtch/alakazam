

# 聊天室

- [快速開始使用服務](#quick-reference)
- [快速建置聊天室服務](#quick-start)
- [依賴工具](#dependencies)
- [架構圖](#architecture)
- [功能](#features)
- [聊天室Web Socket協定](#protocol-body)
- [Web Socket](#web-socket)

## Quick Start

TODO 

## Dependencies

TODO

## Architecture

TODO

## Features

- [x] 用戶單一房間聊天
- [x] 用戶單一房間發紅包
- [x] 用戶切換房間
- [x] 用戶暫時禁言
- [x] 用戶永久禁言
- [x] 用戶自動禁言
- [x] 單一用戶進入多房間(可重覆一樣的房間)
- [x] 限制用戶打碼與充值量聊天
- [x] 用戶封鎖
- [x] 用戶拉取歷史訊息(mysql )
- [x] 房間禁言
- [x] 踢人
- [x] 用戶搶紅包
- [x] 紅包時效
- [x] 用戶查詢紅包明細
- [x] 自動返還過期紅包剩餘金額
- [x] 管理員多房間推送訊息
- [x] 公告(置頂)多房間推送訊息
- [x] 管理員發紅包
- [x] 預定時間發佈紅包
- [x] 建立/更新房間(禁言,紅包,聊天限制)
- [x] 關閉房間
- [x] 房間在線人數
- [x] 遮蔽敏感詞
- [x] 訊息限速(1秒1則，10內不可重覆三次一樣訊息)
- [x] 依據性別切換頭像

TODO

- [ ] json改用msgp
- [ ] Metrics監控
- [ ] 註冊中心(etcd or zk)
- [ ] 整合測試 bash shell
- [ ] golang pprof
- [ ] 歷史訊息redis
- [ ] 歷史訊息 mongodb
- [ ] Kafka and zk config 調整
- [ ] 訊息壓縮
- [ ] cache 優化

## Quick Reference

前台：

1. 如何進入聊天室 [答案](#room)
2. 進入聊天室後的回覆 or 如何知道用戶在聊天室相關權限，如聊天，發紅包等等動作 [答案](#room-reply)
3. 如何進入聊天室有失敗會怎樣 `答案:失敗會直接close連線`
4. 如何在聊天室發訊息 [請看前台訊息推送API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#71c23912-6830-4c42-a675-ea6ae31f5d80)
5. 如何接收聊天室訊息 [答案](#message)
6. 聊天室訊息結構 [答案](#message-reply)
7. 封鎖狀態下進入聊天室會怎樣 [答案](#close-reply)
8. 禁言狀態下聊天會怎樣 [請看前台訊息推送API範例-禁言中](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#71c23912-6830-4c42-a675-ea6ae31f5d80)
9. 如何在聊天室發紅包 [請看前台發紅包API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#f6c5fb74-cd42-40fb-bb66-1c2bde3419af)
10. 聊天室發紅包訊息結構 [答案](#message-reply)
11. 如何搶紅包 [請看前台搶紅包API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#68e59552-6c95-45ab-9f9e-3eaad7c1989d)
12. 如何切換聊天室房間[答案](#change-room)
13. 如何拿到歷史紀錄 [請看前台歷史紀錄API範例](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#6da9dbce-f64a-4dc1-8b15-a015bfd83d1c)
14. 如何跟聊天室做心跳 [答案](#heartbeat)
15. 聊天室心跳週期是多少 `答案:server要求1分鐘一次`
16. 如何產生一個跟websocket溝通的Protocol [答案](#buffer)
17. 如何拿到房間在線人數 [答案](#heartbeat-reply)

## Protocol Body格式

採用websocket做binary傳輸，聊天室推給client訊息內容如下格式

| name      | length  | remork                        | 說明                   |
| --------- | ------- | ----------------------------- | ---------------------- |
| Package   | 4 bytes | header + body length          | 整個Protocol bytes長度 |
| Header    | 2 bytes | protocol header length        | Package  - Boyd        |
| Operation | 4 bytes | [Operation](#operation)       | Protocol的動作         |
| Body      | 不固定  | 傳送的資料16bytes之後就是Body | json格式               |

![arch](./doc/protocol.png)

### Package

用於表示本次傳輸的binary內容總長度是多少(header + body)

### Header

用來說明Protocol Header binary 長度是多少

### Operation

不同的Operation說明本次Protocol資料是什麼，如心跳回覆,訊息等等

| value | 說明                                         | body   |
| ----- | -------------------------------------------- | ------ |
| 1     | [要求連線到某一個房間](#room)                | json   |
| 2     | [連線到某一個房間結果回覆](#room-reply)      | json   |
| 3     | [發送心跳](#heartbeat)                       | 無Body |
| 4     | [回覆心跳結果](#heartbeat-reply)             | int32  |
| 5     | [聊天室批次訊息](#message)                   | json   |
| 6     | [聊天室訊息](#message-raw)                   | json   |
| 7     | [更換房間](#change-room)                     | json   |
| 8     | [回覆更換房間結果](#change-room-reply)       | json   |
| 9     | [取消置頂訊息](#cancle-header-message-reply) | json   |
| 20    | [聊天室踢人](#close-Reply)                   | json   |



### Body

聊天室的訊息內容

### Buffer

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
            // 所以需要for將body內所有訊息取出來，每一則訊息還是依照解析 Protocol
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

#### Room Reply

Operation = `2`=> 連線到某一個房間結果回覆Body

```json
{
    "room_id": 1000,
    "uid": "82ea16cd2d6a49d887440066ef739669",
    "key": "defb108d-3d51-475a-b266-4a7f459e7a59",
    "message": "聊天室目前禁言状态，无法发言",
    "status": true,
    "permission": {
        "is_message": true,
        "is_red_envelope": true
    }
}
```

| name                       | 說明                                                         |
| -------------------------- | ------------------------------------------------------------ |
| uid                        | user uid，發送訊息會用到                                     |
| key                        | 這次web socket連線id，發送訊息會用到                         |
| room_id                    | 房間id                                                       |
| message                    | 用於顯示給用戶當前房間狀況，訊息，通知等等，可放在發話輸入欄做提示 |
| status                     | 進入房間是否成功，只代表進入不代表任何權限                   |
| permission.is_message      | true: 可聊天，false: 不可聊天                                |
| permission.is_red_envelope | true: 可發/搶紅包，false: 不可發/搶紅包                      |

進入房間有幾種狀況，原因都會寫在`message`

1. status => `true` 
   1. 房間無法發話
   2. 用戶無法發話

2. status => false
   1. 用戶被封鎖
   2. 房間不存在
   3. 房間已關閉
   4. 其他例外狀況

#### Close Reply

Operation = `20`=> 踢出房間

如何會員被封鎖 or 被踢，server close連線後會回傳一個Body內容如下

```json
{
    "message": "您在封鎖状态，无法进入聊天室"
}
```



#### Heartbeat Reply

Operation = `4`=> 回覆心跳結果

```
body是內容是該房間在線人數，是一個int32
```

![arch](./doc/heartbeatReply.png)



#### Message Reply

Operation = `6`=> 單筆訊息

![arch](./doc/message_raw.png)

Ｍessage type 

每個message json 內都有一個type來判斷訊息種類

| Value        | 說明           |
| ------------ | -------------- |
| message      | 普通訊息       |
| red_envelope | 紅包訊息       |
| top          | 公告(置頂)訊息 |

普通訊息 or 公告(置頂)訊息

```json
{
    "id": 4001,
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "type": "message",
    "name": "sam78",
    "avatar": "female",
    "message": "測試",
    "time": "12:37:00"
}
```

| name    | 說明                     | 格式     |
| ------- | ------------------------ | -------- |
| Id      | 訊息id                   | Int      |
| uid     | 訊息人uid                | string   |
| type    | 訊息類型                 | string   |
| name    | 訊息人名稱               | string   |
| avatar  | 頭像名稱 [類型](#avatar) | string   |
| message | 訊息                     | string   |
| time    | 發送時間                 | 時:分:秒 |

紅包訊息

```json
{
    "id": 4404,
    "uid": "0d641b03d4d548dbb3a73a2197811261",
    "type": "red_envelope",
    "name": "sam78",
    "avatar": "male",
    "message": "發大財",
    "time": "12:37:00",
    "red_envelope": {
        "id": "0d641b03d4d548dbb3a73a2197811261",
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjY4NzkzMTMsImlkIjoiMWI5MTZiNDc4YzBjNGZjMzhmMGE0MzE1NjMwNjExMTQiLCJ1aWQiOiIwZDY0MWIwM2Q0ZDU0OGRiYjNhNzNhMjE5NzgxMTI2MSJ9.pgyltHiT11XcZySJPiuetV35OXU-wgQ4XtU_UTzwghU",
        "expired": "2019-08-27T12:15:13+08:00"
    }
}
```

| name                 | 說明                     | 格式     |
| -------------------- | ------------------------ | -------- |
| Id                   | 訊息id                   | Int      |
| uid                  | 發紅包人的uid            | string   |
| type                 | 訊息種類                 | string   |
| name                 | 發紅包人名稱             | string   |
| avatar               | 頭像名稱 [類型](#avatar) | 名稱     |
| message              | 紅包說明                 | string   |
| time                 | 發送時間                 | 時:分:秒 |
| red_envelope.id      | 紅包id                   | string   |
| red_envelope.token   | 搶紅包的token            | string   |
| red_envelope.expired | 紅包過期時間             | RF3339   |



#### Change Room Reply

Operation = `8`=> 回覆更換房間結果

![arch](./doc/changeRoomReply.png)

```json
{
    "room_id": 1000,
    "status": true,
    "message": "聊天室目前禁言状态，无法发言",
    "permission": {
        "is_message": true,
        "is_red_envelope": true
    }
}
```

| name                       | 說明                                                         | 格式   |
| -------------------------- | ------------------------------------------------------------ | ------ |
| room_id                    | 新房間id                                                     | int    |
| status                     | 進入房間是否成功                                             | bool   |
| ｍessage                   | 用於顯示給用戶當前房間狀況，訊息，通知等等，可放在發話輸入欄做提示 | String |
| permission.is_message      | true: 可聊天，false: 不可聊天                                | bool   |
| permission.is_red_envelope | true: 可發/搶紅包，false: 不可發/搶紅包                      | bool   |

>`status` 與 `message` 可以參考[進入房間](#room-reply)

#### Cancle Header Message Reply

Operation = `9`=> 取消置頂訊息

```json
{
    "id": 5001
}
```

| name | 說明       | 格式 |
| ---- | ---------- | ---- |
| id   | 置頂消息id | int  |



## Web Socket

### Room

跟websocket建立完連線後將以下json包裝成[Protocol](#protocol-body)發送至websocket，Protocol Operation[參考](#operation)

```json
  {
      "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTcyMTE2NTAsIm5iZiI6MTU1NzIxMTY1MCwiaXNzIjoibG9naW4iLCJzZXNzaW9uX3Rva2VuIjoiZjc2OTYyM2Y0YTNlNDE4MWE4NzAwYWNkYTE3NzE1MmIiLCJkYXRhIjp7InVpZCI6IjEyNTdlN2Q5ZTFjOTQ0ZWY5YTZmMTI5Y2I5NDk1ZDAyIiwidXNlcm5hbWUiOiJyb290In19.7VJxH3tQpnJqWTlPbId7f0Rt7eQoaVvaJmbWxtHTqRU",
      "room_id": 1000
  }
```

| name    | 說明                                                         | 格式   | 限制 |
| ------- | ------------------------------------------------------------ | ------ | ---- |
| token   | 請自行透過[paras服務](https://jetfueltw.postman.co/collections/2999246-70e3e838-def8-4d53-ba12-e1b2797eec57?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#13ab4886-ab04-4e38-90b1-a0e73c4c8e0f)登入後拿到token | string | jwt  |
| room_id | 想要進入的房間id，透過`paras服務`取得                        | Int    |      |

| 結果             | 說明                      |
| ---------------- | ------------------------- |
| 成功             | [Response](#room-reply)   |
| 失敗(非封鎖造成) | server會把websocket close |



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

| 結果 | 說明                         |      |
| ---- | ---------------------------- | ---- |
| 成功 | [Response](#heartbeat-reply) |      |
| 失敗 | 失敗就會close連線            |      |



### Change Room

Boyd內容如下，Protocol Operation[參考](#operation)

```json
{
    "room_id": 1000
}
```



| name    | 說明     | 格式   |      |
| ------- | -------- | ------ | ---- |
| room_id | 新房間id | string |      |

![arch](./doc/changeRoom.png)

| 結果 | 說明                           |      |
| ---- | ------------------------------ | ---- |
| 成功 | [Response](#change-room-reply) |      |
| 失敗 | 失敗就會close連線              |      |

## Avatar

每個訊息頭像，請參考[Message Reply](#message-reply)  `avatar`欄位值

| Value  | 說明                                        |
| ------ | ------------------------------------------- |
| female | 女性頭像 ![arch](./doc/avatar/female.svg)   |
| male   | 男性頭像![arch](./doc/avatar/male.svg)      |
| other  | 其他性別頭像![arch](./doc/avatar/other.svg) |
| root   | 管理員頭像![arch](./doc/avatar/root.svg)    |



## Error Code

| Code               | http code | 說明                                                         |
| ------------------ | --------- | ------------------------------------------------------------ |
| 10024042, 15024042 | 404       | 沒有房間資料                                                 |
| 10024041,12024041  | 404       | 找不到會員資料                                               |
| 10024043           | 404       | 目前房间已关闭                                               |
| 10024291           | 429       | 1秒内只能发一则消息                                          |
| 10024292           | 429       | 10秒内相同讯息3次，自动禁言10分钟                            |
| 10024011, 15024031 | 401       | 用户认证失败 (token解不開)                                   |
| 10024012           | 401       | 用户认证失败 (token 資料有問題)                              |
| 10024013           | 401       | 用户认证失败 (時效已過)                                      |
| 10024014           | 401       | 您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元 |
| 10024015           | 401       | 您在永久禁言状态，无法发言                                   |
| 10024016           | 401       | 您在禁言状态，无法发言                                       |
| 10024017           | 401       | 聊天室目前禁言状态，无法发言                                 |
| 10024220, 15024220 | 422       | 資料格式有問題，如json欄位驗證失敗                           |
| 12024020           | 402       | 余额不足                                                     |
| 15021001           | 400       | 拼手氣紅包金額不能小於包數                                   |
| 15024044, 15024045 | 404       | 红包不存在                                                   |
| 10024000, 15024000 | 400       | json欄位型態錯誤，如string放int                              |
| 10025000, 15025000 | 500       | 伺服器出問題                                                 |



## Version Update

1. 進入某房間成功後回覆的json結構 [參考](#room-reply)
   1. room_id 由string 改成 int
   2. permission.is_banned 更名為permission.permission
2. 增加聊天室強制中斷連線通知 [參考](#close-reply)
   1. 如被封鎖，被後台踢掉，被自動踢掉等情況都會這樣
   2. 會先發一組json內容後，server在將web socket 做close
3. 前台跟聊天室web socket 做心跳請改成1分鐘1次 [參考](#heartbeat)
4. 聊天室推送的訊息增加`type`欄位表示種類，原先是利用Protocol Operation 來表示，[參考](#message-reply)
5. 切換房間原先room_id是string請改成int [參考](#change-room)
6. 紅包訊息結構內的red_envelope.expired原先是時間戳記改為RFC3339
7. [發訊息API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#71c23912-6830-4c42-a675-ea6ae31f5d80) 原先需要帶`key ` and `uid` 現在只需帶`room_id` `message`
8. [發紅包API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#f6c5fb74-cd42-40fb-bb66-1c2bde3419af) 不需要帶`key` and `uid`

