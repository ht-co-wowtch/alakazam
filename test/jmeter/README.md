# Jmeter 壓力測試



## 參數

| 參數         | 預設                                      | 說明                                                         |
| ------------ | ----------------------------------------- | ------------------------------------------------------------ |
| path         | /private/var/www/cpw/alakazam/test/jmeter | 本專案目錄路徑，ex: `/private/var/www/cpw/alakazam/test/jmeter` |
| url          | 192.168.0.138                             | 聊天室 web socket host                                       |
| rapidash_url | 192.168.0.138                             | 認證中心url                                                  |
| run_sec      | 3600                                      | 執行多久，單位是ms(毫秒)                                     |
| send_room_id | 2                                         | 發訊息至某房間，請填房間id，請不要填1，腳本默認會使用這個房間 |
| user_total   | 10                                        | 同時段發送多少訊息                                           |
| user_csv     | members_0_1000.csv                        | user uid等資料，csv檔需放在`data`目錄下                      |

# 連線

1. 同時段連線



## 單ㄧ房間聊天

1. 固定頻率聊天  `one_chat_rate_message.jmx`

   壓測房間最多每秒能夠發送多少訊息，由於聊天室限制一個人每秒只能發一則訊息，所以每人每兩秒才會發一則訊息，所以當`user_total`設定成10，會分成兩組user去發送，每組之間間隔1秒，ex A組10人，B組10人，在線總共20人，A與B組都是每兩秒才執行發訊息，但是A與B之間啟動時間隔1秒，所以才能做到每1秒發送10則訊息

   

   Ex: 房間每1秒就有100個人同時聊天

   

   `jmeter -n -t one_chat_rate_message.jmx -l one_chat_rate_message.jtl -e -o report -J path=<path> -J url=<alakazam websocket> -J rapidash_url=<rapidash url> -J run_sec=<sec> -J send_room_id=<room id> -J user_total=<total> -J user_csv=<user csv>`

   

2. 隨機頻率聊天 `one_chat_random_message`

   Ex: 房間每人每次以1 - 3秒隨機頻率聊天

   

   `jmeter -n -t one_chat_random_message.jmx -l one_chat_random_message.jtl -e -o report -J path=<path> -J url=<alakazam websocket> -J rapidash_url=<rapidash url> -J run_sec=<sec> -J send_room_id=<room id> -J user_total=<total> -J user_csv=<user csv>`

   

3. 固定 and 隨機頻率聊天

   Ex: 承1,2一起執行

   

## 多個房間聊天



# 紅包

# 問題

目前負責聊天的用戶全都集中至房間1，然後對非房間1的房間發話，這樣做是因為要避免出現tcp window full，以下舉例

1. user_1進入房間1
2. user_1固定每秒對所在的房間1做發話
3. 假設tcp window size => 65536
4. 由於user_1也在房間1，所以tcp連線也會接收到訊息
5. user_1不讀訊息，因為在整個流程中只希望負責發話的壓力測試
6. 不讀訊息造成tcp window size越來越小
7. 當tcp window size = 0時，tcp write 會阻塞，所以user_1 tcp也阻塞
8. 由於tcp也阻塞造成心跳機制也會阻塞
9. 造成心跳超時，server自動將user_1 tcp close

目前要解決上述問題，有兩個方案

1. 將發話者跟接收訊息房間做隔離，不然發話者的tcp會接收到訊息
2. 壓力測試流程中每發一次訊息後tcp就要讀一次訊息

目前使用方案1