class ShikoAction {
    constructor(service) {
        this.service = service;
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
            const response = Object.keys(result).map(x => `${x}: ${result[x]}`).join("\n").slice(0, 128);
            return this.service.client.post("statuses/update", {
                in_reply_to_status_id: status.id_str,
                status: `@java_shlt\n${response}`,
            });
        });
    }
}

exports.CreateShikoActions = function (service) {
    return [
        new SqlShikoAction(service),
        new PyuUpdateShikoAction(service),
        new NijieUpdateShikoAction(service),
    ];
};
