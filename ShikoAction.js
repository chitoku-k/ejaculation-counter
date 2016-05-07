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
    invoke(status) {
        if (status.retweeted_status || status.user.id_str !== this.service.ID) {
            return;
        }

        const current = this.service.parseProfile(status.user);
        current.today++;

        return Promise.all([
            this.service.updateProfile(current),
            this.service.db.update(new Date, current.today),
        ]).then(values => {
            console.log(values);
        }, err => {
            console.error(err);
        });
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

class ShindanmakerShikoAction extends ShikoAction {
    get regex() {
        return /ぴゅっぴゅしていい[\?|？]/;
    }

    invoke(status) {
        if (status.retweeted_status || status.user.id_str !== this.service.ID) {
            return;
        }

        // 名前の一部を取り出す
        const name = status.user.name.replace(/(@.+|[\(（].*[\)）])$/g, "");
        request({
            method: "POST",
            uri: "https://shindanmaker.com/a/503598",
            form: {
                u: name,
            },
        }).then(body => {
            const [, result] = body.match(/<textarea(?:[^>]+)>([\s\S]*)<\/textarea>/) || [];
            this.reply(status.id_str, `@${status.user.screen_name}\n${result}`);
        }).catch(err => {
            this.reply(status.id_str, `@${status.user.screen_name} おちんちんぴゅっぴゅ管理官が不在のためぴゅっぴゅしちゃダメです`);
        });
    }
}

class SqlShikoAction extends ShikoAction {
    get regex() {
        return /^SQL:\s?(.+)/;
    }

    invoke(status) {
        if (status.retweeted_status || status.user.id_str !== this.service.ID) {
            return;
        }

        const [, sql] = status.text.match(this.regex) || [];
        if (!sql) {
            return;
        }
        return this.service.db.query(sql).then(result => {
            const response = Object.keys(result).map(x => `${x}: ${result[x]}`)
                                                .join("\n")
                                                .slice(0, 137 - status.user.screen_name.length);
            return this.reply(status.id_str, `@${status.user.screen_name}\n${response}`);
        });
    }
}

exports.CreateShikoActions = function (service) {
    return [
        new SqlShikoAction(service),
        new PyuUpdateShikoAction(service),
        new ShindanmakerShikoAction(service),
        new NijieUpdateShikoAction(service),
    ];
};
