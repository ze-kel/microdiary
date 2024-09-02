I use Obsidian.md to keep a diary: writing about events, self reflection etc. I love the app but quick capture is not it's strength. For quick things I use this bot, and then once a month export and save to obsidian.

### Microdiary

Telegram bot for keeping a diary or making quick notes.

Just write messages to save them. Use `/export` to get zip with markdown file containing all messages and all attachments or `/clear` to get markdown file and clear all messages in db.

Supports text, voice messages(saved as .ogg) and photos(saved as .jpg)

##### Export example

```
## 2024
### July
##### 28
19:00
Test message
![[2024.09.03 — 00-08.ogg]]
```

#### Deploy

To store messages you can use either local sqlite(will be saved in `/app/db/messages.db`) or a remote postgres instance(table names are prefixed with `Microdiary_` to avoid collisions). To store files you can use filesystem(`/app/db/file.ext`) or remote s3(minio) instance.

#### Local db, local files

```
TG_TOKEN=1337:MICRODIARY
TIMEZONE=Europe/Moscow
```

(timezone is needed to set message time when exporting)

`docker build https://github.com/ze-kel/microdiary.git -t microdiary`

`docker run --detach --env-file "./.env" --mount source=microdiarydb,target=/app/db microdiary`

#### Remote db, files in s3

```
TG_TOKEN=1337:MICRODIARY
TIMEZONE=Europe/Moscow
POSTGRES_URL="postgresql://admin:samplepass@localhost:1234/db"
BUCKET_URL="minio-something-something.zekel.io"
BUCKET_NAME="microdiary"
BUCKET_CRED_ID="md"
BUCKET_CRED_SECRET="samplepass"
```

`docker build https://github.com/ze-kel/microdiary.git -t microdiary`

`docker run --detach --env-file "./.env" microdiary`
