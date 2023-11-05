Self hosted telegram bot for keeping a mircoblog diary. Write notes, then ```/export``` a markdown file with all your notes separated by days and with timecodes attached.


```docker build . -t microdiary```
```docker run -e TG_TOKEN="..."```