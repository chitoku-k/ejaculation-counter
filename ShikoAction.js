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
    async shindan(status) {
        // 名前の一部を取り出す
        const name = status.user.name.replace(/(@.+|[\(（].*[\)）])$/g, "");
        const body = await request({
            method: "POST",
            uri: this.uri,
            form: {
                u: name,
            },
        });
        const [ , result ] = body.match(/<textarea id="copy_text_140"(?:[^>]+)>([\s\S]*)<\/textarea>/) || [];
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

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        try {
            const result = await this.shindan(status);
            const message = result.replace(/しこしこして/g, "もふもふさせて")
                                  .replace(/しこしこ|しゅっしゅ/g, "もふもふ")
                                  .replace(/ぴゅっぴゅって/g, "もふもふって")
                                  .replace(/ぴゅっぴゅ|お?ちんちん/g, "おふとん")
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
        return /ちん(ちん|ぽ|こ)(なん[かぞ])?に(勝[たちつてと]|負[かきくけこ])/;
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
        return /(^|[^#＃])ちん(ちん|ぽ|こ)[チﾁ][ャｬ][レﾚ][ンﾝ](ジ|ｼﾞ)/;
    }

    get uri() {
        return "https://shindanmaker.com/656461";
    }

    async invoke(status) {
        if (status.retweeted_status) {
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

class SqlShikoAction extends ShikoAction {
    get regex() {
        return /^SQL:\s?(.+)/;
    }

    async invoke(status) {
        if (status.retweeted_status || status.user.id_str !== this.service.ID) {
            return;
        }

        const [ , sql ] = status.text.match(this.regex) || [];
        if (!sql) {
            return;
        }

        let response;
        try {
            const result = await this.service.db.query(sql);
            const lines = [];
            for (const [ key, value ] of Object.entries(result)) {
                lines.push(`${key}: ${value}`);
            }
            response = lines.join("\n");
        } catch (err) {
            console.error(err);
            response = "エラーが発生しました";
        }

        response = response.slice(0, 140);
        try {
            await this.reply(status.id_str, `@${status.user.screen_name}\n${response}`);
        } catch (err) {
            console.error(err);
        }
    }
}

exports.CreateShikoActions = service => [
    new SqlShikoAction(service),
    new PyuUpdateShikoAction(service),
    new PyuppyuManagerShindanmakerShikoAction(service),
    new OfutonManagerShindanmakerShikoAction(service),
    new BattleChimpoShindanmakerShikoAction(service),
    new ChimpoChallengeShindanmakerShikoAction(service),
    new NijieUpdateShikoAction(service),
    new HorneUpdateShikoAction(service),
];
