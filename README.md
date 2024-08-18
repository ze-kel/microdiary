Telegram bot for keeping a diary or making quick notes.

Just write messages to save them. Use `/export` to get markdown file with all messages or `/clear` to get markdown file and clear all messages in db.

##### Export example

```
## 2024
### July
##### 28
19:00
Test message
```

##### Running in docker

.env exa,ple

```
TG_TOKEN=1337:MICRODIARY
TIMEZONE=Europe/Moscow
# Optional, if you want to store in external pg database
POSTGRES_URL="postgresql://admin:samplepass@localhost:1234/db"
```

Building
`docker build https://github.com/ze-kel/microdiary.git -t microdiary`
sqlite
`docker run --detach --env-file "./.env" --mount source=microdiarydb,target=/app/db microdiary`
postgres
`docker run --detach --env-file "./.env" microdiary`
