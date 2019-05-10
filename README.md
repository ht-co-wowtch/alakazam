# 聊天室
- [快速建置聊天室服務](#quick-start)
  - [編譯](#build)
  - [運行](#run)
- [依賴工具](#dependencies)
- [架構圖](#architecture)
- [如何使用聊天室各服務](#quick-reference)
- [聊天室Web Socket協定](#protocol-body)
- [Web Socket](#web-socket)
- [前台API](#frontend-api)
- [後台 API](#admin-api)
- [會員身份權限](#member-permissions)
- [訊息規則](#message-rule)
- [系統訊息](#system-message)
- [聊天室版本](#tag)
- [Q&A](#q-and-a)

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

    cd kafka
    docker-compose up -d
    
## Architecture

![arch](./doc/architecture.png)


`comet` : 負責管理web端連線與將訊息推送給web

`logic` : 負責接收各種商業邏輯訊息推送

`job` : 負責告知comet要推送什麼訊息以及房間訊息聚合

## Quick Reference

前台：
1. 如何跟進入聊天室
2. 如何在聊天室發訊息
3. 如何接收聊天室訊息
4. 如何知道被禁言
5. 如何知道被封鎖
6. 如何在聊天室發紅包
7. 如何搶紅包
8. 如何在聊天室發跟注
9. 如何跟注
10. 如何切換聊天室房間
11. 如何拿到歷史紀錄
12. 如何知道會員現在可不可以發紅包,跟注等權限操作
13. 如何跟聊天室做心跳

後台：
1. 如何以管理員身份廣播多個聊天室
2. 如何以系統公告身份廣播多個聊天室
3. 如何得到線上所有聊天室清單與在線人數
4. 如何得到某聊天室歷史紀錄
5. 如何禁言某會員
6. 如何封鎖某會員
7. 如何解禁言某會員
8. 如何解封鎖某會員
9. 如何在聊天室發紅包
10. 如何在聊天室發跟注
11. 如何將訊息置頂
12. 如何得到禁言名單
13. 如何得到封鎖名單
14. 如何得到某聊天室名單
15. 如何設定聊天設定(比如充值多少才能聊天)

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
用來說明binary Header是多少

### Operation
不同的Operation說明本次Protocol資料是什麼，如心跳回覆,訊息等等

name | 說明 |
-----|-----|
1|要求連線到某一個房間
2|連線到某一個房間結果回覆
3|發送心跳
4|回覆心跳結果
5|聊天室訊息
6|更換房間
7|回覆更換房間結果

### Body
聊天室的訊息內容

## Web Socket

## Frontend API

##  Admin API

## Member Permissions
會員權限與身份
狀態 |進入聊天室 |查看聊天 | 聊天 |發紅包|搶紅包|跟注|發跟注|
-----|-----|-----|-----|-----|-----|-----|-----|
禁言|yes|yes|no|yes|yes|yes|no|
封鎖|no|no|no|no|no|no|no|

身份 |進入聊天室 |查看聊天 | 聊天 |發紅包|搶紅包|跟注|發跟注|訊息置頂|充值限制聊天|打碼量聊天|
-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|
帶玩帳號|yes|yes|yes|yes|yes|yes|yes|no|
一般帳號|yes|yes|yes|yes|yes|yes|yes|no|
試玩帳號|yes|yes|no|no|no|no|no|no|
假人|no|no|no|no|no|no|yes|no|
後台|no|no|yes|yes|no|no|no|yes|
未登入|no|no|no|no|no|no|no|no|

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
  - 管理原

發送訊息限制
  - 下限1字元
  - 上限100字元
  - 1秒1則

## System Message
異常狀況| 訊息
-----|-----
未登入|请先登入会员
試玩發言|请先登入会员
試玩搶紅包|请先登入会员
試玩發紅包|请先登入会员
後台執行禁言|[会员名称] 被禁言
後台執行封鎖|[会员名称] 被封鎖
被禁言者發言|您在禁言状态，无法发言
不符後台條件發言|您无法发言，当前发言条件：前两天充值不少于[  ]元；打码量不少于[  ]元

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

## Q&A