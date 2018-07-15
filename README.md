# Memo

### Prerequisites

- Golang (version 1.9)
- MySQL (version 5.5)
- Memcache
- Bitcoin node (ABC, Unlimited, etc)

### Setup

- Get repo
    ```sh
    go get github.com/memocash/memo/...
    ```

- Create MySQL database
  - Use charset `utf8_general_ci`

- Create config.yaml in memo diretory

    ```yaml
    MYSQL_HOST: 127.0.0.1
    MYSQL_USER: memo_user
    MYSQL_PASS: memo_password
    MYSQL_DB: memo
    
    MEMCACHE_HOST: 127.0.0.1
    MEMCACHE_PORT: 11211
    
    BITCOIN_NODE_HOST: 127.0.0.1
    BITCOIN_NODE_PORT: 8333
    ```

### Running

```sh
go build

# Run node
./memo main-node

# Separately run web server
./memo web --insecure
```

### Notes
- Can take about 30 minutes for main-node to full sync
- Main node can sometimes disconnect while syncing, just restart
- You may see a few errors, these are usually mal-formed memos and can be ignored


### View

Visit `http://127.0.0.1:8261` in your browser
