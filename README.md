Self hosted telegram bot for keeping a mircoblog diary. Write notes, then `/export` a markdown file with all your notes separated by days and with timecodes attached.

```
docker build https://github.com/ze-kel/microdiary.git -t microdiary
touch .env
echo TG_TOKEN=1337:MICRODIARY > .env
docker run --env-file "./.env" --mount source=microdiarydb,target=/app/db microdiary
```
