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

```
docker build https://github.com/ze-kel/microdiary.git -t microdiary
touch .env
echo TG_TOKEN=1337:MICRODIARY > .env
echo \nTIMEZONE="Europe/Moscow" >> .env  see https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
docker run --detach --env-file "./.env" --mount source=microdiarydb,target=/app/db microdiary
```
