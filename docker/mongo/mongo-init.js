db.createUser({
    user: "channelCrawler",
    pwd: "pass",
    roles: [
        {
            role: "dbOwner",
            db: "crawler"
        }
    ]
})