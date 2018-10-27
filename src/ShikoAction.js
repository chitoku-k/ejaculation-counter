const request = require("request-promise");

class ShikoAction {
    constructor(service) {
        this.limit = 500;
        this.service = service;
    }

    pack(status, target, delimiter, limit) {
        delimiter = delimiter || "\n";
        limit = limit || this.limit - `@${status.account.acct} `.length;

        let content = "";

        for (const item of target) {
            if (content.length + delimiter.length + item.length > limit) {
                break;
            }
            content += `${delimiter}${item}`;
        }

        return content;
    }

    reply(status, content) {
        if (!content.replace(/\n/g, "").length) {
            return;
        }

        const text = `@${status.account.acct} ${content}`.slice(0, this.limit);

        return this.service.client.post("statuses", {
            in_reply_to_id: status.id,
            status: text,
            visibility: status.visibility === "direct" ? "direct" : "private",
        });
    }
}

class UpdateShikoAction extends ShikoAction {
    async invoke(status) {
        if (status.reblog || this.service.IDs.every(x => status.account.id !== x)) {
            return;
        }

        const current = await this.service.getProfile();
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
        return (status.account.display_name || status.account.username).replace(/(@.+|[\(（].+[\)）])$/g, "");
    }

    async shindan(status) {
        // 名前の一部を取り出す
        // 診断メーカーは preg_replace にエスケープなしで名前を渡すためエスケープ
        const name = this.getName(status).replace(/([$\\]{?\d+)/g, "\\$1");
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
            await this.reply(status, result);
        } catch (e) {
            await this.reply(status, "おちんちんぴゅっぴゅ管理官が不在のためぴゅっぴゅしちゃダメです");
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
            await this.reply(status, message);
        } catch (e) {
            await this.reply(status, "ふとんがふっとんだｗ");
            throw e;
        }
    }
}

class BattleChimpoShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /^お?ちん(ちん|ぽ|こ)(なん[かぞ])?に(勝[たちつてと]|負[かきくけこ])/;
    }

    get uri() {
        return "https://shindanmaker.com/a/584238";
    }

    async invoke(status) {
        if (status.reblog) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status, result);
        } catch (e) {
            await this.reply(status, "おちんぽは現在勝負を受け付けていません");
            throw e;
        }
    }
}

class ChimpoChallengeShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /ちん(ちん|ぽ|こ)[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    get uri() {
        return "https://shindanmaker.com/a/656461";
    }

    async invoke(status) {
        if (status.reblog || status.tags.some(x => x.name === "ちんぽチャレンジ")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status, result);
        } catch (e) {
            await this.reply(status, "チャレンジできませんでした……。");
            throw e;
        }
    }
}

class ChimpoInsertionChallengeShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /ちん(ちん|ぽ|こ)挿入[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    get uri() {
        return "https://shindanmaker.com/a/670773";
    }

    async invoke(status) {
        if (status.reblog || status.tags.some(x => x.name === "おちんぽ挿入チャレンジ")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status, result);
        } catch (e) {
            await this.reply(status, "チャレンジできませんでした……。");
        }
    }
}

class SushiShindanmakerShikoAction extends ShindanmakerShikoAction {
    get regex() {
        return /(🍣|寿司|すし|ちん(ちん|ぽ|こ))(握|にぎ)/;
    }

    get emoji() {
        return [
            "ios_big_sushi_1",
            "ios_big_sushi_2",
            "ios_big_sushi_3",
            "ios_big_sushi_4",
            "top_left_sushi",
            "top_center_sushi",
            "top_right_sushi",
            "middle_left_sushi",
            "middle_right_sushi",
            "bottom_left_sushi",
            "bottom_center_sushi",
            "bottom_right_sushi",
        ];
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
            await this.reply(status, result);
        } catch (e) {
            await this.reply(status, "寿司職人がおやすみです……。");
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
        if (status.reblog || status.tags.some(x => x.name === "同人avタイトルジェネレーター")) {
            return;
        }

        try {
            const result = await this.shindan(status);
            await this.reply(status, result);
        } catch (e) {
            await this.reply(status, "AV に出演できませんでした……。");
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
            await this.reply(status, `${result}\n#おふとんチャレンジ`);
        } catch (e) {
            await this.reply(status, "チャレンジできませんでした……。");
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
            const limit = this.limit - `@${status.account.acct} \n${this.uri}`.length;
            await this.reply(status, this.pack(status, result, "\n", limit) + `\n${this.uri}`);
        } catch (e) {
            await this.reply(status, "何かがおかしいよ");
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

            await this.reply(status, this.pack(status, mpyw.result));
        } catch (e) {
            await this.reply(status, "エラーが発生しました。実務経験がないのでしょうか。。。");
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
        if (status.reblog || this.service.IDs.every(x => status.account.id !== x)) {
            return;
        }

        const [ , sql ] = status.content.match(this.regex) || [];
        if (!sql) {
            return;
        }

        const response = await this.query(sql).catch(err => err.message);
        try {
            await this.reply(status, `\n${response}`);
        } catch (e) {
            await this.reply(status, `\nエラーが発生しました`);
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
