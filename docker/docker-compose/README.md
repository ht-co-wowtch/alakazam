# docker-compose
整合 alakazam 的環境與服務。

# 環境需求
docker and docker-compose

# 開始使用
git clone 到任何你喜歡的位子後，先 `cp .env.example .env` 再更改你需要的設定，例如：
```
MYSQL_ROOT_PASS=root
```

## 登入 GitLab Container Registry
先到 Gitlab 產生一個 [Personal Access Tokens](https://gitlab.com/profile/personal_access_tokens)，name 填 Container Registry（或其他好記的名字），scopes 勾選 read_registry，建立後記得把 token 記起來。  
之後 terminal 輸入 `docker login -u yourname@cqcp.com.tw registry.gitlab.com`，密碼則是剛才產生的 token。

## 啟動與停止服務
請指定你要啟動的服務，不然會全部啟動，可以配合 alias 節省打字時間。
```
// 啟動 alakazam，他會自動把依賴的服務也跑起來
docker-compose up -d alakazam

// 停止所有服務
docker-compose down
```

# Database Migration
絕大多數的服務都依賴於資料庫，與正確的 schema 版本，在開始開發前，你需要先初始化資料庫。  
第一次會預設建立 `platform`、`alakazam` 兩個資料庫，如果你在 .env 使用其他名字的話，你必須手動新增資料庫再跑 migrate。
```
// 撤銷所有 migration
make platform.rollback

// 跑還沒跑過的 migration
make platform.migrate

// 塞入預設的必要與測試資料
make platform.seed

// 重設整個資料庫，等同於：rollback + migration + seed
make platform.reset
```

# MySQL initdb
當第一次啟動 MySQL 時，會執行 `docker-entrypoint-initdb.d` 資料夾底下的 `.sh`、`.sql` 與 `.sql.gz`，你可以在裡面放初始化資料庫的語法。第一次啟動會跑一段時間才能訪問，大約 1 分鐘左右。  
如果你想刪除所有資料庫並重跑 init，可以刪除 volume 後再啟動。
```
docker-compose down
rm -rf ./data/mysql
docker-compose up -d mysql
```