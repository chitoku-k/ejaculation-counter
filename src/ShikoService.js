const Masto = require("mastodon");
const WebSocket = require("ws");
const xpath = require("xpath");
const { DOMParser } = require("xmldom");
const { ShikoDatabase } = require("./ShikoDatabase");
const { CronJob } = require("cron");
const { CreateShikoActions } = require("./ShikoAction");
const { fromEvent, from } = require("rxjs");
const { filter, flatMap, map, mergeMap, toArray } = require("rxjs/operators");

exports.ShikoService = class ShikoService {
    constructor() {
        this.ID = process.env.MASTODON_ID;
        this.client = new Masto({
            access_token: process.env.MASTODON_ACCESS_TOKEN,
            api_url: process.env.MASTODON_API_URL,
        });
        this.db = new ShikoDatabase();
        this.job = new CronJob({
            cronTime: "00 00 * * *",
            onTick: () => this.onTick(),
        });
        this.start(CreateShikoActions(this));
    }

    decodeHtml(text) {
        return text.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&amp;/g, "&");
    }

    decodeParagraph(html) {
        // TODO: xmldom error
        // return this.decodeHtml(xpath.select("string(/)", (new DOMParser()).parseFromString(html)));
        const [ , match ] = html.match(/<p>(.*)<\/p>/) || [];
        return this.decodeHtml(match);
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

    start(actions) {
        this.actions = actions;
        this.job.start();

        const stream = new WebSocket(`${process.env.MASTODON_WSS_URL}streaming?access_token=${process.env.MASTODON_ACCESS_TOKEN}&stream=user`);
        fromEvent(stream, "message")
            .pipe(map(x => JSON.parse(x.data)))
            .pipe(filter(x => x.event === "update"))
            .pipe(map(x => JSON.parse(x.payload)))
            .pipe(map(x => this.decodeToot(x)))
            .pipe(mergeMap(toot => from(this.actions)
                .pipe(map(action => ({ match: action.regex.exec(toot.content), action, toot })))
                .pipe(filter(({ match }) => match))
                .pipe(toArray())
                .pipe(map(x => x.sort((a, b) => a.match.index - b.match.index)))
            ))
            .pipe(flatMap(x => x))
            .pipe(map(({ action, toot }) => action.invoke(toot)))
            .subscribe(x => console.log(x));
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
        const { data: profile } = await this.client.get(`accounts/${this.ID}`);
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
