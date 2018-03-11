const request = require("request-promise");

class ShikoAction {
    constructor(service) {
        this.service = service;
    }

    reply(id, status) {
        return this.service.client.post("statuses/update", {
            in_reply_to_status_id: id,
            status: status,
        });
    }
}

class UpdateShikoAction extends ShikoAction {
    async invoke(status) {
        if (status.retweeted_status || status.user.id_str !== this.service.ID) {
            return;
        }

        const current = this.service.parseProfile(status.user);
        current.today++;

        try {
            const [ profile, db ] = await Promise.all([
                this.service.updateProfile(current),
                this.service.db.update(new Date(), current.today),
            ]);
            console.log(profile, db);
        } catch (err) {
            console.error(err);
        }
    }
}

class PyuUpdateShikoAction extends UpdateShikoAction {
    get regex() {
        return /^ぴゅっ♡+($| https:\/\/t\.co)/;
    }
}

class NijieUpdateShikoAction extends UpdateShikoAction {
    get regex() {
        return /ニジエの「.*」で抜きました。 #ニジエ/;
    }
}

class HorneUpdateShikoAction extends UpdateShikoAction {
    get regex() {
        return /ホルネの「.*」でたぎりました。 #ホルネ/;
    }
}

class ShindanmakerShikoAction extends ShikoAction {
    getName(user) {
        return user.name.replace(/(@.+|[\(（].+[\)）])$/g, "");
    }

    async shindan(status) {
        // 名前の一部を取り出す
        const name = this.getName(status.user);
        const body = await request({
            method: "POST",
            uri: this.uri,
            form: {
                u: name,
            },
        });
        const [ , result ] = body.match(/<textarea id="copy_text_140"(?:[^>]+)>([\s\S]*)<\/textarea>/) || [];
        if (!result) {
            throw new Error("No shindan result found.");
        }
        return result;
    }
}

class PyuppyuManagerShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /ぴゅっぴゅしても?[いよ良]い[?？]/;
    }

    get uri() {
        return "https://shindanmaker.com/a/503598";
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} おちんちんぴゅっぴゅ管理官が不在のためぴゅっぴゅしちゃダメです`);
        }
    }
}

class OfutonManagerShindanmakerShikoAction extends PyuppyuManagerShindanmakerShikoAction {
    get regex() {
        return /ふとん(し|(入|はい|い|行|潜|もぐ)っ)ても?[いよ良]い[?？]/;
    }

    getName(user) {
        return super.getName(user) + "ぶとん";
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        try {
            const result = await this.shindan(status);
            const message = result.replace(/しこしこして/g, "もふもふさせて")
                                  .replace(/しこしこ|しゅっしゅ/g, "もふもふ")
                                  .replace(/ぴゅっぴゅって/g, "もふもふって")
                                  .replace(/ぴゅっぴゅ|いじるの|お?ちんちん/g, "おふとん")
                                  .replace(/出せる/g, "もふもふできる")
                                  .replace(/出し/g, "もふもふし")
                                  .replace(/手の平に/g, "朝まで");
            await this.reply(status.id_str, `@${status.user.screen_name} ${message}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} ふとんがふっとんだｗ`);
        }
    }
}

class BattleChimpoShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /^お?ちん(ちん|ぽ|こ)(なん[かぞ])?に(勝[たちつてと]|負[かきくけこ])/;
    }

    get uri() {
        return "https://shindanmaker.com/584238";
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} おちんぽは現在勝負を受け付けていません`);
        }
    }
}

class ChimpoChallengeShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /ちん(ちん|ぽ|こ)[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    get uri() {
        return "https://shindanmaker.com/656461";
    }

    async invoke(status) {
        if (status.retweeted_status || status.text.includes("#ちんぽチャレンジ")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} チャレンジできませんでした……。`);
        }
    }
}

class ChimpoInsertionChallengeShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /ちん(ちん|ぽ|こ)挿入[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    get uri() {
        return "https://shindanmaker.com/670773";
    }

    async invoke(status) {
        if (status.retweeted_status || status.text.includes("#おちんぽ挿入チャレンジ")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} チャレンジできませんでした……。`);
        }
    }
}

class SushiShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /(🍣|寿司|すし|ちん(ちん|ぽ|こ))(握|にぎ)/;
    }

    get uri() {
        return "https://shindanmaker.com/a/577901";
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} 寿司職人がおやすみです……。`);
        }
    }
}

class OfutonChallengeShikoAction extends ShikoAction {
    get regex() {
        return /ふとん[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    async invoke(status) {
        if (status.retweeted_status || status.text.includes("#おふとんチャレンジ")) {
            return;
        }

        try {
            const target = [..."おふとん"];
            const result = target.map(() => target[Math.random() * target.length | 0]).join("");
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}\n#おふとんチャレンジ`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} チャレンジできませんでした……。`);
        }
    }
}

class ThroughShikoAction extends ShikoAction {
    get regex() {
        return /駿河茶|今日の\s*through|through\s*(が|ガ|ｶﾞ)[チﾁ][ャｬ]/;
    }

    get api() {
        return "http://api.chitoku.jp/through/";
    }

    get uri() {
        return "http://user.keio.ac.jp/~rhotta/hellog/2009-06-20-1.html";
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        try {
            const through = await request({
                method: "GET",
                uri: this.api,
                json: true,
            });
            const result = through[Math.random() * through.length | 0];
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}\n${this.uri}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} 何かがおかしいよ`);
        }
    }
}

class MpywShikoAction extends ShikoAction {
    get regex() {
        return /(?:mpyw|まっぴー|実務経験)(?:(\d+)連)?(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]/;
    }

    get api() {
        return "http://mpyw.kb10uy.org";
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        const count = this.regex.exec(status.text)[1] || 1;

        try {
            const mpyw = await request({
                method: "HEAD",
                uri: this.api,
                qs: { count },
                simple: false,
                followRedirect: false,
                resolveWithFullResponse: true,
            });
            if (!mpyw.headers.location) {
                throw new Error("No location header is found.");
            }
            await this.reply(status.id_str, `@${status.user.screen_name} ${mpyw.headers.location}`);
        } catch (e) {
            await this.reply(status.id_str, `@${status.user.screen_name} エラーが発生しました。実務経験がないのでしょうか。。。`);
        }
    }
}

class SqlShikoAction extends ShikoAction {
    get regex() {
        return /^SQL:\s?(.+)/;
    }

    async query(sql) {
        const [ result ] = await this.service.db.query(sql);
        const lines = [];
        for (const [ key, value ] of Object.entries(result)) {
            lines.push(`${key}: ${value}`);
        }
        return lines.join("\n");
    }

    async invoke(status) {
        if (status.retweeted_status || status.user.id_str !== this.service.ID) {
            return;
        }

        const [ , sql ] = status.text.match(this.regex) || [];
        if (!sql) {
            return;
        }

        const response = await this.query(sql).catch(err => err.message).then(x => x.slice(0, 120));
        try {
            await this.reply(status.id_str, `@${status.user.screen_name}\n${response}`);
        } catch (err) {
            console.error(err);
        }
    }
}

exports.CreateShikoActions = service => [
    new SqlShikoAction(service),
    new OfutonChallengeShikoAction(service),
    new PyuUpdateShikoAction(service),
    new PyuppyuManagerShindanmakerShikoAction(service),
    new OfutonManagerShindanmakerShikoAction(service),
    new BattleChimpoShindanmakerShikoAction(service),
    new ChimpoChallengeShindanmakerShikoAction(service),
    new ChimpoInsertionChallengeShindanmakerShikoAction(service),
    new SushiShindanmakerShikoAction(service),
    new ThroughShikoAction(service),
    new MpywShikoAction(service),
    new NijieUpdateShikoAction(service),
    new HorneUpdateShikoAction(service),
];
