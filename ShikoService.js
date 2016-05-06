const ShikoDatabase = require("./ShikoDatabase.js").ShikoDatabase;
const CronJob = require("cron").CronJob;
const TwitterPromise = require("twitter-promise");
const CreateShikoActions = require("./ShikoAction.js").CreateShikoActions;

class ShikoService {
    constructor() {
        this.ID = process.env.TWITTER_ID;
        this.client = new TwitterPromise({
            consumer_key: process.env.TWITTER_CONSUMER_KEY,
            consumer_secret: process.env.TWITTER_CONSUMER_SECRET,
            access_token_key: process.env.TWITTER_ACCESS_TOKEN,
            access_token_secret: process.env.TWITTER_ACCESS_TOKEN_SECRET,
        });
        this.db = new ShikoDatabase();
        this.job = new CronJob({
            cronTime: "00 00 * * *",
            onTick: () => this.onTick(),
        });
        this.start(CreateShikoActions(this));
    }

    start(actions) {
        this.actions = actions;
        this.job.start();
        this.client.stream("user", {}, stream => {
            stream.on("data", data => {
                if (data.text) {
                    data.text = data.text.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&amp;/g, "&");
                    this.actions.filter(x => x.regex.test(data.text)).forEach(x => x.invoke(data));
                }
            });

            stream.on("error", err => {
                console.error(err);
                process.exit(1);
            });
        });
    }

    onTick() {
        return this.getProfile().then(current => {
            return this.updateStatus(current);
        }).then(update => {
            this.updateProfile(update.profile);
            return this.db.update(update.date, update.profile.yesterday);
        }).catch(err => {
            return console.error(err);
        });
    }

    getProfile() {
        return this.client.get("users/show", { id: this.ID }).then(obj => {
            console.log(`name: ${obj.name}`);
            return this.parseProfile(obj);
        });
    }

    parseProfile(profile) {
        const [, name, yesterday, today] = profile.name.match(/(.*)（昨日: (\d+) \/ 今日: (\d+)）/) || [];
        return {
            name,
            yesterday: +yesterday,
            today: +today,
        };
    }

    updateProfile(profile) {
        const name = `${profile.name}（昨日: ${profile.yesterday} / 今日: ${profile.today}）`;
        return this.client.post("account/update_profile", { name }).then(obj => {
            console.log(`name: ${obj.name}`);
            return this.parseProfile(obj);
        });
    }

    updateStatus(profile) {
        const date = new Date;
        date.setDate(date.getDate() - 1);

        let message = `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} `;
        message += profile.yesterday === profile.today ? "も" : "は";

        if (profile.today > 0) {
            message += ` ${profile.today} 回ぴゅっぴゅしました…`;
        } else {
            message += "ぴゅっぴゅしませんでした…";
        }

        return this.client.post("statuses/update", { status: message }).then(obj => {
            console.log(obj.text);
            profile.yesterday = profile.today;
            profile.today = 0;

            return { date, profile };
        });
    }
}

new ShikoService();
