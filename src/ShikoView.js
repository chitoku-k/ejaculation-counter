const Koa = require("koa");
const KoaStatic = require("koa-static");
const KoaProxy = require("koa-proxy");
const path = require("path");

exports.ShikoView = class ShikoView {
    constructor() {
        this.app = new Koa();
        this.app.use(KoaStatic(path.join(__dirname, "../public")));
        this.app.use(KoaProxy({
            host: "http://grafana:3000",
            match: /^\/grafana\//,
        }))
        this.app.listen(3000);
    }
}
