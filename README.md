# Alakazam 聊天室

- [啟動聊天室](./docker/docker-compose/README.md)
- [快速開始介接聊天室](#quick-reference)
- [快速建置聊天室服務](#quick-start)
- [依賴工具](#dependencies)
- [架構](#architecture)
- [功能](#features)
- [聊天室Web Socket協定](#protocol-body)
- [Web Socket](#web-socket)
- [會員權限動作](#permission)
- [監控](./docker/docker-compose/README.md#metrics)
- [注意事項](#precautions)

## Quick Start

``` bash
make run
```

## Architecture

### 系統架構

![arch](./doc/architecture.png)

### 負責功能

1. comet: 用戶連結

   web socket Server

   grpc server

2. logic: 業務邏輯

   http server

   grpc server

3. job: 訊息推送

   grpc server

4. message: 訊息持久化處理

5. seq:

> `comet`．`logic`．`job` 為核心系統。

> 消息系統中的kafka也可視為外部服務。

### 系統流程

![arch](./doc/flow.png)

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
- [x] 用戶拉取歷史訊息(mysql)
- [x] 房間禁言
- [x] 踢人
- [x] 用戶搶紅包
- [x] 紅包時效
- [x] 用戶查詢紅包明細
- [x] 自動返還過期紅包剩餘金額
- [x] 跟投訊息
- [x] 發送跟投訊息
- [x] 管理員多房間推送訊息
- [x] 公告(置頂)多房間推送訊息
- [x] 管理員發紅包
- [x] 預定時間發佈紅包
- [x] 建立/更新房間(禁言,跟投,紅包,聊天限制)
- [x] 關閉房間
- [x] 房間在線人數
- [x] 遮蔽敏感詞
- [x] 訊息限速(1秒1則，10內不可重覆三次一樣訊息)
- [x] 依據性別切換頭像
- [x] 收費房
- [x] 升級通知
- [x] 各服務監控
- [x] kafka監控
- [x] golang pprof
- [x] 歷史訊息redis
- [x] Kafka and zk config 調整
- [x] cache 優化

TODO

- [ ] json改用msgp
- [ ] zk監控
- [ ] 註冊中心(etcd or zk)
- [ ] 整合測試 bash shell
- [ ] 歷史訊息 mongodb
- [ ] 訊息壓縮
- [ ] 訊息持久化速度慢(透過監控offset lag發現)
- [ ] 架構圖
- [ ] 壓測資料
- [ ] 完成壓測腳本
- [ ] websocket替換成原生
- [ ] time替換成原生

## Quick Reference

前台：

1. 如何進入聊天室 [答案](#room)
2. 進入聊天室後的回覆 or 如何知道用戶在聊天室相關權限，如聊天，發紅包等等動作 [答案](#room-reply)
3. 如何進入聊天室有失敗會怎樣 `答案:失敗會直接close連線`
4. 如何在聊天室發訊息 [請看前台訊息推送API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?
   version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#71c23912-6830-4c42-a675-ea6ae31f5d80)
5. 如何接收聊天室訊息 [答案](#message)
6. 聊天室訊息結構 [答案](#message-reply)
7. 封鎖狀態下進入聊天室會怎樣 [答案](#close-reply)
8. 禁言狀態下聊天會怎樣 [請看前台訊息推送API範例-禁言中](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?
   version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#71c23912-6830-4c42-a675-ea6ae31f5d80)
9. 如何在聊天室發紅包 [請看前台發紅包API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?
   version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#f6c5fb74-cd42-40fb-bb66-1c2bde3419af)
10. 聊天室發紅包訊息結構 [答案](#message-reply)
11. 如何搶紅包 [請看前台搶紅包API](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?
    version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#68e59552-6c95-45ab-9f9e-3eaad7c1989d)
12. 如何切換聊天室房間[答案](#change-room)
13. 如何拿到歷史紀錄 [請看前台歷史紀錄API範例](https://jetfueltw.postman.co/collections/6851408-6a660dbe-4cc3-4c3e-94b5-897071b2802b?
    version=latest&workspace=56a5a88a-bfd1-46b5-8102-a2ca97183649#6da9dbce-f64a-4dc1-8b15-a015bfd83d1c)
14. 如何跟聊天室做心跳 [答案](#heartbeat)
15. 聊天室心跳週期是多少 `答案:server要求1分鐘一次`
16. 如何產生一個跟websocket溝通的Protocol [答案](#buffer)
17. 如何拿到房間在線人數 [答案](#heartbeat-reply)

## Protocol Body格式

採用websocket做binary傳輸，聊天室推給client訊息內容如下格式

| name      | length  | remork                        | 說明                   |
| --------- | ------- | ----------------------------- | ---------------------- |
| Package   | 4 bytes | header + body length          | 整個Protocol bytes長度 |
| Header    | 2 bytes | protocol header length        | Package - Boyd        |
| Operation | 4 bytes | [Operation](#operation)       | Protocol的動作         |
| Body      | 不固定  | 傳送的資料16bytes之後就是Body | json格式               |

![arch](./doc/protocol.png)

### Package

用於表示本次傳輸的binary內容總長度是多少(header + body)

### Header

用來說明Protocol Header binary 長度是多少

### Operation

不同的Operation說明本次Protocol資料是什麼，如心跳回覆,訊息等等

| value | 說明                                           | body   |
| ----- | ---------------------------------------------- | ------ |
| 1     | [要求連線到某一個房間](#room)                      | json   |
| 2     | [連線到某一個房間結果回覆](#room-reply)             | json   |
| 3     | [發送心跳](#heartbeat)                           | 無Body |
| 4     | [回覆心跳結果](#heartbeat-reply)                  | int32  |
| 5     | [聊天室批次訊息](#message)                         | json   |
| 6     | [聊天室訊息](#message-reply)                       | json   |
| 7     | [更換房間](#change-room)                           | json   |
| 8     | [回覆更換房間結果](#change-room-reply)              | json   |
| 9     | [取消置頂訊息](#cancle-header-message-reply)       | json   |
| 10    | [發送付費房驗證 - 月卡](#paid-room-expiry)          | 無Body   |
| 11    | [回覆付費房驗證 - 月卡](#paid-room-expiry-reply)    | json   |
| 12    | [發送付費房鑽石付費結果](#paid-room-diamond)          | 無Body   |
| 13    | [回覆付費房鑽石付費結果](#paid-room-diamond-reply)    | json   |
| 20    | [聊天室踢人](#close-reply)                         | json   |

Diamond

### Body

聊天室的訊息內容

### Buffer

如何在前端產生一個Protocol Body，以進入聊天室為例子

> 本例子js只是解釋如何產生一個Protocol，實際寫法請自行實作

```bash
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

```js
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

```bash
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
  "message": "目前房间已关闭",
  "status": false,
  "permission": {
    "is_message": false,
    "is_red_envelope": false,
    "is_bets": false
  },
  "permission_message": {
    "is_message": "聊天室目前禁言状态，无法发言",
    "is_red_envelope": "请先登入会员",
    "is_bets": "请先登入会员"
  }
}
```

| name                               | 說明                                         |
| ---------------------------------- | -------------------------------------------- |
| uid                                | user uid，發送訊息會用到                     |
| key                                | 這次web socket連線id，發送訊息會用到         |
| room_id                            | 房間id                                       |
| message                            | 用於顯示給用戶在進入當前房間提示或是錯誤訊息 |
| status                             | 進入房間是否成功，只代表進入不代表任何權限   |
| permission.is_message              | true: 可聊天，false: 不可聊天                |
| permission.is_red_envelope         | true: 可發/搶紅包，false: 不可發/搶紅包      |
| permission.is_bets                 | true: 可跟注，false: 不可跟注                |
| permission.is_manage               | true: 房管，false: 不是房管                |
| permission_message.is_message      | 進入房間後關於發言相關提示訊息               |
| permission_message.is_red_envelope | 進入房間後關於搶/發紅包相關提示訊息          |
| permission_message.is_bets         | 進入房間後關於跟注相關提示訊息               |
| permission_message.is_manage       | 你已被主播设为房管人员               |

進入房間有幾種狀況，原因都會寫在`message`

1. status => `true`
    1. 房間無法發話
    2. 用戶無法發話
    3. 沒有任何異常，為空值

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

```text
body是內容是該房間在線人數，是一個int32
```

![arch](./doc/heartbeatReply.png)

#### Message Reply

Operation = `6`=> 單筆訊息

![arch](./doc/message_raw.png)

Message type

每個message json 內都有一個type來判斷訊息種類

| Value           | 說明         |
| --------------- | ----------- |
| message         | 一般         |
| private_message | 私訊         |
| top             | 置頂         |
| bulletin        | 進場公告      |
| red_envelope    | 紅包         |
| bets            | 跟投         |
| gift            | 禮物/打賞     |
| hint            | 提示         |
| bets_win_reward | 投注中獎打賞   |
| quiz_bets       | 競猜跟跟投    |
| open_live       | 開播         |
| close_live      | 關播         |
| permission      | 權限         |
| level           | 升級通知      |
| anchor_level    | 主播升級      |

![arch](./doc/message/live_message.png)

| 欄位名稱     | 說明         | 格式          |
| ------------ | ------------ | ------------- |
| id           | 訊息id       | int           |
| type         | 訊息類型     | string        |
| time         | 發送時間     | 時:分:秒      |
| timestamp    | 訊息發送時間 | 時間戳記(int) |
| display      | 顯示訊息資料 |               |
| user         | 發訊息人資料 |               |
| red_envelope | 紅包資料     |               |
| bet          | 跟投         |               |
| gift         | 禮物         |               |
| live_stream  | 直播         |               |
| follow       | 關注         |               |

display 顯示訊息資料格式

| 欄位名稱 | 說明                     | 格式    |
| -------- | ------------------------ | ------- |
| user     | 用戶資料                     | user  |
| level    | 發訊息人等級                 | level |
| title   | 發訊息人身份 | title  |
| message   | 訊息資料 | message  |
| is_manage   | 頭像名稱 [類型](#avatar) | bool  |
| background_color   | 訊息背景顏色 | string  |
| background_image   | 訊息背景圖 | string  |

display.user: 顯示用戶資料

| 欄位名稱 | 說明                     | 格式    |
| -------- | ------------------------ | ------- |
| text     | 人名                     | string  |
| color    | 字體顏色                 | #000000 |
| avatar   | 頭像名稱 [類型](#avatar) | string  |

display.level: 發訊息人等級

| 欄位名稱         | 說明         | 格式    |
| ---------------- | ------------ | ------- |
| text             | 身份名稱     | string  |
| color            | 字體顏色     | #000000 |
| background_color | 字體背景顏色 | #000000 |

display.title: 發訊息人身份

| 欄位名稱         | 說明         | 格式    |
| ---------------- | ------------ | ------- |
| text             | 身份名稱     | string  |
| color            | 字體顏色     | #000000 |
| background_color | 字體背景顏色 | #000000 |

display.message: 訊息資料

| 欄位名稱                | 說明           | 格式      |
| ----------------------- | -------------- | --------- |
| text                    | 訊息           | string    |
| color                   | 字體顏色(預設) | #000000   |
| background_color        | 訊息框顏色     | #0000003f |
| entity                  | 文字實體       | array     |
| entity.type             | 文字類型       | string    |
| entity.offset           | 偏移第幾個字   | int       |
| entity.length           | 字範圍長度     | int       |
| entity.color            | 字體顏色       | #000000   |
| entity.background_color | 標籤背景       | #A680B8   |

display.background_image: 訊息框背景圖

| 欄位名稱 | 說明 | 格式   |
| -------- | ---- | ------ |
| text     | 種類 | string |

display.background_image.`type`=linear-gradient

| 欄位名稱 | 說明                   | 格式   |
| -------- | ---------------------- | ------ |
| to       | 漸層方向 `right``left` | string |
| color    | 顏色                   | array  |

user: 發訊息人資料

| 欄位名稱 | 說明                     | 格式   |
| -------- | ------------------------ | ------ |
| uid      | 發訊息人 會員uuid        | string |
| name     | 發訊息人名稱             | string |
| avatar   | 頭像名稱 [類型](#avatar) | string |
| type     | 身份(訪客`guest` 營運`market` 一般`player` 主播`streamer` 房管`manage`) | string |

red_envelope: 紅包資料

| 欄位名稱 | 說明          | 格式   |
| -------- | ------------- | ------ |
| id       | 紅包id        | string |
| token    | 搶紅包的token | string |
| expired  | 紅包過期時間  | RF3339 |

bet: 跟投

| 欄位名稱         | 說明           | 格式     |
| ---------------- | -------------- | -------- |
| game_id          | 遊戲 id        | int      |
| game_name        | 遊戲 名稱      | string   |
| period_number    | 本注期號       | int      |
| count            | 下注總筆數     | int      |
| total_amount     | 下注總金額     | int      |
| bets.name        | 下注號碼名稱   | string   |
| bets.odds_code   | 賠率代號       | string   |
| bets.items       | 組合號碼       | []string |
| bets.trans_items | 組合號碼(中文) | []string |
| bets.amount      | 下注金額       | int      |

gift: 禮物

| 欄位名稱                  | 說明                   | 格式        |
| ------------------------- | ---------------------- | ----------- |
| gift_id                   | 禮物 id                | int         |
| total_amount              | 今日房間總收禮物金額   | float       |
| amount                    | 本次送禮or打賞金額     | float       |
| message                   | 送禮訊息               | string      |
| combo.id                  | 禮物連擊次數           | int         |
| combo.duration_ms         | 連擊訊息顯示多久(毫秒) | int         |
| show_animation            | 是否顯示動畫           | bool        |
| hint_box.background_image | hint box 背景圖        | string(url) |
| hint_box.duration_ms      | hint box顯示多久(毫秒) | int         |
| entity                    | 文字實體               | array       |
| entity.type               | 文字類型               | string      |
| entity.offset             | 偏移第幾個字           | int         |
| entity.length             | 字範圍長度             | int         |
| entity.color              | 字體顏色               | #000000     |
| entity.background_color   | 標籤背景               | #A680B8     |

permission: 權限

| 欄位名稱 | 說明      | 格式 |
| -------- | --------- | ---- |
| is_message  | 是否禁言 | boolean  |
| is_red_envelope  | 是否發/搶紅包 | boolean  |
| is_bets  | 是否發跟投 | boolean  |
| is_manage  | 是否為房管 | boolean  |

live_stream: 直播

| 欄位名稱 | 說明      | 格式 |
| -------- | --------- | ---- |
| chat_id  | 直播間 id | int  |

entity.type: 文字實體內容

| 值       | 說明     | 格式   |
| -------- | -------- | ------ |
| username | user名稱 | string |
| button   | 按鈕     | string |

##### 一般訊息(用戶)

```json
{
  "id": 4001,
  "type": "message",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": {
      "text": "sam78",
      "color": "#000000",
      "avatar": "female"
    },
    "level": {
      "text": "vip1",
      "color": "#DF3030",
      "background_color": "#C7B9C1"
    },
    "title": {
      "text": "房管",
      "color": "#E5C4C4",
      "background_color": "#332121"
    },
    "message": {
      "text": "測試",
      "color": "#A680B8",
      "background_color": "#0000003f",
      "entity": null
    },
    "is_manage": false,
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
    "type": 2
  }
}
```

##### 一般訊息(房管)

```json
{
  "id": 4001,
  "type": "message",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": {
      "text": "sam78",
      "color": "#000000",
      "avatar": "female"
    },
    "level": {
      "text": "vip1",
      "color": "#DF3030",
      "background_color": "#C7B9C1"
    },
    "title": {
      "text": "房管",
      "color": "#E5C4C4",
      "background_color": "#332121"
    },
    "message": {
      "text": "測試",
      "color": "#A680B8",
      "background_color": "#0000003f",
      "entity": null
    },
    "is_manage": true,
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
    "type": 2
  }
}
```

![arch](./doc/message/user.png)

##### 私訊

```json
{
  "id": 14003,
  "type": "private_message",
  "time": "10:55:17",
  "timestamp": 1599015317,
  "display": {
    "user": {
      "text": "Sam",
      "color": "#FFFFAA",
      "avatar": "male"
    },
    "level": null,
    "title": {
      "text": "私讯",
      "color": "#FFFFFF",
      "background_color": "#F79EB6"
    },
    "message": {
      "text": "測試",
      "color": "#FFFFFF",
      "background_color": "#00000000",
      "entity": null
    },
    "background_color": "#38A2DB7F",
    "background_image": null
  },
  "user": {
    "uid": "79a5ded9a5f34b57bed0786ae4cbb21b",
    "name": "Sam",
    "avatar": "male",
    "type": 2
  }
}
```

##### 公告(管理員)

```json
{
  "id": 59402,
  "type": "bulletin",
  "time": "14:37:50",
  "timestamp": 1591857470,
  "display": {
    "user": null,
    "level": null,
    "title": {
      "text": "系统",
      "color": "#FFFFFF",
      "background_color": "#FC8813"
    },
    "message": {
      "text": "測試",
      "color": "#FFFFAA",
      "background_color": "#00000000",
      "entity": null
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "root",
    "name": "管理员",
    "avatar": "root",
    "type": 2
  }
}
```

![arch](./doc/message/root.png)

##### 置頂訊息

```json
{
  "id": 4001,
  "type": "top",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": {
      "text": "管理員",
      "color": "#FFFFFF",
      "avatar": "female"
    },
    "level": null,
    "title": null,
    "message": {
      "text": "你好嗎？",
      "color": "#FFFFFF",
      "entity": null
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "root",
    "name": "管理員",
    "avatar": "root",
    "type": 2
  }
}
```

![arch](./doc/message/top.png)

##### 系統訊息

```json
{
  "id": 4001,
  "type": "message",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": null,
    "level": null,
    "title": {
      "text": "系統",
      "color": "#FFFFFF",
      "background_color": "#E5CB29"
    },
    "message": {
      "text": "測試",
      "color": "#A680B8",
      "background_color": "#0000003f",
      "entity": null
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": null
}
```

![arch](./doc/message/system.png)

##### 投注中獎打賞訊息

```json
{
  "id": 53402,
  "type": "bets_win_reward",
  "time": "10:54:50",
  "timestamp": 1591757690,
  "display": null,
  "user": {
    "uid": "0d641b03d4d548dbb3a73a2197811261",
    "name": "nickname_1",
    "avatar": "male"
  },
  "bets_win_reward": {
    "text": "恭喜您中奖 金额＄1,010.01 打賞主播",
    "color": "#FFFFAA",
    "background_color": "#F8565699",
    "entity": [
      {
        "type": "button",
        "offset": 18,
        "length": 4,
        "color": "#FFFFAA",
        "background_color": "#F85656"
      }
    ]
  }
}
```

![arch](./doc/message/will.png)

##### 進入房間訊息

```json
{
  "id": 48801,
  "type": "hint",
  "time": "11:58:37",
  "timestamp": 1591243117,
  "display": {
    "user": null,
    "title": null,
    "level": {
      "level": 100,
      "text": "会员",
      "color": "#FFFFFF",
      "background_color": "#7FC355"
    },
    "message": {
      "text": "nickname_1进入聊天室",
      "color": "#FFFFAA",
      "background_color": "#00000000",
      "entity": [
        {
          "type": "username",
          "offset": 0,
          "length": 10,
          "color": "#7CE7EB",
          "background_color": "#00000000"
        }
      ]
    },
    "background_color": null,
    "background_image": {
      "type": "linear-gradient",
      "to": "right",
      "color": {
        "0": "#FC881380",
        "99": "#FC881300"
      }
    }
  },
  "user": null
}
```

<img src="./doc/message/in_room.png" alt="arch" style="zoom:80%;" />

##### 送禮訊息

```json
{
  "id": 4001,
  "type": "gift",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": null,
    "level": null,
    "title": {
      "text": "會員",
      "color": "#FFFFFF",
      "background_color": "#FF0000"
    },
    "message": {
      "text": "sam78 送出禮物x1",
      "color": "#A680B8",
      "background_color": "#0000003f",
      "entity": [
        {
          "type": "username",
          "offset": 0,
          "length": 5,
          "color": "#A680B8",
          "background_color": "#0000003f"
        }
      ]
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
    "type": 2
  },
  "gift": {
    "gift_id": 1,
    "total_amount": 10,
    "amount": 5,
    "combo": {
      "count": 1,
      "duration_ms": 2000
    },
    "hint_box": null,
    "show_animation": false,
    "message": "送出跑車",
    "entity": null
  }
}
```

![arch](./doc/message/gift.png)

![arch](./doc/message/gift_hint.png)

##### 打賞訊息

```json
{
  "id": 4001,
  "type": "gift",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": null,
    "level": null,
    "title": {
      "text": "會員",
      "color": "#FFFFFF",
      "background_color": "#FF0000"
    },
    "message": {
      "text": "sam78 打賞10元",
      "color": "#DF3030",
      "background_color": "#0000003f",
      "entity": [
        {
          "type": "username",
          "offset": 0,
          "length": 5,
          "color": "#A680B8",
          "background_color": "#0000003f"
        }
      ]
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
    "type": 2
  },
  "gift": {
    "gift_id": 0,
    "total_amount": 0,
    "amount": 5,
    "combo": null,
    "show_animation": false,
    "hint_box": {
      "duration_ms": 3000,
      "background_color": "https://assets.292801.com/awcp/logo/hPH6edsqSY.png"
    },
    "message": "sam78 打賞10元",
    "entity": [
      {
        "type": "username",
        "offset": 0,
        "length": 5,
        "color": "#A680B8",
        "background_color": "#0000003f"
      }
    ]
  }
}
```

![arch](./doc/message/reward.png)

![arch](./doc/message/reward_hint.png)

##### 紅包訊息

```json
{
  "id": 4001,
  "type": "red_envelope",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": {
      "text": "sam78",
      "color": "#FFFFFF",
      "avatar": "female"
    },
    "level": null,
    "title": null,
    "message": {
      "text": "紅包",
      "color": "#FFFFFF",
      "background_color": "#0000003f",
      "entity": null
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
    "type": 2
  },
  "red_envelope": {
    "id": "0d641b03d4d548dbb3a73a2197811261",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NjY4NzkzMTMsImlkIjoiMWI5MTZiNDc4YzBjNGZjMzhmMGE0MzE1NjMwNjExMTQiLCJ1aWQiOiIwZDY0MWIwM2Q0ZDU0OGRiYjNhNzNhMjE5NzgxMTI2MSJ9.pgyltHiT11XcZySJPiuetV35OXU-wgQ4XtU_UTzwghU",
    "expired": "2019-08-27T12:15:13+08:00"
  }
}
```

##### 跟投訊息

```json
{
  "id": 4001,
  "type": "bets",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": {
      "text": "sam78",
      "color": "#FFFFFF",
      "avatar": "female"
    },
    "level": null,
    "title": null,
    "message": {
      "text": "跟投",
      "color": "#DF3030",
      "background_color": "#0000003f",
      "entity": [
        {
          "type": "button",
          "offset": 6,
          "length": 2,
          "color": "#E5C4C4",
          "background_color": "#A680B8"
        }
      ]
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
    "type": 2
  },
  "bet": {
    "game_id": 4567,
    "game_name": "六合彩",
    "period_number": 1234,
    "count": 2,
    "total_amount": 25,
    "bets": [
      {
        "name": "冠軍",
        "odds_code": "1.pos.1",
        "items": [
          "9",
          "tiger"
        ],
        "trans_items": [
          "9",
          "虎"
        ],
        "amount": 15
      },
      {
        "name": "亞軍",
        "odds_code": "2.pos.1",
        "items": [
        ],
        "trans_items": [
        ],
        "amount": 10
      }
    ]
  }
}
```

##### 競猜跟投訊息

```json
{
  "id": 4001,
  "type": "quiz_bets",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": {
    "user": null,
    "level": null,
    "title": null,
    "message": {
      "text": "用户sam78在quiz test下注2元 ＋跟注",
      "color": "#DF3030",
      "background_color": "#0000003f",
      "entity": [
        {
          "type": "username",
          "offset": 2,
          "length": 5,
          "color": "#7CE7EB",
          "background_color": "#00000000"
        },
        {
          "type": "button",
          "offset": 21,
          "length": 4,
          "color": "#E5C4C4",
          "background_color": "#A680B8"
        }
      ]
    },
    "background_color": "#0000003f",
    "background_image": null
  },
  "user": {
    "uid": "3d641b03d4d548dbb3a73a2197811261",
    "name": "sam78",
    "avatar": "female",
     "type": "player",
     "lv": 13
  },
  "quiz_bet": {
    "quiz_id": 1,
    "quiz_name": "quiz test",
    "bet_name": "quiz test",
    "bet": "ass234",
    "amount": 2
  }
}
```

##### 關注訊息

```json
{
  "id": 3602,
  "type": "follow",
  "time": "14:13:29",
  "timestamp": 1597126409,
  "display": null,
  "user": {
    "uid": "79a5ded9a5f34b57bed0786ae4cbb21b",
    "name": "Sam",
    "avatar": "",
    "type": 2
  },
  "follow": {
    "total": 10
  }
}
```

##### 開播訊息

```json
{
  "id": 4001,
  "type": "open_live",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": null,
  "user": null,
  "live_stream": {
    "chat_id": 1
  }
}
```

##### 關播訊息

```json
{
  "id": 4001,
  "type": "close_live",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": null,
  "user": null,
  "live_stream": {
    "chat_id": 1
  }
}
```

##### permission

```json
{
  "id": 15404,
  "type": "permission",
  "time": "12:37:00",
  "timestamp": 1567579890,
  "display": null,
  "user": null,
  "permission": {
    "is_message": true,
    "is_red_envelope": true,
    "is_bets": false,
    "is_manage": false
  },
  "permission_message": {
    "is_message": "",
    "is_red_envelope": "",
    "is_bets": "请先登入会员",
    "is_manage": "你已被主播取消房管人员资格"
  }
}
```

##### 升級通知

```json
{
  "id": 28003,
  "type": "level",
  "time": "14:43:45",
  "timestamp": 1634539425,
  "user": {
    "uid": "f33abfce41c745f795262dad34aa9f9b",
    "name": "s***8",
    "avatar": "other",
    "type": "player"
  },
  "level": 100
}
```

##### 主播升級通知

```json
{
  "id": 28003,
  "type": "anchor_level",
  "time": "14:43:45",
  "timestamp": 1634539425,
  "user": {
    "uid": "f33abfce41c745f795262dad34aa9f9b",
    "name": "s***8",
    "avatar": "other",
    "type": "streamer"
  },
  "level": 100
}
```

#### Change Room Reply

Operation = `8`=> 回覆更換房間結果

![arch](./doc/changeRoomReply.png)

```json
{
  "room_id": 1000,
  "uid": "82ea16cd2d6a49d887440066ef739669",
  "key": "defb108d-3d51-475a-b266-4a7f459e7a59",
  "message": "目前房间已关闭",
  "status": false,
  "permission": {
    "is_message": false,
    "is_red_envelope": false,
    "is_bets": false,
    "is_mag": false
  },
  "permission_message": {
    "is_message": "聊天室目前禁言状态，无法发言",
    "is_red_envelope": "请先登入会员",
    "is_bets": "请先登入会员"
  }
}
```

| name                               | 說明                                         |
| ---------------------------------- | -------------------------------------------- |
| uid                                | user uid，發送訊息會用到                     |
| key                                | 這次web socket連線id，發送訊息會用到         |
| room_id                            | 房間id                                       |
| message                            | 用於顯示給用戶在進入當前房間提示或是錯誤訊息 |
| status                             | 進入房間是否成功，只代表進入不代表任何權限   |
| permission.is_message              | true: 可聊天，false: 不可聊天                |
| permission.is_red_envelope         | true: 可發/搶紅包，false: 不可發/搶紅包      |
| permission.is_bets                 | true: 可跟注，false: 不可跟注                |
| permission_message.is_message      | 進入房間後關於發言相關提示訊息               |
| permission_message.is_red_envelope | 進入房間後關於搶/發紅包相關提示訊息          |
| permission_message.is_bets         | 進入房間後關於跟注相關提示訊息               |

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

```js
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

### Paid Room Expiry

Operation = `10`=> 付費房月卡效期驗證。

根據需求中要求週期定期發送月卡效期驗證，確保會員月卡尚載效期內。請利用送一個herder Operation值為`10`，body為空的[Protocol](#protocol-body)，以下是一個簡單的js範例。

```js
var headerBuf = new ArrayBuffer(rawHeaderLen);
var headerView = new DataView(headerBuf, 0);
headerView.setInt32(packetOffset, rawHeaderLen);
headerView.setInt16(headerOffset, rawHeaderLen);
headerView.setInt32(opOffset, 10);
```

### Paid Room Expiry Reply

Operation = `11`=> 回覆付費房月卡效期驗證結果。

Boyd內容如下

```json
{
  "expire": "2021-09-26T08:47:01+08:00",
  "is_allow": true
}
```

| name     | 說明        | 格式               |
| -------- | ---------- | ------------------ |
| expire   | 月卡效期    | date time (RFC3339) |
| is_allow | 驗證是否通過 | bool               |

### Paid Room Diamond

Operation = `12`=> 付費房 - 鑽石付費。

根據需求中要求週期定期發送付費房鑽石付費請求。請利用送一個herder Operation值為`12`，body為空的[Protocol](#protocol-body)，以下是一個簡單的js範例。

`PS. 一分鐘內只會採用針對第一筆付費請求進行扣款。`

```js
var headerBuf = new ArrayBuffer(rawHeaderLen);
var headerView = new DataView(headerBuf, 0);
headerView.setInt32(packetOffset, rawHeaderLen);
headerView.setInt16(headerOffset, rawHeaderLen);
headerView.setInt32(opOffset, 12);
```

### Paid Room Diamond Reply

Operation = `13`=> 回覆付費房鑽石付費結果。

Boyd內容如下

付費成功

```json
{
  "is_allow": true,
  "diamond": 990.11,
  "paid_time": "2021-09-08T10:06:09+08:00"
}
```

`PS. 因為一分鐘內只會採用針對第一筆付費請求進行扣款，而上次扣款成功一分鐘內的付款請求會返回上次付款成功時的付款時間與剩餘鑽石數量。`

付費失敗

```json
{
  "is_allow": false,
  "error": "余额不足"
}
```

| name      | 說明           | 格式             |
| --------- | -------------- | ---------------- |
| is_allow  | 付費成功        | bool             |
| diamond   | 用附目前鑽石數量 | float            |
| paid_time | 付費時間        | string (RFC3339) |
| error     | 錯誤訊息        | string  |

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
| 10025001, 15025001 | 404       | 沒有房間資料                                                 |
| 10022006           | 401       | 请先登入会员                                                 |
| 10022001,12024041  | 404       | 找不到會員資料                                               |
| 10025002           | 404       | 目前房间已关闭                                               |
| 10025003           | 429       | 1秒内只能发一则消息                                          |
| 10025004           | 429       | 10秒内相同讯息3次，自动禁言10分钟                            |
| 10022003, 15024031 | 401       | 用户认证失败 (token解不開)                                   |
| 10022004           | 401       | 用户认证失败 (token 資料有問題)                              |
| 10022005           | 401       | 用户认证失败 (時效已過)                                      |
| 10024014           | 401       | 您无法发言，当前发言条件：前%d天充值不少于%d元；打码量不少于%d元 |
| 10022008           | 401       | 您在永久禁言状态，无法发言                                   |
| 10022009           | 401       | 您在禁言状态，无法发言                                       |
| 10025006           | 401       | 聊天室目前禁言状态，无法发言                                 |
| 10024220, 15024220 | 422       | 資料格式有問題，如json欄位驗證失敗                           |
| 12024020           | 402       | 余额不足                                                     |
| 15021001           | 400       | 拼手氣紅包金額不能小於包數                                   |
| 15024001, 15024002 | 404       | 红包不存在                                                   |
| 10021003, 15021003 | 400       | json欄位型態錯誤，如string放int                              |
| 10020000, 15020000 | 500       | 伺服器出問題                                                 |

## Permission

角色說明

1. 遊客：非本站會員且未登入
2. 試玩：以試玩身份登入但不是本站會員
3. 會員：以註冊成為本站會員
4. 帶玩(營運)：後台創造的會員，下注等金錢動作不計入帳務
5. 管理員：操作後台人員
6. 系統：由各服務系統本身商業邏輯自動產生訊息，ex：注單、開獎資訊等等

會員權限對照

1. 進入房間，yes代表可以，管理員、系統本身不進入房間

   | 狀態       | 遊客 | 試玩 | 會員 | 帶玩(營運) |
               | ---------- | ---- | ---- | ---- | ---------- |
   | 正常       |      | Yes  | Yes  | Yes        |
   | 房間已關閉 |      |      |      |            |
   | 房間禁言   |      | Yes  | Yes  | Yes        |
   | 會員禁言   |      | Yes  | Yes  | Yes        |
   | 會員封鎖   |      |      |      |            |


2. 聊天(發送普通訊息)，yes代表可以

   | 狀態       | 遊客 | 試玩 | 會員 | 帶玩(營運) | 管理員 | 系統 |
               | ---------- | ---- | ---- | ---- | ---------- | ------ | ---- |
   | 正常       |      |      | Yes  | Yes        | Yes    | Yes  |
   | 房間已關閉 |      |      |      |            |        |      |
   | 房間禁言   |      |      |      |            | Yes    | Yes  |
   | 會員禁言   |      |      |      |            | Yes    | Yes  |
   | 會員封鎖   |      |      |      |            | Yes    | Yes  |


3. 發紅包，yes代表可以

   | 狀態       | 遊客 | 試玩 | 會員 | 帶玩(營運) | 管理員 | 系統 |
               | ---------- | ---- | ---- | ---- | ---------- | ------ | ---- |
   | 正常       |      |      | Yes  | Yes        | Yes    | Yes  |
   | 房間已關閉 |      |      |      |            |        |      |
   | 房間禁言   |      |      | Yes  | Yes        | Yes    | Yes  |
   | 會員禁言   |      |      | Yes  | Yes        | Yes    | Yes  |
   | 會員封鎖   |      |      |      |            | Yes    | Yes  |


4. 搶紅包，yes代表可以

   | 狀態       | 遊客 | 試玩 | 會員 | 帶玩(營運) |
               | ---------- | ---- | ---- | ---- | ---------- |
   | 正常       |      |      | Yes  | Yes        |
   | 房間已關閉 |      |      |      |            |
   | 房間禁言   |      |      | Yes  | Yes        |
   | 會員禁言   |      |      | Yes  | Yes        |
   | 會員封鎖   |      |      |      |            |


5. 發跟投，yes代表可以，什麼單子要發出來讓人跟投由營運or系統自動決定，不對外開放

   | 狀態       | 遊客 | 試玩 | 會員 | 帶玩(營運) | 管理員 | 系統 |
               | ---------- | ---- | ---- | ---- | ---------- | ------ | ---- |
   | 正常       |      |      |      |            | Yes    | Yes  |
   | 房間已關閉 |      |      |      |            |        |      |
   | 房間禁言   |      |      |      |            | Yes    | Yes  |
   | 會員禁言   |      |      |      |            | Yes    | Yes  |
   | 會員封鎖   |      |      |      |            | Yes    | Yes  |
   | 跟投關閉   |      |      |      |            |        |      |

6. 跟投，yes代表可以

   | 狀態       | 遊客 | 試玩 | 會員 | 帶玩(營運) |
               | ---------- | ---- | ---- | ---- | ---------- |
   | 正常       |      |      | Yes  | Yes        |
   | 房間已關閉 |      |      |      |            |
   | 房間禁言   |      |      | Yes  | Yes        |
   | 會員禁言   |      |      | Yes  | Yes        |
   | 會員封鎖   |      |      |      |            |
   | 跟投關閉   |      |      |      |            |

## Precautions

1. 為什麼聊天室每次連線1分鐘後會立即斷線

   目前聊天室心跳是依照logic.yml做設定，默認是5分鐘，所以正常來說5分鐘沒心跳連線才會斷，但如果前端以大於等於1分鐘的頻率做心跳，發現每次要發心跳時都會失敗，可能原因是連線已斷，但正常來說5分鐘後才會斷，這種情況可能的原因是連線是使用nginx or aws alb這種代理，默認`read` and `send`
   都是60s，所以必須調整這些代理的timeout大於5分鐘一點即可正常．

   nginx : `proxy_read_timeout` `proxy_send_timeout`
