const { createPool } = require("mysql");

exports.ShikoDatabase = class ShikoDatabase {
    constructor() {
        this.user = 1;
        this.pool = createPool({
            connectionLimit: 4,
            host: process.env.MYSQL_HOST,
            user: process.env.MYSQL_USER,
            password: process.env.MYSQL_PASSWORD,
            database: process.env.MYSQL_DATABASE,
        });
    }

    query(...args) {
        return new Promise((resolve, reject) =>
            this.pool.query(...args)
                     .on("error", err => reject(err))
                     .on("result", rows => resolve(rows))
        );
    }

    async update(date, count) {
        const check = "SELECT COUNT(*) AS `total` FROM `counts` WHERE `user` = ? AND `date` = ?";
        const insert = "INSERT INTO `counts` (`user`, `count`, `date`) VALUES (?, ?, ?)";
        const update = "UPDATE `counts` SET `count` = ? WHERE `date` = ?";

        const dateString = `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`;
        const rows = await this.query(check, [this.user, dateString]);
        if (rows.total === 1) {
            await this.query(update, [count, dateString]);
        } else {
            await this.query(insert, [this.user, count, dateString]);
        }

        return {
            date,
            count,
        };
    }
};
