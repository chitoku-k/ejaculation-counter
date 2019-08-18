const decode = require("decode-html");
const striptags = require("striptags");
const Masto = require("mastodon");
const { ShikoDatabase } = require("./ShikoDatabase");
const { ShikoStream } = require("./ShikoStream");
const { ShikoView } = require("./ShikoView");
const { CronJob } = require("cron");
const { CreateShikoActions } = require("./ShikoAction");

exports.ShikoService = class ShikoService {
    constructor() {
        // REST API
        this.IDs = process.env.MASTODON_ID.split(" ");
        this.client = new Masto({
            access_token: process.env.MASTODON_ACCESS_TOKEN,
            api_url: process.env.MASTODON_API_URL,
        });

        // Streaming
        this.stream = new ShikoStream(this, CreateShikoActions(this));
        this.stream.create();

        // Database
        this.db = new ShikoDatabase();

        // View
        this.view = new ShikoView();

        // Cron
        this.job = new CronJob({
            cronTime: "00 00 * * *",
            onTick: () => this.onTick(),
        });
        this.job.start();
    }

    decodeHtml(text) {
        return decode(text);
    }

    decodeParagraph(html) {
        return this.decodeHtml(striptags(html));
    }

    decodeToot(toot) {
        return {
            ...toot,
            account: {
                ...toot.account,
                note: this.decodeParagraph(toot.account.note),
            },
            content: this.decodeParagraph(toot.content),
        };
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
        const { data: profile } = await this.client.get(`accounts/${this.IDs[0]}`);
        console.log(`name: ${profile.display_name}`);
        return this.parseProfile(profile);
    }

    parseProfile(profile) {
        const [ , name, yesterday, today ] = profile.display_name.match(/(.*)（昨日: (\d+) \/ 今日: (\d+)）/) || [];
        return {
            name,
            yesterday: +yesterday,
            today: +today,
        };
    }

    async updateProfile(profile) {
        const name = `${profile.name}（昨日: ${profile.yesterday} / 今日: ${profile.today}）`;
        const { data: update } = await this.client.patch("accounts/update_credentials", { display_name: name });
        console.log(`name: ${update.display_name}`);
        return this.parseProfile(update);
    }

    async updateStatus(profile) {
        const date = new Date();
        date.setDate(date.getDate() - 1);

        let message = `${date.getFullYear()}/${date.getMonth() + 1}/${date.getDate()} `;
        message += profile.yesterday === profile.today ? "も" : "は";

        if (profile.today > 0) {
            message += ` ${profile.today} 回ぴゅっぴゅしました…`;
        } else {
            message += "ぴゅっぴゅしませんでした…";
        }

        const { data: status } = await this.client.post("statuses", { status: message });
        console.log(status.content);
        profile.yesterday = profile.today;
        profile.today = 0;

        return { date, profile };
    }
}
