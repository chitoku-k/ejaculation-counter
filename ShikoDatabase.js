const createConnection = require("mysql").createConnection;

exports.ShikoDatabase = class ShikoDatabase {
    constructor() {
        this.connection = createConnection({
            host: process.env.MYSQL_HOST,
            user: process.env.MYSQL_USER,
            password: process.env.MYSQL_PASSWORD,
            database: process.env.MYSQL_DATABASE,
        });
    }

    query(...args) {
        return new Promise((resolve, reject) => this.connection.query(...args).on("error", err => reject(err)).on("result", rows => resolve(rows)));
    }

    update(date, count) {
        const check = "SELECT COUNT(*) AS `total` FROM `counts` WHERE `user` = ? AND `date` = ?"
        const insert = "INSERT INTO `counts` (`user`, `count`, `date`) VALUES (?, ?, ?)";
        const update = "UPDATE `counts` SET `count` = ? WHERE `date` = ?";

        const dateString = `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`;
        return this.query(check, [user, dateString]).then(rows => {
            if (rows.total === 1) {
                this.query(update, [count, dateString]);
            } else {
                this.query(insert, [user, count, dateString]);
            }
        }).then(rows => ({ date, count }));
    }
};
