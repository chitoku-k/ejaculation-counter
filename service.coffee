require("./db.js")
db = new ShikoDatabase

ID = process.env.TWITTER_ID

client = require("twitter-promise")(
    consumer_key:        process.env.TWITTER_CONSUMER_KEY
    consumer_secret:     process.env.TWITTER_CONSUMER_SECRET
    access_token_key:    process.env.TWITTER_ACCESS_TOKEN
    access_token_secret: process.env.TWITTER_ACCESS_TOKEN_SECRET
)


CronJob = require("cron").CronJob
job = new CronJob(
    cronTime: "00 00 * * *"
    onTick: () ->
        GetProfile().then (current) ->
            UpdateStatus current
        .then (update) ->
            UpdateProfile update.profile
            db.update update.date, update.profile.yesterday
        .catch (err) ->
            console.error err
    start: true
)


ParseProfile = (profile) ->
    matches = profile.name.match /(.*)（昨日: (\d+) \/ 今日: (\d+)）/

    name: matches[1]
    yesterday: +matches[2]
    today: +matches[3]


GetProfile = ->
    client.get("users/show", id: ID).then (obj) ->
        console.log "name: #{obj.name}"
        ParseProfile obj


UpdateProfile = (profile) ->
    name = "#{profile.name}（昨日: #{profile.yesterday} / 今日: #{profile.today}）"
    client.post("account/update_profile", name: name).then (obj) ->
        console.log "name: #{obj.name}"
        ParseProfile obj


UpdateStatus = (profile) ->
    date = new Date
    date.setDate date.getDate() - 1

    msg = "#{date.getFullYear()}/#{date.getMonth() + 1}/#{date.getDate()} "
    msg += if profile.yesterday is profile.today then "も" else "は"

    if profile.today > 0
        msg += " #{profile.today} 回ぴゅっぴゅしました…"
    else
        msg += "ぴゅっぴゅしませんでした…"

    client.post("statuses/update", status: msg).then (obj) ->
        console.log obj.text
        profile.yesterday = profile.today
        profile.today = 0

        date: date
        profile: profile


Object.defineProperty(Object.prototype, "enumerate",
    enumerable: false
    configurable: false
    writable: false
    value: ->
        Object.keys(@).map (elm) =>
            "#{elm}: #{@[elm]}"
)


client.stream("user", {}, (stream) ->
    stream.on("data", (data) ->
        if not data.text or data.retweeted_status or data.user.id_str isnt ID
            return

        text = data.text.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&amp;/g, "&")
        if (matches = text.match /^SQL:\s?(.+)/) isnt null
            db.query(matches[1]).then ((rows) -> (rows)), ((err) -> (err))
            .then (result) ->
                client.post "statuses/update", (
                    in_reply_to_status_id: data.id_str
                    status: "@java_shlt\n" + result.enumerate().join("\n").slice 0, 128
                )

        if not /^ぴゅっ♡+($| https:\/\/t\.co)/.test text
            return

        current = ParseProfile data.user
        current.today++

        Promise.all([
            UpdateProfile current
            db.update new Date, current.today
        ]).then (values) ->
            console.log values
    )

    stream.on("error", (err) ->
        console.error err
    )
)
