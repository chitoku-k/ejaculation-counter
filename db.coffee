class global.ShikoDatabase
    connection = require("mysql").createConnection(
        host:     process.env.MYSQL_HOST
        user:     process.env.MYSQL_USER
        password: process.env.MYSQL_PASSWORD
        database: process.env.MYSQL_DATABASE
    )

    query: ->
        args = arguments
        new Promise (resolve, reject) ->
            connection.query.apply(connection, args).on "error", (err) ->
                reject err
            .on "result", (rows) ->
                resolve rows

    update: (date, count) ->
        check  = "SELECT COUNT(*) AS `total` FROM `counts` WHERE `date` = ?"
        insert = "INSERT INTO `counts` (`count`, `date`) VALUES (?, ?)"
        update = "UPDATE `counts` SET `count` = ? WHERE `date` = ?"

        dateString = "#{date.getFullYear()}-#{date.getMonth() + 1}-#{date.getDate()}"
        @.query(check, [dateString]).then (rows) =>
            if rows.total is 1
                @.query update, [count, dateString]
            else
                @.query insert, [count, dateString]
        .then (rows) =>
            date: date
            count: count
