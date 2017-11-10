const { createPool } = require("mysql");

exports.ShikoDatabase = class ShikoDatabase {
    constructor() {
        this.user = process.env.SHIKO_USER;
        this.pool = createPool({
            connectionLimit: 4,
            host: process.env.MYSQL_HOST,
            user: process.env.MYSQL_USER,
            password: process.env.MYSQL_PASSWORD,
            database: process.env.MYSQL_DATABASE,
        });
    }

    query(...args) {
        return new Promise((resolve, reject) => {
            this.pool.query(...args, (error, results, fields) => {
                if (error) {
                    reject(error);
                    return;
                }
                if (!Array.isArray(results)) {
                    resolve([ { message: results.message } ]);
                    return;
                }
                return resolve(results);
            });
        });
    }

    async update(date, count) {
        const check = "SELECT COUNT(*) AS `total` FROM `counts` WHERE `user` = ? AND `date` = DATE_FORMAT(?, '%Y-%m-%d')";
        const insert = "INSERT INTO `counts` (`user`, `count`, `date`) VALUES (?, ?, ?)";
        const update = "UPDATE `counts` SET `count` = ? WHERE `date` = ?";

        const [{ total }] = await this.query(check, [this.user, date]);
        if (total === 1) {
            await this.query(update, [count, date]);
        } else {
            await this.query(insert, [this.user, count, date]);
        }

        return {
            date,
            count,
        };
    }
};
