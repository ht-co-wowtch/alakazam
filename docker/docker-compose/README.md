# docker-compose

整合 alakazam 的環境與服務。

- [目錄解說](#目錄解說)
- [開始使用](#開始使用)
  - [登入 GitLab Container Registry](#登入-gitlab-container-registry)
  - [Database Migration](#database-migration)
  - [MySQL initdb](#mysql-initdb)
  - [啟動與停止服務](#啟動與停止服務)
  - [啟動metrics](#啟動metrics)

## 目錄解說

[目錄](https://gitlab.com/jetfueltw/cpw/alakazam/tree/develop/docker/docker-compose)

```bash
.
├── Makefile
├── README.md
├── data
├── docker-compose.yml
├── kafka
├── metrics
├── mysql
└── wait-for
```

`Makefile` : 提供關於database `migrate` `rollback` `seed` `reset`指令來操作mysql docker，詳情看Makefile內容

`data` : mysql與redis位於docker內保存的資料目錄

`kafka` : 存放 kafka docker的資料與jmx設定檔目錄

`mysql` : 存放mysql Dockerfile與初始化sql資料

## 開始使用

git clone 到任何你喜歡的位子後，先 `cp .env.example .env` 再更改你需要的設定，例如：

```
MYSQL_ROOT_PASS=root
```

### 登入 GitLab Container Registry

先到 Gitlab 產生一個 [Personal Access Tokens](https://gitlab.com/-/profile/personal_access_tokens)，name 填 Container Registry（或其他好記的名字），scopes 勾選 read_registry，建立後記得把 token 記起來。  
之後 terminal 輸入 `docker login -u <your email of docker hub> registry.gitlab.com`，密碼則是剛才產生的 token。

### Database Migration

絕大多數的服務都依賴於資料庫，與正確的 schema 版本，在開始開發前，你需要先初始化資料庫。  
第一次會預設建立 `platform`、`alakazam` 兩個資料庫，如果你在 .env 使用其他名字的話，你必須手動新增資料庫再跑 migrate。

```bash
// 撤銷所有 migration
make platform.rollback

// 跑還沒跑過的 migration
make platform.migrate

// 塞入預設的必要與測試資料
make platform.seed

// 重設整個資料庫，等同於：rollback + migration + seed
make platform.reset
```

### MySQL init db

當第一次啟動 MySQL 時，會執行 `docker-entrypoint-initdb.d` 資料夾底下的 `.sh`、`.sql` 與 `.sql.gz`，你可以在裡面放初始化資料庫的語法。第一次啟動會跑一段時間才能訪問。  
如果你想刪除所有資料庫並重跑 init，可以刪除 volume 後再啟動。

```bash
docker-compose down
rm -rf ./data/mysql
docker-compose up -d mysql
```

### 啟動與停止服務

請指定你要啟動的服務，不然會全部啟動，可以配合 alias 節省打字時間。

```bash
// 啟動 

docker-compose up -d kafka 
docker-compose up -d alakazam 

// 停止所有服務

docker-compose down
```

### 啟動metrics

實際作法[參考](./metrics/README.md)

```bash
cp ./metrics/prometheus-example.yml ./metrics/prometheus.yml 

docker-compose up -d burrow_prometheus 
```
