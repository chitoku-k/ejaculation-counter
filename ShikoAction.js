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
                this.service.db.update(new Date, current.today),
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
    get regex() {
        return /ぴゅっぴゅしても?いい[\?|？]/;
    }

    async invoke(status) {
        if (status.retweeted_status) {
            return;
        }

        // 名前の一部を取り出す
        const name = status.user.name.replace(/(@.+|[\(（].*[\)）])$/g, "");
        try {
            const body = await request({
                method: "POST",
                uri: "https://shindanmaker.com/a/503598",
                form: {
                    u: name,
                },
            });
            const [ , result ] = body.match(/<textarea(?:[^>]+)>([\s\S]*)<\/textarea>/) || [];
            await this.reply(status.id_str, `@${status.user.screen_name} ${result}`);
        } catch (err) {
            console.error(err);
            await this.reply(status.id_str, `@${status.user.screen_name} おちんちんぴゅっぴゅ管理官が不在のためぴゅっぴゅしちゃダメです`);
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
            for (const [ key, value] of Object.entries(result)) {
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

exports.CreateShikoActions = function (service) {
    return [
        new SqlShikoAction(service),
        new PyuUpdateShikoAction(service),
        new ShindanmakerShikoAction(service),
        new NijieUpdateShikoAction(service),
        new HorneUpdateShikoAction(service),
    ];
};
