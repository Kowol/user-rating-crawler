db.createUser({
    user: "channelCrawlerTest",
    pwd: "pass",
    roles: [
        {
            role: "dbOwner",
            db: "crawlerTest"
        }
    ]
})