const request = require("request-promise");

class ShikoAction {
    constructor(service) {
        this.service = service;
    }

    reply(id, visibility, status) {
        return this.service.client.post("statuses", {
            in_reply_to_id: id,
            status: status,
            visibility: visibility === "direct" ? "direct" : "private",
        });
    }
}

class UpdateShikoAction extends ShikoAction {
    async invoke(status) {
        if (status.reblog || status.account.id !== this.service.ID) {
            return;
        }

        const current = this.service.parseProfile(status.account);
        current.today++;

        const [ profile, db ] = await Promise.all([
            this.service.updateProfile(current),
            this.service.db.update(new Date(), current.today),
        ]);
        console.log(profile, db);
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
    getName(status) {
        return status.account.display_name.replace(/(@.+|[\(（].+[\)）])$/g, "");
    }

    async shindan(status) {
        // 名前の一部を取り出す
        const name = this.getName(status);
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

        // 二重エスケープ回避
        return this.service.decodeHtml(
            this.service.decodeHtml(result)
        );
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
        if (status.reblog) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} おちんちんぴゅっぴゅ管理官が不在のためぴゅっぴゅしちゃダメです`);
            throw e;
        }
    }
}

class OfutonManagerShindanmakerShikoAction extends PyuppyuManagerShindanmakerShikoAction {
    get regex() {
        return /ふとん(し|(入|はい|い|行|潜|もぐ)っ)ても?[いよ良]い[?？]/;
    }

    getName(status) {
        return super.getName(status) + "ぶとん";
    }

    async invoke(status) {
        if (status.reblog) {
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
            await this.reply(status.id, status.visibility, `@${status.account.username} ${message}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} ふとんがふっとんだｗ`);
            throw e;
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
        if (status.reblog) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} おちんぽは現在勝負を受け付けていません`);
            throw e;
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
        if (status.reblog || status.tags.some(x => x.name === "ちんぽチャレンジ")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} チャレンジできませんでした……。`);
            throw e;
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
        if (status.reblog || status.tags.some(x => x.name === "おちんぽ挿入チャレンジ")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} チャレンジできませんでした……。`);
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
        if (status.reblog) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} 寿司職人がおやすみです……。`);
            throw e;
        }
    }
}

class AVShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /([^,.、。，．]+?)\s*(くん|ちゃん)?の\s*AV/;
    }

    getName(status) {
        const [ , name ] = status.content.match(this.regex);
        return name;
    }

    get uri() {
        return "https://shindanmaker.com/a/794363";
    }

    async invoke(status) {
        if (status.reblog || status.tags.some(x => x.name === "同人AVタイトルジェネレーター")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} AV に出演できませんでした……。`);
            throw e;
        }
    }
}

class OfutonChallengeShikoAction extends ShikoAction {
    get regex() {
        return /ふとん[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    async invoke(status) {
        if (status.reblog || status.tags.some(x => x.name === "おふとんチャレンジ")) {
            return;
        }

        try {
            const target = [..."おふとん"];
            const result = target.map(() => target[Math.random() * target.length | 0]).join("");
            await this.reply(status.id, status.visibility, `@${status.account.username} ${result}\n#おふとんチャレンジ`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} チャレンジできませんでした……。`);
            throw e;
        }
    }
}

class ThroughShikoAction extends ShikoAction {
    get regex() {
        return /(?:\s*(\d+)\s*連)?駿河茶|今日の\s*through|through\s*(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]/;
    }

    get api() {
        return "https://api.chitoku.jp/through/";
    }

    get uri() {
        return "http://user.keio.ac.jp/~rhotta/hellog/2009-06-20-1.html";
    }

    async invoke(status) {
        if (status.reblog) {
            return;
        }

        try {
            const length = this.regex.exec(status.content)[1] || 1;
            const through = await request({
                method: "GET",
                uri: this.api,
                json: true,
            });
            const result = Array.from({ length }, () => through[Math.random() * through.length | 0]);
            await this.reply(status.id, status.visibility, `@${status.account.username}\n${result.join("\n")}\n${this.uri}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} 何かがおかしいよ`);
            throw e;
        }
    }
}

class MpywShikoAction extends ShikoAction {
    get regex() {
        return /(?:mpyw|まっぴー|実務経験)(?:\s*(\d+)\s*連)?(?:が|ガ|ｶﾞ)[チﾁ][ャｬ]/;
    }

    get api() {
        return "https://mpyw.kb10uy.org/api";
    }

    async invoke(status) {
        if (status.reblog) {
            return;
        }

        try {
            const count = this.regex.exec(status.content)[1] || 1;
            const mpyw = await request({
                method: "GET",
                uri: this.api,
                qs: { count },
                json: true,
            });
            await this.reply(status.id, status.visibility, `@${status.account.username}\n${mpyw.result.join("\n")}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username} エラーが発生しました。実務経験がないのでしょうか。。。`);
            throw e;
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
        if (status.reblog || status.account.id !== this.service.ID) {
            return;
        }

        const [ , sql ] = status.content.match(this.regex) || [];
        if (!sql) {
            return;
        }

        const response = await this.query(sql).catch(err => err.message).then(x => x.slice(0, 120));
        try {
            await this.reply(status.id, status.visibility, `@${status.account.username}\n${response}`);
        } catch (e) {
            await this.reply(status.id, status.visibility, `@${status.account.username}\nエラーが発生しました`);
            throw e;
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
    new AVShindanmakerShikoAction(service),
    new ThroughShikoAction(service),
    new MpywShikoAction(service),
    new NijieUpdateShikoAction(service),
    new HorneUpdateShikoAction(service),
];
