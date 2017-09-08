const Twit = require("twit");
const { ShikoDatabase } = require("./ShikoDatabase");
const { CronJob } = require("cron");
const { CreateShikoActions } = require("./ShikoAction");

exports.ShikoService = class ShikoService {
    constructor() {
        this.ID = process.env.TWITTER_ID;
        this.client = new Twit({
            consumer_key: process.env.TWITTER_CONSUMER_KEY,
            consumer_secret: process.env.TWITTER_CONSUMER_SECRET,
            access_token: process.env.TWITTER_ACCESS_TOKEN,
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

        const stream = this.client.stream("user");
        stream.on("tweet", async data => {
            data.text = data.text.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&amp;/g, "&");

            const target = this.actions.map(action => ({ match: action.regex.exec(data.text), action }))
                                       .filter(x => x.match)
                                       .sort((x, y) => x.match.index - y.match.index);

            for (const { action } of target) {
                await action.invoke(data);
            }
        });
        stream.on("error", err => {
            console.error(err);
            process.exit(1);
        });
    }

    async onTick() {
        try {
            const profile = await this.getProfile();
            const update = await this.updateStatus(profile);
            await Promise.all([
                this.updateProfile(update.profile),
                this.db.update(update.date, update.profile.yesterday),
            ]);
        } catch (err) {
            console.error(err);
        }
    }

    async getProfile() {
        const { data: profile } = await this.client.get("users/show/:id", { id: this.ID });
        console.log(`name: ${profile.name}`);
        return this.parseProfile(profile);
    }

    parseProfile(profile) {
        const [ , name, yesterday, today ] = profile.name.match(/(.*)（昨日: (\d+) \/ 今日: (\d+)）/) || [];
        return {
            name,
            yesterday: +yesterday,
            today: +today,
        };
    }

    async updateProfile(profile) {
        const name = `${profile.name}（昨日: ${profile.yesterday} / 今日: ${profile.today}）`;
        const { data: update } = await this.client.post("account/update_profile", { name });
        console.log(`name: ${update.name}`);
        return this.parseProfile(update);
    }

    async updateStatus(profile) {
        const date = new Date;
        date.setDate(date.getDate() - 1);

        let message = `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} `;
        message += profile.yesterday === profile.today ? "も" : "は";

        if (profile.today > 0) {
            message += ` ${profile.today} 回ぴゅっぴゅしました…`;
        } else {
            message += "ぴゅっぴゅしませんでした…";
        }

        const { data: status } = await this.client.post("statuses/update", { status: message });
        console.log(status.text);
        profile.yesterday = profile.today;
        profile.today = 0;

        return { date, profile };
    }
}
